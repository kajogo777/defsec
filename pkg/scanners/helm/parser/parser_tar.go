package parser

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/liamg/memoryfs"

	"github.com/aquasecurity/defsec/pkg/detection"
)

func (p *Parser) addTarToFS(path string) (fs.FS, error) {

	var file io.ReadCloser
	var err error

	tarFS := memoryfs.CloneFS(p.workingFS)
	file, err = tarFS.Open(path)
	if err != nil {
		return nil, err
	}

	if detection.IsZip(path) {
		if file, err = gzip.NewReader(file); err != nil {
			return nil, err
		}
	}

	defer func() { _ = file.Close() }()

	tr := tar.NewReader(file)

	for {
		header, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		// get the individual path and extract to the current directory
		path := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			if err := tarFS.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return nil, err
			}
		case tar.TypeReg:
			p.debug.Log("Untarring %s", path)
			_ = tarFS.MkdirAll(filepath.Dir(path), fs.ModePerm)
			content := []byte{}
			writer := bytes.NewBuffer(content)

			if err != nil {
				return nil, err
			}
			for {
				_, err := io.CopyN(writer, tr, 1024)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return nil, err
				}
			}
			if err := tarFS.WriteFile(path, writer.Bytes(), fs.ModePerm); err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("could not untar the section")
		}
	}

	// force close the file for Windows so we can remove it from FS
	_ = file.Close()

	// remove the tarball from the fs
	if err := tarFS.Remove(path); err != nil {
		return nil, err
	}

	return tarFS, nil
}
