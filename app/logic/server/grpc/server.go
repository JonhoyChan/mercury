package grpc

import (
	"context"
	"github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"
	"mercury/app/logic/api"
	"mercury/app/logic/config"
	"mercury/app/logic/service"
	"mercury/x/microx"

	"github.com/micro/go-micro/v2"
)

type grpcServer struct {
	s *service.Service
}

// 注册服务
func Init(c config.Provider, srv *service.Service) {
	opts := append(microx.InitMicroOptions(c), micro.WrapHandler(
		ratelimit.NewHandlerWrapper(1024),
		srv.AuthenticateClientToken,
	))
	microServer := micro.NewService(opts...)
	microServer.Init()

	s := &grpcServer{
		s: srv,
	}

	if err := api.RegisterChatAdminHandler(microServer.Server(), s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}
	if err := api.RegisterChatHandler(microServer.Server(), s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}

	go func() {
		if err := microServer.Run(); err != nil {
			panic("unable to start service:" + err.Error())
		}
	}()
}

func (s *grpcServer) GenerateToken(ctx context.Context, req *api.GenerateTokenReq, resp *api.TokenResp) error {
	token, lifetime, err := s.s.GenerateToken(ctx, req)
	if err != nil {
		return err
	}

	resp.Token = token
	resp.Lifetime = lifetime
	return nil
}

func (s *grpcServer) CreateClient(ctx context.Context, req *api.CreateClientReq, resp *api.CreateClientResp) error {
	clientID, clientSecret, err := s.s.CreateClient(ctx, req)
	if err != nil {
		return err
	}

	resp.ClientID = clientID
	resp.ClientSecret = clientSecret
	return nil
}

func (s *grpcServer) UpdateClient(ctx context.Context, req *api.UpdateClientReq, resp *api.Empty) error {
	if err := s.s.UpdateClient(ctx, req); err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) DeleteClient(ctx context.Context, req *api.DeleteClientReq, resp *api.Empty) error {
	if err := s.s.DeleteClient(ctx); err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) CreateUser(ctx context.Context, req *api.CreateUserReq, resp *api.CreateUserResp) error {
	uid, err := s.s.CreateUser(ctx, req)
	if err != nil {
		return err
	}

	resp.UID = uid
	return nil
}

func (s *grpcServer) UpdateActivated(ctx context.Context, req *api.UpdateActivatedReq, resp *api.Empty) error {
	err := s.s.UpdateActivated(ctx, req.UID, req.Activated)
	if err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) DeleteUser(ctx context.Context, req *api.DeleteUserReq, resp *api.Empty) error {
	if err := s.s.DeleteUser(ctx, req.UID); err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) GenerateUserToken(ctx context.Context, req *api.GenerateUserTokenReq, resp *api.TokenResp) error {
	token, lifetime, err := s.s.GenerateUserToken(ctx, req.UID)
	if err != nil {
		return err
	}

	resp.Token = token
	resp.Lifetime = lifetime
	return nil
}

func (s *grpcServer) AddFriend(ctx context.Context, req *api.AddFriendReq, resp *api.Empty) error {
	err := s.s.AddFriend(ctx, req.UID, req.FriendUID)
	if err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) GetFriends(ctx context.Context, req *api.GetFriendsReq, resp *api.GetFriendsResp) error {
	friends, err := s.s.GetFriends(ctx, req.UID)
	if err != nil {
		return err
	}

	resp.Friends = friends
	return nil
}

func (s *grpcServer) DeleteFriend(ctx context.Context, req *api.DeleteFriendReq, resp *api.Empty) error {
	err := s.s.DeleteFriend(ctx, req.UID, req.FriendUID)
	if err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) CreateGroup(ctx context.Context, req *api.CreateGroupReq, resp *api.CreateGroupResp) error {
	group, err := s.s.CreateGroup(ctx, req)
	if err != nil {
		return err
	}

	resp.Group = group
	return nil
}

func (s *grpcServer) GetGroups(ctx context.Context, req *api.GetGroupsReq, resp *api.GetGroupsResp) error {
	groups, err := s.s.GetGroups(ctx, req.UID)
	if err != nil {
		return err
	}

	resp.Groups = groups
	return nil
}

func (s *grpcServer) AddMember(ctx context.Context, req *api.AddMemberReq, resp *api.Empty) error {
	err := s.s.AddMember(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) GetMembers(ctx context.Context, req *api.GetMembersReq, resp *api.GetMembersResp) error {
	members, err := s.s.GetMembers(ctx, req.GID)
	if err != nil {
		return err
	}

	resp.Members = members
	return nil
}

func (s *grpcServer) Listen(ctx context.Context, req *api.ListenReq, stream api.ChatAdmin_ListenStream) error {
	err := s.s.Listen(ctx, req.Token, stream)
	if err != nil {
		return err
	}
	return nil
}

func (s *grpcServer) Connect(ctx context.Context, req *api.ConnectReq, resp *api.ConnectResp) error {
	clientID, uid, err := s.s.Connect(ctx, req)
	if err != nil {
		return err
	}

	resp.ClientID = clientID
	resp.UID = uid
	return nil
}

func (s *grpcServer) Disconnect(ctx context.Context, req *api.DisconnectReq, resp *api.Empty) error {
	err := s.s.Disconnect(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) Heartbeat(ctx context.Context, req *api.HeartbeatReq, resp *api.Empty) error {
	err := s.s.Heartbeat(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) PushMessage(ctx context.Context, req *api.PushMessageReq, resp *api.PushMessageResp) error {
	id, sequence, err := s.s.PushMessage(ctx, req)
	if err != nil {
		return err
	}

	resp.MessageId = id
	resp.Sequence = sequence
	return nil
}

func (s *grpcServer) PullMessage(ctx context.Context, req *api.PullMessageReq, resp *api.PullMessageResp) error {
	topicMessages, err := s.s.PullMessage(ctx, req)
	if err != nil {
		return err
	}

	resp.TopicMessages = topicMessages
	return nil
}

func (s *grpcServer) ReadMessage(ctx context.Context, req *api.ReadMessageReq, resp *api.Empty) error {
	err := s.s.ReadMessage(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *grpcServer) Keypress(ctx context.Context, req *api.KeypressReq, resp *api.Empty) error {
	err := s.s.Keypress(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
