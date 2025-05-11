package serializer

import (
	"fmt"
	"io"
)

type Message interface {
	Command() string
	MaxPayloadLength(uint32) uint32
}

type ChainAdapter interface {
	SerializeVersion(w io.Writer, msg *VersionPayload, pver uint32, encoding MessageEncoding) error
	UnserializeVersion(r io.Reader, msg *VersionPayload, pver uint32, encoding MessageEncoding) error

	SerializeVerAck(w io.Writer, msg *VerAckPayload, pver uint32, encoding MessageEncoding) error
	UnserializeVerAck(r io.Reader, msg *VerAckPayload, pver uint32, encoding MessageEncoding) error

	SerializeGetAddr(w io.Writer, msg *GetAddrPayload, pver uint32, encoding MessageEncoding) error
	UnserializeGetAddr(r io.Reader, msg *GetAddrPayload, pver uint32, encoding MessageEncoding) error

	SerializeInv(w io.Writer, msg *InvPayload, pver uint32, encoding MessageEncoding) error
	UnserializeInv(r io.Reader, msg *InvPayload, pver uint32, encoding MessageEncoding) error

	SerializePing(w io.Writer, msg *PingPayload, pver uint32, encoding MessageEncoding) error
	UnserializePing(r io.Reader, msg *PingPayload, pver uint32, encoding MessageEncoding) error

	SerializePong(w io.Writer, msg *PongPayload, pver uint32, encoding MessageEncoding) error
	UnserializePong(r io.Reader, msg *PongPayload, pver uint32, encoding MessageEncoding) error

	SerializeBlock(w io.Writer, msg *BlockPayload, pver uint32, encoding MessageEncoding) error
	UnserializeBlock(r io.Reader, msg *BlockPayload, pver uint32, encoding MessageEncoding) error

	SerializeTx(w io.Writer, msg *TxPayload, pver uint32, encoding MessageEncoding) error
	UnserializeTx(r io.Reader, msg *TxPayload, pver uint32, encoding MessageEncoding) error
}

type Serializer struct {
	adapter ChainAdapter
}

func (s *Serializer) Serialize(w io.Writer, msg Message, pver uint32, encoding MessageEncoding) error {
	switch m := msg.(type) {
	case *VersionPayload:
		return s.adapter.SerializeVersion(w, m, pver, encoding)
	case *VerAckPayload:
		return s.adapter.SerializeVerAck(w, m, pver, encoding)
	case *GetAddrPayload:
		return s.adapter.SerializeGetAddr(w, m, pver, encoding)
	case *InvPayload:
		return s.adapter.SerializeInv(w, m, pver, encoding)
	case *PingPayload:
		return s.adapter.SerializePing(w, m, pver, encoding)
	case *PongPayload:
		return s.adapter.SerializePong(w, m, pver, encoding)
	case *BlockPayload:
		return s.adapter.SerializeBlock(w, m, pver, encoding)
	case *TxPayload:
		return s.adapter.SerializeTx(w, m, pver, encoding)
	}

	return fmt.Errorf("serialize: unknown command: %s", msg.Command())
}

func (s *Serializer) Deserialize(r io.Reader, msg Message, pver uint32, encoding MessageEncoding) error {
	switch m := msg.(type) {
	case *VersionPayload:
		return s.adapter.UnserializeVersion(r, m, pver, encoding)
	case *VerAckPayload:
		return s.adapter.UnserializeVerAck(r, m, pver, encoding)
	case *GetAddrPayload:
		return s.adapter.UnserializeGetAddr(r, m, pver, encoding)
	case *InvPayload:
		return s.adapter.UnserializeInv(r, m, pver, encoding)
	case *PingPayload:
		return s.adapter.UnserializePing(r, m, pver, encoding)
	case *PongPayload:
		return s.adapter.UnserializePong(r, m, pver, encoding)
	case *BlockPayload:
		return s.adapter.UnserializeBlock(r, m, pver, encoding)
	case *TxPayload:
		return s.adapter.UnserializeTx(r, m, pver, encoding)
	}

	return fmt.Errorf("deserialize: unknown command: %s", msg.Command())
}
