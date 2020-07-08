package service

import (
	"context"
	"outgoing/app/gateway/account/model"
	uApi "outgoing/app/service/main/account/api"
	aApi "outgoing/app/service/main/auth/api"
	"outgoing/x/ecode"
	"outgoing/x/log"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
)

type Service struct {
	log            log.Logger
	accountService uApi.AccountService
	authService    aApi.AuthService
}

func NewService(log log.Logger) *Service {
	opts := []client.Option{
		// 重试存在计数BUG，因为从0开始计算，所以实际的重试次数为 i + 1
		client.Retries(2),
		client.Retry(ecode.RetryOnMicroError),
		client.WrapCall(ecode.MicroCallFunc),
	}

	c := grpc.NewClient(opts...)

	return &Service{
		log:            log,
		accountService: uApi.NewAccountService("account.srv", c),
		authService:    aApi.NewAuthService("auth.srv", c),
	}
}

func (s *Service) generateToken(ctx context.Context, uid string) (resp *aApi.GenerateTokenResp, err error) {
	record := &aApi.Record{
		Uid:   uid,
		Level: aApi.AuthLevel_AuthLevelAuth,
		State: aApi.UserState_UserStateNormal,
	}
	resp, err = s.authService.GenerateToken(ctx, &aApi.GenerateTokenReq{
		HandlerType: aApi.HandlerType_HandlerTypeJWT,
		Record:      record,
	})

	return
}

func (s *Service) Register(ctx context.Context, mobile, ip string) (*model.RegisterResp, error) {
	rResp, err := s.accountService.Register(ctx, &uApi.RegisterReq{Mobile: mobile, Ip: ip})
	if err != nil {
		s.log.Error("User-Service-Register", "error", err)
		return nil, err
	}

	gResp, err := s.generateToken(ctx, rResp.UID)
	if err != nil {
		s.log.Error("User-Service-Register", "error", err)
		return nil, err
	}

	return &model.RegisterResp{
		VID:   rResp.VID,
		Token: gResp.Token,
	}, nil
}

func (s *Service) Login(ctx context.Context, input, captcha, password, ip string) (*model.LoginResp, error) {
	rResp, err := s.accountService.Login(ctx, &uApi.LoginReq{
		Input:    input,
		Captcha:  captcha,
		Password: password,
		Version:  "",
		DeviceId: "",
		Ip:       ip,
	})
	if err != nil {
		s.log.Error("User-Service-Login", "error", err)
		return nil, err
	}

	gResp, err := s.generateToken(ctx, rResp.UID)
	if err != nil {
		s.log.Error("User-Service-Login", "error", err)
		return nil, err
	}

	return &model.LoginResp{
		VID:   rResp.VID,
		Token: gResp.Token,
	}, nil
}
