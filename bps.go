package bps

// Message represents a single pub/sub message.
type Message struct {
	// ID is a message identifier.
	ID string `json:"id,omitempty"`

	// Data is the message payload.
	Data []byte `json:"data,omitempty"`

	// Attributes contains key-value labels.
	Attributes map[string]string `json:"attributes,omitempty"`
}
