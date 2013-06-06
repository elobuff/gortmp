package rtmp

import (
	"github.com/jcoene/gologger"
	"time"
)

var log logger.Logger = *logger.NewLogger(logger.LOG_LEVEL_DEBUG, "rtmp")

func GetCurrentTimestamp() uint32 {
	return uint32(time.Now().UnixNano()/int64(1000000)) % TIMESTAMP_MAX
}
