package entity

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
	ClientID  int64  `gorm:"column:client_id"`
	Name      string `gorm:"type:VARCHAR;column:name"`
	// Unique ID of the user.
	UID       string `gorm:"not null;type:UUID;default:gen_random_uuid();column:uid"`
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
	ClientID  int64  `gorm:"column:client_id"`
	Name      string `gorm:"type:VARCHAR;column:name"`
	// Unique ID of the group.
	GID          string `gorm:"not null;type:UUID;default:gen_random_uuid();column:gid"`
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
	ClientID  int64  `gorm:"column:client_id"`
	// Unique ID of the message.
	MID string `gorm:"not null;type:UUID;default:gen_random_uuid();column:mid"`
	// Message type. e.g. (1: single chat message, 2:group chat message)
	MessageType uint8  `gorm:"type:SMALLINT;column:message_type"`
	Sender      uint64 `gorm:"column:sender"`
	// If it is a single chat, the value is user ID.
	// If it is a group chat, the value is group ID.
	Receiver uint64 `gorm:"column:receiver"`
	// Message body. e.g. {"content": "Hello, World!". "content_type": 10}
	Body     string `gorm:"type:TEXT;column:body"`
	Status   uint8  `gorm:"not null;type:SMALLINT;column:status"`
	Mentions string `gorm:"column:mentions"`
}

/*
{
	"content": "Hello, World!",
	"content_type": 10
}
*/
type TextMessage struct {
	Content string `json:"content"`
	// Content type. e.g. (10: text, 20:image, 30:location, 40:audio, 50:video, 60:file)
	ContentType uint8 `json:"content_type"`
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
	FileStat FileStat `json:"file_stat"`
	Hash     string   `json:"hash"`
	// Content type. e.g. (10: text, 20:image, 30:location, 40:audio, 50:video, 60:file)
	ContentType uint8 `json:"content_type"`
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
	Address   string  `gorm:"type:VARCHAR;column:address"`
	Longitude float64 `gorm:"type:DECIMAL(10,6);column:longitude"`
	Latitude  float64 `gorm:"type:DECIMAL(10,6);column:latitude"`
	// Content type. e.g. (10: text, 20:image, 30:location, 40:audio, 50:video, 60:file)
	ContentType uint8 `json:"content_type"`
}

type AudioMessage struct {
	FileStat FileStat `json:"file_stat"`
	// Voice length (unit: second)
	Length int32  `json:"length"`
	Hash   string `json:"hash"`
	// Content type. e.g. (10: text, 20:image, 30:location, 40:audio, 50:video, 60:file)
	ContentType uint8 `json:"content_type"`
}

type VideoMessage struct {
	FileStat FileStat `json:"file_stat"`
	// Video length (unit: second)
	Length    int32  `json:"length"`
	Thumbnail string `json:"thumbnail"`
	Hash      string `json:"hash"`
	// Content type. e.g. (10: text, 20:image, 30:location, 40:audio, 50:video, 60:file)
	ContentType uint8 `json:"content_type"`
}

type FileMessage struct {
	FileStat FileStat `json:"file_stat"`
	Hash     string   `json:"hash"`
	// Content type. e.g. (10: text, 20:image, 30:location, 40:audio, 50:video, 60:file)
	ContentType uint8 `json:"content_type"`
}

type FileStat struct {
	Filename string `gorm:"type:VARCHAR;column:filename"`
	Size     int64  `gorm:"column:size"`
	Width    int32  `gorm:"column:width,omitempty"`
	Height   int32  `gorm:"column:height,omitempty"`
}
