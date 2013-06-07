package rtmp

/*
func (c *Client) Invoke(destination string, operation string, objects ...interface{}) (tid uint32, err error) {
  rmh := *amf.MakeObject()
  rmh["DSRequestTimeout"] = 60
  rmh["DSId"] = c.connectionId
  rmh["DSEndpoint"] = "my-rtmps"

  rm := *amf.MakeTypedObject()
  rm.Type = "flex.messaging.messages.RemotingMessage"
  rm.Object["destination"] = destination
  rm.Object["operation"] = operation
  rm.Object["messageId"] = uuid.New()
  rm.Object["source"] = nil
  rm.Object["timestamp"] = 0
  rm.Object["timeToLive"] = 0
  rm.Object["headers"] = rmh
  rm.Object["body"] = objects

  buf := new(bytes.Buffer)
  tid = c.NextTransactionId()

  amf.WriteMarker(buf, 0x00)                           // AMF3
  amf.WriteNull(buf)                                   // command
  amf.WriteDouble(buf, float64(tid))                   // tid
  amf.WriteMarker(buf, amf.AMF0_ACMPLUS_OBJECT_MARKER) // amf3

  log.Debug("rm: %+v", rm)
  log.Debug("buf before: %+v", buf.Bytes())
  _, err = amf.AMF3_WriteValue(buf, rm)
  if err != nil {
    log.Debug("error: %s", err)
    return
  }
  log.Debug("buf after: %s", buf.Bytes())

  m := &Message{
    Type:          MESSAGE_TYPE_AMF3,
    ChunkStreamId: CHUNK_STREAM_ID_COMMAND,
    Length:        uint32(buf.Len()),
    Buffer:        buf,
  }

  c.outMessages <- msg

  return
}
*/
