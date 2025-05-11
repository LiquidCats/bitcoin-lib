package serializer

type VerAckPayload struct{}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *VerAckPayload) Command() string {
	return CmdVerAck
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *VerAckPayload) MaxPayloadLength(pver uint32) uint32 {
	return 0
}
