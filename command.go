package rtmp

type Command struct {
	Name          string
	TransactionId float64
	Version       uint8
	Objects       []interface{}
}
