package rtmp

import (
  "time"
  "github.com/elobuff/gologger"
)

var log logger.Logger = *logger.NewLogger(logger.LOG_LEVEL_DEBUG, "rtmp")

func GetCurrentTimestamp() uint32 {
  return uint32(time.Now().UnixNano()/int64(1000000)) % TIMESTAMP_MAX
}
