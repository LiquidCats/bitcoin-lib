package connector

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/LiquidCats/bitcoin-lib/p2p/serializer"
	"github.com/LiquidCats/bitcoin-lib/p2p/serializer/bitcoin"
	"github.com/LiquidCats/bitcoin-lib/p2p/serializer/payload"
	"github.com/LiquidCats/bitcoin-lib/params"
	"github.com/decred/dcrd/lru"
)

var setNonce = lru.NewCache(50)

type HashFunc func() (hash *serializer.Hash, height int32, err error)

type Serializer interface {
	Serialize(w io.Writer, msg payload.Message, pver uint32, encoding serializer.MessageEncoding) (int, error)
	Deserialize(w io.Writer, msg payload.Message, pver uint32, encoding serializer.MessageEncoding) (int, error)
}

type Connector struct {
	host       string
	port       uint16
	params     *params.Params
	serializer Serializer
	cfg        Config
}

type Config struct {
	Address           string
	ProtocolVersion   uint32
	Params            *params.Params
	DisableRelayTx    bool
	Services          serializer.ServiceFlag
	UserAgentName     string
	UserAgentVersion  string
	UserAgentComments []string
	NewestBlock       HashFunc
	_                 any
}

func NewConnector(
	cfg Config,
	serializer Serializer,
) (*Connector, error) {
	host, portStr, err := net.SplitHostPort(cfg.Address)
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, err
	}
	return &Connector{
		host:       host,
		port:       uint16(port),
		params:     cfg.Params,
		serializer: serializer,
		cfg:        cfg,
	}, nil
}

func (c *Connector) ProtocolVersion() uint32 {
	return c.cfg.ProtocolVersion
}

// writeMessage creates a complete message from a command and payload (both as hex strings)
// following the protocol header format.
func (c *Connector) writeMessage(w io.Writer, msg payload.Message) (int, error) {
	totalBytes := 0

	body := bytes.NewBuffer([]byte{})
	_, err := c.serializer.Serialize(body, msg, c.ProtocolVersion(), serializer.LatestEncoding)
	if err != nil {
		return totalBytes, err
	}

	bodyBytes := body.Bytes()

	header := bytes.NewBuffer([]byte{})

	msgHeader := &bitcoin.MessageHeader{
		Magic:   c.params.Net,
		Command: msg.Command(),
		Size:    msg.MaxPayloadLength(c.ProtocolVersion()),
	}
	copy(msgHeader.Checksum[:], payload.Checksum(bodyBytes)[0:4])

	_, err = msgHeader.Serialize(header)
	if err != nil {
		return totalBytes, err
	}

	n, err := w.Write(header.Bytes())
	if err != nil {
		return 0, err
	}
	totalBytes += n
	if body.Len() > 0 {
		n, err = w.Write(body.Bytes())
		if err != nil {
			return 0, err
		}
		totalBytes += n
	}

	return totalBytes, nil
}

// readMessage creates a complete message from a command and payload (both as hex strings)
// following the protocol header format.
func (c *Connector) readMessage(r io.Reader, msg payload.Message) (int, error) {
	return 0, nil
}

// readBytes reads exactly n bytes from conn.
func readBytes(conn net.Conn, n int) []byte {
	buf := make([]byte, n)
	if _, err := io.ReadFull(conn, buf); err != nil {
		log.Fatal(err)
	}
	return buf
}

func (c *Connector) createLocalVersionMessage() (*payload.VersionPayload, error) {
	var blockNum int32
	if c.cfg.NewestBlock != nil {
		var err error
		_, blockNum, err = c.cfg.NewestBlock()
		if err != nil {
			return nil, err
		}
	}

	nonce := uint64(rand.Int63())
	setNonce.Add(nonce)

	versionMsg := &payload.VersionPayload{
		ProtocolVersion: int32(c.ProtocolVersion()),
		Services:        c.cfg.Services,
		Timestamp:       time.Unix(time.Now().Unix(), 0),
		AddrYou: serializer.NetAddress{
			Timestamp: time.Now(),
			Services:  0,
			IP:        net.ParseIP(c.host),
			Port:      c.port,
		},
		AddrMe: serializer.NetAddress{
			Services: 0,
		},
		Nonce:          nonce,
		LastBlock:      blockNum,
		DisableRelayTx: false,
	}
	_ = versionMsg.AddUserAgent(c.cfg.UserAgentName, c.cfg.UserAgentVersion, c.cfg.UserAgentComments...)

	return versionMsg, nil
}

func (c *Connector) Connect(_ context.Context) error {
	// Open TCP connection to the node.
	conn, err := net.Dial("tcp", c.cfg.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 2. Send Version Message
	versionMsg, err := c.createLocalVersionMessage()
	if err != nil {
		return err
	}

	_, err = c.writeMessage(conn, versionMsg)
	if err != nil {
		return err
	}

	// 3. Receive Version Message.
	// Read the header: magic (4 bytes), command (12 bytes), size (4 bytes), checksum (4 bytes).
	magic := readBytes(conn, 4)
	commandBytes := readBytes(conn, 12)
	sizeBytes := readBytes(conn, 4)
	checksumBytes := readBytes(conn, 4)

	fmt.Println("<-version")
	fmt.Println("magic_bytes: " + hex.EncodeToString(magic))
	// Remove padding (null bytes) from the command string.
	fmt.Println("command:     " + strings.TrimRight(string(commandBytes), "\x00"))
	// Interpret size as little-endian uint32.
	sizeVal := binary.LittleEndian.Uint32(sizeBytes)
	fmt.Println("size:        ", sizeVal)
	fmt.Println("checksum:    " + hex.EncodeToString(checksumBytes))

	// Read payload.
	payloadData := readBytes(conn, int(sizeVal))
	fmt.Println("payload:     " + hex.EncodeToString(payloadData))
	fmt.Println()

	// 4. Receive VerAck Message.
	magic = readBytes(conn, 4)
	commandBytes = readBytes(conn, 12)
	sizeBytes = readBytes(conn, 4)
	checksumBytes = readBytes(conn, 4)

	fmt.Println("<-verack")
	fmt.Println("magic_bytes: " + hex.EncodeToString(magic))
	fmt.Println("command:     " + strings.TrimRight(string(commandBytes), "\x00"))
	sizeVal = binary.LittleEndian.Uint32(sizeBytes)
	fmt.Println("size:        ", sizeVal)
	fmt.Println("checksum:    " + hex.EncodeToString(checksumBytes))
	payloadData = readBytes(conn, int(sizeVal))
	fmt.Println("payload:     " + hex.EncodeToString(payloadData))
	fmt.Println()

	// 5. Send VeAck
	_, err = c.writeMessage(conn, &payload.VerAckPayload{})

	// 6. Continuously read messages.
	for {
		// Search for the magic bytes (4 bytes represented as 8 hex characters).
		buffer := ""
		for {
			b := readBytes(conn, 1)
			if len(b) == 0 {
				fmt.Println("Read a nil byte from the socket. Remote node disconnected.")
				os.Exit(1)
			}
			buffer += hex.EncodeToString(b)
			if len(buffer) == 8 {
				if buffer == "f9beb4d9" {
					break
				}
				buffer = ""
			}
		}
		// Read rest of the header.
		cmdBytes := readBytes(conn, 12)
		cmd := strings.TrimRight(string(cmdBytes), "\x00")
		sizeB := readBytes(conn, 4)
		msgSize := binary.LittleEndian.Uint32(sizeB)
		chkB := readBytes(conn, 4)
		payloadData := readBytes(conn, int(msgSize))

		fmt.Printf("<-%s\n", cmd)
		fmt.Println("magic_bytes: " + buffer)
		fmt.Println("command:     " + cmd)
		fmt.Println("size:        ", msgSize)
		fmt.Println("checksum:    " + hex.EncodeToString(chkB))
		fmt.Println("payload:     " + hex.EncodeToString(payloadData))
		fmt.Println()

		// Respond to "inv" messages with a "getdata" message.
		if cmd == "inv" {
			newCommand := "getdata"
			// Use the same payload as received.
			newMsg := buildMessage(newCommand, hex.EncodeToString(payloadData))
			newMsgBytes, err := hex.DecodeString(newMsg)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s->\n", newCommand)
			fmt.Println("magic_bytes: f9beb4d9")
			fmt.Println("command:     " + newCommand)
			fmt.Println("size:        ", len(payloadData))
			fmt.Println("checksum:    " + checksum(hex.EncodeToString(payloadData)))
			fmt.Println("payload:     " + hex.EncodeToString(payloadData))
			fmt.Println()
			_, err = conn.Write(newMsgBytes)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Respond to "ping" messages with a "pong" message.
		if cmd == "ping" {
			newCommand := "pong"
			newMsg := buildMessage(newCommand, hex.EncodeToString(payloadData))
			newMsgBytes, err := hex.DecodeString(newMsg)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s->\n", newCommand)
			fmt.Println("magic_bytes: f9beb4d9")
			fmt.Println("command:     " + newCommand)
			fmt.Println("size:        ", len(payloadData))
			fmt.Println("checksum:    " + checksum(hex.EncodeToString(payloadData)))
			fmt.Println("payload:     " + hex.EncodeToString(payloadData))
			fmt.Println()
			_, err = conn.Write(newMsgBytes)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
