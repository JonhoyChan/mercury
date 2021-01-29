package main

import (
	"flag"
	"mercury/app/admin/config"
	"mercury/app/admin/server/http"
	"mercury/app/admin/service"
	"mercury/x/log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", ".config/mercury-admin.yml", "Path to config file.")
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

	// Signal handler
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-signalChan
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Info("[MercuryAdmin] service shutdown")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
