package main

import (
	"mercury/app/comet/config"
	"mercury/app/comet/server/grpc"
	"mercury/app/comet/server/http"
	"mercury/app/comet/service"
	"mercury/x/log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Initialize configuration
	config.Init()

	c := config.NewViperProvider()

	// Initialize log
	lvl, _ := log.LvlFromString(c.LogMode())
	log.Root().SetHandler(log.LvlFilterHandler(lvl, log.StreamHandler(os.Stdout, log.TerminalFormat(true))))

	srv := service.NewService(c.Logger())

	http.Init(c, srv)
	grpc.Init(c, srv)

	// Signal handler
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-signalChan
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("[MercuryComet] service shutdown")
			srv.SessionStore.Shutdown()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
