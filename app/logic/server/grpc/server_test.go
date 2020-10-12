package grpc

import (
	"context"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/stretchr/testify/require"
	"mercury/app/logic/api"
	"mercury/x/ecode"
	"testing"
	"time"
)

var (
	apiAdminClient api.ChatAdminService
	apiClient      api.ChatService
	ctx            context.Context
)

func init() {
	apiAdminClient = api.NewChatAdminService("mercury.logic", grpc.NewClient(
		client.Retry(ecode.RetryOnMicroError),
		client.WrapCall(ecode.MicroCallFunc)),
	)
	apiClient = api.NewChatService("mercury.logic", grpc.NewClient(
		client.Retry(ecode.RetryOnMicroError),
		client.WrapCall(ecode.MicroCallFunc)),
	)
	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)
}

func TestGrpcServer_Client(t *testing.T) {
	var clientID, clientSecret string
	t.Run("Create Client", func(t *testing.T) {
		resp, err := apiAdminClient.CreateClient(ctx, &api.CreateClientReq{
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
		resp, err := apiAdminClient.GenerateToken(ctx, &api.GenerateTokenReq{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		})
		require.Nil(t, err)

		t.Logf("token: %s, lifetime: %s", resp.Token, resp.Lifetime)
		token = resp.Token
	})

	t.Run("Update Client", func(t *testing.T) {
		_, err := apiAdminClient.UpdateClient(ctx, &api.UpdateClientReq{
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
		_, err := apiAdminClient.DeleteClient(ctx, &api.DeleteClientReq{
			Token: token,
		})
		require.Nil(t, err)
	})
}

func TestGrpcServer_User(t *testing.T) {
	var token string
	t.Run("Generate token", func(t *testing.T) {
		resp, err := apiAdminClient.GenerateToken(ctx, &api.GenerateTokenReq{
			ClientID:     "c4ff4ca9-6f2f-4fdb-8481-8460c3ace3b0",
			ClientSecret: "Aiz_KDuKV3Bkdv9OYtF3Tv1ZYU",
		})
		require.Nil(t, err)

		t.Logf("token: %s, lifetime: %s", resp.Token, resp.Lifetime)
		token = resp.Token
	})

	var uid string
	t.Run("Create user", func(t *testing.T) {
		resp, err := apiAdminClient.CreateUser(ctx, &api.CreateUserReq{
			Token: "7b2245787069726573223a313630303136353338362c2244617461223a22597a526d5a6a526a59546b744e6d59795a6930305a6d52694c5467304f4445744f4451324d474d7a59574e6c4d324977227d1c249a99733d0ae591455211e62a4fa9843018d84ff583785c61a7a396343995",
			Name:  "Molly",
		})
		require.Nil(t, err)

		t.Logf("uid: %s", resp.UID)
		uid = resp.UID
	})

	t.Run("Generate user token", func(t *testing.T) {
		resp, err := apiAdminClient.GenerateUserToken(ctx, &api.GenerateUserTokenReq{
			Token: "7b2245787069726573223a313630333730333031322c2244617461223a22597a526d5a6a526a59546b744e6d59795a6930305a6d52694c5467304f4445744f4451324d474d7a59574e6c4d324977227dd3a76fd7edac25c916cf2b0e61a85a4b25fd64c3cf3a1ef9f017ce00e8e78720",
			UID:   "uidOwbRDvaLyaw",
		})
		require.Nil(t, err)

		t.Logf("token: %s, lifetime: %s", resp.Token, resp.Lifetime)
	})

	t.Run("Update user activated", func(t *testing.T) {
		_, err := apiAdminClient.UpdateActivated(ctx, &api.UpdateActivatedReq{
			Token:     token,
			UID:       uid,
			Activated: false,
		})
		require.Nil(t, err)
	})

	t.Run("Delete user", func(t *testing.T) {
		_, err := apiAdminClient.DeleteUser(ctx, &api.DeleteUserReq{
			Token: token,
			UID:   uid,
		})
		require.Nil(t, err)
	})

	t.Run("Add friend", func(t *testing.T) {
		_, err := apiAdminClient.AddFriend(ctx, &api.AddFriendReq{
			Token:     "7b2245787069726573223a313630303136353338362c2244617461223a22597a526d5a6a526a59546b744e6d59795a6930305a6d52694c5467304f4445744f4451324d474d7a59574e6c4d324977227d1c249a99733d0ae591455211e62a4fa9843018d84ff583785c61a7a396343995",
			UID:       "uid7KA8fY5Jb3A",
			FriendUID: "uidOwbRDvaLyaw",
		})
		require.Nil(t, err)
	})

	t.Run("Get friends", func(t *testing.T) {
		resp, err := apiAdminClient.GetFriends(ctx, &api.GetFriendsReq{
			Token: "7b2245787069726573223a313630303136353338362c2244617461223a22597a526d5a6a526a59546b744e6d59795a6930305a6d52694c5467304f4445744f4451324d474d7a59574e6c4d324977227d1c249a99733d0ae591455211e62a4fa9843018d84ff583785c61a7a396343995",
			UID:   "uid7KA8fY5Jb3A",
		})
		require.Nil(t, err)

		t.Logf("friends: %+v", resp.Friends)
	})

	t.Run("Delete user", func(t *testing.T) {
		_, err := apiAdminClient.DeleteFriend(ctx, &api.DeleteFriendReq{
			Token:     "7b2245787069726573223a313630303136353338362c2244617461223a22597a526d5a6a526a59546b744e6d59795a6930305a6d52694c5467304f4445744f4451324d474d7a59574e6c4d324977227d1c249a99733d0ae591455211e62a4fa9843018d84ff583785c61a7a396343995",
			UID:       "uid7KA8fY5Jb3A",
			FriendUID: "uiduN_f_2oWkUQ",
		})
		require.Nil(t, err)
	})
}

func TestGrpcServer_Group(t *testing.T) {
	//var gid string
	t.Run("Create group", func(t *testing.T) {
		resp, err := apiAdminClient.CreateGroup(ctx, &api.CreateGroupReq{
			Token:        "7b2245787069726573223a313630303136353338362c2244617461223a22597a526d5a6a526a59546b744e6d59795a6930305a6d52694c5467304f4445744f4451324d474d7a59574e6c4d324977227d1c249a99733d0ae591455211e62a4fa9843018d84ff583785c61a7a396343995",
			Name:         "陈府",
			Introduction: "相亲相爱一家人",
			Owner:        "uiduN_f_2oWkUQ",
		})
		require.Nil(t, err)

		t.Logf("group: %v", resp.Group)
		//gid = resp.GID
	})

	t.Run("Get groups", func(t *testing.T) {
		resp, err := apiAdminClient.GetGroups(ctx, &api.GetGroupsReq{
			Token: "7b2245787069726573223a313630303136353338362c2244617461223a22597a526d5a6a526a59546b744e6d59795a6930305a6d52694c5467304f4445744f4451324d474d7a59574e6c4d324977227d1c249a99733d0ae591455211e62a4fa9843018d84ff583785c61a7a396343995",
			UID:   "uid7KA8fY5Jb3A",
		})
		require.Nil(t, err)

		t.Logf("groups: %+v", resp.Groups)
	})

	t.Run("Add member", func(t *testing.T) {
		_, err := apiAdminClient.AddMember(ctx, &api.AddMemberReq{
			Token: "7b2245787069726573223a313630303136353338362c2244617461223a22597a526d5a6a526a59546b744e6d59795a6930305a6d52694c5467304f4445744f4451324d474d7a59574e6c4d324977227d1c249a99733d0ae591455211e62a4fa9843018d84ff583785c61a7a396343995",
			GID:   "gid4Fl1QvXZpM4",
			UID:   "uidOwbRDvaLyaw",
		})
		require.Nil(t, err)
	})

	t.Run("Get members", func(t *testing.T) {
		resp, err := apiAdminClient.GetMembers(ctx, &api.GetMembersReq{
			Token: "7b2245787069726573223a313630303136353338362c2244617461223a22597a526d5a6a526a59546b744e6d59795a6930305a6d52694c5467304f4445744f4451324d474d7a59574e6c4d324977227d1c249a99733d0ae591455211e62a4fa9843018d84ff583785c61a7a396343995",
			GID:   "gid4Fl1QvXZpM4",
		})
		require.Nil(t, err)

		t.Logf("members: %+v", resp.Members)
	})
}

func TestGrpcServer_PullMessage(t *testing.T) {
	t.Run("pull message", func(t *testing.T) {
		resp, err := apiClient.PullMessage(ctx, &api.PullMessageReq{
			UID: "uiduN_f_2oWkUQ",
		})
		require.Nil(t, err)

		t.Logf("messages: %+v", resp.TopicMessages)
	})
}

func TestGrpcServer_PushMessage(t *testing.T) {
	t.Run("push message", func(t *testing.T) {
		resp, err := apiClient.PushMessage(ctx, &api.PushMessageReq{
			ClientID:    "c4ff4ca9-6f2f-4fdb-8481-8460c3ace3b0",
			MessageType: api.MessageTypeSingle,
			Sender:      "uiduN_f_2oWkUQ",
			Receiver:    "uid7KA8fY5Jb3A",
			ContentType: api.ContentTypeText,
			Body:        []byte(`{"content": "Hello, World!"}`),
			Mentions:    nil,
		})
		require.Nil(t, err)

		t.Logf("sequence: %d", resp.Sequence)
	})
}

func TestGrpcServer_Listen(t *testing.T) {
	t.Run("listen", func(t *testing.T) {
		resp, err := apiAdminClient.Listen(context.Background(), &api.ListenReq{
			Token: "7b2245787069726573223a313630303136353338362c2244617461223a22597a526d5a6a526a59546b744e6d59795a6930305a6d52694c5467304f4445744f4451324d474d7a59574e6c4d324977227d1c249a99733d0ae591455211e62a4fa9843018d84ff583785c61a7a396343995",
		})
		require.Nil(t, err)

		i := 0
		for {
			i++
			message, err := resp.Recv()
			require.Nil(t, err)

			time.Sleep(1 * time.Second)
			if i > 5 {
				break
			}

			t.Logf("message: %+v", message)
		}

	})
}
