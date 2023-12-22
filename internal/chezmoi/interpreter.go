package chezmoi

import (
	"os/exec"

	"golang.org/x/exp/slog"
)

// An Interpreter interprets scripts.
type Interpreter struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}

// ExecCommand returns the *exec.Cmd to interpret name.
func (i *Interpreter) ExecCommand(name string) *exec.Cmd {
	if i.None() {
		return exec.Command(name)
	}
	return exec.Command(i.Command, append(i.Args, name)...) //nolint:gosec
}

// None returns if i represents no interpreter.
func (i *Interpreter) None() bool {
	return i == nil || i.Command == ""
}

// LogValue implements golang.org/x/exp/slog.LogValuer.
func (i *Interpreter) LogValue() slog.Value {
	var attrs []slog.Attr
	if i.Command != "" {
		attrs = append(attrs, slog.String("command", i.Command))
	}
	if i.Args != nil {
		attrs = append(attrs, slog.Any("args", i.Args))
	}
	return slog.GroupValue(attrs...)
}
