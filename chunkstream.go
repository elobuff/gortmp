package rtmp

type OutboundChunkStream struct {
  Id                        uint32
  lastHeader                *Header
  lastOutAbsoluteTimestamp  uint32
  lastInAbsoluteTimestamp   uint32
  startAtTimestamp          uint32
}

type InboundChunkStream struct {
  Id                        uint32
  lastHeader                *Header
  lastOutAbsoluteTimestamp  uint32
  lastInAbsoluteTimestamp   uint32
  currentMessage            *Message
}

func NewOutboundChunkStream(id uint32) *OutboundChunkStream {
  return &OutboundChunkStream {
    Id: id,
  }
}

func NewInboundChunkStream(id uint32) *InboundChunkStream {
  return &InboundChunkStream {
    Id: id,
  }
}

func (cs *OutboundChunkStream) NewOutboundHeader(m *Message) *Header {
  h := &Header {
    ChunkStreamId: cs.Id,
    MessageLength: uint32(m.Buffer.Len()),
    MessageTypeId: m.Type,
    MessageStreamId: m.StreamId,
  }

  ts := m.Timestamp
  if ts == TIMESTAMP_AUTO {
    ts = cs.GetTimestamp()
    m.Timestamp = ts
    m.AbsoluteTimestamp = ts
  }

  deltaTimestamp := uint32(0)
  if cs.lastOutAbsoluteTimestamp < m.Timestamp {
    deltaTimestamp = m.Timestamp - cs.lastOutAbsoluteTimestamp
  }

  if cs.lastHeader == nil {
    h.Format = HEADER_FORMAT_FULL
    h.Timestamp = ts
  } else {
    if h.MessageStreamId == cs.lastHeader.MessageStreamId {
      if h.MessageTypeId == cs.lastHeader.MessageTypeId && h.MessageLength == cs.lastHeader.MessageLength {
        switch cs.lastHeader.Format {
        case HEADER_FORMAT_FULL:
          h.Format = HEADER_FORMAT_SAME_LENGTH_AND_STREAM
          h.Timestamp = deltaTimestamp
        case HEADER_FORMAT_SAME_STREAM:
          fallthrough
        case HEADER_FORMAT_SAME_LENGTH_AND_STREAM:
          fallthrough
        case HEADER_FORMAT_CONTINUATION:
          if cs.lastHeader.Timestamp == deltaTimestamp {
            h.Format = HEADER_FORMAT_CONTINUATION
          } else {
            h.Format = HEADER_FORMAT_SAME_LENGTH_AND_STREAM
            h.Timestamp = deltaTimestamp
          }
        }
      } else {
        h.Format = HEADER_FORMAT_SAME_STREAM
        h.Timestamp = ts
      }
    }
  }

  if h.Timestamp >= TIMESTAMP_EXTENDED {
    h.ExtendedTimestamp = m.Timestamp
    h.Timestamp = TIMESTAMP_EXTENDED
  } else {
    h.ExtendedTimestamp = 0
  }

  cs.lastHeader = h
  cs.lastOutAbsoluteTimestamp = ts

  return h
}

func (cs *OutboundChunkStream) GetTimestamp() uint32 {
  if cs.startAtTimestamp == uint32(0) {
    cs.startAtTimestamp = GetCurrentTimestamp()
    return uint32(0)
  }

  return GetCurrentTimestamp() - cs.startAtTimestamp
}
