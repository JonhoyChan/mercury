package persistence

type ClientTokenConfig struct {
	TokenSecret string
	TokenExpire int32
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
