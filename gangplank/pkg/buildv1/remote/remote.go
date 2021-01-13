package remote

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/coreos/gangplank/pkg/buildv1/minio"
	"github.com/coreos/gangplank/pkg/cosa"
)

// File is an object to fetch from a remote server
type File struct {
	Bucket     string         `json:"bucket,omitempty"`
	Object     string         `json:"object,omitempty"`
	Minio      *minio.Server  `json:"remote,omitempty"`
	Compressed bool           `json:"comptempty"`
	Artifact   *cosa.Artifact `json:"artifact,omitempty"`
}

// WriteToPath fetches the remote file and writes it locally.
func (r *File) WriteToPath(ctx context.Context, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	dest, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0755)
	if err != nil {
		return err
	}
	defer dest.Close()

	err = r.Minio.Fetcher(ctx, r.Bucket, r.Object, dest)
	return err
}

// Extract decompresses the remote file to the path
func (r *File) Extract(ctx context.Context, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmpf, err := ioutil.TempFile("", "obj")
	if err != nil {
		return err
	}
	defer os.Remove(tmpf.Name())
	defer tmpf.Close()
	if err := r.Minio.Fetcher(ctx, r.Bucket, r.Object, tmpf); err != nil {
		return err
	}
	// sync and then seek back to zero
	if err := tmpf.Sync(); err != nil {
		return fmt.Errorf("oof, unable to sync the file...this is bad: %w", err)
	}
	if _, err := tmpf.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("double oof, srly? %w", err)
	}
	return Decompress(tmpf, path)
}

// Decompress takes an open file and extracts its to directory.
func Decompress(in *os.File, dir string) error {
	log.Info("Receiving binary source from STDIN as archive ...")
	args := []string{"-x", "-v", "-o", "-m", "-f", "-", "-C", dir}
	cmd := exec.Command("bsdtar", args...)
	cmd.Stdin = in
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Info("Extracting...")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("unable to extract binary build input, must be a zip, tar, or gzipped tar, or specified as a file: %v", err)
	}
	return nil
}
