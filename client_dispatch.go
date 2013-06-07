package rtmp

import (
	"encoding/binary"
)

func (c *Client) dispatchLoop() {
	for {
		m := <-c.inMessages

		switch m.ChunkStreamId {
		case CHUNK_STREAM_ID_PROTOCOL:
			c.dispatchProtocolMessage(m)
		case CHUNK_STREAM_ID_COMMAND:
			c.dispatchCommandMessage(m)
		default:
			log.Warn("discarding message on unknown chunk stream %d: +%v", m.ChunkStreamId, m)
		}
	}
}

func (c *Client) dispatchProtocolMessage(m *Message) {
	switch m.Type {
	case MESSAGE_TYPE_CHUNK_SIZE:
		var err error
		var size uint32
		err = binary.Read(m.Buffer, binary.BigEndian, &size)
		if err != nil {
			log.Error("error decoding chunk size: %s", err)
			return
		}

		c.inChunkSize = size
		log.Debug("received chunk size, setting to %d", c.inChunkSize, size)

	case MESSAGE_TYPE_ACK_SIZE:
		log.Debug("received ack size, discarding")

	case MESSAGE_TYPE_BANDWIDTH:
		log.Debug("received bandwidth, discarding")

	default:
		log.Debug("received protocol message %d, discarding", m.Type)

	}
}

func (c *Client) dispatchCommandMessage(m *Message) {
	cmd, err := m.DecodeCommand(&c.dec)
	if err != nil {
		log.Error("unable to decode message type %d on stream %d into command, discarding: %s", m.Type, m.ChunkStreamId, err)
		return
	}

	if cmd.Name == "_result" && cmd.TransactionId == 1 {
		c.handleConnectResult(cmd)
	} else {
		c.handler.OnRtmpCommand(cmd)
		return
	}
}
