package nodes

import (
	"github.com/dsxack/gitfs/internal/testdata"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestCommitsNode_Readdir(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	expectedDirs := []string{
		"3991e5a92b70e6a4e91ce48d2165a92b8b056cdd",
		"3d97e28b20f8babc2182f2e67ba7d51397dd0ff5",
		"90309e427770fdea9e8a63b09069f2c51df7134b",
	}

	entries, err := os.ReadDir(filepath.Join(mountPoint, "commits"))
	require.NoError(t, err)
	require.Len(t, entries, len(expectedDirs))

	for _, e := range entries {
		require.Contains(t, expectedDirs, e.Name())
	}
}

func TestCommitsNode_Lookup(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	expectedFiles := map[string]map[string]string{
		"commits/3991e5a92b70e6a4e91ce48d2165a92b8b056cdd": {
			"testfile1": "testfile1 content\n",
		},
		"commits/3d97e28b20f8babc2182f2e67ba7d51397dd0ff5": {
			"testfile1": "testfile1 content\n",
			"testfile2": "testfile2 content\n",
		},
		"commits/90309e427770fdea9e8a63b09069f2c51df7134b": {
			"testfile1": "testfile1 content\n",
			"testfile2": "testfile2 content\n",
			"testfile3": "testfile3 content\n",
		},
	}

	for dir, files := range expectedFiles {
		entries, err := os.ReadDir(filepath.Join(mountPoint, dir))
		require.NoError(t, err)
		require.Len(t, entries, len(files))

		for _, e := range entries {
			require.Contains(t, files, e.Name())
			content, _ := os.ReadFile(filepath.Join(mountPoint, dir, e.Name()))
			require.Equal(t, files[e.Name()], string(content))
		}
	}
}
