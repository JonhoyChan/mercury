package config

type Hasher struct {
	Argon2 HasherArgon2 `json:"argon2"`
	BCrypt HasherBCrypt `json:"bcrypt"`
}

type HasherArgon2 struct {
	Memory      int `json:"memory"`
	Iterations  int `json:"iterations"`
	Parallelism int `json:"parallelism"`
	SaltLength  int `json:"salt_length"`
	KeyLength   int `json:"key_length"`
}

type HasherBCrypt struct {
	Cost int `json:"cost"`
}

func DefaultHasher() *Hasher {
	return &Hasher{
		Argon2: HasherArgon2{
			Memory:      8192,
			Iterations:  2,
			Parallelism: 4,
			SaltLength:  16,
			KeyLength:   16,
		},
		BCrypt: HasherBCrypt{
			Cost: 10,
		},
	}
}
