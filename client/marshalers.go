package client

import (
	"encoding/json"
	"errors"
)

func Marshal(from Payload, enc Encoding) (interface{}, error) {
	switch enc {
	case JSON:
		return marshalJSON(from)
	default:
		return nil, errors.New("Encoding is not supported!")
	}
}

func marshalJSON(from Payload) (interface{}, error) {
	// Copy the object, we don;t want to operate on the same payload
	newPayload := from
	// First encode the inner data payload
	newPayload.Encode()
	return json.Marshal(newPayload)
}
