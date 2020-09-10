package main

import (
	"flag"
	"os"
	"os/signal"
	"outgoing/app/comet/config"
	"outgoing/app/comet/server/grpc"
	"outgoing/app/comet/server/http"
	"outgoing/app/comet/service"
	"outgoing/x"
	"outgoing/x/log"
	"path/filepath"
	"runtime"
	"syscall"
)

var configFile string

func init() {
	executable, _ := os.Executable()

	// All relative paths are resolved against the executable path, not against current working directory.
	// Absolute paths are left unchanged.
	rootPath, _ := filepath.Split(executable)

	path := x.ToAbsolutePath(rootPath, "mercury-comet.yml")

	flag.StringVar(&configFile, "config", path, "Path to config file.")
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Initialize configuration
	config.Init(configFile)

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
			log.Info("[ChatGateway] service shutdown")
			srv.SessionStore.Shutdown()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
