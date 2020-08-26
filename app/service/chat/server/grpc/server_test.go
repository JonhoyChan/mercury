package grpc

import (
	"context"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/stretchr/testify/require"
	"outgoing/app/service/chat/api"
	"outgoing/x/ecode"
	"testing"
	"time"
)

var (
	apiClient api.ChatService
	ctx       context.Context
)

func init() {
	apiClient = api.NewChatService("service.chat.logic", grpc.NewClient(
		client.Retry(ecode.RetryOnMicroError),
		client.WrapCall(ecode.MicroCallFunc)),
	)
	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)
}

func TestGrpcServer_CreateClient(t *testing.T) {
	resp, err := apiClient.CreateClient(ctx, &api.CreateClientReq{
		Name:        "mercury",
		TokenSecret: "6ZG~izEhm1wGfITYR2Sx6cClCC",
		TokenExpire: 604800,
	})
	require.Nil(t, err)

	t.Logf("client id: %s, client secret: %s", resp.ClientID, resp.ClientSecret)
}

func TestGrpcServer_GenerateToken(t *testing.T) {
	resp, err := apiClient.GenerateToken(ctx, &api.GenerateTokenReq{
		ClientID:     "c55f97da-1125-4a0e-88ba-6d418aa0ced5",
		ClientSecret: "Z3SF-0_azXa-2wK9H8L9-2~zW0",
	})
	require.Nil(t, err)

	t.Logf("token: %s, lifetime: %s", resp.Token, resp.Lifetime)
}

func TestGrpcServer_UpdateClient(t *testing.T) {
	token := "7b2245787069726573223a313539393633393031392c2244617461223a22597a55315a6a6b335a4745744d5445794e5330305954426c4c546734596d45744e6d51304d5468685954426a5a575131227d813320d785192809a2d2110b7e339aa13505649eb6bcbea4b56c64219c67f927"
	_, err := apiClient.UpdateClient(ctx, &api.UpdateClientReq{
		Token: token,
	})
	require.Nil(t, err)
}

func TestGrpcServer_DeleteClient(t *testing.T) {
	token := "7b2245787069726573223a313539393633393031392c2244617461223a22597a55315a6a6b335a4745744d5445794e5330305954426c4c546734596d45744e6d51304d5468685954426a5a575131227d813320d785192809a2d2110b7e339aa13505649eb6bcbea4b56c64219c67f927"
	_, err := apiClient.DeleteClient(ctx, &api.DeleteClientReq{
		Token: token,
	})
	require.Nil(t, err)
}
