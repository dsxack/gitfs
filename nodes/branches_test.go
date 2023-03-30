package nodes

import (
	"github.com/dsxack/gitfs/internal/testdata"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestBranchesNode_Readdir(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	expectedDirs := []string{
		"master",
		"nested",
		"test",
	}

	entries, err := os.ReadDir(filepath.Join(mountPoint, "branches"))
	require.NoError(t, err)
	require.Len(t, entries, len(expectedDirs))

	for _, e := range entries {
		require.Contains(t, expectedDirs, e.Name())
	}
}

func TestBranchesNode_Lookup(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	expectedFiles := map[string]map[string]string{
		"branches/test": {
			"testfile1": "testfile1 content\n",
		},
		"branches/master": {
			"testfile1": "testfile1 content\n",
			"testfile2": "testfile2 content\n",
		},
		"branches/nested/test": {
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
