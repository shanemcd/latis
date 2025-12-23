package quic

// StreamType identifies the purpose of a QUIC stream.
// The first byte written to a stream indicates its type.
type StreamType byte

const (
	// StreamTypeControl is for control plane operations (lifecycle, health, etc.)
	StreamTypeControl StreamType = 0x01

	// StreamTypeA2A is for A2A protocol traffic (agent-to-agent communication)
	StreamTypeA2A StreamType = 0x02
)

func (t StreamType) String() string {
	switch t {
	case StreamTypeControl:
		return "control"
	case StreamTypeA2A:
		return "a2a"
	default:
		return "unknown"
	}
}
