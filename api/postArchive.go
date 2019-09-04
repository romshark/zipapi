package api

import (
	"archive/zip"
	"compress/flate"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"time"

	"github.com/romshark/zipapi/store"

	"github.com/pkg/errors"
)

func (srv *server) postArchive(
	out http.ResponseWriter,
	in *http.Request,
) error {
	// Make sure the content-type header is set
	contentTypeHeader := in.Header.Get("Content-Type")
	if contentTypeHeader == "" {
		http.Error(
			out,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
		return nil
	}

	// Validate content-type
	contentType, _, err := mime.ParseMediaType(contentTypeHeader)
	if err != nil {
		return errors.Wrap(err, "parsing content-type header")
	}
	if contentType != "multipart/form-data" {
		http.Error(
			out,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
		return nil
	}

	startTime := time.Now()
	userAgent := in.Header.Get("User-Agent")

	// Limit total form size
	in.Body = http.MaxBytesReader(
		out,
		in.Body,
		int64(srv.conf.App.MaxReqSize),
	)

	// Parse inputs
	if err := in.ParseMultipartForm(
		int64(srv.conf.App.MaxMultipartMembuf),
	); err != nil {
		// This is damn ugly, but there seems to be no way around
		// comparing the error string
		if err.Error() == "http: request body too large" {
			http.Error(
				out,
				"request body too large",
				http.StatusBadRequest,
			)
			return nil
		}
		return errors.Wrap(err, "parsing multipart/form-data")
	}

	// Init zip archive writer
	arch := zip.NewWriter(out)
	defer arch.Close()
	arch.RegisterCompressor(
		//TODO: make this compression optional
		zip.Deflate,
		func(out io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(out, flate.BestCompression)
		},
	)
	defer arch.Close()

	files := make([]store.File, 0, len(in.MultipartForm.File))

	if len(in.MultipartForm.File) < 1 {
		// Missing files
		http.Error(
			out,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
		return nil
	}

	for flName, fl := range in.MultipartForm.File {
		// Check file size
		if uint64(fl[0].Size) > srv.conf.App.MaxFileSize {
			http.Error(
				out,
				fmt.Sprintf(
					"file '%s' exceeds max file size (%d)",
					flName,
					srv.conf.App.MaxFileSize,
				),
				http.StatusBadRequest,
			)
			return nil
		}

		file, err := fl[0].Open()
		if err != nil {
			return errors.Wrapf(
				err,
				"opening file '%s' from multipart/form-data",
				flName,
			)
		}

		// Read file
		contents, err := ioutil.ReadAll(file)
		if err != nil {
			return errors.Wrapf(
				err,
				"reading file '%s' multipart/form-data",
				flName,
			)
		}

		files = append(files, store.File{
			Upload: store.UploadInfo{
				Time:        startTime,
				ClientAgent: userAgent,
			},
			Name:     flName,
			Contents: contents,
		})

		// Add file to archive
		fout, err := arch.Create(flName)
		if err != nil {
			return errors.Wrap(err, "creating archive file")
		}

		// Write file
		_, err = fout.Write(contents)
		if err != nil {
			return errors.Wrap(err, "copying multipart file to archive")
		}
	}

	// Save file to store
	if err := srv.store.SaveFiles(files...); err != nil {
		return errors.Wrap(err, "saving files to store")
	}

	return nil
}
