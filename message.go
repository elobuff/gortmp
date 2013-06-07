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
	Length            uint32
	Buffer            *bytes.Buffer
}

func (m *Message) RemainingBytes() uint32 {
	if m.Buffer == nil {
		return m.Length
	}

	return m.Length - uint32(m.Buffer.Len())
}

func (m *Message) DecodeCommand(dec *amf.Decoder) (*Command, error) {
	var err error
	var obj interface{}

	cmd := new(Command)
	cmd.Version = AMF0

	if m.ChunkStreamId != CHUNK_STREAM_ID_COMMAND {
		return cmd, Error("message is not a command message")
	}

	switch m.Type {
	case MESSAGE_TYPE_AMF3:
		cmd.Version = AMF3
		_, err = m.Buffer.ReadByte()
		if err != nil {
			return cmd, Error("unable to read first byte of amf3 message")
		}
		fallthrough

	case MESSAGE_TYPE_AMF0:
		cmd.Name, err = dec.DecodeAmf0String(m.Buffer, true)
		if err != nil {
			return cmd, Error("unable to read command from amf message")
		}

		cmd.TransactionId, err = dec.DecodeAmf0Number(m.Buffer, true)
		if err != nil {
			return cmd, Error("unable to read tid from amf message")
		}

		for m.Buffer.Len() > 0 {
			obj, err = dec.Decode(m.Buffer, 0)
			if err != nil {
				return cmd, Error("unable to read object from amf message: %s", err)
			}

			cmd.Objects = append(cmd.Objects, obj)
		}
	default:
		return cmd, Error("unable to decode message: %+v", m)
	}

	log.Debug("command decoded: %+v", cmd)

	return cmd, err
}
