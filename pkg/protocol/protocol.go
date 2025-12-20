// Package protocol defines the wire protocol for communication between
// cmdr and units. Uses protobuf schemas with length-prefixed framing.
package protocol

// MessageType identifies the type of protocol message.
type MessageType int

const (
	// Controller -> Unit
	MessageTypeSessionCreate MessageType = iota
	MessageTypeSessionResume
	MessageTypeSessionDestroy
	MessageTypePromptSend
	MessageTypePromptCancel
	MessageTypeStateGet
	MessageTypeStateSubscribe

	// Unit -> Controller
	MessageTypeAck
	MessageTypeResponseChunk
	MessageTypeResponseComplete
	MessageTypeResponseCancelled
	MessageTypeStateUpdate
	MessageTypeError
)
