package rtmp

import (
	"github.com/elobuff/goamf"
)

type Result struct {
	Name          string
	TransactionId float64
	Objects       []interface{}
}

type ResultError struct {
	Type         string
	ErrorCode    string
	Message      string
	Substitution string
}

func (r *Result) IsResult() bool {
	return r.Name == "_result"
}

func (r *Result) IsError() bool {
	return r.Name == "_error"
}

func (r *Result) DecodeError() (result ResultError, err error) {
	for _, obj := range r.Objects {
		if tobj, ok := obj.(amf.TypedObject); ok == true {
			if tobj.Type == "flex.messaging.messages.ErrorMessage" {
				if rc := tobj.Object["rootCause"]; rc != nil {
					if rootCause, ok := rc.(amf.TypedObject); ok == true {
						result = *new(ResultError)
						result.Type = rootCause.Type

						if tmp, ok := rootCause.Object["errorCode"].(string); ok {
							result.ErrorCode = tmp
						}

						if tmp, ok := rootCause.Object["message"].(string); ok {
							result.Message = tmp
						}

						if sa := rootCause.Object["substitutionArguments"]; sa != nil {
							if subArgs, ok := sa.(amf.Array); ok {
								len := len(subArgs)
								if tmp, ok := subArgs[len-1].(string); ok {
									result.Substitution = tmp
								}
							}
						}

						return
					}
				}
			}
		}
	}

	return result, Error("Could not decode object error")
}
