package serializer

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/LiquidCats/bitcoin-lib/params"
)

// MessageHeader represents the parsed header of a protocol message.
type MessageHeader struct {
	Magic    params.BitcoinNet
	Command  string
	Size     uint32
	Checksum [4]byte
}

func (h *MessageHeader) Serialize(w io.Writer) (numBytes int, err error) {
	err = binary.Write(w, littleEndian, uint32(h.Magic))
	if err != nil {
		return
	}

	// 2) Command padded to 12 bytes
	cmdBytes := make([]byte, len(h.Command))
	copy(cmdBytes, h.Command)
	_, err = w.Write(cmdBytes)
	if err != nil {
		return
	}

	// 3) Payload length
	err = binary.Write(w, littleEndian, h.Size)
	if err != nil {
		return
	}

	// 4) Checksum: first 4 bytes of double SHA256
	_, err = w.Write(h.Checksum[:])
	if err != nil {
		return
	}

	return
}

func (h *MessageHeader) Deserialize(r io.Reader) (numBytes int, err error) {
	var headerBytes [MessageHeaderSize]byte

	numBytes, err = io.ReadFull(r, headerBytes[:])
	if err != nil {
		return
	}

	hr := bytes.NewReader(headerBytes[:])

	var command [CommandSize]byte
	err = ReadElements(hr, &h.Magic, &command, &h.Size, &h.Checksum)
	if err != nil {
		return
	}

	// Strip trailing zeros from command string.
	h.Command = string(bytes.TrimRight(command[:], "\x00"))

	return int(h.Size), nil
}
