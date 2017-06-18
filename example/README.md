The examples are just here to show usage.
Currently we have a server, cli-client and web-client(webserver with small angular app).

Currently NSQD is needed to run the examples, since that is the only pubsub implemented at this time.

* nsqnnd - The daemon. It will connect to nordnet, both with public and private feeds, and then provide access over nsq.
* nsqnnc - CLI client. Can run the commands mannually. Type 'help' for list of commands. The ones starting with / is via the api-client, and the other ones are generated from the api.Transport information from the server.  I will move towards not having the api-client implementation, since adhoc features is easier to add without.
* nsqnnwebapp - A webserver that connects as a client, and provide SockJS access for the webapp.
* nsqnnwebapp/web - The angular app itself. Not ment to look good, just provide the basic features. Dependencies are only sockjs, angularjs and lazyjs, all via cdn.
* nn_feed_prices - A small script that displays feed prices in a top-like style.

nn_feed_prices have no dependency to yanngo, it is a mere example of benefits of having the data available in different manners. To run it, you can eighter first stream everything to a file, and then tail that one, or just: nsq_tail -topic nordnet.feed --nsqd-tcp-address 127.0.0.1:5150 | go run main.go

* omxtime - A small tool to convert or check time. Locale is hardcoded to Stockholm, regardless of system locale

To install omxtime for example, for easier use, then just:
go get -u github.com/Forau/yanngo/example/omxtime


