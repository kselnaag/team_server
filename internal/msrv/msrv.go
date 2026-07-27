package msrv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	T "team_server/internal/types"
	"time"
)

var _ T.ImSrv = (*MSrv)(nil)

type MSrv struct {
	appdir string
	cfg    T.ICfg
	log    T.ILog
}

func NewMediaServer(cfg T.ICfg, log T.ILog, appdir string) *MSrv {
	return &MSrv{
		appdir: appdir,
		cfg:    cfg,
		log:    log,
	}
}

func (msrv *MSrv) sendPatchReq(url string, payload map[string]any) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		err = fmt.Errorf("sendPatchReq(): Ошибка сериализации json: %w", err)
		msrv.log.LogError(err)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(jsonData))
	if err != nil {
		err = fmt.Errorf("sendPatchReq(): Ошибка создания PATCH запроса: %w", err)
		msrv.log.LogError(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("sendPatchReq(): Ошибка отправки PATCH запроса: %w", err)
		msrv.log.LogError(err)
		return err
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("sendPatchReq(): Ошибка чтения PATCH ответа: %w", err)
		msrv.log.LogError(err)
		return err
	}
	return nil
}

func (msrv *MSrv) PathInit(path string) error {
	url := "http://localhost:9997/v3/config/paths/patch/" + path
	payload := map[string]any{
		"runOnInitRestart": true,
		"alwaysAvailable":  true,
	}
	err := msrv.sendPatchReq(url, payload)
	if err != nil {
		err = fmt.Errorf("PathInit("+path+"): Ошибка Patch запроса: %w", err)
		msrv.log.LogInfo(err.Error())
	}
	return err
}

func (msrv *MSrv) PathFini(path string) error {
	url := "http://localhost:9997/v3/config/paths/patch/" + path
	payload := map[string]any{
		"runOnInitRestart": false,
		"alwaysAvailable":  false,
	}
	err := msrv.sendPatchReq(url, payload)
	if err != nil {
		err = fmt.Errorf("PathFini("+path+"): Ошибка Patch запроса: %w", err)
		msrv.log.LogInfo(err.Error())
	}
	return err
}

func (msrv *MSrv) Start() func(err error) {
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
				cmd.Env = append(os.Environ(), "GOMAXPROCS=1")
				cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
				if err := cmd.Start(); err != nil {
					msrv.log.LogError(fmt.Errorf("mediamtx can not started with error: %w", err))
					procCancel()
					continue
				}
				msrv.log.LogInfo("mediamtx started")
				if err := cmd.Wait(); err != nil {
					msrv.log.LogDebug("mediamtx stoped with error: %w", err)
				}
			}
			msrv.log.LogInfo("mediamtx stoped")
		}
	}(checkCtx)
	time.Sleep(100 * time.Millisecond)
	msrv.log.LogInfo("mSrv started")
	return func(err error) { // MSrvStop
		checkCancel()
		time.Sleep(100 * time.Millisecond)
		procCancel()
		if err != nil {
			msrv.log.LogError(fmt.Errorf("%s: %w", "mSrv stoped with error", err))
		} else {
			msrv.log.LogInfo("mSrv stoped")
		}
	}
}
