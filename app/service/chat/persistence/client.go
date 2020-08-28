package persistence

type Client struct {
	ID          string
	Name        string
	TokenSecret string
	TokenExpire int64
}

type ClientCreate struct {
	ID          string
	Name        string
	TokenSecret string
	Credential  string
	TokenExpire int64
}

type ClientUpdate struct {
	ID          string
	Name        *string
	TokenSecret *string
	TokenExpire *int64
}
