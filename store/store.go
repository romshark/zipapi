package store

import "time"

// UploadInfo represents information about an uploader of a file
type UploadInfo struct {
	Time        time.Time
	ClientAgent string
}

// File represents an uploaded file
type File struct {
	Upload   UploadInfo
	Name     string
	Contents []byte
}

// Store represents an abstract store
type Store interface {
	// Init initializes the store
	Init() error

	// SaveFiles saves the given files to the store
	//
	// This method is thread-safe and can safely be used by
	// multiple goroutines concurrently
	SaveFiles(files ...File) error
}
