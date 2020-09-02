package types

import (
	"outgoing/x/ecode"
	"strings"
)

/* ---------------------------------------- Message type ---------------------------------------- */
type MessageType uint8

const (
	MessageTypeSingle MessageType = iota
	MessageTypeGroup
)

// MarshalText converts MessageType to a slice of bytes wit
func (t MessageType) MarshalText() ([]byte, error) {
	switch t {
	case MessageTypeSingle:
		return []byte("user"), nil
	case MessageTypeGroup:
		return []byte("group"), nil
	default:
		return nil, ecode.NewError("invalid message type")
	}
}

// UnmarshalText parses MessageType from a string. the name of the MessageType.
func (t *MessageType) UnmarshalText(b []byte) error {
	switch strings.ToLower(string(b)) {
	case "user":
		*t = MessageTypeSingle
		return nil
	case "group":
		*t = MessageTypeGroup
		return nil
	default:
		return ecode.NewError("unrecognized")
	}
}

// MarshalJSON converts MessageType to a quoted string.
func (t MessageType) MarshalJSON() ([]byte, error) {
	res, err := t.MarshalText()
	if err != nil {
		return nil, err
	}

	return append(append([]byte{'"'}, res...), '"'), nil
}

// UnmarshalJSON reads MessageType from a quoted string.
func (t *MessageType) UnmarshalJSON(b []byte) error {
	if b[0] != '"' || b[len(b)-1] != '"' {
		return ecode.NewError("syntax error")
	}

	return t.UnmarshalText(b[1 : len(b)-1])
}

func (t MessageType) String() string {
	s, err := t.MarshalText()
	if err != nil {
		return "unknown"
	}
	return string(s)
}

/* ---------------------------------------- Message status ---------------------------------------- */
// 0: normal, 1:recalled, 2:deleted
type MessageStatus uint8

const (
	MessageStatusNormal MessageStatus = iota
	MessageStatusRecalled
	MessageStatusDeleted
)

// MarshalText converts MessageStatus to a slice of bytes wit
func (t MessageStatus) MarshalText() ([]byte, error) {
	switch t {
	case MessageStatusNormal:
		return []byte("normal"), nil
	case MessageStatusRecalled:
		return []byte("recalled"), nil
	case MessageStatusDeleted:
		return []byte("deleted"), nil
	default:
		return nil, ecode.NewError("invalid message status")
	}
}

// UnmarshalText parses MessageStatus from a string. the name of the MessageStatus.
func (t *MessageStatus) UnmarshalText(b []byte) error {
	switch strings.ToLower(string(b)) {
	case "normal":
		*t = MessageStatusNormal
		return nil
	case "recalled":
		*t = MessageStatusRecalled
		return nil
	case "deleted":
		*t = MessageStatusDeleted
		return nil
	default:
		return ecode.NewError("unrecognized")
	}
}

// MarshalJSON converts MessageStatus to a quoted string.
func (t MessageStatus) MarshalJSON() ([]byte, error) {
	res, err := t.MarshalText()
	if err != nil {
		return nil, err
	}

	return append(append([]byte{'"'}, res...), '"'), nil
}

// UnmarshalJSON reads MessageStatus from a quoted string.
func (t *MessageStatus) UnmarshalJSON(b []byte) error {
	if b[0] != '"' || b[len(b)-1] != '"' {
		return ecode.NewError("syntax error")
	}

	return t.UnmarshalText(b[1 : len(b)-1])
}

func (t MessageStatus) String() string {
	s, err := t.MarshalText()
	if err != nil {
		return "unknown"
	}
	return string(s)
}

/* ---------------------------------------- Content type ---------------------------------------- */
// 10: text, 20:image, 30:location, 40:audio, 50:video, 60:file
type ContentType uint8

const (
	ContentTypeText ContentType = (iota + 1) * 10
	ContentTypeImage
	ContentTypeLocation
	ContentTypeAudio
	ContentTypeVideo
	ContentTypeFile
)

// MarshalText converts ContentType to a slice of bytes wit
func (t ContentType) MarshalText() ([]byte, error) {
	switch t {
	case ContentTypeText:
		return []byte("text"), nil
	case ContentTypeImage:
		return []byte("image"), nil
	case ContentTypeAudio:
		return []byte("location"), nil
	case ContentTypeVideo:
		return []byte("audio"), nil
	case ContentTypeFile:
		return []byte("video"), nil
	case ContentTypeLocation:
		return []byte("file"), nil
	default:
		return nil, ecode.NewError("invalid content type")
	}
}

// UnmarshalText parses ContentType from a string. the name of the ContentType.
func (t *ContentType) UnmarshalText(b []byte) error {
	switch strings.ToLower(string(b)) {
	case "text":
		*t = ContentTypeText
		return nil
	case "image":
		*t = ContentTypeImage
		return nil
	case "location":
		*t = ContentTypeAudio
		return nil
	case "audio":
		*t = ContentTypeVideo
		return nil
	case "video":
		*t = ContentTypeFile
		return nil
	case "file":
		*t = ContentTypeLocation
		return nil
	default:
		return ecode.NewError("unrecognized")
	}
}

// MarshalJSON converts ContentType to a quoted string.
func (t ContentType) MarshalJSON() ([]byte, error) {
	res, err := t.MarshalText()
	if err != nil {
		return nil, err
	}

	return append(append([]byte{'"'}, res...), '"'), nil
}

// UnmarshalJSON reads ContentType from a quoted string.
func (t *ContentType) UnmarshalJSON(b []byte) error {
	if b[0] != '"' || b[len(b)-1] != '"' {
		return ecode.NewError("syntax error")
	}

	return t.UnmarshalText(b[1 : len(b)-1])
}

func (t ContentType) String() string {
	s, err := t.MarshalText()
	if err != nil {
		return "unknown"
	}
	return string(s)
}

/* ---------------------------------------- Content ---------------------------------------- */

type Message struct {
	Id          int64       `json:"id,string"`
	CreatedAt   int64       `json:"created_at,string"`
	MessageType MessageType `json:"message_type"`
	Sender      string      `json:"sender"`
	Receiver    string      `json:"receiver"`
	Topic       string      `json:"topic"`
	Sequence    int64       `json:"sequence,string"`
	ContentType ContentType `json:"content_type"`
	Body        []byte      `json:"body"`
	Mentions    []string    `json:"mentions,omitempty"`
}

/*
{
	"content": "Hello, World!"
}
*/
type TextMessage struct {
	Content string `json:"content"`
}

/*
{
	"file_stat": {
		"filename": "v2-f7ea6b00ebcfbd1b774434bf7e839ac6.jpg",
		"size": 115360,
		"width": 1253,
		"height": 1253
	},
	"hash": "bafykbzacedjkodrxars66qrsonrca7y6advofhrqfdtuxpkksvofu2l6slwjo"
}
*/
type ImageMessage struct {
	FileStat FileStat `json:"file_stat"`
	Hash     string   `json:"hash"`
}

/*
{
	"address": "西城区西便门桥",
	"longitude": 116.36302,
	"latitude": 39.9053
}
*/
type LocationMessage struct {
	Address   string  `json:"address"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type AudioMessage struct {
	FileStat FileStat `json:"file_stat"`
	// Voice length (unit: second)
	Length int32  `json:"length"`
	Hash   string `json:"hash"`
}

type VideoMessage struct {
	FileStat FileStat `json:"file_stat"`
	// Video length (unit: second)
	Length    int32  `json:"length"`
	Thumbnail string `json:"thumbnail"`
	Hash      string `json:"hash"`
}

type FileMessage struct {
	FileStat FileStat `json:"file_stat"`
	Hash     string   `json:"hash"`
}

type QuoteMessage struct {
	QuotedMessageID int64  `json:"quoted_message_id"`
	Content         string `json:"content"`
}

type FileStat struct {
	Filename string `gorm:"type:VARCHAR;column:filename"`
	Size     int64  `gorm:"column:size"`
	Width    int32  `gorm:"column:width,omitempty"`
	Height   int32  `gorm:"column:height,omitempty"`
}
