package nodes

import (
	"github.com/dsxack/gitfs/internal/testdata"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestLookup(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	for path, expected := range testdata.ExpectedFiles {
		t.Run(path, func(t *testing.T) {
			actual, _ := os.ReadFile(filepath.Join(mountPoint, path))
			require.Equal(t, string(expected.Data), string(actual))
		})
	}
}

func TestReaddir(t *testing.T) {
	mountPoint := testdata.Initialize(t, NewRootNode)

	require.NoError(t, fs.WalkDir(testdata.ExpectedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			expectedEntries, err := fs.ReadDir(testdata.ExpectedFiles, path)
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
