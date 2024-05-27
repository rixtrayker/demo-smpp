package worker

import (
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
)

type Message interface {
    proto.Message
}

type Decoder struct {
}

func NewDecoder() *Decoder {
    return &Decoder{}
}

func (d *Decoder) DecodeJSON(data []byte) (interface{}, error) {
    return d.Decode(data, FormatJSON)
}


func (d *Decoder) Decode(data []byte, format Format) (interface{}, error) {
    switch format {
    case FormatProtobuf:
        var msg proto.Message
        err := proto.Unmarshal(data, msg)
        if err != nil {
            return nil, fmt.Errorf("failed to decode protobuf message: %w", err)
        }
        return msg, nil
    case FormatJSON:
        var msg JSONMessage
        err := json.Unmarshal(data, &msg)
        if err != nil {
            return nil, fmt.Errorf("failed to decode JSON message: %w", err)
        }
        return msg, nil
    default:
        return nil, errors.New("unsupported format")
    }
}

// Format defines the expected message format (Protobuf or JSON)
type Format string

const (
    FormatProtobuf Format = "protobuf"
    FormatJSON     Format = "json"
)

// JSONMessage represents a message in JSON format
type JSONMessage struct {
    Provider    string `json:"provider"`
    PhoneNumber string `json:"phone_number"`
    Text        string `json:"text"`
}