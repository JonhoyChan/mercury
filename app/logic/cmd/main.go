package main

import (
	"mercury/app/logic/config"
	"mercury/app/logic/server/grpc"
	"mercury/app/logic/service"
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

	srv, err := service.NewService(c)
	if err != nil {
		panic("unable to initialize service:" + err.Error())
	}

	grpc.Init(c, srv)

	// Signal handler
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-signalChan
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("[MercuryLogic] service shutdown")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
