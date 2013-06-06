package rtmp

import (
	"encoding/binary"
	"errors"
	"io"
)

type Header struct {
	Format            uint8
	ChunkStreamId     uint32
	MessageLength     uint32
	MessageTypeId     uint8
	MessageStreamId   uint32
	Timestamp         uint32
	ExtendedTimestamp uint32
}

func NewHeader() *Header {
	return &Header{}
}

func ReadHeader(r io.Reader) (Header, error) {
	h := *NewHeader()
	u8 := make([]byte, 1)
	u16 := make([]byte, 2)
	u32 := make([]byte, 4)

	// The first byte we read from the header will indicate the
	// format of the packet and chunk stream id
	_, err := r.Read(u8)
	if err != nil {
		return h, err
	}

	// Determine the packet format from the byte
	h.Format = (u8[0] & 0xC0) >> 6

	// Determine Chunk Stream ID using the remainder of the byte
	h.ChunkStreamId = uint32(u8[0] & 0x3F)

	switch h.ChunkStreamId {
	case 0:
		// A Chunk Stream ID of 0 indicates that the real value
		// is between 64-319, which is reached by adding 64 to the
		// next byte.
		_, err = r.Read(u8)
		if err != nil {
			return h, err
		}
		h.ChunkStreamId = uint32(64) + uint32(u8[0])

	case 1:
		// A Chunk Stream ID of 1 indicates that the real value
		// is between 64-65599 and can be reached by adding 64 to
		// the next byte and then multiplying the one after it
		// by 256.
		_, err = r.Read(u16)
		if err != nil {
			return h, err
		}
		h.ChunkStreamId = uint32(u16[0]) + (256 * uint32(u16[1]))
	}

	// If the header is full, same length, or same length
	// and stream, then we only need to extract the timestamp.
	if h.Format <= HEADER_FORMAT_SAME_LENGTH_AND_STREAM {
		_, err = r.Read(u32[1:])
		if err != nil {
			return h, err
		}
		h.Timestamp = binary.BigEndian.Uint32(u32)
	}

	// If the header is full or same stream, then we also
	// need to extract the message size and message type.
	if h.Format <= HEADER_FORMAT_SAME_STREAM {
		_, err = r.Read(u32[1:])
		if err != nil {
			return h, err
		}
		h.MessageLength = binary.BigEndian.Uint32(u32)

		_, err = r.Read(u8)
		if err != nil {
			return h, err
		}
		h.MessageTypeId = uint8(u8[0])
	}

	// If the header is full, we also need to extract
	// the message stream id.
	if h.Format <= HEADER_FORMAT_FULL {
		_, err = r.Read(u32)
		if err != nil {
			return h, err
		}
		h.MessageStreamId = binary.LittleEndian.Uint32(u32)
	}

	// If the header has an extended timestamp, we need to
	// extract that as well.
	if h.Timestamp == TIMESTAMP_EXTENDED {
		_, err = r.Read(u32)
		if err != nil {
			return h, err
		}

		h.ExtendedTimestamp = binary.BigEndian.Uint32(u32)
	}

	return h, nil
}

func (h *Header) Write(w io.Writer) (n int, err error) {
	m := 0
	u8 := make([]byte, 1)
	u32 := make([]byte, 4)

	switch {
	case h.ChunkStreamId <= 63:
		u8[0] = byte(h.Format<<6) | byte(h.ChunkStreamId)
		_, err = w.Write(u8)
		if err != nil {
			return
		}
		n += 1

	case h.ChunkStreamId <= 319:
		u8[0] = byte(h.Format << 6)
		_, err = w.Write(u8)
		if err != nil {
			return
		}
		n += 1

		u8[0] = byte(h.ChunkStreamId - 64)
		_, err = w.Write(u8)
		if err != nil {
			return
		}
		n += 1

	case h.ChunkStreamId <= 65599:
		u8[0] = byte(h.Format<<6) | byte(0x01)
		_, err = w.Write(u8)
		if err != nil {
			return
		}
		n += 1

		tmp := uint16(h.ChunkStreamId - 64)
		err = binary.Write(w, binary.BigEndian, &tmp)
		if err != nil {
			return
		}
		n += 2

	default:
		return n, errors.New("chunk stream too large")
	}

	if h.Format <= HEADER_FORMAT_SAME_LENGTH_AND_STREAM {
		binary.BigEndian.PutUint32(u32, h.Timestamp)
		m, err = w.Write(u32[1:])
		if err != nil {
			return
		}
		n += m
	}

	if h.Format <= HEADER_FORMAT_SAME_STREAM {
		binary.BigEndian.PutUint32(u32, h.MessageLength)
		m, err = w.Write(u32[1:])
		if err != nil {
			return
		}
		n += m

		u8[0] = byte(h.MessageTypeId)
		_, err = w.Write(u8)
		if err != nil {
			return
		}
		n += 1
	}

	if h.Format == HEADER_FORMAT_FULL {
		err = binary.Write(w, binary.LittleEndian, &h.MessageStreamId)
		if err != nil {
			return
		}
		n += 4
	}

	if h.Timestamp >= TIMESTAMP_EXTENDED {
		err = binary.Write(w, binary.BigEndian, &h.ExtendedTimestamp)
		if err != nil {
			return
		}
		n += 4
	}

	return
}

func (h *Header) CalculateTimestamp() uint32 {
	if h.Timestamp >= TIMESTAMP_MAX {
		return h.ExtendedTimestamp
	}

	return h.Timestamp
}
