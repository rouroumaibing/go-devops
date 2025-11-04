package versions

import (
	"strings"
)

const (
	KindStatus   = "Status"
	APIVersionV1 = "v1"
	KindAccount  = "Account"
)

type TypeMeta struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
}

type Response struct {
	TypeMeta     `json:",inline"`
	Status       string `json:"status"`
	Code         int32  `json:"code"`
	Reason       string `json:"reason"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Message      string `json:"message"`
}

func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (r Response) WithMessage(message string) Response {
	r.Message = FirstUpper(message)
	return r
}

func (r Response) Error() string {
	if r.Message != "" {
		return r.Message
	}
	return r.ErrorMessage
}
