package lib

import (
	"context"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-plugins/broker/stan/v2"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"
	"mercury/app/logic/api"
	"mercury/app/logic/service"
	"mercury/config"
	"mercury/x"
	"mercury/x/ecode"
	"mercury/x/log"
	"mercury/x/microx"
	"strings"
)

type LogicServer struct {
	inst *Instance
	log  log.Logger
	srv  service.Servicer
}

func NewLogicServer(inst *Instance, l log.Logger) *LogicServer {
	return &LogicServer{
		inst: inst,
		log:  l,
	}
}

func (s *LogicServer) Serve(ctx context.Context) error {
	cfg := config.NewProviderConfig(s.inst.cfg)
	var err error
	if s.srv, err = service.NewService(cfg, s.log.New("service", "mercury.logic")); err != nil {
		return err
	}

	srvCfg, founded := cfg.GetService("mercury.logic")
	if !founded {
		return ecode.NewError("can not found \"mercury.job\" service config")
	}

	opts := microx.DefaultMicroOptions(srvCfg)
	opts = append(opts, micro.WrapHandler(
		ratelimit.NewHandlerWrapper(1024),
		s.srv.AuthenticateClientToken,
	))

	// 判断是否使用了etcd作为服务注册
	if cfg.Registry().ETCD.Enable {
		r := etcdv3.NewRegistry(func(op *registry.Options) {
			var addresses []string
			for _, v := range cfg.Registry().ETCD.Addresses {
				v = strings.TrimSpace(v)
				addresses = append(addresses, x.ReplaceHttpOrHttps(v))
			}

			op.Addrs = addresses
		})
		opts = append(opts, micro.Registry(r))
	}

	if cfg.Broker().Stan.Enable {
		// 创建一个新stanBroker实例
		b := stan.NewBroker(
			// 设置stan集群的地址
			broker.Addrs(cfg.Broker().Stan.Addresses...),
			stan.ConnectRetry(true),
			// 设置stan集群标识
			stan.ClusterID(cfg.Broker().Stan.ClusterID),
			// 设置订阅者使用的永久名
			stan.DurableName(cfg.Broker().Stan.DurableName),
		)

		if err := b.Init(); err != nil {
			panic("unable to init stan broker:" + err.Error())
		}

		if err := b.Connect(); err != nil {
			panic("unable to connect to stan broker:" + err.Error())
		}

		opts = append(opts, micro.Broker(b))
	}

	microServer := micro.NewService(opts...)
	microServer.Init()

	if err := api.RegisterChatAdminHandler(microServer.Server(), s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}
	if err := api.RegisterChatClientAdminHandler(microServer.Server(), s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}
	if err := api.RegisterChatHandler(microServer.Server(), s); err != nil {
		panic("unable to register grpc service:" + err.Error())
	}

	return microServer.Run()
}

func (s *LogicServer) GenerateToken(ctx context.Context, req *api.GenerateTokenReq, resp *api.TokenResp) error {
	token, lifetime, err := s.srv.GenerateToken(ctx, req)
	if err != nil {
		return err
	}

	resp.Token = token
	resp.Lifetime = lifetime
	return nil
}

func (s *LogicServer) GetClient(ctx context.Context, req *api.GetClientReq, resp *api.GetClientResp) error {
	client, err := s.srv.GetClient(ctx)
	if err != nil {
		return err
	}

	resp.Client = client
	return nil
}

func (s *LogicServer) CreateClient(ctx context.Context, req *api.CreateClientReq, resp *api.CreateClientResp) error {
	clientID, clientSecret, err := s.srv.CreateClient(ctx, req)
	if err != nil {
		return err
	}

	resp.ClientID = clientID
	resp.ClientSecret = clientSecret
	return nil
}

func (s *LogicServer) UpdateClient(ctx context.Context, req *api.UpdateClientReq, resp *api.Empty) error {
	if err := s.srv.UpdateClient(ctx, req); err != nil {
		return err
	}

	return nil
}

func (s *LogicServer) DeleteClient(ctx context.Context, req *api.DeleteClientReq, resp *api.Empty) error {
	if err := s.srv.DeleteClient(ctx); err != nil {
		return err
	}

	return nil
}

func (s *LogicServer) CreateUser(ctx context.Context, req *api.CreateUserReq, resp *api.CreateUserResp) error {
	uid, err := s.srv.CreateUser(ctx, req)
	if err != nil {
		return err
	}

	resp.UID = uid
	return nil
}

func (s *LogicServer) UpdateActivated(ctx context.Context, req *api.UpdateActivatedReq, resp *api.Empty) error {
	err := s.srv.UpdateActivated(ctx, req.UID, req.Activated)
	if err != nil {
		return err
	}

	return nil
}

func (s *LogicServer) DeleteUser(ctx context.Context, req *api.DeleteUserReq, resp *api.Empty) error {
	if err := s.srv.DeleteUser(ctx, req.UID); err != nil {
		return err
	}

	return nil
}

func (s *LogicServer) GenerateUserToken(ctx context.Context, req *api.GenerateUserTokenReq, resp *api.TokenResp) error {
	token, lifetime, err := s.srv.GenerateUserToken(ctx, req.UID)
	if err != nil {
		return err
	}

	resp.Token = token
	resp.Lifetime = lifetime
	return nil
}

func (s *LogicServer) AddFriend(ctx context.Context, req *api.AddFriendReq, resp *api.Empty) error {
	err := s.srv.AddFriend(ctx, req.UID, req.FriendUID)
	if err != nil {
		return err
	}

	return nil
}

func (s *LogicServer) GetFriends(ctx context.Context, req *api.GetFriendsReq, resp *api.GetFriendsResp) error {
	friends, err := s.srv.GetFriends(ctx, req.UID)
	if err != nil {
		return err
	}

	resp.Friends = friends
	return nil
}

func (s *LogicServer) DeleteFriend(ctx context.Context, req *api.DeleteFriendReq, resp *api.Empty) error {
	err := s.srv.DeleteFriend(ctx, req.UID, req.FriendUID)
	if err != nil {
		return err
	}

	return nil
}

func (s *LogicServer) CreateGroup(ctx context.Context, req *api.CreateGroupReq, resp *api.CreateGroupResp) error {
	group, err := s.srv.CreateGroup(ctx, req)
	if err != nil {
		return err
	}

	resp.Group = group
	return nil
}

func (s *LogicServer) GetGroups(ctx context.Context, req *api.GetGroupsReq, resp *api.GetGroupsResp) error {
	groups, err := s.srv.GetGroups(ctx, req.UID)
	if err != nil {
		return err
	}

	resp.Groups = groups
	return nil
}

func (s *LogicServer) AddMember(ctx context.Context, req *api.AddMemberReq, resp *api.Empty) error {
	err := s.srv.AddMember(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *LogicServer) GetMembers(ctx context.Context, req *api.GetMembersReq, resp *api.GetMembersResp) error {
	members, err := s.srv.GetMembers(ctx, req.GID)
	if err != nil {
		return err
	}

	resp.Members = members
	return nil
}

func (s *LogicServer) Listen(ctx context.Context, req *api.ListenReq, stream api.ChatClientAdmin_ListenStream) error {
	err := s.srv.Listen(ctx, req.Token, stream)
	if err != nil {
		return err
	}
	return nil
}

func (s *LogicServer) Connect(ctx context.Context, req *api.ConnectReq, resp *api.ConnectResp) error {
	clientID, uid, err := s.srv.Connect(ctx, req)
	if err != nil {
		return err
	}

	resp.ClientID = clientID
	resp.UID = uid
	return nil
}

func (s *LogicServer) Disconnect(ctx context.Context, req *api.DisconnectReq, resp *api.Empty) error {
	err := s.srv.Disconnect(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *LogicServer) Heartbeat(ctx context.Context, req *api.HeartbeatReq, resp *api.Empty) error {
	err := s.srv.Heartbeat(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *LogicServer) PushMessage(ctx context.Context, req *api.PushMessageReq, resp *api.PushMessageResp) error {
	id, sequence, err := s.srv.PushMessage(ctx, req)
	if err != nil {
		return err
	}

	resp.MessageId = id
	resp.Sequence = sequence
	return nil
}

func (s *LogicServer) PullMessage(ctx context.Context, req *api.PullMessageReq, resp *api.PullMessageResp) error {
	topicMessages, err := s.srv.PullMessage(ctx, req)
	if err != nil {
		return err
	}

	resp.TopicMessages = topicMessages
	return nil
}

func (s *LogicServer) ReadMessage(ctx context.Context, req *api.ReadMessageReq, resp *api.Empty) error {
	err := s.srv.ReadMessage(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *LogicServer) Keypress(ctx context.Context, req *api.KeypressReq, resp *api.Empty) error {
	err := s.srv.Keypress(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
