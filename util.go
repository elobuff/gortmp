package rtmp

import (
	"errors"
	"fmt"
	"github.com/jcoene/gologger"
	"time"
)

var log logger.Logger = *logger.NewLogger(logger.LOG_LEVEL_DEBUG, "rtmp")

func Error(f string, v ...interface{}) error {
	return errors.New(fmt.Sprintf(f, v...))
}

func GetCurrentTimestamp() uint32 {
	return uint32(time.Now().UnixNano()/int64(1000000)) % TIMESTAMP_MAX
}
