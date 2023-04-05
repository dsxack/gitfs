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
	"testing/fstest"
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

var commits = []string{
	"3991e5a92b70e6a4e91ce48d2165a92b8b056cdd",
	"3d97e28b20f8babc2182f2e67ba7d51397dd0ff5",
	"90309e427770fdea9e8a63b09069f2c51df7134b",
	"d41f14efa3c0423eb22040766e25cfef81dd9081",
}

var commitFiles = map[string]fstest.MapFS{
	commits[0]: {
		"testfile1": mapFile("testfile1 content\n"),
	},
	commits[1]: {
		"testfile1": mapFile("testfile1 content\n"),
		"testfile2": mapFile("testfile2 content\n"),
	},
	commits[2]: {
		"testfile1": mapFile("testfile1 content\n"),
		"testfile2": mapFile("testfile2 content\n"),
		"testfile3": mapFile("testfile3 content\n"),
	},
	commits[3]: {
		"testfile1":         mapFile("testfile1 content\n"),
		"testfile2":         mapFile("testfile2 content\n"),
		"testfile3":         mapFile("testfile3 content\n"),
		"testdir/testfile4": mapFile("content of testfile4\n"),
	},
}

var ExpectedFiles = combineMapFS(
	prefixMapFS("branches/test/", commitFiles[commits[0]]),
	prefixMapFS("branches/master/", commitFiles[commits[1]]),
	prefixMapFS("branches/nested/test/", commitFiles[commits[2]]),
	prefixMapFS("branches/nested-dir/", commitFiles[commits[3]]),
	prefixMapFS("tags/v1.0.0/", commitFiles[commits[0]]),
	prefixMapFS("tags/v1.0.1/", commitFiles[commits[1]]),
	prefixMapFS("tags/nested/v1.0.2/", commitFiles[commits[2]]),
	prefixMapFS("commits/"+commits[0]+"/", commitFiles[commits[0]]),
	prefixMapFS("commits/"+commits[1]+"/", commitFiles[commits[1]]),
	prefixMapFS("commits/"+commits[2]+"/", commitFiles[commits[2]]),
	prefixMapFS("commits/"+commits[3]+"/", commitFiles[commits[3]]),
)

func mapFile(content string) *fstest.MapFile {
	return &fstest.MapFile{
		Data: []byte(content),
		Mode: 0666,
	}
}

func prefixMapFS(prefix string, fs fstest.MapFS) fstest.MapFS {
	newFS := fstest.MapFS{}
	for k, v := range fs {
		newFS[prefix+k] = v
	}
	return newFS
}

func combineMapFS(fs ...fstest.MapFS) fstest.MapFS {
	newFS := fstest.MapFS{}
	for _, f := range fs {
		for k, v := range f {
			newFS[k] = v
		}
	}
	return newFS
}
