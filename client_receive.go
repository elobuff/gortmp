package rtmp

import (
	"bytes"
	"io"
)

func (c *Client) receiveLoop() {
	for {
		// Read the next header from the connection
		h, err := ReadHeader(c)
		if err != nil {
			if c.IsAlive() {
				log.Warn("unable to receive next header while connected: %s", err)
				c.Reset()
			} else {
				log.Debug("client receive: connection closed")
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
			c.Reset()
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
				Type:              h.MessageTypeId,
				ChunkStreamId:     h.ChunkStreamId,
				StreamId:          h.MessageStreamId,
				Timestamp:         h.CalculateTimestamp(),
				AbsoluteTimestamp: ts,
				Length:            h.MessageLength,
				Buffer:            new(bytes.Buffer),
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
				c.Reset()
			}

			return
		}

		if m.RemainingBytes() == 0 {
			cs.currentMessage = nil
			log.Trace("receive sending message to router: %#v", m)
			c.inMessages <- m
		} else {
			cs.currentMessage = m
		}
	}
}
