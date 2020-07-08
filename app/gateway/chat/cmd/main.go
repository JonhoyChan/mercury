package main

import (
	"flag"
	"os"
	"os/signal"
	"outgoing/app/gateway/chat/config"
	"outgoing/app/gateway/chat/http"
	"outgoing/app/gateway/chat/session"
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

	path := x.ToAbsolutePath(rootPath, "chat-gateway.yml")

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
	log.Root().SetHandler(log.LvlFilterHandler(lvl, log.StreamHandler(os.Stdout, log.LogfmtFormat())))

	// Initialize http server
	http.Init(c)

	// Signal handler
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-signalChan
		log.Info("[chat-gateway] get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			session.GlobalSessionStore.Shutdown()
			log.Info("[chat-gateway] exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
