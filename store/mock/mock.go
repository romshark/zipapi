package mock

import (
	"sync"

	"github.com/romshark/zipapi/store"
)

// Store represents an in-memory store mock-implementation
type Store struct {
	lock       *sync.RWMutex
	savedFiles []store.File
}

// SavedFiles returns copies of all stored files
func (str *Store) SavedFiles() []store.File {
	str.lock.RLock()
	defer str.lock.RUnlock()

	cp := make([]store.File, len(str.savedFiles))
	for ix, fl := range str.savedFiles {
		flc := fl
		copy(flc.Contents, fl.Contents)
		cp[ix] = flc
	}
	return cp
}

// Init implements the Store interface
func (str *Store) Init() error {
	str.lock = &sync.RWMutex{}
	str.savedFiles = make([]store.File, 0)

	return nil
}

// SaveFiles implements the Store interface
func (str *Store) SaveFiles(files ...store.File) error {
	str.lock.Lock()
	str.savedFiles = append(str.savedFiles, files...)
	str.lock.Unlock()
	return nil
}
