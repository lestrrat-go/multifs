package multifs_test

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lestrrat-go/multifs"
	"github.com/stretchr/testify/require"
)

var _ fs.FS = &multifs.FS{}
var _ fs.ReadDirFS = &multifs.FS{}

func TestMultiFS(t *testing.T) {
	root, err := os.MkdirTemp("", "multifs-*")
	require.NoError(t, err, `os.MkdirTemp should succeed`)

	files := []struct {
		Path    string
		Content string
	}{
		{
			Path:    "foo/1.txt",
			Content: strings.Repeat("1", 100),
		},
		{
			Path:    "foo/2.txt",
			Content: strings.Repeat("2", 100),
		},
		{
			Path:    "bar/a.txt",
			Content: strings.Repeat("a", 100),
		},
		{
			Path:    "bar/b.txt",
			Content: strings.Repeat("b", 100),
		},
	}

	for _, file := range files {
		path := filepath.Join(root, file.Path)
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); err != nil {
			require.NoError(t, os.MkdirAll(dir, 0755), `creating directory %s should succeed`, dir)
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		require.NoError(t, err, `os.OpenFile should succeed`)

		io.WriteString(f, file.Content)

		_ = f.Close()
	}

	foo := os.DirFS(filepath.Join(root, "foo"))
	bar := os.DirFS(filepath.Join(root, "bar"))

	var mfs multifs.FS
	require.NoError(t, mfs.Mount("/quux", foo), `mfs.Mount(/quux) should succeed`)
	require.NoError(t, mfs.Mount("/corge", bar), `mfs.Mount(/corge) should succeed`)

	paths := make(map[string]struct{})
	for _, file := range files {
		file := file
		var path string
		if strings.HasPrefix(file.Path, "foo/") {
			path = "/quux/" + strings.TrimPrefix(file.Path, "foo/")
		} else {
			path = "/corge/" + strings.TrimPrefix(file.Path, "bar/")
		}

		paths[path] = struct{}{}
		t.Run(fmt.Sprintf("Open %q", path), func(t *testing.T) {
			f, err := mfs.Open(path)
			if f != nil {
				defer f.Close()
			}
			require.NoError(t, err, `fs.Open should succeed`)
		})
	}

	fs.WalkDir(&mfs, ".", func(name string, d fs.DirEntry, err error) error {
		delete(paths, "/"+name)
		return nil
	})
	require.Len(t, paths, 0, `paths should be empty`)

	require.Error(t, mfs.Unmount("/non-existent"), `fs.Unmoun(/non-existent) should fail`)
	require.NoError(t, mfs.Unmount("/corge"), `fs.Unmount(/corge) should succeed`)
	require.NoError(t, mfs.Unmount("/quux"), `fs.Unmount(/quux) should succeed`)
	require.Error(t, mfs.Unmount("/corge"), `fs.Unmount(/corge) a second time should fail`)
}
