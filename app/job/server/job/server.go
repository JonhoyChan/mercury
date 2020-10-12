package job

import (
	"github.com/micro/go-micro/v2/server"
	"mercury/app/job/config"
	"mercury/app/job/service"
	"mercury/x/microx"
)

// 注册服务
func Init(c config.Provider, srv *service.Service) {
	microServer := server.NewServer(microx.InitServerOptions(c)...)
	if err := microServer.Init(); err != nil {
		panic("unable to initialize server:" + err.Error())
	}
	srv.Init(microServer.Options())

	// Run service
	go func() {
		if err := microServer.Start(); err != nil {
			panic("unable to start service:" + err.Error())
		}
	}()
}
