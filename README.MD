hashServer
==============
This program hash and encodes passwords over HTTP using SHA512 encryption.  The program is also able to return ID information to the client so that they can easily lookup their passwords.  Another functionality is that clients can request statistics for the server over HTTP.  Lastly, a graceful shutdown request is available to clients that shutsdown the server after all open requests are finished.

**Examples**
- Run the server by building the source and typing in: ./hashServer < port > into your command line (note: if no port is provided default is 8080)
- curl --data "password=< password >" http://localhost:< port >/hash/ will return a hash ID the client can use to request their hashed password (ex return value: 3968902113 when password is set to: 'angryMonkey')
- curl http://localhost:< port >/hash/?id=< id > will return the SHA512 password. (ex return value: ZEHhWB65gUlzdVwtDQArEyx-KVLzp_aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A-gf7Q== when id is set to: 3968902113)
- curl http://localhost:< port >/stats/ will return a JSON statistic object reporting the total number of password hash requests to the server, and the average duration of each request in milliseconds (ex return value: {"total":2,"average":0.02772})
- curl -X POST http://localhost:< port >/shutdown will gracefully shutdown the server

