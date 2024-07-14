package queue

import (
	"encoding/json"
	"fmt"

	"github.com/phuslu/log"
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
	PhoneNumbers []int64 	     `json:"phone_numbers"`
	Text         string          `json:"text"`
	GatewayHistory []string 	 `json:"gateway_history"`
}

type MessageData struct {
	Id             int64    `json:"id"`
	Gateway        string   `json:"gateway"`
	Sender         string   `json:"sender"`
	Number         string   `json:"number"`
	Text           string   `json:"text"`
	GatewayHistory []string `json:"gateway_history"`
}

func (m *QueueMessage) MarshalObject(e *log.Entry) {
	e.Str("message_id", m.MessageID).Str("provider", m.Provider).Str("sender", m.Sender).Ints64("phone_numbers", m.PhoneNumbers).Str("text", m.Text).Strs("gateway_history", m.GatewayHistory)
}

func (m *MessageData) MarshalObject(e *log.Entry) {
	e.Int64("id", m.Id).Str("gateway", m.Gateway).Str("sender", m.Sender).Str("number", m.Number).Str("text", m.Text).Strs("gateway_history", m.GatewayHistory)
}

func (m *QueueMessage) Deflate() []MessageData {
	var data []MessageData
	for _, number := range m.PhoneNumbers {
		data = append(data, MessageData{
			Id:             0,
			Gateway:        m.Provider,
			Sender:         m.Sender,
			Number:         fmt.Sprintf("%d", number),
			Text:           m.Text,
			GatewayHistory: m.GatewayHistory,
		})
	}
	return data
}
