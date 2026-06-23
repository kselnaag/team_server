package msrv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	T "team_server/internal/types"
	"time"
)

var _ T.IMsrv = (*Msrv)(nil)

type Msrv struct {
	appdir string
	cfg    T.ICfg
	log    T.ILog
}

func NewMediaServer(cfg T.ICfg, log T.ILog, appdir string) *Msrv {
	return &Msrv{
		appdir: appdir,
		cfg:    cfg,
		log:    log,
		// HTTP client
	}
}

func (msrv *Msrv) PathInit(path string) error {

	return nil
}

func (msrv *Msrv) PathFini(path string) error {

	return nil
}

func (msrv *Msrv) Start() func(err error) {
	var (
		procCtx    context.Context
		procCancel context.CancelFunc
	)
	ctx := context.Background()
	checkCtx, checkCancel := context.WithCancel(ctx)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				procCtx, procCancel = context.WithCancel(ctx)
				cmd := exec.CommandContext(procCtx, filepath.Join(msrv.appdir, "mediamtx"))
				cmd.Dir = msrv.appdir
				cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
				if err := cmd.Start(); err != nil {
					procCancel()
					continue
				}
				msrv.log.LogInfo("mediamtx started")
				_ = cmd.Wait()
			}
			msrv.log.LogInfo("mediamtx stoped")
		}
	}(checkCtx)
	msrv.log.LogInfo("Msrv started")
	return func(err error) { // MsrvStop
		checkCancel()
		time.Sleep(100 * time.Millisecond)
		procCancel()
		if err != nil {
			msrv.log.LogError(fmt.Errorf("%s: %w", "Msrv stoped with error", err))
		} else {
			msrv.log.LogInfo("Msrv stoped")
		}
	}
}
