package rtmp

import (
  "bytes"
  "crypto/tls"
  "errors"
  "fmt"
  "io"
  "net"
  "net/url"
)

type ClientHandler interface {
  OnConnect()
  OnDisconnect()
  OnReceive(message *Message)
}

type Client struct {
  url               string

  handler           ClientHandler
  connected         bool

  conn              net.Conn

  outBytes          uint32
  outMessages       chan *Message
  outWindowSize     uint32
  outChunkSize      uint32
  outChunkStreams   map[uint32]*OutboundChunkStream

  inBytes           uint32
  inMessages        chan *Message
  inNotify          chan uint8
  inWindowSize      uint32
  inChunkSize       uint32
  inChunkStreams    map[uint32]*InboundChunkStream

  lastTid           uint32
}

func NewClient(url string) (*Client, error) {
  c := &Client{
    url: url,

    connected:        false,

    outMessages:      make(chan *Message),
    outChunkSize:     DEFAULT_CHUNK_SIZE,
    outWindowSize:    DEFAULT_WINDOW_SIZE,
    outChunkStreams:  make(map[uint32]*OutboundChunkStream),

    inMessages:       make(chan *Message, 100),
    inChunkSize:      DEFAULT_CHUNK_SIZE,
    inWindowSize:     DEFAULT_WINDOW_SIZE,
    inChunkStreams:   make(map[uint32]*InboundChunkStream),
  }

  err := c.Connect()
  if err != nil {
    return c, err
  }

  return c, err
}

func (c *Client) Connect() (err error) {
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
    return errors.New(fmt.Sprintf("Unsupported scheme: %s", url.Scheme))
  }

  err = c.handshake()
  if err != nil {
    return err
  }

  c.connected = true

  go c.dispatchLoop()
  go c.receiveLoop()
  go c.sendLoop()

  log.Info("connected to %s", c.url)

  return nil
}

func (c *Client) Disconnect() {
  c.connected = false
  c.conn.Close()

  log.Info("disconnected from %s", c.url, c.outBytes, c.inBytes)
}

func (c *Client) dispatchLoop() {
  for {
    m := <- c.inMessages

    switch m.ChunkStreamId {
    case CHUNK_STREAM_ID_PROTOCOL:
      log.Debug("dispatch protocol message")
    case CHUNK_STREAM_ID_COMMAND:
      log.Debug("dispatch command message")
    }
  }
}

func (c *Client) sendLoop() {
  for {
    m := <- c.outMessages

    var cs *OutboundChunkStream = c.outChunkStreams[m.ChunkStreamId]
    if cs == nil {
      cs = NewOutboundChunkStream(m.ChunkStreamId)
    }

    h := cs.NewOutboundHeader(m)

    var n int64 = 0
    var err error
    var ws uint32 = 0
    var rem uint32 = m.Length

    for rem > 0 {
      log.Debug("rem is %d", rem)
      log.Debug("send message header: %+v", h)
      _, err = h.Write(c)
      if err != nil {
        if c.connected {
          log.Warn("unable to send header: %v", err)
          c.Disconnect()
        }
        return
      }

      ws = rem
      if ws > c.outChunkSize {
        ws = c.outChunkSize
      }

      n, err = io.CopyN(c, m.Buffer, int64(ws))
      if err != nil {
        if c.connected {
          log.Warn("unable to send message")
          c.Disconnect()
        }
        return
      }

      rem -= uint32(n)

      // Set the header to continuation only for the
      // next iteration (if it happens).
      h.Format = HEADER_FORMAT_CONTINUATION
    }

    log.Debug("finished sending message")

  }
}

func (c *Client) receiveLoop() {
  for {
    // Read the next header from the connection
    h, err := ReadHeader(c)
    if err != nil {
      if c.connected {
        log.Warn("unable to receive next header while connected")
        c.Disconnect()
      }
      return
    }

    // Determine whether or not we already have a chunk stream
    // allocated for this ID. If we don't, create one.
    var cs *InboundChunkStream = c.inChunkStreams[h.ChunkStreamId]
    if cs == nil {
      cs = NewInboundChunkStream(h.ChunkStreamId)
      c.inChunkStreams[h.ChunkStreamId] = cs
    }

    var ts uint32
    var m *Message

    if (cs.lastHeader == nil) && (h.Format != HEADER_FORMAT_FULL) {
      log.Warn("unable to find previous header on chunk stream %d", h.ChunkStreamId)
      c.Disconnect()
      return
    }

    switch h.Format {
    case HEADER_FORMAT_FULL:
      // If it's an entirely new header, replace the reference in
      // the chunk stream and set the working timestamp from
      // the header.
      cs.lastHeader = &h
      ts = h.Timestamp

    case HEADER_FORMAT_SAME_STREAM:
      // If it's the same stream, use the last message stream id,
      // but otherwise use values from the header.
      h.MessageStreamId = cs.lastHeader.MessageStreamId
      cs.lastHeader = &h
      ts = cs.lastInAbsoluteTimestamp + h.Timestamp

    case HEADER_FORMAT_SAME_LENGTH_AND_STREAM:
      // If it's the same length and stream, copy values from the
      // last header and replace it.
      h.MessageStreamId = cs.lastHeader.MessageStreamId
      h.MessageLength = cs.lastHeader.MessageLength
      h.MessageTypeId = cs.lastHeader.MessageTypeId
      cs.lastHeader = &h
      ts = cs.lastInAbsoluteTimestamp + h.Timestamp

    case HEADER_FORMAT_CONTINUATION:
      // A full continuation of the previous stream. Copy all values.
      h.MessageStreamId = cs.lastHeader.MessageStreamId
      h.MessageLength = cs.lastHeader.MessageLength
      h.MessageTypeId = cs.lastHeader.MessageTypeId
      h.Timestamp = cs.lastHeader.Timestamp
      ts = cs.lastInAbsoluteTimestamp + cs.lastHeader.Timestamp

      // If there's a message already started, use it.
      if cs.currentMessage != nil {
        m = cs.currentMessage
      }
    }

    if m == nil {
      m = &Message{
        Type: h.MessageTypeId,
        ChunkStreamId: h.ChunkStreamId,
        StreamId: h.MessageStreamId,
        Timestamp: h.CalculateTimestamp(),
        AbsoluteTimestamp: ts,
        Length: h.MessageLength,
        Buffer: new(bytes.Buffer),
      }
    }

    cs.lastInAbsoluteTimestamp = ts

    rs := m.RemainingBytes()
    if rs > c.inChunkSize {
      rs = c.inChunkSize
    }

    _, err = io.CopyN(m.Buffer, c, int64(rs))
    if err != nil {
      if c.connected {
        log.Warn("unable to copy %d message bytes from buffer", rs)
        c.Disconnect()
      }

      return
    }

    if m.RemainingBytes() == 0 {
      cs.currentMessage = nil
      c.inMessages <- m
    } else {
      cs.currentMessage = m
    }
  }
}

func (c *Client) Read(p []byte) (n int, err error) {
  n, err = c.conn.Read(p)
  c.inBytes += uint32(n)
  log.Debug("read %d", n)
  return n, err
}

func (c *Client) Write(p []byte) (n int, err error) {
  n, err = c.conn.Write(p)
  c.outBytes += uint32(n)
  log.Debug("write %d", n)
  return n, err
}
