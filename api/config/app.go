package config

// App represents the application configurations
type App struct {
	// MaxReqSize defines the maximum request size in bytes
	MaxReqSize uint64

	// MaxFileSize defines the maximum file size in bytes
	MaxFileSize uint64

	// MaxMultipartMembuf defines the maximum multipart/form-data memory buffer
	MaxMultipartMembuf uint64
}
