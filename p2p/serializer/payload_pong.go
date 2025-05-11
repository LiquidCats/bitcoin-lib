package serializer

type PongPayload struct {
	// Unique value associated with message that is used to identify
	// specific ping message.
	Nonce uint64
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *PongPayload) Command() string {
	return CmdPong
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *PongPayload) MaxPayloadLength(pver uint32) uint32 {
	plen := uint32(0)
	// The pong message did not exist for BIP0031Version and earlier.
	// NOTE: > is not a mistake here.  The BIP0031 was defined as AFTER
	// the version unlike most others.
	if pver > BIP0031Version {
		// Nonce 8 bytes.
		plen += 8
	}

	return plen
}
