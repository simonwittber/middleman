# middleman
A PubSub & Request/Response WebSocket Server, in Go.


MiddleMan distributes messages between connected services and clients. It
allows stateless and stateful services from diverse platforms to be
quickly integrated into a running cluster of services.

This is the middleman protocol.


1. When a client connects to middleman, the the first message is the API key.

A servce API key enables all messages, a client API key only enables
messages that a service has explicitly enabled.

Eg:
```
MyServerKey
```

The middleman will then respond with:

```
MM OK
```

if the key was accepted. If the key is not accepted, the connection is closed.


2. Each message then follows this structure:

```
COMMAND TARGET
Header:Value

Body 
.
```

Note that the message body is terminated with a single dot and newline, just like SMTP.


3. Commands are:

PUB: Publish a message with a name (target). Message is received by 
 all subscribers.

SUB: Request that all published messages of (target) are delivered.

REQ: Request something from a service. Must have ReqID header. This 
 is delivered to only one service, even if multiple services have subscribed to the target.

RES: Respond to a request. Must have a ReqID header.

EPUB: Allow clients to publish to this target.

EREQ: Allow clients to request from this target.

ESUB: Allow clients to subscribe to this target.

A PUB command distributes to all services and clients. A REQ command
is effectively load balanced between any waiting services.


<pre> 
                             +------------------+
                             |                  |          +--------------+
  +-------------+      +-----+-------+  +-------------+----+  Service #1  |
  |  Client #1  +------+  Client Key |  | Server Key  |    +--------------+
  +-------------+      |             |  |             |
                       |             |  |             |
                       ++-+----------+  +----------+-++
                        | |  |                  |  | |     +--------------+
                        | |  |                  |  | +-----+  Service #2  |
  +-------------+       | |  |     MiddleMan    |  |       +--------------+
  |  Client #2  +-------+ |  |                  |  |
  +-------------+         |  |                  |  |
                          |  +------------------+  |
                          |                        |       +--------------+
                          |                        +-------+  Service #N  |
                          |                                +--------------+
  +-------------+         |
  |  Client #N  +---------+
  +-------------+
 

     PUB #Name                                           EPUB #Name
     #Body
                                                         ESUB #Name
     SUB #Name
                                                         EREQ #Name
     REQ #Name
     #Body

     RES #Name
     #Body


</pre>
