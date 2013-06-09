package rtmp

func (c *Client) routeLoop() {
	for {
		msg, open := <-c.inMessages

		log.Trace("client route: received message: %#v", msg)

		if !open {
			log.Trace("client route: channel closed, exiting")
			return
		}

		switch msg.ChunkStreamId {
		case CHUNK_STREAM_ID_PROTOCOL:
			c.handleProtocolMessage(msg)
		case CHUNK_STREAM_ID_COMMAND:
			c.routeCommandMessage(msg)
		default:
			log.Warn("discarding message on unknown chunk stream %d: +%v", msg.ChunkStreamId, msg)
		}
	}
}

func (c *Client) routeCommandMessage(msg *Message) {
	response, err := msg.DecodeResponse(&c.dec)
	if err != nil {
		log.Error("unable to decode message type %d on stream %d into command, discarding: %s", msg.Type, msg.ChunkStreamId, err)
		return
	}

	tid := uint32(response.TransactionId)

	c.responsesMutex.Lock()
	c.responses[tid] = response
	c.responsesMutex.Unlock()
}
