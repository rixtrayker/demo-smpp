package queue

import (
	"encoding/json"
	"fmt"
	// "google.golang.org/protobuf/proto"
)

// type Message interface {
//     proto.Message
// }

type Decoder struct {
}

func NewDecoder() *Decoder {
    return &Decoder{}
}

func (d *Decoder) DecodeJSON(data []byte) (QueueMessage, error) {
    result, err := d.Decode(data, FormatJSON)
    if err != nil {
        return QueueMessage{}, err
    }
    return result.(QueueMessage), nil

}


func (d *Decoder) Decode(data []byte, format Format) (interface{}, error) {
    switch format {
    // case FormatProtobuf:
    //     var msg proto.Message
    //     err := proto.Unmarshal(data, msg)
    //     if err != nil {
    //         return nil, fmt.Errorf("failed to decode protobuf message: %w", err)
    //     }
    //     return msg, nil
    default:
        var msg QueueMessage
        err := json.Unmarshal(data, &msg)
        if err != nil {
            return nil, fmt.Errorf("failed to decode JSON message: %w", err)
        }
        return msg, nil
    }
}

// func (d *Decoder) DecodeJSON(data []byte) (QueueMessage, error) {
//     var msg QueueMessage
//     err := json.Unmarshal(data, &msg)
//     if err != nil {
//         return QueueMessage{}, fmt.Errorf("failed to decode JSON message: %w", err)
//     }
//     return msg, nil
// }


type Format string

const (
    FormatProtobuf Format = "protobuf"
    FormatJSON     Format = "json"
)

type QueueMessage struct {
    Provider    string `json:"provider"`
    Sender      string `json:"sender"`
    PhoneNumbers map[int64]int64 `json:"phone_numbers"`
    Text        string `json:"text"`
}