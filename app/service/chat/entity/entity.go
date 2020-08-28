package entity

import (
	"outgoing/x/ecode"
	"strings"
)

type Client struct {
	ID          string `gorm:"primary_key;type:UUID;column:id"`
	CreatedAt   int64  `gorm:"column:created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at"`
	Name        string `gorm:"not null;type:VARCHAR;column:name"`
	TokenSecret string `gorm:"not null;type:VARCHAR;column:token_secret"`
	TokenExpire int64  `gorm:"type:VARCHAR;column:token_expire"`
	Credential  string `gorm:"type:VARCHAR;column:credential"`
	UserCount   int32  `gorm:"column:user_count"`
}

type User struct {
	ID        int64  `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	ClientID  string `gorm:"type:VARCHAR;column:client_id"`
	Name      string `gorm:"type:VARCHAR;column:name"`
	// Unique ID of the user.
	UID       string `gorm:"not null;type:VARCHAR;column:uid"`
	Activated bool   `gorm:"default:true;column:activated"`
}

type Friend struct {
	ID           uint64 `gorm:"primary_key;column:id"`
	CreatedAt    int64  `gorm:"column:created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at"`
	UserID       int64  `gorm:"column:user_id"`
	FriendUserID int64  `gorm:"column:friend_user_id"`
}

type Group struct {
	ID        int64  `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	ClientID  string `gorm:"type:VARCHAR;column:client_id"`
	Name      string `gorm:"type:VARCHAR;column:name"`
	// Unique ID of the group.
	GID          string `gorm:"not null;type:VARCHAR;column:gid"`
	Introduction string `gorm:"type:VARCHAR;column:introduction"`
	Owner        int64  `gorm:"column:owner"`
	// Group type e.g. (0: public group)
	Type        uint8 `gorm:"not null;type:SMALLINT;column:type"`
	Activated   bool  `gorm:"default:true;column:activated"`
	MemberCount int32 `gorm:"column:member_count"`
}

type GroupMember struct {
	ID        int64 `gorm:"primary_key;column:id"`
	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
	GroupID   int64 `gorm:"column:group_id"`
	UserID    int64 `gorm:"column:user_id"`
}

type Message struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	ClientID  string `gorm:"type:VARCHAR;column:client_id"`
	// Message type. e.g. (1: single chat message, 2:group chat message)
	MessageType uint8  `gorm:"type:SMALLINT;column:message_type"`
	Sender      uint64 `gorm:"column:sender"`
	// If a single chat, the value is user ID.
	// If a group chat, the value is group ID.
	Receiver uint64 `gorm:"column:receiver"`
	// Message body. e.g. {"content": "Hello, World!". "content_type": 10}
	Body     string `gorm:"type:JSON;column:body"`
	Status   uint8  `gorm:"not null;type:SMALLINT;column:status"`
	Mentions string `gorm:"column:mentions"`
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

/*
{
	"content": "Hello, World!",
	"content_type": 10
}
*/
type TextMessage struct {
	Content     string      `json:"content"`
	ContentType ContentType `json:"content_type"`
}

/*
{
	"file_stat": {
		"filename": "v2-f7ea6b00ebcfbd1b774434bf7e839ac6.jpg",
		"size": 115360,
		"width": 1253,
		"height": 1253
	},
	"hash": "bafykbzacedjkodrxars66qrsonrca7y6advofhrqfdtuxpkksvofu2l6slwjo",
	"content_type": 20
}
*/
type ImageMessage struct {
	FileStat    FileStat    `json:"file_stat"`
	Hash        string      `json:"hash"`
	ContentType ContentType `json:"content_type"`
}

/*
{
	"address": "西城区西便门桥",
	"longitude": 116.36302,
	"latitude": 39.9053,
	"content_type": 30
}
*/
type LocationMessage struct {
	Address     string      `gorm:"type:VARCHAR;column:address"`
	Longitude   float64     `gorm:"type:DECIMAL(10,6);column:longitude"`
	Latitude    float64     `gorm:"type:DECIMAL(10,6);column:latitude"`
	ContentType ContentType `json:"content_type"`
}

type AudioMessage struct {
	FileStat FileStat `json:"file_stat"`
	// Voice length (unit: second)
	Length      int32       `json:"length"`
	Hash        string      `json:"hash"`
	ContentType ContentType `json:"content_type"`
}

type VideoMessage struct {
	FileStat FileStat `json:"file_stat"`
	// Video length (unit: second)
	Length      int32       `json:"length"`
	Thumbnail   string      `json:"thumbnail"`
	Hash        string      `json:"hash"`
	ContentType ContentType `json:"content_type"`
}

type FileMessage struct {
	FileStat    FileStat    `json:"file_stat"`
	Hash        string      `json:"hash"`
	ContentType ContentType `json:"content_type"`
}

type FileStat struct {
	Filename string `gorm:"type:VARCHAR;column:filename"`
	Size     int64  `gorm:"column:size"`
	Width    int32  `gorm:"column:width,omitempty"`
	Height   int32  `gorm:"column:height,omitempty"`
}
