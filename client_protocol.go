package rtmp

import (
	"encoding/binary"
)

func (c *Client) handleProtocolMessage(m *Message) {
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
