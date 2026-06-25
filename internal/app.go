package app

import (
	"fmt"
	"os"
	"path/filepath"

	C "team_server/internal/cfg"
	L "team_server/internal/log"
	M "team_server/internal/msrv"
	T "team_server/internal/types"
)

type App struct {
	appname string
	cfg     T.ICfg
	log     T.ILog
	msrv    T.ImSrv
	//tg      T.ITG
}

func execPathAndFname() (string, string) {
	path, _ := os.Executable()
	return filepath.Split(path)
}

func NewApp() *App {
	appdir, appname := execPathAndFname()
	cfg := C.NewCfgMaps(appdir, appname).Parse()
	log := L.NewLogFprintf(cfg, 0, 0)
	msrv := M.NewMediaServer(cfg, log, appdir)
	// tg := TG.NewTGBot(cfg, log, msrv)
	return &App{
		appname: appname,
		cfg:     cfg,
		log:     log,
		msrv:    msrv,
		// tg:      tg,
	}
}

func (a *App) Start() func(err error) {
	logStop := a.log.Start()
	mSrvStop := a.msrv.Start()
	// TGStop := a.tg.Start()
	a.log.LogInfo(a.appname + " app started")
	return func(err error) { // AppStop
		// TGStop(err)
		mSrvStop(err)
		if err != nil {
			a.log.LogError(fmt.Errorf("%s: %w", a.appname+" app stoped with error", err))
		} else {
			a.log.LogInfo(a.appname + " app stoped")
		}
		logStop()
	}
}
