/* Package multifs provides a simple in-memory abstraction that allows
 * multiple fs.FS objects to act as if they are mounted under a common
 * filesystem.
 */

package multifs

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"
)

type FS struct {
	mu sync.RWMutex

	// mountPoints holds a sorted list of names so that we can
	// match paths from longest to shortest
	mountPoints []string
	fsmap       map[string]fs.FS
}

// New creates an empty multifs.FS object. You will need to call Mount()
// to add other filesystems
func New() *FS {
	return &FS{}
}

func (mfs *FS) initNoLock() {
	if mfs.fsmap == nil {
		mfs.fsmap = make(map[string]fs.FS)
	}
}

// Mount associates prefix with another fs.FS. For example
// if you mount "/foo" with an fs.FS that contains files
// with names such as "/bar/baz.txt", you would be able to
// access them via "/foo/bar/baz.txt".
//
// Mount currently only understands linux-style paths (technically
// it uses "path" package).
func (mfs *FS) Mount(prefix string, other fs.FS) error {
	// The prefix must be normalized.
	prefix = path.Clean(prefix)
	if !strings.HasPrefix(prefix, "/") {
		return fmt.Errorf(`invalid prefix (path was normalized to %q)`, prefix)
	}

	mfs.mu.Lock()
	defer mfs.mu.Unlock()

	mfs.initNoLock()

	if _, ok := mfs.fsmap[prefix]; ok {
		return fmt.Errorf(`prefix %q has already been mounted`, prefix)
	}

	mountPoints := append(mfs.mountPoints, prefix)

	// TODO: Yeah... obviously we can optimize this so that we don't
	// have to sort it every time. Patches welcome
	sort.Slice(mountPoints, func(i, j int) bool {
		// longest matches come first
		return len(mountPoints[i]) > len(mountPoints[j])
	})

	mfs.mountPoints = mountPoints
	mfs.fsmap[prefix] = other
	return nil
}

func (mfs *FS) Open(name string) (fs.File, error) {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()
	name = path.Clean(name)

	for _, prefix := range mfs.mountPoints {
		if !strings.HasPrefix(name, prefix+"/") {
			continue
		}

		src := mfs.fsmap[prefix]
		return src.Open(strings.TrimPrefix(name, prefix+"/"))
	}
	return nil, fmt.Errorf(`file %q was not found`, name)
}

func (mfs *FS) Unmount(prefix string) error {
	// The prefix must be normalized.
	prefix = path.Clean(prefix)
	if !strings.HasPrefix(prefix, "/") {
		return fmt.Errorf(`invalid prefix (path was normalized to %q)`, prefix)
	}

	mfs.mu.Lock()
	defer mfs.mu.Unlock()

	mfs.initNoLock()

	if _, ok := mfs.fsmap[prefix]; !ok {
		return fmt.Errorf(`prefix %q has not been mounted`, prefix)
	}

	for i, n := range mfs.mountPoints {
		if n != prefix {
			continue
		}

		// TODO: inefficient
		mfs.mountPoints = append(mfs.mountPoints[:i], mfs.mountPoints[i+1:]...)

		delete(mfs.fsmap, prefix)
		break
	}
	return nil
}

func (mfs *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()

	// emulation required for these
	if src, ok := mfs.fsmap[name]; ok {
		return fs.ReadDir(src, name)
	}

	for _, prefix := range mfs.mountPoints {
		if !strings.HasPrefix(name, prefix+"/") {
			continue
		}

		src := mfs.fsmap[prefix]
		return fs.ReadDir(src, name)
	}

	return nil, fmt.Errorf(`no such directory %q`, name)
}
