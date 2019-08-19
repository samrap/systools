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
func CreateTarball(source, targetDir string) (tarPath string, err error) {
	// baseName is the base directory of `source`, which must be included in the
	// file name of the headers written to the tarball. For example, if the
	// source is directory `/foo/bar`, and `bar` contains several files
	// which will be put in the archive, their header's name field
	// must include their base directory, `bar`, such that the
	// header contains the entire path relative to `bar`.
	//
	// If `source` is not a directory, then `baseName` should be an empty string.
	var baseName string

	// Get file info for the source file.
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return tarPath, err
	}

	// We only need to specify the base directory if `source` is a directory.
	if sourceInfo.IsDir() {
		baseName = filepath.Base(source)
	}

	// Create a tar file to write to.
	tarPath = filepath.Join(targetDir, fmt.Sprintf("%s.tar", sourceInfo.Name()))
	tarfile, err := os.Create(tarPath)
	if err != nil {
		return tarPath, err
	}
	defer tarfile.Close()

	// Create a tarball to write to.
	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	return tarPath, filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		// link is the destination that path points to if it is a symbolic link.
		var link string

		if err != nil {
			return err
		}

		// If we are dealing with a symbolic link, then we need to get the path
		// that it points to in order to write a valid header for this file.
		if info.Mode()&os.ModeSymlink != 0 {
			if link, err = os.Readlink(path); err != nil {
				return err
			}
		}

		// Create the header.
		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}

		// We need to include the base name of source if it is a directory.
		if baseName != "" {
			header.Name = filepath.Join(baseName, strings.TrimPrefix(path, source))
		}

		// Write the header.
		if err := tarball.WriteHeader(header); err != nil {
			return err
		}

		// If we are dealing with a non-regular file, our work is done here.
		if !info.Mode().IsRegular() {
			return nil
		}

		// For regular files, we will copy their contents into the tarball.
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(tarball, file)
		return err
	})
}

// ExtractTarball takes a reader containing tarball bytes and attempts to extract
// the archive into `targetDir`.
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
			// If the file is a directory, we will simply create the directory
			// with the original permissions.
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		} else if info.Mode()&os.ModeSymlink != 0 {
			// If the file is a symlink, we will simply create the link using
			// the link name stored in the header.
			if err = os.Symlink(header.Linkname, path); err != nil {
				return err
			}
			continue
		}

		// Finally, if we are dealing with a regular file we will copy the
		// contents from the tar reader into a newly created file.
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
