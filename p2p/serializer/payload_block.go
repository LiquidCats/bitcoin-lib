package serializer

import (
	"time"
)

type BlockPayload struct {
	Header       BlockHeaderPayload
	Transactions []*TxPayload
}

type BlockHeaderPayload struct {
	Version int32

	// Hash of the previous block header in the block chain.
	PrevBlock Hash

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot Hash

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Timestamp time.Time

	// Difficulty target for the block.
	Bits uint32

	// Nonce used to generate the block.
	Nonce uint32
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *BlockPayload) Command() string {
	return CmdBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *BlockPayload) MaxPayloadLength(pver uint32) uint32 {
	// Block header at 80 bytes + transaction count + max transactions
	// which can vary up to the MaxBlockPayload (including the block header
	// and transaction count).
	return MaxBlockPayload
}
