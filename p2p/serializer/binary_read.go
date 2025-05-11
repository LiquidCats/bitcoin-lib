package serializer

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/LiquidCats/bitcoin-lib/params"
)

// Uint32Time represents a unix timestamp encoded with a uint32.  It is used as
// a way to signal the ReadElement function how to decode a timestamp into a Go
// time.Time since it is otherwise ambiguous.
type Uint32Time time.Time

// Int64Time represents a unix timestamp encoded with an int64.  It is used as
// a way to signal the ReadElement function how to decode a timestamp into a Go
// time.Time since it is otherwise ambiguous.
type Int64Time time.Time

// ReadElement reads the next sequence of bytes from r using little endian
// depending on the concrete type of element pointed to.
func ReadElement(r io.Reader, element interface{}) error {
	// Attempt to read the element based on the concrete type via fast
	// type assertions first.
	switch e := element.(type) {
	case *int32:
		rv, err := binarySerializer.Uint32(r, littleEndian)
		if err != nil {
			return err
		}
		*e = int32(rv)
		return nil

	case *uint32:
		rv, err := binarySerializer.Uint32(r, littleEndian)
		if err != nil {
			return err
		}
		*e = rv
		return nil

	case *int64:
		rv, err := binarySerializer.Uint64(r, littleEndian)
		if err != nil {
			return err
		}
		*e = int64(rv)
		return nil

	case *uint64:
		rv, err := binarySerializer.Uint64(r, littleEndian)
		if err != nil {
			return err
		}
		*e = rv
		return nil

	case *bool:
		rv, err := binarySerializer.Uint8(r)
		if err != nil {
			return err
		}
		if rv == 0x00 {
			*e = false
		} else {
			*e = true
		}
		return nil

	// Unix timestamp encoded as a uint32.
	case *Uint32Time:
		rv, err := binarySerializer.Uint32(r, binary.LittleEndian)
		if err != nil {
			return err
		}
		*e = Uint32Time(time.Unix(int64(rv), 0))
		return nil

	// Unix timestamp encoded as an int64.
	case *Int64Time:
		rv, err := binarySerializer.Uint64(r, binary.LittleEndian)
		if err != nil {
			return err
		}
		*e = Int64Time(time.Unix(int64(rv), 0))
		return nil

	// Message header checksum.
	case *[4]byte:
		_, err := io.ReadFull(r, e[:])
		if err != nil {
			return err
		}
		return nil

	// Message header command.
	case *[CommandSize]uint8:
		_, err := io.ReadFull(r, e[:])
		if err != nil {
			return err
		}
		return nil

	// IP address.
	case *[16]byte:
		_, err := io.ReadFull(r, e[:])
		if err != nil {
			return err
		}
		return nil

	case *Hash:
		_, err := io.ReadFull(r, e[:])
		if err != nil {
			return err
		}
		return nil

	case *ServiceFlag:
		rv, err := binarySerializer.Uint64(r, littleEndian)
		if err != nil {
			return err
		}
		*e = ServiceFlag(rv)
		return nil

	case *InvType:
		rv, err := binarySerializer.Uint32(r, littleEndian)
		if err != nil {
			return err
		}
		*e = InvType(rv)
		return nil

	case *params.BitcoinNet:
		rv, err := binarySerializer.Uint32(r, littleEndian)
		if err != nil {
			return err
		}
		*e = params.BitcoinNet(rv)
		return nil
	}

	// Fall back to the slower binary.Read if a fast path was not available
	// above.
	return binary.Read(r, littleEndian, element)
}

// ReadElements reads multiple items from r.  It is equivalent to multiple
// calls to ReadElement.
func ReadElements(r io.Reader, elements ...interface{}) error {
	for _, element := range elements {
		err := ReadElement(r, element)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadNetAddress reads an encoded NetAddress from r depending on the protocol
// version and whether or not the timestamp is included per ts.  Some messages
// like version do not include the timestamp.
func ReadNetAddress(r io.Reader, pver uint32, na *NetAddress, ts bool) error {
	buf := binarySerializer.Borrow()
	defer binarySerializer.Return(buf)

	err := ReadNetAddressBuf(r, pver, na, ts, buf)
	return err
}

// ReadNetAddressBuf reads an encoded NetAddress from r depending on the
// protocol version and whether or not the timestamp is included per ts.  Some
// messages like version do not include the timestamp.
//
// If b is non-nil, the provided buffer will be used for serializing small
// values.  Otherwise a buffer will be drawn from the binarySerializer's pool
// and return when the method finishes.
//
// NOTE: b MUST either be nil or at least an 8-byte slice.
func ReadNetAddressBuf(r io.Reader, pver uint32, na *NetAddress, ts bool,
	buf []byte) error {

	var (
		timestamp time.Time
		services  ServiceFlag
		ip        [16]byte
		port      uint16
	)

	// NOTE: The bitcoin protocol uses a uint32 for the timestamp so it will
	// stop working somewhere around 2106.  Also timestamp wasn't added until
	// protocol version >= NetAddressTimeVersion
	if ts && pver >= NetAddressTimeVersion {
		if _, err := io.ReadFull(r, buf[:4]); err != nil {
			return err
		}
		timestamp = time.Unix(int64(littleEndian.Uint32(buf[:4])), 0)
	}

	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	services = ServiceFlag(littleEndian.Uint64(buf))

	if _, err := io.ReadFull(r, ip[:]); err != nil {
		return err
	}

	// Sigh.  Bitcoin protocol mixes little and big endian.
	if _, err := io.ReadFull(r, buf[:2]); err != nil {
		return err
	}
	port = bigEndian.Uint16(buf[:2])

	*na = NetAddress{
		Timestamp: timestamp,
		Services:  services,
		IP:        ip[:],
		Port:      port,
	}
	return nil
}

// ReadVarString reads a variable length string from r and returns it as a Go
// string.  A variable length string is encoded as a variable length integer
// containing the length of the string followed by the bytes that represent the
// string itself.  An error is returned if the length is greater than the
// maximum block payload size since it helps protect against memory exhaustion
// attacks and forced panics through malformed messages.
func ReadVarString(r io.Reader, pver uint32) (string, error) {
	buf := binarySerializer.Borrow()
	defer binarySerializer.Return(buf)

	str, err := readVarStringBuf(r, pver, buf)
	return str, err
}

// readVarStringBuf reads a variable length string from r and returns it as a Go
// string.  A variable length string is encoded as a variable length integer
// containing the length of the string followed by the bytes that represent the
// string itself.  An error is returned if the length is greater than the
// maximum block payload size since it helps protect against memory exhaustion
// attacks and forced panics through malformed messages.
//
// If b is non-nil, the provided buffer will be used for serializing small
// values.  Otherwise a buffer will be drawn from the binarySerializer's pool
// and return when the method finishes.
//
// NOTE: b MUST either be nil or at least an 8-byte slice.
func readVarStringBuf(r io.Reader, pver uint32, buf []byte) (string, error) {
	count, err := ReadVarIntBuf(r, pver, buf)
	if err != nil {
		return "", err
	}

	// Prevent variable length strings that are larger than the maximum
	// message size.  It would be possible to cause memory exhaustion and
	// panics without a sane upper bound on this count.
	if count > MaxMessagePayload {
		return "", fmt.Errorf("ReadVarString: variable length string is too long  [count %d, max %d]", count, MaxMessagePayload)
	}

	str := make([]byte, count)
	_, err = io.ReadFull(r, str)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

// ReadVarIntBuf reads a variable length integer from r using a preallocated
// scratch buffer and returns it as a uint64.
//
// NOTE: buf MUST at least an 8-byte slice.
func ReadVarIntBuf(r io.Reader, _ uint32, buf []byte) (uint64, error) {
	if _, err := io.ReadFull(r, buf[:1]); err != nil {
		return 0, err
	}
	discriminant := buf[0]

	var rv uint64
	switch discriminant {
	case 0xff:
		if _, err := io.ReadFull(r, buf); err != nil {
			return 0, err
		}
		rv = littleEndian.Uint64(buf)

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		min := uint64(0x100000000)
		if rv < min {
			return 0, fmt.Errorf("ReadVarInt: non-canonical varint %x - discriminant %x must encode a value greater than %x", rv, discriminant, min)
		}

	case 0xfe:
		if _, err := io.ReadFull(r, buf[:4]); err != nil {
			return 0, err
		}
		rv = uint64(littleEndian.Uint32(buf[:4]))

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		min := uint64(0x10000)
		if rv < min {
			return 0, fmt.Errorf("ReadVarInt: non-canonical varint %x - discriminant %x must encode a value greater than %x", rv, discriminant, min)
		}

	case 0xfd:
		if _, err := io.ReadFull(r, buf[:2]); err != nil {
			return 0, err
		}
		rv = uint64(littleEndian.Uint16(buf[:2]))

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		min := uint64(0xfd)
		if rv < min {
			return 0, fmt.Errorf("ReadVarInt: non-canonical varint %x - discriminant %x must encode a value greater than %x", rv, discriminant, min)
		}

	default:
		rv = uint64(discriminant)
	}

	return rv, nil
}
