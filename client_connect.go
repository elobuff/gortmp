package rtmp

import (
	"bytes"

	"github.com/elobuff/goamf"
	"github.com/pborman/uuid"
)

func (c *Client) connect() (id string, err error) {
	var msg *Message
	msg, err = EncodeConnect(c)
	if err != nil {
		return id, Error("client connect: unable to encode connect command: %s", err)
	}

	var response *Response
	response, err = c.Call(msg, 10)
	if err != nil {
		return id, Error("client connect: unable to complete connect: %s", err)
	}

	if response.Name != "_result" {
		return id, Error("client connect: connect result unsuccessful: %#v", response)
	}

	obj, ok := response.Objects[1].(amf.Object)
	if !ok {
		return id, Error("client connect: unable to find connect response: %#v", response)
	}

	if obj["code"] != "NetConnection.Connect.Success" {
		return id, Error("client connect: connection was unsuccessful: %#v", response)
	}

	id = obj["id"].(string)

	return
}

func EncodeConnect(c *Client) (msg *Message, err error) {
	tid := c.NextTransactionId()

	opts := make(amf.Object)
	opts["app"] = ""
	opts["flashVer"] = "WIN 10,1,85,3"
	opts["swfUrl"] = "app://mod_ser.dat"
	opts["tcUrl"] = c.url
	opts["fpad"] = false
	opts["capabilities"] = 239
	opts["audioCodecs"] = 3191
	opts["videoCodecs"] = 252
	opts["videoFunction"] = 1
	opts["pageUrl"] = nil
	opts["objectEncoding"] = 3

	cmh := make(amf.Object)
	cmh["DSMessagingVersion"] = 1
	cmh["DSId"] = "my-rtmps"

	cm := *amf.NewTypedObject()
	cm.Type = "flex.messaging.messages.CommandMessage"
	cm.Object["destination"] = ""
	cm.Object["operation"] = 5
	cm.Object["correlationId"] = ""
	cm.Object["timestamp"] = 0
	cm.Object["timeToLive"] = 0
	cm.Object["messageId"] = uuid.New()
	cm.Object["body"] = nil
	cm.Object["headers"] = cmh

	enc := new(amf.Encoder)
	buf := new(bytes.Buffer)
	if _, err = enc.Encode(buf, "connect", 0); err != nil {
		return
	}
	if _, err = enc.Encode(buf, tid, 0); err != nil {
		return
	}
	if _, err = enc.Encode(buf, opts, 0); err != nil {
		return
	}
	if _, err = enc.Encode(buf, false, 0); err != nil {
		return
	}
	if _, err = enc.Encode(buf, "nil", 0); err != nil {
		return
	}
	if _, err = enc.Encode(buf, "", 0); err != nil {
		return
	}
	if err = enc.EncodeAmf0Amf3Marker(buf); err != nil {
		return
	}
	if _, err = enc.Encode(buf, cm, 3); err != nil {
		return
	}

	return &Message{
		Type:          MESSAGE_TYPE_AMF0,
		ChunkStreamId: CHUNK_STREAM_ID_COMMAND,
		TransactionId: tid,
		Length:        uint32(buf.Len()),
		Buffer:        buf,
	}, err
}
