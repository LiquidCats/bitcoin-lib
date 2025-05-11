package serializer

import (
	"io"
	"math"
)

// WriteNetAddress serializes a NetAddress to w depending on the protocol
// version and whether or not the timestamp is included per ts.  Some messages
// like version do not include the timestamp.
func WriteNetAddress(w io.Writer, pver uint32, na *NetAddress, ts bool) error {
	buf := binarySerializer.Borrow()
	defer binarySerializer.Return(buf)
	err := WriteNetAddressBuf(w, pver, na, ts, buf)

	return err
}

// WriteNetAddressBuf serializes a NetAddress to w depending on the protocol
// version and whether or not the timestamp is included per ts.  Some messages
// like version do not include the timestamp.
//
// If b is non-nil, the provided buffer will be used for serializing small
// values.  Otherwise a buffer will be drawn from the binarySerializer's pool
// and return when the method finishes.
//
// NOTE: b MUST either be nil or at least an 8-byte slice.
func WriteNetAddressBuf(w io.Writer, pver uint32, na *NetAddress, ts bool, buf []byte) error {
	// NOTE: The bitcoin protocol uses a uint32 for the timestamp so it will
	// stop working somewhere around 2106.  Also timestamp wasn't added until
	// until protocol version >= NetAddressTimeVersion.
	if ts && pver >= NetAddressTimeVersion {
		littleEndian.PutUint32(buf[:4], uint32(na.Timestamp.Unix()))
		if _, err := w.Write(buf[:4]); err != nil {
			return err
		}
	}

	littleEndian.PutUint64(buf, uint64(na.Services))
	if _, err := w.Write(buf); err != nil {
		return err
	}

	// Ensure to always write 16 bytes even if the ip is nil.
	var ip [16]byte
	if na.IP != nil {
		copy(ip[:], na.IP.To16())
	}
	if _, err := w.Write(ip[:]); err != nil {
		return err
	}

	// Sigh.  Bitcoin protocol mixes little and big endian.
	bigEndian.PutUint16(buf[:2], na.Port)
	_, err := w.Write(buf[:2])

	return err
}

// WriteVarString serializes str to w as a variable length integer containing
// the length of the string followed by the bytes that represent the string
// itself.
func WriteVarString(w io.Writer, pver uint32, str string) error {
	buf := binarySerializer.Borrow()
	defer binarySerializer.Return(buf)

	err := writeVarStringBuf(w, pver, str, buf)
	return err
}

// writeVarStringBuf serializes str to w as a variable length integer containing
// the length of the string followed by the bytes that represent the string
// itself.
//
// If b is non-nil, the provided buffer will be used for serializing small
// values.  Otherwise a buffer will be drawn from the binarySerializer's pool
// and return when the method finishes.
//
// NOTE: b MUST either be nil or at least an 8-byte slice.
func writeVarStringBuf(w io.Writer, pver uint32, str string, buf []byte) error {
	err := WriteVarIntBuf(w, pver, uint64(len(str)), buf)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(str))
	return err
}

// WriteVarIntBuf serializes val to w using a variable number of bytes depending
// on its value using a preallocated scratch buffer.
//
// NOTE: buf MUST at least an 8-byte slice.
func WriteVarIntBuf(w io.Writer, pver uint32, val uint64, buf []byte) error {
	switch {
	case val < 0xfd:
		buf[0] = uint8(val)
		_, err := w.Write(buf[:1])
		return err

	case val <= math.MaxUint16:
		buf[0] = 0xfd
		littleEndian.PutUint16(buf[1:3], uint16(val))
		_, err := w.Write(buf[:3])
		return err

	case val <= math.MaxUint32:
		buf[0] = 0xfe
		littleEndian.PutUint32(buf[1:5], uint32(val))
		_, err := w.Write(buf[:5])
		return err

	default:
		buf[0] = 0xff
		if _, err := w.Write(buf[:1]); err != nil {
			return err
		}

		littleEndian.PutUint64(buf, val)
		_, err := w.Write(buf)
		return err
	}
}
