package rtmp

import (
	"io"
)

func (c *Client) sendLoop() {
	for {
		m := <-c.outMessages

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
			log.Debug("send header: %+v", h)
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

			log.Debug("send bytes: %d", ws)

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

		log.Debug("send complete")

	}
}
