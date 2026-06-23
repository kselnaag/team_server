/*
	Fprintf log module:

- universal DI interface (see types/log.go)
- structured log to JSON
- manual key positions into JSON object
- 8 Log levels (trace, debug, info, warn, error, panic, fatal, nolog)
- stack trace in Panic and Fatal log messages (os.Exit(1) on Fatal)
- multi-target message sending with io.Writer interface (if empty - os.Stderr)
- log batching with timeout (if 0 - no batching)
- log metrics from "runtime/metrics" with timeout (if 0 - no metrics)
- Error, Panic, Fatal has filepath and line number
- ASSERT, TODO, UNREACHABLE dev messages are sending to os.Stdout
*/
package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/metrics"
	"strings"
	"sync"
	T "team_server/internal/types"
	"time"
)

var _ T.ILog = (*LogFprintf)(nil)

type LogFprintf struct {
	loglvl     T.LogLevel
	host       string
	svc        string
	targets    []io.Writer
	batchTime  time.Duration
	logbuf     strings.Builder
	mu         sync.Mutex
	metricTime time.Duration
}

func NewLogFprintf(cfg T.ICfg, metricTime time.Duration, batchTime time.Duration, targets ...io.Writer) *LogFprintf {
	if metricTime != 0 {
		debug.SetTraceback("all")
	}
	if len(targets) == 0 {
		targets = append(targets, os.Stderr)
	}

	var lvl T.LogLevel
	switch cfg.GetEnvVal(T.TS_LOG_LEVEL) {
	case T.StrTrace:
		lvl = T.Trace
	case T.StrDebug:
		lvl = T.Debug
	case T.StrInfo:
		lvl = T.Info
	case T.StrWarn:
		lvl = T.Warn
	case T.StrError:
		lvl = T.Error
	case T.StrPanic:
		lvl = T.Panic
	case T.StrFatal:
		lvl = T.Fatal
	default:
		lvl = T.NoLog
	}
	return &LogFprintf{
		loglvl:     lvl,
		host:       cfg.GetEnvVal(T.TS_APP_IP),
		svc:        cfg.GetEnvVal(T.TS_APP_NAME),
		targets:    targets,
		batchTime:  batchTime,
		metricTime: metricTime,
	}
}

func (l *LogFprintf) writeBatch() {
	l.mu.Lock()
	if l.logbuf.Len() != 0 {
		for _, point := range l.targets {
			fmt.Fprint(point, l.logbuf.String())
		}
		l.logbuf.Reset()
	}
	l.mu.Unlock()
}

func (l *LogFprintf) Start() func() {
	var wg sync.WaitGroup
	ctx, ctxCancel := context.WithCancel(context.Background())
	if l.batchTime != 0 {
		wg.Add(1)
		go func() {
			for {
				select {
				case <-time.After(l.batchTime):
					l.writeBatch()
				case <-ctx.Done():
					l.writeBatch()
					wg.Done()
					return
				}
			}
		}()
	}
	if l.metricTime != 0 {
		wg.Add(1)
		go func() {
			for {
				select {
				case <-time.After(l.metricTime):
					l.logMetrics(l.getMetrics())
				case <-ctx.Done():
					wg.Done()
					return
				}
			}
		}()
	}
	return func() {
		ctxCancel()
		wg.Wait()
	}
}

func replaceEOL(str string) string {
	return strings.ReplaceAll(str, "\n", "\t")
}

func (l *LogFprintf) logMessage(lvl, host, svc, mess string) {
	timenow := time.Now().Format(time.RFC3339Nano)
	formatstr := `{"T":"%s","L":"%s","H":"%s","S":"%s","M":"%s"}` + "\n" // TIME LOGLEVEL HOST SERVISE MESSAGE
	if l.batchTime == 0 {
		for _, point := range l.targets {
			fmt.Fprintf(point, formatstr, timenow, lvl, host, svc, replaceEOL(mess))
		}
	} else {
		l.mu.Lock()
		l.logbuf.WriteString(fmt.Sprintf(formatstr, timenow, lvl, host, svc, mess))
		l.mu.Unlock()
	}
}

func (l *LogFprintf) LogTrace(format string, v ...any) {
	if l.loglvl <= T.Trace {
		l.logMessage(T.StrTrace, l.host, l.svc, fmt.Sprintf(format, v...))
	}
}

func (l *LogFprintf) LogDebug(format string, v ...any) {
	if l.loglvl <= T.Debug {
		l.logMessage(T.StrDebug, l.host, l.svc, fmt.Sprintf(format, v...))
	}
}

func (l *LogFprintf) LogInfo(format string, v ...any) {
	if l.loglvl <= T.Info {
		l.logMessage(T.StrInfo, l.host, l.svc, fmt.Sprintf(format, v...))
	}
}

func (l *LogFprintf) LogWarn(format string, v ...any) {
	if l.loglvl <= T.Warn {
		l.logMessage(T.StrWarn, l.host, l.svc, fmt.Sprintf(format, v...))
	}
}

func (l *LogFprintf) LogError(err error) {
	if l.loglvl <= T.Error {
		l.logMessage(T.StrError, l.host, l.svc, fmt.Sprintf("%s: %s", getLineNumber(), err.Error()))
	}
}

func (l *LogFprintf) LogPanic(err error) {
	if l.loglvl <= T.Panic {
		l.logMessage(T.StrPanic, l.host, l.svc, fmt.Sprintf("%s: %s: %s", getLineNumber(), err.Error(), debug.Stack()))
	}
}

func (l *LogFprintf) LogFatal(err error) {
	if l.loglvl <= T.Fatal {
		l.logMessage(T.StrFatal, l.host, l.svc, fmt.Sprintf("%s: %s: %s", getLineNumber(), err.Error(), debug.Stack()))
		if l.batchTime != 0 {
			l.writeBatch()
		}
		os.Exit(1)
	}
}

func getLineNumber() string {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return ""
}

func (l *LogFprintf) logMetrics(msg string) {
	l.logMessage("METRIC", l.host, l.svc, msg)
}

func (l *LogFprintf) medianBucket(h *metrics.Float64Histogram) float64 {
	total := uint64(0)
	for _, count := range h.Counts {
		total += count
	}
	thresh := total / 2
	total = 0
	for i, count := range h.Counts {
		total += count
		if total >= thresh {
			return h.Buckets[i]
		}
	}
	panic("log module: should not happen")
}

func (l *LogFprintf) getMetrics() string {
	descs := metrics.All()
	samples := make([]metrics.Sample, len(descs))
	for i := range samples {
		samples[i].Name = descs[i].Name
	}
	metrics.Read(samples)
	var metr strings.Builder
	metr.WriteString("\n")
	for _, sample := range samples {
		name, value := sample.Name, sample.Value
		switch value.Kind() {
		case metrics.KindUint64:
			fmt.Fprintf(&metr, "%s: %d\n", name, value.Uint64())
		case metrics.KindFloat64:
			fmt.Fprintf(&metr, "%s: %f\n", name, value.Float64())
		case metrics.KindFloat64Histogram:
			fmt.Fprintf(&metr, "%s: %f\n", name, l.medianBucket(value.Float64Histogram()))
		case metrics.KindBad:
			l.LogPanic(fmt.Errorf("%s", "bad value kind in log module using runtime/metrics package!"))
		default:
			fmt.Fprintf(&metr, "%s: unexpected metric Kind: %v\n", name, value.Kind())
		}
	}
	return metr.String()
}

func ASSERT_(cond bool, msg string) {
	if !cond {
		fmt.Fprintf(os.Stdout, "\n[ASSERT] %s %s\n\n", getLineNumber(), msg)
	}
}

func ASSERT(cond bool, msg string) {
	ASSERT_(cond, msg)
	os.Exit(1)
}

func TODO(msg string) {
	fmt.Fprintf(os.Stdout, "\n[TODO] %s %s\n\n", getLineNumber(), msg)
}

func UNREACHABLE(msg string) {
	fmt.Fprintf(os.Stdout, "\n[UNREACHABLE] %s %s\n\n", getLineNumber(), msg)
}
