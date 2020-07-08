package main

import (
	"flag"
	"os"
	"os/signal"
	"outgoing/app/service/auth/config"
	"outgoing/app/service/auth/server/grpc"
	"outgoing/app/service/auth/service"
	"outgoing/x"
	"outgoing/x/log"
	"path/filepath"
	"syscall"
)

var configFile string

func init() {
	executable, _ := os.Executable()

	// All relative paths are resolved against the executable path, not against current working directory.
	// Absolute paths are left unchanged.
	rootPath, _ := filepath.Split(executable)

	path := x.ToAbsolutePath(rootPath, "auth-service.yml")

	flag.StringVar(&configFile, "config", path, "Path to config file.")
}

func main() {
	flag.Parse()

	// Initialize configuration
	config.Init(configFile)

	c := config.NewViperProvider()

	// Initialize log
	lvl, _ := log.LvlFromString(c.LogMode())
	log.Root().SetHandler(log.LvlFilterHandler(lvl, log.StreamHandler(os.Stdout, log.LogfmtFormat())))

	srv, err := service.NewService(c)
	if err != nil {
		panic("unable to initialize service:" + err.Error())
	}

	// Initialize grpc server
	grpc.Init(c, srv)

	// Signal handler
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-signalChan
		log.Info("[auth-service] get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("[auth-service] exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
