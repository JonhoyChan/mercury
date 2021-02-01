package service

import (
	"context"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/registry"
	"mercury/app/admin/model"
	chatApi "mercury/app/logic/api"
	"mercury/x/ecode"
	"mercury/x/log"
	"strconv"
	"strings"
	"time"
)

type Servicer interface {
	GetClient(ctx context.Context, clientID string) (*model.Client, error)
	CreateClient(ctx context.Context, req *chatApi.CreateClientReq) (*model.NewClient, error)
	UpdateClient(ctx context.Context, clientID string, req *chatApi.UpdateClientReq) error
	DeleteClient(ctx context.Context, clientID string) error
}

type Service struct {
	log     log.Logger
	service chatApi.ChatAdminService
}

func NewService(log log.Logger) (*Service, error) {
	opts := []client.Option{
		client.Retries(2),
		client.Retry(ecode.RetryOnMicroError),
		client.WrapCall(ecode.MicroCallFunc),
	}

	c := grpc.NewClient(opts...)

	return &Service{
		log:     log,
		service: chatApi.NewChatAdminService("mercury.logic", c),
	}, nil
}

func beforeCall(clientID string) client.CallWrapper {
	return func(fn client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
			if !strings.HasPrefix(req.Endpoint(), "ChatAdmin.") {
				return ecode.ErrBadRequest
			}

			timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
			m := metadata.Metadata{
				"Timestamp": timestamp,
				"Issuer":    "Mercury",
				"Id":        clientID,
			}
			ctx = metadata.NewContext(ctx, m)
			return fn(ctx, node, req, rsp, opts)
		}
	}
}

func withCallWrapper(clientID string) client.CallOption {
	return client.WithCallWrapper(beforeCall(clientID))
}

func (s *Service) GetClient(ctx context.Context, clientID string) (*model.Client, error) {
	resp, err := s.service.GetClient(ctx, &chatApi.GetClientReq{}, withCallWrapper(clientID))
	if err != nil {
		return nil, err
	}

	var data model.Client
	data.Fill(resp.Client)
	return &data, nil
}

func (s *Service) CreateClient(ctx context.Context, req *chatApi.CreateClientReq) (*model.NewClient, error) {
	resp, err := s.service.CreateClient(ctx, req, withCallWrapper(""))
	if err != nil {
		return nil, err
	}

	return &model.NewClient{
		ID:     resp.ClientID,
		Secret: resp.ClientSecret,
	}, nil
}

func (s *Service) UpdateClient(ctx context.Context, clientID string, req *chatApi.UpdateClientReq) error {
	_, err := s.service.UpdateClient(ctx, req, withCallWrapper(clientID))
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteClient(ctx context.Context, clientID string) error {
	_, err := s.service.DeleteClient(ctx, &chatApi.DeleteClientReq{}, withCallWrapper(clientID))
	if err != nil {
		return err
	}

	return nil
}
