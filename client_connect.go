package rtmp

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"github.com/elobuff/goamf"
)

func (c *Client) invokeConnect() (err error) {
	buf := new(bytes.Buffer)
	tid := c.NextTransactionId()

	c.enc.Encode(buf, "connect", 0)
	c.enc.Encode(buf, float64(tid), 0)

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

	c.enc.Encode(buf, opts, 0)
	c.enc.Encode(buf, false, 0)
	c.enc.Encode(buf, "nil", 0)
	c.enc.Encode(buf, "", 0)

	cmh := make(amf.Object)
	cmh["DSMessagingVersion"] = 1
	cmh["DSId"] = "my-rtmps"

	cm := amf.NewTypedObject()
	cm.Type = "flex.messaging.messages.CommandMessage"
	cm.Object["destination"] = ""
	cm.Object["operation"] = 5
	cm.Object["correlationId"] = ""
	cm.Object["timestamp"] = 0
	cm.Object["timeToLive"] = 0
	cm.Object["messageId"] = uuid.New()
	cm.Object["body"] = nil
	cm.Object["headers"] = cmh

	c.enc.Encode(buf, cm, 0)

	m := &Message{
		ChunkStreamId: CHUNK_STREAM_ID_COMMAND,
		Type:          MESSAGE_TYPE_AMF0,
		Length:        uint32(buf.Len()),
		Buffer:        buf,
	}

	c.outMessages <- m

	return
}

func (c *Client) handleConnectResult(cmd *Command) {
	log.Debug("connect response received from %s", c.url)

	obj, ok := cmd.Objects[1].(amf.Object)
	if !ok {
		log.Error("unable to find connect result in command response")
		c.Disconnect()
		return
	}

	if obj["code"] != "NetConnection.Connect.Success" {
		log.Error("connect unsuccessful: %s", obj["code"])
		c.Disconnect()
		return
	}

	c.connected = true
	c.connectionId = obj["id"].(string)
	c.handler.OnRtmpConnect()
	log.Info("connected to %s (%s)", c.url, c.connectionId)

	return
}
