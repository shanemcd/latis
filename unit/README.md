# unit

The agent endpoint daemon for Latis.

## Responsibilities

- **Protocol handling**: Parse incoming messages, emit responses
- **Agent wrapping**: Interface with underlying AI agents (Claude, GPT, local models, custom code)
- **Session state**: Maintain conversation context and agent state
- **Streaming**: Push response chunks back to cmdr in real-time

## Operation Modes

A unit can run as:

- **Daemon**: Long-running process accepting connections
- **On-demand**: Spawned by connector (e.g., via SSH), runs for session duration
- **Embedded**: Library mode, integrated into other applications

## Protocol Messages Handled

```
← session.create     → ack + session_id
← session.resume     → ack + state
← session.destroy    → ack
← prompt.send        → response.chunk* + response.complete
← prompt.cancel      → ack
← state.get          → state.update
← state.subscribe    → state.update*
```

## Agent Adapters

Units don't implement AI directly — they wrap agents:

```
unit
 └── agent adapter
      └── actual agent (claude, gpt, llama, custom, etc.)
```

The adapter interface is minimal: receive prompt, stream response, report state.

## Design Notes

Units are designed to be lightweight. They should be easy to deploy anywhere — a remote server, a container, a Raspberry Pi. The heavy lifting happens in the wrapped agent; the unit just handles protocol and plumbing.
