[![Go Report Card](https://goreportcard.com/badge/github.com/Forau/yanngo)](https://goreportcard.com/report/github.com/Forau/yanngo)
[![Build Status](https://travis-ci.org/Forau/yanngo.svg?branch=master)](https://travis-ci.org/Forau/yanngo)
# yanngo
Yet Another NordNet Go API

Inspired by https://github.com/denro/nordnet, but will have changes in the api to make it easier to rpc, and aggregate some logic.

There is a fork of denro's work where I have a branch demonstrating some of the changes I will make at https://github.com/Forau/nordnet.
However, I have not make pull-request since I am not sure thats the direction most people want to go.  Thanks goes to denro for his nice implementation, and some of the code here is derrived from his work.

## Why another API?

I do not want my client's to need to know my login to nordnet. They only need credentials to my daemon.  Also, since we should not create many sessions to nordnet, the daemon needs to keep track of more information, and provide a slightly different api to the clients.


## Folder structure

The reason for making sub packages is to keep it easy for removing third party dependencies.
There will be a few third party liberies in the default build, but the parts with the dependencies will be easily replaced if needed / wanted.
There will also be some packages for optional features, to avoid the dependencies being linked in if not used.

## Status

Under heavy development. Do not use straight of for now.  Everything can change without notice. Version 0.0.0b.
I am also making this for myself, so focus is gettign all featurs in, and thus some tests and documentation is lacking.

Currently there is a redesign under way..

* api - Interfaces for the api and transports
* crypto - Helper functions for credentials
* example - Fairly basic examples. 
* example/nsqnnd - New structured deamon, with nsq as eventbus
* example/nsqnnc - New structured client, with nsq as eventbus
* example/nsqnnwebapp - New structured webserver with a small trading app. NSQ as eventbus, and SockJS for web stuff.
* feed - Basic feed.  (Will have some redesign)
* httpcli - Http-client helper.  Current implementation depends on resty, but a pure standard one would be an easy change.
* remote - Interfaces to unify remote calls, like RPC or eventbus'es. Wrappers to provide functionality for unificatgion.
* remote/nsqconn - Providing what is needed for the 'remote' interfaces when using NSQ as channel.
* swagger - Generated swagger model. Only scripted changes, so it can be updated if nordnet changes its api.
* transports - Implementation of api/transports interface.

What should work on any given checkin is the tests, and the examples.

-- Current restructure --
The focus is to shift from the api (IE the class with all the methods to call), to let transports define their own api in runtime, and thus be able to hook in custom 'helper' commands, or custom caching methods.
