package message

import (
	"encoding/json"
	b "wizeBlockchain/code_network/block"
	"errors"
)

type MessageType int

const (
	MessageTypeQueryLatest MessageType = iota
	MessageTypeQueryAll    MessageType = iota
	MessageTypeBlocks      MessageType = iota
)

var (

	ErrInvalidChain       = errors.New("invalid chain")
	ErrInvalidBlock       = errors.New("invalid block")
	ErrUnknownMessageType = errors.New("unknown message type")
)

func (ms MessageType) Name() string {
	switch ms {
	case MessageTypeQueryLatest:
		return "QUERY_LATEST"
	case MessageTypeQueryAll:
		return "QUERY_ALL"
	case MessageTypeBlocks:
		return "BLOCKS"
	default:
		return "UNKNOWN"
	}
}

type Message struct {
	Type MessageType `json:"type"`
	Data string      `json:"data"`
}

func NewQueryLatestMessage() *Message {
	return &Message{
		Type: MessageTypeQueryLatest,
	}
}

func NewQueryAllMessage() *Message {
	return &Message{
		Type: MessageTypeQueryAll,
	}
}

func NewBlocksMessage(blocks b.Blocks) (*Message, error) {
	b, err := json.Marshal(blocks)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type: MessageTypeBlocks,
		Data: string(b),
	}, nil
}
