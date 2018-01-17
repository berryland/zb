package zb

import "fmt"

type ApiError struct {
	Code    uint16
	Message string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("fail to invoke api (%d, %v)", e.Code, e.Message)
}
