package nodes

import (
	"github.com/dsxack/gitfs/internal/testdata"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestBranchSegmentNode_Readdir(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	expectedDirs := []string{
		"test",
	}

	entries, err := os.ReadDir(filepath.Join(mountPoint, "branches/nested"))
	require.NoError(t, err)
	require.Len(t, entries, len(expectedDirs))

	for _, e := range entries {
		require.Contains(t, expectedDirs, e.Name())
	}
}
