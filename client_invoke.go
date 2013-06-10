package rtmp

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"github.com/elobuff/goamf"
)

func (c *Client) EncodeInvokeCommand(destination interface{}, operation interface{}, body interface{}) (msg *Message, err error) {
	return c.EncodeInvoke("flex.messaging.messages.CommandMessage", destination, operation, body)
}

func (c *Client) EncodeInvokeRemote(destination interface{}, operation interface{}, body interface{}) (msg *Message, err error) {
	return c.EncodeInvoke("flex.messaging.messages.RemotingMessage", destination, operation, body)
}

func (c *Client) EncodeInvoke(className string, destination interface{}, operation interface{}, body interface{}) (msg *Message, err error) {
	tid := c.NextTransactionId()

	rmh := make(amf.Object)
	rmh["DSRequestTimeout"] = 60
	rmh["DSId"] = c.connectionId
	rmh["DSEndpoint"] = "my-rtmps"

	rm := *amf.NewTypedObject()
	rm.Type = className
	rm.Object["destination"] = destination
	rm.Object["operation"] = operation
	rm.Object["messageId"] = uuid.New()
	rm.Object["source"] = nil
	rm.Object["timestamp"] = 0
	rm.Object["timeToLive"] = 0
	rm.Object["headers"] = rmh
	rm.Object["body"] = body

	buf := new(bytes.Buffer)

	// amf3 empty byte
	if err = amf.WriteMarker(buf, 0x00); err != nil {
		return msg, Error("client invoke: could not encode amf3 0x00 byte: %s", err)
	}

	// amf0 command
	if _, err = c.enc.EncodeAmf0Null(buf, true); err != nil {
		return msg, Error("client invoke: could not encode amf0 command: %s", err)
	}

	// amf0 tid
	if _, err = c.enc.EncodeAmf0Number(buf, float64(tid), true); err != nil {
		return msg, Error("client invoke: could not encode amf0 tid: %s", err)
	}

	// amf0 nil?
	if err = amf.WriteMarker(buf, 0x05); err != nil {
		return msg, Error("client invoke: could not encode amf3 0x05 byte: %s", err)
	}

	// amf0 amf3
	if err = c.enc.EncodeAmf0Amf3Marker(buf); err != nil {
		return msg, Error("client invoke: could not encode amf3 0x11 byte: %s", err)
	}

	// amf3 object
	if _, err = c.enc.EncodeAmf3(buf, rm); err != nil {
		return msg, Error("client invoke: could not encode amf3 object: %s", err)
	}

	return &Message{
		Type:          MESSAGE_TYPE_AMF3,
		ChunkStreamId: CHUNK_STREAM_ID_COMMAND,
		TransactionId: tid,
		Length:        uint32(buf.Len()),
		Buffer:        buf,
	}, nil
}
