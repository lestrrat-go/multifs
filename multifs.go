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
	mu       sync.RWMutex
	mounts   []*mountPoint
	prefixes map[string]struct{} // used solely for deduping
}

// New creates an empty multifs.FS object. You will need to call Mount()
// to add other filesystems
func New() *FS {
	return &FS{}
}

type mountPoint struct {
	prefix string
	fs     fs.FS
}

func (fs *FS) initNoLock() {
	if fs.prefixes == nil {
		fs.prefixes = make(map[string]struct{})
	}
}

// Mount associates prefix with another fs.FS. For example
// if you mount "/foo" with an fs.FS that contains files
// with names such as "/bar/baz.txt", you would be able to
// access them via "/foo/bar/baz.txt".
//
// Mount currently only understands linux-style paths (technically
// it uses "path" package).
func (fs *FS) Mount(prefix string, other fs.FS) error {
	// The prefix must be normalized.
	prefix = path.Clean(prefix)
	if !strings.HasPrefix(prefix, "/") {
		return fmt.Errorf(`invalid prefix (path was normalized to %q)`, prefix)
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.initNoLock()

	if _, ok := fs.prefixes[prefix]; ok {
		return fmt.Errorf(`prefix %q has already been mounted`, prefix)
	}

	mounts := append(fs.mounts, &mountPoint{
		prefix: prefix,
		fs:     other,
	})

	// TODO: Yeah... obviously we can optimize this so that we don't
	// have to sort it every time. Patches welcome
	sort.Slice(mounts, func(i, j int) bool {
		return mounts[i].prefix < mounts[j].prefix
	})

	fs.mounts = mounts
	fs.prefixes[prefix] = struct{}{}
	return nil
}

func (fs *FS) Open(name string) (fs.File, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	name = path.Clean(name)
	for _, mount := range fs.mounts {
		if !strings.HasPrefix(name, mount.prefix) {
			continue
		}

		return mount.fs.Open(strings.TrimPrefix(name, mount.prefix+"/"))
	}
	return nil, fmt.Errorf(`file %q was not found`, name)
}
