package types

import "errors"

type HttpResponse struct {
	Status int
	Data   string
	Err    error
}

var StatusCodeError error = errors.New("Not 200.")
