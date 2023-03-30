package nodes

import (
	"github.com/dsxack/gitfs/internal/testdata"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestTagsSegmentNode_Readdir(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	expectedDirs := []string{
		"v1.0.2",
	}

	entries, err := os.ReadDir(filepath.Join(mountPoint, "tags/nested"))
	require.NoError(t, err)
	require.Len(t, entries, len(expectedDirs))

	for _, e := range entries {
		require.Contains(t, expectedDirs, e.Name())
	}
}
