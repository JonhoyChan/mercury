package service

import (
	"encoding/json"
	"outgoing/x/types"
)

type HandshakeRequest struct {
	// client-provided message id
	MID string `json:"mid,omitempty"`
	// version of the wire protocol supported by the client
	Version string `json:"version" validate:"required"`
	// user agent identifying client software
	UserAgent string `json:"user_agent" validate:"required"`
	// connected device for the purpose of push notifications
	Platform string `json:"platform,omitempty"`
	// human language of the client device
	Language string `json:"language,omitempty"`
	// client's unique device ID
	DeviceID string `json:"device_id"`
	// authentication token
	Token string `json:"token" validate:"required"`
}

func (r *HandshakeRequest) Marshal() (data []byte, err error) {
	return json.Marshal(r)
}

func (r *HandshakeRequest) Unmarshal(data []byte) error {
	return json.Unmarshal(data, r)
}

type HeartbeatRequest struct {
	// client-provided message id
	MID string `json:"mid,omitempty"`
}

func (r *HeartbeatRequest) Marshal() (data []byte, err error) {
	return json.Marshal(r)
}

func (r *HeartbeatRequest) Unmarshal(data []byte) error {
	return json.Unmarshal(data, r)
}

type ConnectRequest struct {
	// client-provided message id
	MID string `json:"mid,omitempty"`
	// authentication token
	Token string `json:"token" validate:"required"`
}

func (r *ConnectRequest) Marshal() (data []byte, err error) {
	return json.Marshal(r)
}

func (r *ConnectRequest) Unmarshal(data []byte) error {
	return json.Unmarshal(data, r)
}

type PushMessageRequest struct {
	// client-provided message id
	MID string `json:"mid,omitempty"`
	// the type of the message. e.g. (1: single chat message, 2:group chat message)
	MessageType types.MessageType `json:"message_type" validate:"required"`
	// message receiver
	// if a single chat, the value is user ID.
	// if a group chat, the value is group ID.
	Receiver string `json:"receiver" validate:"required"`
	// the type of the message content. e.g. (0: normal, 1:recall, 2:quote, 40:audio, 50:video, 60:file)
	ContentType types.ContentType `json:"content_type" validate:"required"`
	// the body of the message, will change according to the content type
	Body json.RawMessage `json:"body" validate:"required"`
	// list of user IDs mentioned in the message
	Mentions []string `json:"mentions,omitempty"`
}

func (r *PushMessageRequest) GetBody() ([]byte, error) {
	switch r.ContentType {
	case types.ContentTypeText:
		var v types.TextMessage
		if err := json.Unmarshal(r.Body, &v); err != nil {
			return nil, err
		}
	case types.ContentTypeImage:
		var v types.ImageMessage
		if err := json.Unmarshal(r.Body, &v); err != nil {
			return nil, err
		}
	case types.ContentTypeLocation:
		var v types.LocationMessage
		if err := json.Unmarshal(r.Body, &v); err != nil {
			return nil, err
		}
	case types.ContentTypeAudio:
		var v types.AudioMessage
		if err := json.Unmarshal(r.Body, &v); err != nil {
			return nil, err
		}
	case types.ContentTypeVideo:
		var v types.VideoMessage
		if err := json.Unmarshal(r.Body, &v); err != nil {
			return nil, err
		}
	case types.ContentTypeFile:
		var v types.FileMessage
		if err := json.Unmarshal(r.Body, &v); err != nil {
			return nil, err
		}
	}
	return r.Body, nil
}

func (r *PushMessageRequest) Marshal() (data []byte, err error) {
	return json.Marshal(r)
}

func (r *PushMessageRequest) Unmarshal(data []byte) error {
	return json.Unmarshal(data, r)
}
