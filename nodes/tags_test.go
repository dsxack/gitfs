package nodes

import (
	"github.com/dsxack/gitfs/internal/testdata"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestTagsNode_Readdir(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	expectedDirs := []string{
		"v1.0.0",
		"v1.0.1",
		"nested",
	}

	entries, err := os.ReadDir(filepath.Join(mountPoint, "tags"))
	require.NoError(t, err)
	require.Len(t, entries, len(expectedDirs))

	for _, e := range entries {
		require.Contains(t, expectedDirs, e.Name())
	}
}

func TestTagsNode_Lookup(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	expectedFiles := map[string]map[string]string{
		"tags/v1.0.0": {
			"testfile1": "testfile1 content\n",
		},
		"tags/v1.0.1": {
			"testfile1": "testfile1 content\n",
			"testfile2": "testfile2 content\n",
		},
		"tags/nested/v1.0.2": {
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
