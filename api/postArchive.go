package api

import (
	"archive/zip"
	"compress/flate"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"zipapi/store"

	"github.com/pkg/errors"
)

func (srv *server) postArchive(
	out http.ResponseWriter,
	in *http.Request,
) error {
	startTime := time.Now()
	userAgent := in.Header.Get("User-Agent")

	// Limit total form size to 32 mb
	in.Body = http.MaxBytesReader(out, in.Body, 1024*1024*32)

	// Parse inputs
	if err := in.ParseMultipartForm(1024 * 1024 * 8); err != nil {
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

	for flName, fl := range in.MultipartForm.File {
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
