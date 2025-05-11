package serializer

type GetAddrPayload struct{}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *GetAddrPayload) Command() string {
	return CmdGetAddr
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *GetAddrPayload) MaxPayloadLength(_ uint32) uint32 {
	return 0
}
