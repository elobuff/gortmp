package rtmp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jcoene/gologger"
	"time"
)

var log logger.Logger = *logger.NewLogger(logger.LOG_LEVEL_WARN, "rtmp")

func DumpBytes(label string, buf []byte, size int) {
	fmt.Printf("Dumping %s (%d bytes):\n", label, size)
	for i := 0; i < size; i++ {
		fmt.Printf("0x%02x ", buf[i])
	}
	fmt.Printf("\n")
}

func Dump(label string, val interface{}) error {
	json, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return Error("Error dumping %s: %s", label, err)
	}

	fmt.Printf("Dumping %s:\n%s\n", label, json)
	return nil
}

func Error(f string, v ...interface{}) error {
	return errors.New(fmt.Sprintf(f, v...))
}

func GetCurrentTimestamp() uint32 {
	return uint32(time.Now().UnixNano()/int64(1000000)) % TIMESTAMP_MAX
}
