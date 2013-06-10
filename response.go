package rtmp

import (
	"fmt"
	"github.com/elobuff/goamf"
)

type Response struct {
	Name          string
	TransactionId float64
	Objects       []interface{}
}

type ResponseError struct {
	Type         string
	ErrorCode    string
	Message      string
	Substitution string
}

func (r *Response) IsResult() bool {
	return r.Name == "_result"
}

func (r *Response) IsError() bool {
	return r.Name == "_error"
}

func (r *Response) DecodeBody() (result interface{}, err error) {
	for _, obj := range r.Objects {
		if tobj, ok := obj.(amf.TypedObject); ok == true {
			if body := tobj.Object["body"]; body != nil {
				return body, nil
			}
		}
	}

	return result, Error("Could not extract body")
}

func (r *Response) DecodeError() (result ResponseError, err error) {
	for _, obj := range r.Objects {
		if tobj, ok := obj.(amf.TypedObject); ok == true {
			if tobj.Type == "flex.messaging.messages.ErrorMessage" {
				if rc := tobj.Object["rootCause"]; rc != nil {
					if rootCause, ok := rc.(amf.TypedObject); ok == true {
						result = *new(ResponseError)
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

	result.Message = fmt.Sprintf("%+v", r)

	return result, Error("Could not decode object error")
}
