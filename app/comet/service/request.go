package service

import (
	jsoniter "github.com/json-iterator/go"
	"outgoing/x/types"
)

type requester interface {
	Validate() bool
	Unmarshal(data []byte) error
}

type HandshakeRequest struct {
	// Client-provided message id
	MID string `json:"mid,omitempty"`
	// Version of the wire protocol supported by the client
	Version string `json:"version" validate:"required"`
	// User agent identifying client software
	UserAgent string `json:"user_agent" validate:"required"`
	// Connected device for the purpose of push notifications
	Platform string `json:"platform,omitempty"`
	// Human language of the client device
	Language string `json:"language,omitempty"`
	// Client's unique device ID
	DeviceID string `json:"device_id"`
	// Authentication token
	Token string `json:"token" validate:"required"`
}

func (r *HandshakeRequest) Validate() bool {
	if err := validate.Struct(r); err != nil {
		return false
	}
	return true
}

func (r *HandshakeRequest) Unmarshal(data []byte) error {
	return jsoniter.Unmarshal(data, r)
}

type HeartbeatRequest struct {
	// Client-provided message id
	MID string `json:"mid,omitempty"`
}

func (r *HeartbeatRequest) Validate() bool {
	return true
}

func (r *HeartbeatRequest) Unmarshal(data []byte) error {
	return jsoniter.Unmarshal(data, r)
}

type ConnectRequest struct {
	// Client-provided message id
	MID string `json:"mid,omitempty"`
	// Authentication token
	Token string `json:"token" validate:"required"`
}

func (r *ConnectRequest) Validate() bool {
	if err := validate.Struct(r); err != nil {
		return false
	}
	return true
}

func (r *ConnectRequest) Unmarshal(data []byte) error {
	return jsoniter.Unmarshal(data, r)
}

type PullMessageRequest struct {
	// Client-provided message id
	MID string `json:"mid,omitempty"`
}

func (r *PullMessageRequest) Validate() bool {
	return true
}

func (r *PullMessageRequest) Unmarshal(data []byte) error {
	return jsoniter.Unmarshal(data, r)
}

type PushMessageRequest struct {
	// Client-provided message id
	MID string `json:"mid,omitempty"`
	// The type of the message. e.g. (0: single chat message, 1:group chat message)
	MessageType types.MessageType `json:"message_type"`
	// Message receiver
	// If a single chat, the value is user ID.
	// If a group chat, the value is group ID.
	Receiver string `json:"receiver" validate:"required,is-id"`
	// The type of the message content. e.g. (10: text, 20:image, 30:location, 40:audio, 50:video, 60:file)
	ContentType types.ContentType `json:"content_type"`
	// The body of the message, will change according to the content type
	Body types.Content `json:"body" validate:"required"`
	// List of user IDs mentioned in the message
	Mentions []string `json:"mentions,omitempty"`
}

func (r *PushMessageRequest) ValidateBody() bool {
	var v interface{}
	switch r.ContentType {
	case types.ContentTypeText:
		var m types.TextMessage
		v = &m
	case types.ContentTypeImage:
		var m types.ImageMessage
		v = &m
	case types.ContentTypeLocation:
		var m types.LocationMessage
		v = &m
	case types.ContentTypeAudio:
		var m types.AudioMessage
		v = &m
	case types.ContentTypeVideo:
		var m types.VideoMessage
		v = &m
	case types.ContentTypeFile:
		var m types.FileMessage
		v = &m
	}
	if err := jsoniter.Unmarshal(r.Body, v); err != nil {
		return false
	}
	if err := validate.Struct(v); err != nil {
		return false
	}
	return true
}

func (r *PushMessageRequest) Validate() bool {
	if err := validate.Struct(r); err != nil {
		return false
	}
	return r.ValidateBody()
}

func (r *PushMessageRequest) Unmarshal(data []byte) error {
	return jsoniter.Unmarshal(data, r)
}

type NotificationRequest struct {
	// Client-provided message id
	MID      string         `json:"mid,omitempty"`
	What     types.WhatType `json:"what"`
	Topic    string         `json:"topic" validate:"required"`
	Sequence int64          `json:"sequence,int64,omitempty"`
}

func (r *NotificationRequest) Validate() bool {
	if err := validate.Struct(r); err != nil {
		return false
	}
	return true
}

func (r *NotificationRequest) Unmarshal(data []byte) error {
	return jsoniter.Unmarshal(data, r)
}
