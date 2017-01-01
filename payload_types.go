package tcf

import (
	"errors"
	"time"
)

type PayloadType string

const (
	PayloadDefaultPayload PayloadType = "PayloadDefaultPayload"
)

func NewPayload(msg interface{}) (Payload, error) {
	switch TCFConfig.PayloadType {
	case PayloadDefaultPayload:
		return &DefaultPayload{Message: msg, Encoding: JSON, Time: time.Now().Unix()}, nil
	default:
		return nil, errors.New("Payload type not supported")
	}
}