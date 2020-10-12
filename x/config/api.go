package config

import (
	"context"
	"encoding/json"
	"mercury/app/infra/api"
	"mercury/version"
	"mercury/x/secretboxer"
)

func LoadConfig(name string) (map[string]interface{}, error) {
	resp, err := api.New(api.GetInfraURLFromEnv()).LoadConfig(context.Background(), name)
	if err != nil {
		return nil, err
	}

	boxer := secretboxer.NewPassphraseBoxer(version.String, secretboxer.EncodingTypeStd)
	data, err := boxer.Open(resp.Ciphertext)
	if err != nil {
		return nil, err
	}

	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	return v, nil
}
