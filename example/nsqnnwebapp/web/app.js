(function() {

  var nnApp = angular.module("nnApp", []).
      run(['$rootScope', '$window', '$q', '$timeout', function($rootScope, $window, $q, $timeout) {
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

        var lastCloseTime = 0;
        function bindSockJsToRootScope(sockName) {
          var sock = new $window.SockJS(origin+'/'+sockName, undefined, options);
          var deferred = $q.defer();
          $rootScope['$'+sockName+'_open_promise'] = deferred.promise;
          console.log('Creating new SockJS: %o',sock);
          
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
              sock.onopen = sock.onmessage = sock.onclose = function(){};
              var delay = 0;
              if( (Date.now() - lastCloseTime) < 10000 ) {
                delay = 5000;
              }
              lastCloseTime = Date.now();
              $timeout(bindSockJsToRootScope.bind(this,sockName), delay);
            });
          };
          sock.onmessage = function(e) {
            $rootScope.$apply(function(){
              msg = JSON.parse(e.data);
              $rootScope.$broadcast('$'+sockName+'.msg', msg);
            });
          };

        }
        bindSockJsToRootScope('sockjs');
        bindSockJsToRootScope('feed');

        $rootScope.$on('$sockjs.send', function(event, message) {
          $rootScope['$sockjs_open_promise'].then(function(sock) {
            sock.send(message);
          }, function(err) {
            console.log('Unable to send to sockjs: %o',err)
          });
        });
        
        $rootScope.Lazy = $window.Lazy; // Since we have access to rootScope more often, and this is not a serious example.
        $rootScope.alert = function(msg) { $window.alert(msg); };
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
      return this.sendRawCmd({ cmd: cmd, args: args})
    };
    this.sendRawCmd = function(msg) {
      msg.id = Math.random()*Math.pow(2,63);
      var deferred = $q.defer();
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
      
      $scope.$on('$sockjs.msg', function(event, message) {
        console.log('Got message %o',message);
        $scope.lastMsg = message;
      });

      $scope.feedMsgs = {};
      $scope.$on('$feed.msg', function(event, message) {
        console.log('Got FEED message %o',msg);
        $scope.feedMsgs[msg.type+':'+msg.data.i+':'+msg.data.m] = msg;
      });
      
    }]);

  nnApp.directive('accounts', [ '$rootScope', 'nnApiHelper', function($rootScope, nnApiHelper) {
    return {
      scope: {
      },
      template: '<table ng-repeat="acc in accounts" border="1"><tr><th>Accno</th><th>Type</th><th>Alias</th><th>Credits available<button ng-click="$parent.updateAccount(acc.accno)">Refresh</button> </th><tr>'+
        '<td>{{acc.accno}}</td><td>{{acc.type}}</td><td>{{acc.alias}}</td>'+
        '<td>{{acc.trading_power.value}} of {{acc.account_sum.value}} {{acc.account_sum.currency}}</td>'+
        '</tr><tr><td colspan="4"><positions accno="acc.accno"></positions></td>'+
        '</tr><tr><td colspan="4"><orders accno="acc.accno" auto="false"></orders></td>'+
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

    nnApp.directive('orders', [ '$rootScope', 'nnApiHelper', function($rootScope, nnApiHelper) {
    return {
      scope: {
        accno: '=',
        auto: '@?'
      },
      template: '<div><label>Show deleted orders <input type="checkbox" ng-model="showDeletedOrders"/></label>'+
        '<br/><button ng-click="loadOrders()">Reload from API</button><br/><lu><li ng-repeat="order in getOrders()">'+
        '{{order.side}}: {{order.price.value}} {{order.price.currency}} {{order.volume}}st - '+
        'Tradable: {{order.tradable.identifier}}:{{order.tradable.market_id}}, State: '+
        '{{order.action_state}} {{order.order_state}} '+
        '<button ng-click="$parent.orderAction(\'DeleteOrder\',order.order_id)">Delete</button>'+
        '<button ng-if="order.order_state === \'LOCAL\'" ng-click="$parent.orderAction(\'ActivateOrder\',order.order_id)">Activate</button>'+
        '<button ng-if="order.order_state !== \'DELETED\'" ng-click="$parent.editOrder(order)">Edit</button>'+
        '</li></lu></div>',
      replace: true,
      transclude: true,
      controller: function($scope) {
        $scope.orders = [];
        $scope.loadOrders = function() {
          nnApiHelper.sendCmd('AccountOrders',{accno:$scope.accno}).then(function(msg) {
            console.log('Orders[%s]: %o',$scope.accno, msg);
            $scope.orders = msg.payload;
          });
        }
        function loadFeedOrders() {
          nnApiHelper.sendCmd('FeedGetOrders',{accno:$scope.accno}).then(function(msg) {
            console.log('FeedOrders[%s]: %o',$scope.accno, msg);
            $scope.orders = msg.payload;
          });
        }
        
        loadFeedOrders();
        $scope.$on('$feed.msg', function(event, message) {
          if(msg.type === 'order') {
            var newOrders = $rootScope.Lazy($scope.orders).reject({order_id: msg.data.order_id}).concat([msg.data]).toArray();
            $scope.orders = newOrders;
          }
        });
        
        $scope.getOrders = function() {
          return $rootScope.Lazy($scope.orders).reject(function(order) {
            return order.order_state === 'DELETED' && !$scope.showDeletedOrders;
          }).toArray();          
        };

        $scope.editOrder = function(order) {
          $rootScope.$broadcast('cmdForm', {cmd: 'UpdateOrder', args: {
            accno: order.accno, order_id: order.order_id, price: order.price.value,
            currency: order.price.currency, volume: order.volume
          }})
        };

        $scope.orderAction = function(action, oid) {
          nnApiHelper.sendCmd(action,{accno:$scope.accno, order_id:oid}).then(function(msg) {
            console.log('%s: %o',action,msg);
            $rootScope.alert(JSON.stringify(msg));
          }, function(err) {
            console.log('Error %s: %o',action, err);
          });
        };
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
        '<tr ng-if="lastPrice.i"><td></td><td colspan="5">Ask[{{lastPrice.ask_volume}}]: {{lastPrice.ask}} Bid[{{lastPrice.bid_volume}}]: {{lastPrice.bid}} High: {{lastPrice.high}} Low: {{lastPrice.low}} Last: {{lastPrice.last}} Tick: {{lastPrice.tick_timestamp | date:\'dd/MM HH:mm:ss\' }} Trade: {{lastPrice.trade_timestamp | date:\'dd/MM HH:mm:ss\' }}</td></tr>'+
        '<tr ng-if="lastTrade.i"><td></td><td colspan="5">Price: {{lastTrade.price}} Volume: {{lastTrade.volume}}</td></tr>'+
        '<tr ng-if="lastTradingStatus.i"><td></td><td colspan="5">{{lastTradingStatus|json}}</td></tr>'+
        '</table>',
      replace: true,
      transclude: true,
      controller: function($scope) {
        console.log('instrument.... %o', $scope.ins);
        $scope.tradable = $rootScope.Lazy($scope.ins.tradables).find({market_id:11}) || $scope.ins.tradables[0];

        $scope.lastPrice = $scope.lastDepth = $scope.lastTrade = $scope.lastTradingStatus = {};

        function addFeedMessage(msg) {
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
        }
        
        $scope.$on('$feed.msg', function(event, msg) {
          addFeedMessage(msg);
        });

        // This will be very spammy, but its just a example, so who cares...
        nnApiHelper.sendCmd('FeedGetState',{id: $scope.tradable.identifier, market: $scope.tradable.market_id}).
          then(function(msg) {
            console.log('FeedState: %o',msg);
            $rootScope.Lazy(msg.payload).each(function(m) {
              addFeedMessage(m);
            });
          });
        
        
        $scope.trade = function() {
          var args = angular.copy($scope.tradable);
          args.currency = $scope.ins.currency;
          $rootScope.$broadcast('cmdForm', {cmd: 'CreateOrder', args: args });         
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

  nnApp.directive('commandForm', [ '$rootScope', 'nnApiHelper', function($rootScope, nnApiHelper) {
    return {
      scope: {      
      },
      template: '<form name="commandList" ng-submit="doSend()"><label for="listCommand"> List selection: </label>'+
        '<select name="listCommand" id="listCommand" ng-model="selectedCmd" ng-change="onSelectCmd(selectedCmd)">'+
        '<option value="">---Please select command ---</option> '+
        '<option ng-repeat="cmd in commands|orderBy:\'cmd\'" value="{{cmd}}">{{cmd.cmd}}</option></select><br/>'+
        '{{cmd.cmd}}: {{cmd.desc}}<table><tr ng-repeat="arg in cmd.args">'+
        '<td>{{arg.name}}</td><td></td><td>'+
        '<input ng-if="!$parent.isSelectField(arg)" name="{{arg.name}}" ng-model="arg.value"></input>'+
        '<select ng-if="$parent.isSelectField(arg)" name="{{arg.name}}" ng-model="arg.value"><option ng-repeat="o in arg.opts" value="{{o}}">{{o}}</option></select>'+
        '</td><td>{{$parent.getArgmuentMarkers(arg)}} {{arg.desc}}</td></tr></table><input type="submit"/><br\><br\>'+
        '<pre style="background:lightgrey;border-style: inset;margin:5px;overflow: auto;height:400px;">{{cmd.resp}}{{cmd.error}}</pre>'+
        '</form>',
      replace: true,
      transclude: true,
      controller: function($scope) {
        $scope.commands = [];
        $scope.cmd = {};

        nnApiHelper.sendCmd('TransportRespondsTo').then(function(msg) {
          console.log('Commands: %o',msg);
          $scope.commands = msg.payload;
        });
        $scope.onSelectCmd = function(cmd) {
          console.log('Select %o',cmd)
          $scope.cmd = JSON.parse(cmd);
        };
        $scope.isSelectField = function(field) {
          return !!field.opts && field.opts.length > 0
        };
        $scope.getArgmuentMarkers = function(arg) {
          if(arg.optional) {
            return "(optional)";
          }
          return "*";
        }

        $scope.$on('cmdForm', function(evt,msg) {
          console.log('Got form msg: %o',msg);
          var cmd = $rootScope.Lazy($scope.commands).find({cmd:msg.cmd});
          if(cmd) {
            cmd = angular.copy(cmd);
            $rootScope.Lazy(cmd.args).each(function(a) { if(msg.args[a.name]) { a.value = String(msg.args[a.name]); } });
            $scope.cmd = cmd;
          }
        });
        
        $scope.doSend = function() {
          if($scope.cmd && $scope.cmd.cmd) {
            var req = {cmd:$scope.cmd.cmd}
            req.args = $rootScope.Lazy($scope.cmd.args).map(function(a) { return [a.name,a.value]; }).toObject();
            console.log('Do send %o: %o',$scope,req);
            nnApiHelper.sendRawCmd(req).then(function(msg) {
              console.log('Command response: %o', msg);
              $scope.cmd.resp = JSON.stringify(msg.payload,null, "  ") ;
              $scope.cmd.error = JSON.stringify(msg.error,null, "  ");
            });
          }
        };
      }
    };
  }]);
  
})();
