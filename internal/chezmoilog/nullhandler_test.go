package chezmoilog

import "golang.org/x/exp/slog"

var _ slog.Handler = NullHandler{}
