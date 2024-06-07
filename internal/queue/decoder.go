package queue

import (
	"encoding/json"
	"fmt"
)

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
	default:
		var msg QueueMessage
		err := json.Unmarshal(data, &msg)
		if err != nil {
			return nil, fmt.Errorf("failed to decode JSON message: %w", err)
		}
		return msg, nil
	}
}

type Format string

const (
	FormatProtobuf Format = "protobuf"
	FormatJSON     Format = "json"
)

type QueueMessage struct {
	MessageID    string          `json:"message_id"`
	Provider     string          `json:"provider"`
	Sender       string          `json:"sender"`
	PhoneNumbers map[int64]int64 `json:"phone_numbers"`
	Text         string          `json:"text"`
}

type MessageData struct {
	Id             int      `json:"id"`
	Gateway        string   `json:"gateway"`
	Sender         string   `json:"sender"`
	Number         string   `json:"number"`
	Text           string   `json:"text"`
	GatewayHistory []string `json:"gateway_history"`
}

func (m *QueueMessage) Deflate() []MessageData {
	var data []MessageData
	for number, count := range m.PhoneNumbers {
		for i := int64(0); i < count; i++ {
			data = append(data, MessageData{
				Id:             0,
				Gateway:        m.Provider,
				Sender:         m.Sender,
				Number:         fmt.Sprintf("%d", number),
				Text:           m.Text,
				GatewayHistory: []string{},
			})
		}
	}
	return data
}
