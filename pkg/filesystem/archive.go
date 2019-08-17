package filesystem

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateTarball takes a source directory and walks the directory recursively,
// constructing a tarball stored in `targetDir`. It returns `tarpath`, the
// full path to the created tarfile and `err` if the archival fails.
func CreateTarball(src, targetDir string) (tarpath string, err error) {
	// Get file info for the source file.
	srcInfo, err := os.Stat(src)
	if err != nil {
		return tarpath, err
	}

	// Create a tar file to write to.
	tarpath = filepath.Join(targetDir, fmt.Sprintf("%s.tar", srcInfo.Name()))
	tarfile, err := os.Create(tarpath)
	if err != nil {
		return tarpath, err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	var baseDir string
	if srcInfo.IsDir() {
		baseDir = filepath.Base(src)
	}

	return tarpath, filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, src))
		}

		if err := tarball.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(tarball, file)
		return err
	})
}

// ExtractTarball takes a reader containing a tarball and attempts to extract
// it into `targetDir`.
func ExtractTarball(reader io.Reader, targetDir string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(targetDir, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}
