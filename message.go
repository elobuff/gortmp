package rtmp

import (
	"bytes"
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
