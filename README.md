# OCCA (Overly Complex Chat Application)

## Components

### Client CLI application
Used to interact with system.

### Gateway
HTTP frontdoor to the backend system:
- authenticates users.
- resolves which chat server client should connect to.
- provides online members functionality and allows to start a chat with one.
- resolves the event server where the recipient is connected to and forwards the message

### Event server
Keeps a persistent connection to a client for relaying real time events to user.

### Archiver
Collect messages sent between client and stores them into durable storage. Allows to view chat history.