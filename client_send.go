package rtmp

import (
	"io"
	"time"
)

func (c *Client) sendLoop() {
	for {
		m, open := <-c.outMessages

		if !open {
			log.Debug("client send: channel closed, exiting")
			return
		}

		log.Trace("client send: processing message: %#v", m)

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
			log.Trace("client send: send header: %+v", h)
			_, err = h.Write(c)
			if err != nil {
				if c.IsAlive() {
					log.Warn("unable to send header: %v", err)
					c.Reset()
				}
				return
			}

			ws = rem
			if ws > c.outChunkSize {
				ws = c.outChunkSize
			}

			log.Trace("client send: send bytes: %d", ws)

			n, err = io.CopyN(c, m.Buffer, int64(ws))
			if err != nil {
				if c.IsAlive() {
					log.Warn("unable to send message")
					c.Reset()
				}
				return
			}

			rem -= uint32(n)

			// Set the header to continuation only for the
			// next iteration (if it happens).
			h.Format = HEADER_FORMAT_CONTINUATION
		}

		log.Trace("client send: send complete")

	}
}
