package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	APP "team_server/internal"
)

func main() {
	runtime.GOMAXPROCS(1)

	myApp := APP.NewApp()
	myAppStop := myApp.Start()

	defer func() {
		if err := recover(); err != nil {
			myAppStop(err.(error))
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	for {
		switch <-sig {
		case syscall.SIGHUP: // kill -SIGHUP <pid> // restarting all for sake of config reload
			myAppStop(nil)
			myApp = APP.NewApp()
			myAppStop = myApp.Start()
		default:
			myAppStop(nil)
			os.Exit(0)
		}
	}
}
