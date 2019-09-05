package apitest

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/romshark/zipapi/api/config"
	"github.com/romshark/zipapi/apitest/setup"
	"github.com/romshark/zipapi/store"
	mockstore "github.com/romshark/zipapi/store/mock"

	"github.com/stretchr/testify/require"
)

type File struct {
	Name     string
	Contents []byte
	Params   map[string]string
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(
	t *testing.T,
	files ...File,
) *http.Request {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for _, fl := range files {
		part, err := writer.CreateFormFile(fl.Name, fl.Name)
		require.NoError(t, err)
		_, err = io.Copy(part, bytes.NewBuffer(fl.Contents))
		require.NoError(t, err)

		for key, val := range fl.Params {
			require.NoError(t, writer.WriteField(key, val))
		}
	}

	require.NoError(t, writer.Close())

	req, err := http.NewRequest("POST", "", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func checkFiles(
	ts *setup.TestSetup,
	expectedFiles []File,
	actualArchive []byte,
) {
	t := ts.T()

	str := ts.APIServer().Store().(*mockstore.Store)

	findFile := func(fl File) *store.File {
		for _, savedFile := range str.SavedFiles {
			if savedFile.Name == fl.Name {
				return &savedFile
			}
		}
		return nil
	}

	// Make sure the files were saved to the store
	require.Len(t, str.SavedFiles, len(expectedFiles))
	for _, expectedFile := range expectedFiles {
		actualFile := findFile(expectedFile)
		require.NotNil(t, actualFile)
	}

	// Unarchive files
	archReader, err := zip.NewReader(
		bytes.NewReader(actualArchive),
		int64(len(actualArchive)),
	)
	require.NoError(t, err)

	findExpected := func(flName string) *File {
		for _, expected := range expectedFiles {
			if expected.Name == flName {
				return &expected
			}
		}
		return nil
	}

	for _, fl := range archReader.File {
		expected := findExpected(fl.Name)
		require.NotNil(t, expected)

		flReader, err := fl.Open()
		require.NoError(t, err)
		defer flReader.Close()

		bt, err := ioutil.ReadAll(flReader)
		require.NoError(t, err)

		require.Equal(t, expected.Contents, bt)
	}
}

// TestPostArchive tests POST /archive sending 2 .txt files
func TestPostArchive(t *testing.T) {
	ts := setup.New(t, nil)
	defer ts.Teardown()

	files := []File{
		File{
			Name:     "foo.txt",
			Contents: []byte("foo foo foo"),
		}, File{
			Name:     "bar.txt",
			Contents: []byte("bar bar bar bar"),
		},
	}

	// Prepare input files
	req := newfileUploadRequest(t, files...)

	// Make request
	req.URL.Path = "/archive"
	resp := ts.Guest().Do(req)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	actual, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	checkFiles(ts, files, actual)
}

// TestPostArchiveErr tests POST /archive errors
func TestPostArchiveErr(t *testing.T) {
	// NonMultipart
	t.Run("NonMultipart", func(t *testing.T) {
		mimeTypes := []string{
			"",
			"application/json",
			"text/plain",
			"image/jpeg",
		}

		for _, mimeType := range mimeTypes {
			t.Run(fmt.Sprintf("'%s'", mimeType), func(t *testing.T) {
				ts := setup.New(t, nil)
				defer ts.Teardown()

				// Prepare input files
				req, err := http.NewRequest("POST", "", nil)
				require.NoError(t, err)
				req.Header.Set("Content-Type", mimeType)
				require.Equal(t, mimeType, req.Header.Get("Content-Type"))

				// Make request
				req.URL.Path = "/archive"
				resp := ts.Guest().Do(req)

				require.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		}
	})

	// NoFiles tests sending an empty multipart/form-data request
	t.Run("NoFiles", func(t *testing.T) {
		ts := setup.New(t, nil)
		defer ts.Teardown()

		// Prepare input files
		req := newfileUploadRequest(t)

		// Make request
		req.URL.Path = "/archive"
		resp := ts.Guest().Do(req)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// FileTooBig tests sending a file exceeding the file-size limit
	t.Run("FileTooBig", func(t *testing.T) {
		conf := &config.Config{
			App: config.App{
				MaxFileSize:        1024,
				MaxReqSize:         2048,
				MaxMultipartMembuf: 2048,
			},
		}
		ts := setup.New(t, conf)
		defer ts.Teardown()

		// Prepare input files
		req := newfileUploadRequest(t, File{
			Name:     "toolarge.txt",
			Contents: make([]byte, conf.App.MaxFileSize+1),
		})

		// Make request
		req.URL.Path = "/archive"
		resp := ts.Guest().Do(req)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// ReqTooBig tests sending a request exceeding the req-size limit
	t.Run("ReqTooBig", func(t *testing.T) {
		conf := &config.Config{
			App: config.App{
				MaxFileSize:        1024,
				MaxReqSize:         2048,
				MaxMultipartMembuf: 2048,
			},
		}
		ts := setup.New(t, conf)
		defer ts.Teardown()

		// Prepare input files
		req := newfileUploadRequest(t, File{
			Name:     "first.txt",
			Contents: make([]byte, conf.App.MaxFileSize),
		}, File{
			Name:     "second.txt",
			Contents: make([]byte, conf.App.MaxFileSize),
		}, File{
			Name:     "third.txt",
			Contents: make([]byte, conf.App.MaxFileSize),
		})

		// Make request
		req.URL.Path = "/archive"
		resp := ts.Guest().Do(req)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
