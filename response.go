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
	Code         string
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
		if obj, ok := obj.(amf.Object); ok == true {
			if body := obj["body"]; body != nil {
				return body, nil
			}
		}
	}

	return nil, nil
}

func (r *Response) DecodeError() (result ResponseError, err error) {
	for _, obj := range r.Objects {
		if obj, ok := obj.(amf.Object); ok == true {
			if rc := obj["rootCause"]; rc != nil {
				if rcobj, ok := rc.(amf.Object); ok == true {
					result = *new(ResponseError)

					if tmp, ok := rcobj["errorCode"].(string); ok {
						result.Code = tmp
					}

					if tmp, ok := rcobj["message"].(string); ok {
						result.Message = tmp
					}

					if sa := rcobj["substitutionArguments"]; sa != nil {
						if subArgs, ok := sa.(amf.Array); ok {
							length := len(subArgs)
							if length > 0 {
								if tmp, ok := subArgs[length-1].(string); ok {
									result.Substitution = tmp
								}
							}
						}
					}

					return
				}
			}
		}
	}

	result.Message = fmt.Sprintf("%+v", r)

	return result, Error("Could not decode object error")
}
