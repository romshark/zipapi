package mock

import (
	"github.com/romshark/zipapi/store"
)

// Store represents an in-memory store mock-implementation
type Store struct {
	SavedFiles []store.File
}

// New creates a new store mock-implementation instance
func New() *Store {
	return &Store{
		SavedFiles: make([]store.File, 0),
	}
}

// Init implements the Store interface
func (str *Store) Init() error { return nil }

// SaveFiles implements the Store interface
func (str *Store) SaveFiles(files ...store.File) error {
	str.SavedFiles = append(str.SavedFiles, files...)
	return nil
}
