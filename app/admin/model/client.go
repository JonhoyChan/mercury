package model

import (
	"mercury/app/logic/api"
)

type Client struct {
	ID          string `json:"id"`
	CreatedAt   int64  `json:"created_at,string"`
	UpdatedAt   int64  `json:"updated_at,string"`
	Name        string `json:"name"`
	TokenSecret string `json:"token_secret"`
	TokenExpire int64  `json:"token_expire,string"`
	UserCount   int64  `json:"user_count,string"`
	GroupCount  int64  `json:"group_count,string"`
}

func (c *Client) Fill(v *api.Client) {
	*c = Client{
		ID:          v.ID,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
		Name:        v.Name,
		TokenSecret: string(v.TokenSecret),
		TokenExpire: v.TokenExpire,
		UserCount:   v.UserCount,
		GroupCount:  v.GroupCount,
	}
}

type CreateClientReq struct {
	Name        string `json:"name"`
	TokenSecret string `json:"token_secret"`
	TokenExpire int64  `json:"token_expire"`
}

func (r *CreateClientReq) FillToProto() *api.CreateClientReq {
	return &api.CreateClientReq{
		Name:        r.Name,
		TokenSecret: r.TokenSecret,
		TokenExpire: r.TokenExpire,
	}
}

type NewClient struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

type UpdateClientReq struct {
	ID          string  `json:"id"`
	Name        *string `json:"name"`
	TokenSecret *string `json:"token_secret"`
	TokenExpire *int64  `json:"token_expire"`
}

func (r *UpdateClientReq) FillToProto() *api.UpdateClientReq {
	req := &api.UpdateClientReq{}
	if r.Name != nil {
		req.Name = &api.StringValue{
			Value: *r.Name,
		}
	}
	if r.TokenSecret != nil {
		req.TokenSecret = &api.StringValue{
			Value: *r.TokenSecret,
		}
	}
	if r.TokenExpire != nil {
		req.TokenExpire = &api.Int64Value{
			Value: *r.TokenExpire,
		}
	}
	return req
}
