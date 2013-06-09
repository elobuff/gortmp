package rtmp

type Result struct {
	Name          string
	TransactionId float64
	Objects       []interface{}
}

func (r *Result) IsResult() bool {
	return r.Name == "_result"
}

func (r *Result) IsError() bool {
	return r.Name == "_error"
}
