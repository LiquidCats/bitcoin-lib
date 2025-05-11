package serializer

type TxPayload struct {
	Version  int32
	TxIn     []*TxIn
	TxOut    []*TxOut
	LockTime uint32
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *TxPayload) Command() string {
	return CmdTx
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *TxPayload) MaxPayloadLength(_ uint32) uint32 {
	return MaxBlockPayload
}

type OutPoint struct {
	Hash  Hash
	Index uint32
}

type TxIn struct {
	PreviousOutPoint OutPoint
	SignatureScript  []byte
	Witness          Witness
	Sequence         uint32
}

type TxOut struct {
	Value    int64
	PkScript []byte
}
