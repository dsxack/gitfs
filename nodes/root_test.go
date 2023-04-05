package nodes

import (
	"github.com/dsxack/gitfs/internal/testdata"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

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

var expectedFS = combineMapFS(
	prefixMapFS("branches/test/", commitFiles[commits[0]]),
	prefixMapFS("branches/master/", commitFiles[commits[1]]),
	prefixMapFS("branches/nested/dir/test/", commitFiles[commits[3]]),
	prefixMapFS("tags/v1.0.0/", commitFiles[commits[0]]),
	prefixMapFS("tags/v1.0.1/", commitFiles[commits[1]]),
	prefixMapFS("tags/nested/dir/test/", commitFiles[commits[3]]),
	prefixMapFS("commits/"+commits[0]+"/", commitFiles[commits[0]]),
	prefixMapFS("commits/"+commits[1]+"/", commitFiles[commits[1]]),
	prefixMapFS("commits/"+commits[2]+"/", commitFiles[commits[2]]),
	prefixMapFS("commits/"+commits[3]+"/", commitFiles[commits[3]]),
)

func TestLookup(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	for path, expected := range expectedFS {
		t.Run(path, func(t *testing.T) {
			actual, _ := os.ReadFile(filepath.Join(mountPoint, path))
			require.Equal(t, string(expected.Data), string(actual))
		})
	}
}

func TestReaddir(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	require.NoError(t, fs.WalkDir(expectedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			expectedEntries, err := fs.ReadDir(expectedFS, path)
			require.NoError(t, err)
			expectedEntriesNames := dirEntriesNames(expectedEntries)

			actualEntries, err := os.ReadDir(filepath.Join(mountPoint, path))
			require.NoError(t, err)
			actualEntriesNames := dirEntriesNames(actualEntries)

			require.Len(t, actualEntriesNames, len(expectedEntriesNames))
			for _, entry := range actualEntriesNames {
				require.Contains(t, expectedEntriesNames, entry)
			}
		})

		return nil
	}))
}

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

type dirEntry interface {
	Name() string
}

func dirEntriesNames[T dirEntry](entries []T) []string {
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names
}
