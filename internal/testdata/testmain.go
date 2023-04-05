package testdata

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"github.com/go-git/go-git/v5"
	"github.com/hanwen/go-fuse/v2/fs"
	"io"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testrepo.zip
var testRepositoryZip []byte

type rootNodeFactory[T fs.InodeEmbedder] func(*git.Repository) T

func Initialize[T fs.InodeEmbedder](t *testing.T, factory rootNodeFactory[T]) string {
	t.Helper()
	repoPath, err := unzipTestRepository(t, testRepositoryZip)
	if err != nil {
		t.Fatalf("failed to unzip test repository: %v", err)
	}
	mountPoint, err := createMountPoint(t)
	if err != nil {
		t.Fatalf("failed to create mount point: %v", err)
	}
	err = mount(t, repoPath, mountPoint, factory)
	if err != nil {
		t.Fatalf("failed to mount filesystem: %v", err)
	}
	return mountPoint
}

func mount[T fs.InodeEmbedder](t *testing.T, repoPath string, mountPoint string, factory rootNodeFactory[T]) error {
	repository, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}
	server, err := fs.Mount(mountPoint, factory(repository), &fs.Options{})
	if err != nil {
		return err
	}
	t.Cleanup(func() {
		err := server.Unmount()
		if err != nil {
			t.Errorf("failed to unmount filesystem: %v", err)
		}
	})
	return err
}

func unzipTestRepository(t *testing.T, testfile []byte) (string, error) {
	destination, err := os.MkdirTemp("", "gitfs")
	if err != nil {
		return "", err
	}
	t.Cleanup(func() {
		err = os.RemoveAll(destination)
		if err != nil {
			t.Errorf("failed to remove test repository directory: %v", err)
		}
	})
	reader, err := zip.NewReader(bytes.NewReader(testfile), int64(len(testfile)))
	if err != nil {
		return "", err
	}
	for _, f := range reader.File {
		err := unzipFile(f, destination)
		if err != nil {
			return "", err
		}
	}

	return destination, nil
}

func unzipFile(f *zip.File, destination string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	fpath := filepath.Join(destination, f.Name)
	if f.FileInfo().IsDir() {
		os.MkdirAll(fpath, os.ModePerm)
	} else {
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}
	}
	return nil
}

func createMountPoint(t *testing.T) (string, error) {
	mountPoint, err := os.MkdirTemp("", "test")
	if err != nil {
		return "", err
	}
	t.Cleanup(func() {
		err := os.RemoveAll(mountPoint)
		if err != nil {
			t.Errorf("failed to remove mount point: %v", err)
		}
	})
	return mountPoint, nil
}
