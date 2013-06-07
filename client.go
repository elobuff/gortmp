package rtmp

import (
	"crypto/tls"
	"github.com/elobuff/goamf"
	"net"
	"net/url"
	"sync/atomic"
)

type ClientHandler interface {
	OnRtmpConnect()
	OnRtmpDisconnect()
	OnRtmpCommand(command *Command)
}

type Client struct {
	url string

	handler   ClientHandler
	connected bool

	conn net.Conn
	enc  amf.Encoder
	dec  amf.Decoder

	outBytes        uint32
	outMessages     chan *Message
	outWindowSize   uint32
	outChunkSize    uint32
	outChunkStreams map[uint32]*OutboundChunkStream

	inBytes        uint32
	inMessages     chan *Message
	inNotify       chan uint8
	inWindowSize   uint32
	inChunkSize    uint32
	inChunkStreams map[uint32]*InboundChunkStream

	lastTransactionId uint32
	connectionId      string
}

func NewClient(url string, handler ClientHandler) (*Client, error) {
	c := &Client{
		url: url,

		connected:       false,
		handler:         handler,
		enc:             *new(amf.Encoder),
		dec:             *new(amf.Decoder),
		outMessages:     make(chan *Message, 100),
		outChunkSize:    DEFAULT_CHUNK_SIZE,
		outWindowSize:   DEFAULT_WINDOW_SIZE,
		outChunkStreams: make(map[uint32]*OutboundChunkStream),
		inMessages:      make(chan *Message, 100),
		inChunkSize:     DEFAULT_CHUNK_SIZE,
		inWindowSize:    DEFAULT_WINDOW_SIZE,
		inChunkStreams:  make(map[uint32]*InboundChunkStream),
	}

	err := c.Connect()
	if err != nil {
		return c, err
	}

	return c, err
}

func (c *Client) Connect() (err error) {
	log.Info("connecting to %s", c.url)

	url, err := url.Parse(c.url)
	if err != nil {
		return err
	}

	switch url.Scheme {
	case "rtmp":
		c.conn, err = net.Dial("tcp", url.Host)
	case "rtmps":
		config := &tls.Config{InsecureSkipVerify: true}
		c.conn, err = tls.Dial("tcp", url.Host, config)
	default:
		return Error("Unsupported scheme: %s", url.Scheme)
	}

	log.Debug("handshaking with %s", c.url)

	err = c.handshake()
	if err != nil {
		return err
	}

	log.Debug("sending connect command to %s", c.url)

	err = c.invokeConnect()
	if err != nil {
		return err
	}

	go c.dispatchLoop()
	go c.receiveLoop()
	go c.sendLoop()

	return nil
}

func (c *Client) NextTransactionId() uint32 {
	return atomic.AddUint32(&c.lastTransactionId, 1)
}

func (c *Client) Disconnect() {
	c.connected = false
	c.conn.Close()
	c.handler.OnRtmpDisconnect()

	log.Info("disconnected from %s", c.url, c.outBytes, c.inBytes)
}

func (c *Client) Read(p []byte) (n int, err error) {
	n, err = c.conn.Read(p)
	c.inBytes += uint32(n)
	log.Trace("read %d", n)
	return n, err
}

func (c *Client) Write(p []byte) (n int, err error) {
	n, err = c.conn.Write(p)
	c.outBytes += uint32(n)
	log.Trace("write %d", n)
	return n, err
}
