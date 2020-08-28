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
	apiClient api.ChatAdminService
	ctx       context.Context
)

func init() {
	apiClient = api.NewChatAdminService("service.chat.logic", grpc.NewClient(
		client.Retry(ecode.RetryOnMicroError),
		client.WrapCall(ecode.MicroCallFunc)),
	)
	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)
}

func TestGrpcServer_Client(t *testing.T) {
	var clientID, clientSecret string
	t.Run("Create Client", func(t *testing.T) {
		resp, err := apiClient.CreateClient(ctx, &api.CreateClientReq{
			Name:        "mercury",
			TokenSecret: "6ZG~izEhm1wGfITYR2Sx6cClCC",
			TokenExpire: 604800,
		})
		require.Nil(t, err)

		t.Logf("client id: %s, client secret: %s", resp.ClientID, resp.ClientSecret)
		clientID, clientSecret = resp.ClientID, resp.ClientSecret
	})

	var token string
	t.Run("Generate token", func(t *testing.T) {
		resp, err := apiClient.GenerateToken(ctx, &api.GenerateTokenReq{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		})
		require.Nil(t, err)

		t.Logf("token: %s, lifetime: %s", resp.Token, resp.Lifetime)
		token = resp.Token
	})

	t.Run("Update Client", func(t *testing.T) {
		_, err := apiClient.UpdateClient(ctx, &api.UpdateClientReq{
			Token: token,
			TokenSecret: &api.StringValue{
				Value: "chQBriRm7i0bOGbDhTGTCeNzGd",
			},
			TokenExpire: &api.Int64Value{
				Value: 1209600,
			},
		})
		require.Nil(t, err)
	})

	t.Run("Delete Client", func(t *testing.T) {
		_, err := apiClient.DeleteClient(ctx, &api.DeleteClientReq{
			Token: token,
		})
		require.Nil(t, err)
	})
}

func TestGrpcServer_User(t *testing.T) {
	var token string
	t.Run("Generate token", func(t *testing.T) {
		resp, err := apiClient.GenerateToken(ctx, &api.GenerateTokenReq{
			ClientID:     "c4ff4ca9-6f2f-4fdb-8481-8460c3ace3b0",
			ClientSecret: "Aiz_KDuKV3Bkdv9OYtF3Tv1ZYU",
		})
		require.Nil(t, err)

		t.Logf("token: %s, lifetime: %s", resp.Token, resp.Lifetime)
		token = resp.Token
	})

	var uid string
	t.Run("Create user", func(t *testing.T) {
		resp, err := apiClient.CreateUser(ctx, &api.CreateUserReq{
			Token: token,
			Name:  "Jonhoy",
		})
		require.Nil(t, err)

		t.Logf("uid: %s", resp.UID)
		uid = resp.UID
	})

	t.Run("Generate user token", func(t *testing.T) {
		resp, err := apiClient.GenerateUserToken(ctx, &api.GenerateUserTokenReq{
			Token: token,
			UID:   uid,
		})
		require.Nil(t, err)

		t.Logf("token: %s, lifetime: %s", resp.Token, resp.Lifetime)
	})

	t.Run("Update user activated", func(t *testing.T) {
		_, err := apiClient.UpdateActivated(ctx, &api.UpdateActivatedReq{
			Token:     token,
			UID:       uid,
			Activated: false,
		})
		require.Nil(t, err)
	})

	t.Run("Delete user", func(t *testing.T) {
		_, err := apiClient.DeleteUser(ctx, &api.DeleteUserReq{
			Token: token,
			UID:   uid,
		})
		require.Nil(t, err)
	})
}
