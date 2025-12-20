# connector

Transport abstraction layer for Latis.

## Responsibilities

- **Transport bytes**: Move protocol messages between cmdr and units
- **Connection lifecycle**: Establish, maintain, and teardown connections
- **Pluggable interface**: Any transport mechanism can be a connector

## Interface

Every connector implements the same contract:

```
connect(address)      → establishes connection to a unit
send(message)         → sends a protocol message
receive()             → receives a protocol message (blocking or streaming)
close()               → tears down the connection
```

The connector doesn't understand the protocol messages — it just moves them. Serialization and deserialization happen at the protocol layer.

## Planned Connectors

- **ssh**: Shell into remote hosts, communicate via stdin/stdout
- **local**: Spawn and communicate with local processes
- **container**: Exec into containers (podman, docker)
- **websocket**: Persistent bidirectional connections
- **http**: Request/response for stateless interactions

## Design Notes

Connectors are intentionally dumb. They know how to establish a channel and push bytes through it. Authentication, encryption, and reliability are connector concerns — but protocol semantics are not.
