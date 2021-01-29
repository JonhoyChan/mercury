package config

type Generator struct {
	ID IDGenerator `json:"id"`
}

type IDGenerator struct {
	WorkID int64  `json:"work_id"`
	Key    []byte `json:"key"`
}

func DefaultGenerator() *Generator {
	return &Generator{
		ID: IDGenerator{
			WorkID: 1,
			Key:    []byte("la6YsO+bNX/+XIkO"),
		},
	}
}
