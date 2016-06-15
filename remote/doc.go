/*
Package remote contains optional functionality.
Interfaces and utilities will be in package remote, and implementation that depend on 3:rd party will be in sub-packages
For example dependencies to event-bus implementations should not be manadatory.
In case you want to use, for example NSQ,or NATS or something, then it should be easy to include the extra-package
for that, and link in.
If the utils in here grows too big, they will become their own repositorys.
*/
package remote
