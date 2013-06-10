package rtmp

import (
	"bytes"
	"github.com/elobuff/goamf"
)

type Message struct {
	Type              uint8
	ChunkStreamId     uint32
	StreamId          uint32
	Timestamp         uint32
	AbsoluteTimestamp uint32
	TransactionId     uint32
	Length            uint32
	Buffer            *bytes.Buffer
}

func (m *Message) RemainingBytes() uint32 {
	if m.Buffer == nil {
		return m.Length
	}

	return m.Length - uint32(m.Buffer.Len())
}

func (m *Message) DecodeResponse(dec *amf.Decoder) (response *Response, err error) {
	response = new(Response)

	if m.ChunkStreamId != CHUNK_STREAM_ID_COMMAND {
		return response, Error("message is not a command message")
	}

	switch m.Type {
	case MESSAGE_TYPE_AMF3:
		_, err = m.Buffer.ReadByte()
		if err != nil {
			return response, Error("unable to read first byte of amf3 message")
		}
		fallthrough

	case MESSAGE_TYPE_AMF0:
		response.Name, err = dec.DecodeAmf0String(m.Buffer, true)
		if err != nil {
			return response, Error("unable to read command from amf message")
		}

		response.TransactionId, err = dec.DecodeAmf0Number(m.Buffer, true)
		if err != nil {
			return response, Error("unable to read tid from amf message")
		}

		var obj interface{}
		for m.Buffer.Len() > 0 {
			obj, err = dec.Decode(m.Buffer, 0)
			if err != nil {
				return response, Error("unable to read object from amf message: %s", err)
			}

			response.Objects = append(response.Objects, obj)
		}
	default:
		return response, Error("unable to decode message: %+v", m)
	}

	log.Trace("command decoded: %+v", response)

	return response, err
}
