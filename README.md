# OCCA (Overly Complex Chat Application)

## Components

### Client CLI application
Used to interact with system.

### Gateway
HTTP frontdoor to the backend system:
- resolves which chat server client should connect to.
- provides online members functionality and allows to start a chat with one.

### Auth server
Central place of authentication of the system:
- Allows to authenticate via username/password and returns a JWT token.
- Valides whether authentication header is valid.


### Chat server
Stateful distributed component which distributes active users via consistent hashing and keeps
a connection with the client. Forward messages to the correct receiver.

### Archiver
Collect messages sent between client and stores them into durable storage. Allows to view older messages with friends.