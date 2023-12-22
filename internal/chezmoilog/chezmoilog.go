// Package chezmoilog contains support for chezmoi logging.
package chezmoilog

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

const few = 64

// An OSExecCmdLogObject wraps an *os/exec.Cmd and adds
// golang.org/x/exp/slog.LogValuer functionality.
type OSExecCmdLogObject struct {
	*exec.Cmd
}

// An OSExecExitErrorLogObject wraps an *os/exec.ExitError and adds
// golang.org/x/exp/slog.LogValuer.
type OSExecExitErrorLogObject struct {
	*exec.ExitError
}

// An OSProcessStateLogObject wraps an *os.ProcessState and adds
// golang.org/x/exp/slog.LogValuer functionality.
type OSProcessStateLogObject struct {
	*os.ProcessState
}

// LogValue implements golang.org/x/exp/slog.LogValuer.
func (cmd OSExecCmdLogObject) LogValuer() slog.Value {
	var attrs []slog.Attr
	if cmd.Path != "" {
		attrs = append(attrs, slog.String("path", cmd.Path))
	}
	if len(cmd.Args) != 0 {
		attrs = append(attrs, slog.Any("args", cmd.Args))
	}
	if cmd.Dir != "" {
		attrs = append(attrs, slog.String("dir", cmd.Dir))
	}
	if len(cmd.Env) != 0 {
		attrs = append(attrs, slog.Any("env", cmd.Env))
	}
	return slog.GroupValue(attrs...)
}

// LogValue implements golang.org/x/exp/slog.LogValuer.
func (err OSExecExitErrorLogObject) LogValuer() slog.Value {
	attrs := []slog.Attr{
		slog.Any("processState", OSProcessStateLogObject{err.ExitError.ProcessState}),
	}
	if osExecExitError := (&exec.ExitError{}); errors.As(err, &osExecExitError) {
		attrs = append(attrs, slog.String("stderr", string(err.ExitError.Stderr)))
	}
	return slog.GroupValue(attrs...)
}

// LogValue implements golang.org/x/exp/slog.LogValuer.
func (p OSProcessStateLogObject) LogValue() slog.Value {
	var attrs []slog.Attr
	if p.ProcessState != nil {
		if p.Exited() {
			if !p.Success() {
				attrs = append(attrs, slog.Int("exitCode", p.ExitCode()))
			}
		} else {
			attrs = append(attrs, slog.Int("pid", p.Pid()))
		}
		if userTime := p.UserTime(); userTime != 0 {
			attrs = append(attrs, slog.Duration("userTime", userTime))
		}
		if systemTime := p.SystemTime(); systemTime != 0 {
			attrs = append(attrs, slog.Duration("systemTime", systemTime))
		}
	}
	return slog.GroupValue(attrs...)
}

// FirstFewBytes returns the first few bytes of data in a human-readable form.
func FirstFewBytes(data []byte) []byte {
	if len(data) > few {
		data = slices.Clone(data[:few])
		data = append(data, '.', '.', '.')
	}
	return data
}

// LogHTTPRequest calls httpClient.Do, logs the result to logger, and returns
// the result.
func LogHTTPRequest(
	logger *slog.Logger,
	client *http.Client,
	req *http.Request,
) (*http.Response, error) {
	start := time.Now()
	resp, err := client.Do(req)
	args := []any{
		slog.Duration("duration", time.Since(start)),
		slog.String("method", req.Method),
		Stringer("url", req.URL),
	}
	if resp != nil {
		args = append(args,
			slog.Int("contentLength", int(resp.ContentLength)),
			slog.String("status", resp.Status),
			slog.Int("statusCode", resp.StatusCode),
		)
	}
	if err != nil {
		args = append(args, slog.Any("err", err))
		logger.Error("HTTPRequest", args...)
	} else {
		logger.Info("HTTPRequest", args...)
	}
	return resp, err
}

// LogCmdCombinedOutput calls cmd.CombinedOutput, logs the result, and returns the result.
func LogCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	start := time.Now()
	combinedOutput, err := cmd.CombinedOutput()
	attrs := []any{
		slog.Any("cmd", OSExecCmdLogObject{Cmd: cmd}),
		slog.Duration("duration", time.Since(start)),
		slog.Any("combinedOutput", Output(combinedOutput, err)),
		slog.Int("size", len(combinedOutput)),
	}
	for _, attr := range AppendExitErrorAttrs(nil, err) {
		attrs = append(attrs, attr)
	}
	if err != nil {
		slog.Error("Output", attrs...)
	} else {
		slog.Info("Output", attrs...)
	}
	return combinedOutput, err
}

// LogCmdOutput calls cmd.Output, logs the result, and returns the result.
func LogCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	start := time.Now()
	output, err := cmd.Output()
	attrs := []any{
		slog.Any("cmd", OSExecCmdLogObject{Cmd: cmd}),
		slog.Duration("duration", time.Since(start)),
		slog.Any("output", Output(output, err)),
		slog.Int("size", len(output)),
	}
	for _, attr := range AppendExitErrorAttrs(nil, err) {
		attrs = append(attrs, attr)
	}
	if err != nil {
		slog.Error("Output", attrs...)
	} else {
		slog.Info("Output", attrs...)
	}
	return output, err
}

// LogCmdRun calls cmd.Run, logs the result, and returns the result.
func LogCmdRun(cmd *exec.Cmd) error {
	start := time.Now()
	err := cmd.Run()
	attrs := []any{
		slog.Any("cmd", OSExecCmdLogObject{Cmd: cmd}),
		slog.Duration("duration", time.Since(start)),
	}
	for _, attr := range AppendExitErrorAttrs(nil, err) {
		attrs = append(attrs, attr)
	}
	if err != nil {
		slog.Error("Run", attrs...)
	} else {
		slog.Info("Run", attrs...)
	}
	return err
}

// LogCmdStart calls cmd.Start, logs the result, and returns the result.
func LogCmdStart(cmd *exec.Cmd) error {
	start := time.Now()
	err := cmd.Start()
	attrs := []any{
		slog.Any("cmd", OSExecCmdLogObject{Cmd: cmd}),
		slog.Time("start", start),
	}
	for _, attr := range AppendExitErrorAttrs(nil, err) {
		attrs = append(attrs, attr)
	}
	if err != nil {
		slog.Error("Start", attrs...)
	} else {
		slog.Info("Start", attrs...)
	}
	return err
}

// LogCmdWait calls cmd.Wait, logs the result, and returns the result.
func LogCmdWait(cmd *exec.Cmd) error {
	err := cmd.Wait()
	end := time.Now()
	attrs := []any{
		slog.Any("cmd", OSExecCmdLogObject{Cmd: cmd}),
		slog.Time("end", end),
	}
	for _, attr := range AppendExitErrorAttrs(nil, err) {
		attrs = append(attrs, attr)
	}
	if err != nil {
		slog.Error("Wait", attrs...)
	} else {
		slog.Info("Wait", attrs...)
	}
	return err
}

// Output returns the first few bytes of output if err is nil, otherwise it
// returns the full output.
func Output(data []byte, err error) []byte {
	if err != nil {
		return data
	}
	return FirstFewBytes(data)
}

func InfoOrError(logger *slog.Logger, msg string, err error, args ...any) {
	switch {
	case logger == nil:
		return
	case err != nil:
		logger.Error(msg, append([]any{"err", err}, args...)...)
	default:
		logger.Info(msg, args...)
	}
}

// FIXME this should use []any
func AppendExitErrorAttrs(attrs []slog.Attr, err error) []slog.Attr {
	var execExitError *exec.ExitError
	if !errors.As(err, &execExitError) {
		return append(attrs, slog.Any("err", err))
	}

	if execExitError.ProcessState != nil {
		if execExitError.Exited() {
			attrs = append(attrs, slog.Int("exitCode", execExitError.ExitCode()))
		} else {
			attrs = append(attrs, slog.Int("pid", execExitError.Pid()))
		}
		if userTime := execExitError.UserTime(); userTime != 0 {
			attrs = append(attrs, slog.Duration("userTime", userTime))
		}
		if systemTime := execExitError.SystemTime(); systemTime != 0 {
			attrs = append(attrs, slog.Duration("systemTime", systemTime))
		}
	}

	return attrs
}

func Stringer(key string, s fmt.Stringer) slog.Attr {
	return slog.String(key, s.String())
}
