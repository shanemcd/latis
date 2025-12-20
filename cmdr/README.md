# cmdr

The control plane and CLI for Latis.

## Responsibilities

- **CLI interface**: User-facing commands (`latis connect`, `latis prompt`, etc.)
- **Session management**: Create, resume, destroy agent sessions
- **Routing**: Direct messages to the right units via connectors
- **Orchestration**: Coordinate multi-agent tasks
- **State tracking**: Know what's running where

## Usage

```
latis connect <unit-address>
latis session new [--connector ssh|ws|local]
latis prompt "your message here"
latis agents list
latis coordinate --agents unit-1,unit-2 "work together on this"
```

## Design Notes

The cmdr doesn't know how to transport messages — it delegates to connectors. It doesn't know how to execute agent tasks — that's the unit's job. It only knows the protocol and how to orchestrate.
