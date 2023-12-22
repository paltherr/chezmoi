package chezmoi

import (
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
	"golang.org/x/exp/slog"
)

// A DebugPersistentState logs calls to a PersistentState.
type DebugPersistentState struct {
	logger          *slog.Logger
	persistentState PersistentState
}

// NewDebugPersistentState returns a new debugPersistentState that logs methods
// on persistentState to logger.
func NewDebugPersistentState(
	persistentState PersistentState,
	logger *slog.Logger,
) *DebugPersistentState {
	return &DebugPersistentState{
		logger:          logger,
		persistentState: persistentState,
	}
}

// Close implements PersistentState.Close.
func (s *DebugPersistentState) Close() error {
	err := s.persistentState.Close()
	chezmoilog.InfoOrError(s.logger, "Close", err)
	return err
}

// CopyTo implements PersistentState.CopyTo.
func (s *DebugPersistentState) CopyTo(p PersistentState) error {
	err := s.persistentState.CopyTo(p)
	chezmoilog.InfoOrError(s.logger, "CopyTo", err)
	return err
}

// Data implements PersistentState.Data.
func (s *DebugPersistentState) Data() (any, error) {
	data, err := s.persistentState.Data()
	chezmoilog.InfoOrError(s.logger, "Data", err, "data", data)
	return data, err
}

// Delete implements PersistentState.Delete.
func (s *DebugPersistentState) Delete(bucket, key []byte) error {
	err := s.persistentState.Delete(bucket, key)
	chezmoilog.InfoOrError(s.logger, "Delete", err, "bucket", bucket, "key", key)
	return err
}

// DeleteBucket implements PersistentState.DeleteBucket.
func (s *DebugPersistentState) DeleteBucket(bucket []byte) error {
	err := s.persistentState.DeleteBucket(bucket)
	chezmoilog.InfoOrError(s.logger, "DeleteBucket", err, "bucket", bucket)
	return err
}

// ForEach implements PersistentState.ForEach.
func (s *DebugPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	err := s.persistentState.ForEach(bucket, func(k, v []byte) error {
		err := fn(k, v)
		chezmoilog.InfoOrError(s.logger, "ForEach", err, "bucket", bucket, "key", k, "value", v)
		return err
	})
	chezmoilog.InfoOrError(s.logger, "ForEach", err, "bucket", bucket)
	return err
}

// Get implements PersistentState.Get.
func (s *DebugPersistentState) Get(bucket, key []byte) ([]byte, error) {
	value, err := s.persistentState.Get(bucket, key)
	chezmoilog.InfoOrError(s.logger, "Get", err, "bucket", bucket, "key", key, "value", value)
	return value, err
}

// Set implements PersistentState.Set.
func (s *DebugPersistentState) Set(bucket, key, value []byte) error {
	err := s.persistentState.Set(bucket, key, value)
	chezmoilog.InfoOrError(s.logger, "Set", err, "bucket", bucket, "key", key, "value", value)
	return err
}
