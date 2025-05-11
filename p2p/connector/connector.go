package connector

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/LiquidCats/bitcoin-lib/p2p/serializer"
	"github.com/LiquidCats/bitcoin-lib/params"
	"github.com/decred/dcrd/container/lru"
)

var setNonce = lru.NewSet[uint64](50)

type HashFunc func() (hash *serializer.Hash, height int32, err error)

type Connector struct {
	host       string
	port       uint16
	params     *params.Params
	serializer serializer.Serializer
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
	serializer serializer.Serializer,
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
func (c *Connector) writeMessage(w io.Writer, msg serializer.Message) (totalBytes int, err error) {
	body := bytes.NewBuffer([]byte{})
	err = c.serializer.Serialize(body, msg, c.ProtocolVersion(), serializer.LatestEncoding)
	if err != nil {
		return
	}

	bodyBytes := body.Bytes()

	header := bytes.NewBuffer([]byte{})

	msgHeader := &serializer.MessageHeader{
		Magic:   c.params.Net,
		Command: msg.Command(),
		Size:    msg.MaxPayloadLength(c.ProtocolVersion()),
	}
	copy(msgHeader.Checksum[:], serializer.Checksum(bodyBytes)[0:4])

	n, err := msgHeader.Serialize(header)
	if err != nil {
		return
	}
	totalBytes += n

	n, err = w.Write(header.Bytes())
	if err != nil {
		return
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
func (c *Connector) readMessageHeader(r io.Reader) (totalBytes int, msgHeader *serializer.MessageHeader, err error) {
	msgHeader = &serializer.MessageHeader{}

	n, err := msgHeader.Deserialize(r)
	if err != nil {
		return
	}
	totalBytes += n

	// Enforce maximum message payload.
	if msgHeader.Size > serializer.MaxMessagePayload {
		err = fmt.Errorf("ReadMessage: message payload is too large - header  indicates %d bytes, but max message payload is %d bytes.", msgHeader.Size, serializer.MaxMessagePayload)
		return

	}

	// Check for messages from the wrong bitcoin network.
	if msgHeader.Magic != c.params.Net {
		discardInput(r, msgHeader.Size)
		err = fmt.Errorf("ReadMessage: message from other network [%v]", msgHeader.Magic)
		return
	}

	// Check for malformed commands.
	command := msgHeader.Command
	if !utf8.ValidString(command) {
		discardInput(r, msgHeader.Size)
		err = fmt.Errorf("ReadMessage: invalid command %v", []byte(command))
		return
	}

	return
}

func (c *Connector) createLocalVersionMessage() (*serializer.VersionPayload, error) {
	var blockNum int32
	if c.cfg.NewestBlock != nil {
		var err error
		_, blockNum, err = c.cfg.NewestBlock()
		if err != nil {
			return nil, err
		}
	}

	nonce := uint64(rand.Int63())
	setNonce.Put(nonce)

	versionMsg := &serializer.VersionPayload{
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
	_, err = c.readMessage(conn, &serializer.VersionPayload{})
	if err != nil {
		return err
	}

	// 4. Receive VerAck Message.
	_, err = c.readMessage(conn, &serializer.VerAckPayload{})
	if err != nil {
		return err
	}

	// 5. Send VeAck
	_, err = c.writeMessage(conn, &serializer.VerAckPayload{})
	if err != nil {
		return err
	}

	// 6. Continuously read messages.
	for {

		// Read rest of the header.
	msgHeader:
		c.readMessageHeader(conn)
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

// discardInput reads n bytes from reader r in chunks and discards the read
// bytes.  This is used to skip payloads when various errors occur and helps
// prevent rogue nodes from causing massive memory allocation through forging
// header length.
func discardInput(r io.Reader, n uint32) {
	maxSize := uint32(10 * 1024) // 10k at a time
	numReads := n / maxSize
	bytesRemaining := n % maxSize
	if n > 0 {
		buf := make([]byte, maxSize)
		for i := uint32(0); i < numReads; i++ {
			io.ReadFull(r, buf)
		}
	}
	if bytesRemaining > 0 {
		buf := make([]byte, bytesRemaining)
		io.ReadFull(r, buf)
	}
}
