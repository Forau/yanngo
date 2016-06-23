(function() {

  var nnApp = angular.module("nnApp", []).
      run(['$rootScope', '$window', '$q', function($rootScope, $window, $q) {
        if (!$window.location.origin) { 
          $window.location.origin = $window.location.protocol + '//' + $window.location.hostname + ($window.location.port ? (':' + $window.location.port) : '');
        }        
        var origin = $window.location.origin;

        var options = {
          debug: true,
          devel: true,
          protocols_whitelist: ['websocket', 'xdr-streaming', 'xhr-streaming',
                                'iframe-eventsource', 'iframe-htmlfile',
                                'xdr-polling', 'xhr-polling', 'iframe-xhr-polling', 'jsonp-polling']
        };

        function bindSockJsToRootScope(sockName) {
          var sock = new $window.SockJS(origin+'/'+sockName, undefined, options);
          var deferred = $q.defer();
          $rootScope['$'+sockName+'_open_promise'] = deferred.promise;

          sock.onopen = function() {
            $rootScope.$apply(function(){
              console.log('connection open[%s] %o', sockName, sock);
              $rootScope['$'+sockName+'_open'] = true;
              deferred.resolve(sock);
            });
          };
          sock.onclose = function() {
            $rootScope.$apply(function(){
              console.log('connection closed[%s] %o', sockName, sock);
              $rootScope['$'+sockName+'_open'] = false;
              deferred.reject('closed'); // TODO: We probably need a new promise....
            });
          };
          sock.onmessage = function(e) {
            $rootScope.$apply(function(){
              msg = JSON.parse(e.data);
              $rootScope.$broadcast('$'+sockName+'.msg', msg);
            });
          };

          $rootScope.$on('$'+sockName+'.send', function(event, message) {
            if(!$rootScope['$'+sockName+'_open']) {
              // Simple reconnect, but no resend.
              bindSockJsToRootScope(sockName);
            } else {
              sock.send(message);
            }
          });

        }
        bindSockJsToRootScope('sockjs');
        bindSockJsToRootScope('feed');
        $rootScope.Lazy = $window.Lazy; // Since we have access to rootScope more often, and this is not a serious example.
      }]);

  nnApp.service('nnApiHelper', ['$rootScope', '$q', '$timeout', function($rootScope, $q, $timeout) {
    var self = this;
    var pendingMap = {}

    this.rejectPending = function(id,err) {
      var deferred = pendingMap[id];
      if(deferred) {
        deferred.reject(err);
        delete pendingMap[id];        
      }
    };
    
    this.resolvePending = function(id,msg) {
      var deferred = pendingMap[id];
      if(deferred) {
        deferred.resolve(msg);
        delete pendingMap[id];        
      }
    };
    
    this.sendCmd = function(cmd,args) {
      if(args) {
        // Make sure all values are strings, since nordnet uses http-forms instead of rest.
        args = $rootScope.Lazy(args).pairs().tap(function(kv) { kv[1] = String(kv[1]); }).toObject();
      }
      var deferred = $q.defer();
      var msg = { cmd: cmd, args: args, id: Math.random()*Math.pow(2,63) };
      pendingMap[msg.id] = deferred;
      $rootScope.$sockjs_open_promise.then(function(sj) {
        $rootScope.$emit('$sockjs.send', JSON.stringify(msg));
      }).catch(function(err) {
        self.rejectPending(msg.id, err);
      });

      $timeout(function() { self.rejectPending(msg.id, '{"err": "Timeout after 10 sec"}'); }, 10000);
      
      return deferred.promise;
    };

    $rootScope.$on('$sockjs.msg', function(event, msg) {
      self.resolvePending(msg.id, msg);
    });

  }]);

  nnApp.controller("NnCtrl", [
    '$scope', '$rootScope', 
    function($scope, $rootScope) {
      console.log('Scope: %o',$scope);
      
      $scope.sendMsg = function(msg) {
        console.log('Send[%o]: %o', msg, $scope);
        $rootScope.$emit('$sockjs.send',msg);
      };

      $scope.$on('$sockjs.msg', function(event, message) {
        console.log('Got message %o',message);
        $scope.lastMsg = message;
      });

      $scope.feedMsgs = {};
      $scope.$on('$feed.msg', function(event, message) {
        console.log('Got message %o',msg);
        $scope.feedMsgs[msg.type+':'+msg.i+':'+msg.m] = msg;
      });
      
    }]);

  nnApp.directive('accounts', [ '$rootScope', 'nnApiHelper', function($rootScope, nnApiHelper) {
    return {
      scope: {
      },
      template: '<table ng-repeat="acc in accounts" border="1"><tr><th>Accno</th><th>Type</th><th>Alias</th><th>Credits available<button ng-click="$parent.updateAccount(acc.accno)">Refresh</button> </th><tr>'+
        '<td>{{acc.accno}}</td><td>{{acc.type}}</td><td>{{acc.alias}}</td>'+
        '<td>{{acc.trading_power.value}} of {{acc.account_sum.value}} {{acc.account_sum.currency}}</td>'+
        '</tr><tr>'+
        '<td colspan="4"><positions accno="acc.accno"</td>'+
        '</tr></table>',
      replace: true,
      transclude: true,
      controllerAs: 'ctrl',
      controller: function($scope) {
        var self = this;
        console.log('Accounts....');
        $scope.accounts = {};
        nnApiHelper.sendCmd('Accounts').then(function(msg) {
          console.log('Accounts: %o',msg);
          $rootScope.Lazy(msg.payload).each(function(acc) {
            $scope.accounts[acc.accno] = acc;
            $scope.updateAccount(acc.accno);
          });
        });

        $scope.updateAccount = function(accno) {
          console.log('About to update %d',accno);
          $scope.accounts[accno].trading_power = $scope.accounts[accno].account_sum = {};
          nnApiHelper.sendCmd('Account',{accno: accno}).then(function(msg) {
            console.log('Account: %o',msg);
            $scope.accounts[accno].trading_power = msg.payload.trading_power;
            $scope.accounts[accno].account_sum = msg.payload.account_sum;
          });
        };
      }
    };
  }]);

    nnApp.directive('positions', [ '$rootScope', 'nnApiHelper', function($rootScope, nnApiHelper) {
    return {
      scope: {
        accno: '='
      },
      template: '<lu><li ng-repeat="pos in positions">{{pos.instrument.name}}: {{pos.qty}} - GAV: {{pos.acq_price_acc.value}}  {{pos.acq_price_acc.currency}} <instrument ins="pos.instrument"></instrument></li></lu>',
      replace: true,
      transclude: true,
      controller: function($scope) {
        console.log('pos.... %s', $scope.accno);
        $scope.positions = [];
        nnApiHelper.sendCmd('AccountPositions',{accno:$scope.accno}).then(function(msg) {
          console.log('Pos: %o',msg);
          $scope.positions = msg.payload;
        });
      }
    };
  }]);

    nnApp.directive('instrument', [ '$rootScope', 'nnApiHelper', function($rootScope, nnApiHelper) {
    return {
      scope: {
        ins: '='
      },
      template: '<table border="1"><tr>'+
        '<td>{{ins.symbol}}</td><td>{{ins.name}}</td><td>{{tradable.identifier}} {{tradable.market_id}}</td>'+
        '<td><button ng-click="trade()">Trade</button><button ng-click="subFeed()">+Feed</button></tr>'+
        '<tr ng-if="lastPrice.i"><td></td><td colspan="5">Ask: {{lastPrice.ask}} Bid: {{lastPrice.bid}} High: {{lastPrice.high}} Low: {{lastPrice.low}} Last: {{lastPrice.last}} Tick: {{lastPrice.tick_timestamp | date:\'dd/MM HH:mm:ss\' }} Trade: {{lastPrice.trade_timestamp | date:\'dd/MM HH:mm:ss\' }}</td></tr>'+
        '<tr ng-if="lastTrade.i"><td></td><td colspan="5">Price: {{lastPtrade.price}} Volume: {{lastTrade.volume}}</td></tr>'+
        '<tr ng-if="lastTradingStatus.i"><td></td><td colspan="5">{{lastTradingStatus|json}}</td></tr>'+
        '</table>',
      replace: true,
      transclude: true,
      controller: function($scope) {
        console.log('instrument.... %o', $scope.ins);
        $scope.tradable = $rootScope.Lazy($scope.ins.tradables).find({market_id:11}) || $scope.ins.tradables[0];

        $scope.lastPrice = $scope.lastDepth = $scope.lastTrade = $scope.lastTradingStatus = {};

        $scope.$on('$feed.msg', function(event, msg) {
          if($scope.tradable) {
            if(msg.data.i === $scope.tradable.identifier && msg.data.m === $scope.tradable.market_id) {
              switch(msg.type) {
              case 'price':
                $scope.lastPrice = msg.data;
                break;
              case 'depth':
                $scope.lastDepth = msg.data;
                break;
              case 'trade':
                $scope.lastTrade = msg.data;
                break;
              case 'trading_status':
                $scope.lastTradingStatus = msg.data;
                break;
              default:
                console.log('Matched, but dont know type %o',msg);
              }
            }
          }
        });

        $scope.trade = function() {
          $rootScope.$broadcast('trade', {instrument: $scope.ins});
        };
        
        $scope.subFeed = function() {
          nnApiHelper.sendCmd('FeedSubscribe',{type:'price', id:$scope.tradable.identifier, market:$scope.tradable.market_id}).then(
            function(res) { console.log('price %o',res); });
          nnApiHelper.sendCmd('FeedSubscribe',{type:'depth', id:$scope.tradable.identifier, market:$scope.tradable.market_id}).then(
            function(res) { console.log('depth %o',res); });
          nnApiHelper.sendCmd('FeedSubscribe',{type:'trade', id:$scope.tradable.identifier, market:$scope.tradable.market_id}).then(
            function(res) { console.log('trade %o',res); });
          nnApiHelper.sendCmd('FeedSubscribe',{type:'trading_status', id:$scope.tradable.identifier, market:$scope.tradable.market_id}).then(function(res) { console.log('trading_status %o',res); });

        };
      }
    };
    }]);

  nnApp.directive('selectList', [ '$rootScope', 'nnApiHelper', function($rootScope, nnApiHelper) {
    return {
      scope: {       
      },
      template: '<form name="selectList"><label for="listSelect"> List selection: </label>'+
        '<select name="listSelect" id="listSelect" ng-model="selectedList" ng-change="onSelectList(selectedList)">'+
        '<option value="">---Please select list---</option> '+
        '<option ng-repeat="list in lists" value="{{list.list_id}}">{{list.symbol}}</option>'+
        '</select><lu><li ng-repeat="ins in selectedListItems"><instrument ins="ins"></instrument></li></lu></form>',
      replace: true,
      transclude: true,
      controller: function($scope) {
        console.log('pos.... %s', $scope.accno);
        $scope.lists = [];
        nnApiHelper.sendCmd('Lists').then(function(msg) {
          console.log('Lists: %o',msg);
          $scope.lists = msg.payload;
        });

        $scope.onSelectList = function(listId) {
          console.log('Selected list: %s', listId);
          nnApiHelper.sendCmd('List',{id:listId}).then(function(msg) {
            console.log('Got list %s: %o',listId, msg);
            $scope.selectedListItems = msg.payload;
          })
        };
      }
    };
  }]);

  
})();
