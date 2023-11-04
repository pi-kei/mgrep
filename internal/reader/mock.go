package reader

import (
	"errors"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/pi-kei/mgrep/internal/base"
)

// This structure is not preventing you from a file that have children entries.
// Just keep it in mind and try to avoid it.
// Key is a single path. Parts of a path separeted by /.
// Do not put / at the start or end of a path.
// Do not put multiple / in a row.
type MockEntries = map[string]MockEntry

type MockEntry struct{
	ModTime time.Time   // modification time
	Content *string     // files must have this not equal to nil, dirs must have this equal to nil
	Err error           // error reading this entry or nil
	children []string   // if present then it means it has precalculated children. for dirs only
}

type mockReader struct {
	entries MockEntries
}

func NewMockReader(entries MockEntries) base.Reader {
	return &mockReader{entries}
}

func (r *mockReader) OpenFile(fileEntry base.DirEntry) (io.ReadCloser, error) {
	entry, ok := r.entries[fileEntry.Path]
	if !ok {
		return nil, errors.New("path does not exist")
	}
	if entry.Err != nil {
		return nil, entry.Err
	}
	if entry.Content == nil {
		return nil, errors.New("path is not a file")
	}
	return io.NopCloser(strings.NewReader(*entry.Content)), nil
}

func (r *mockReader) ReadDir(dirEntry base.DirEntry) (base.Iterator[base.DirEntry], error) {
	entry, ok := r.entries[dirEntry.Path]
	if !ok {
		return nil, errors.New("path does not exist")
	}
	if entry.Err != nil {
		return nil, entry.Err
	}
	if entry.Content != nil {
		return nil, errors.New("path is not a directory")
	}
	if entry.children != nil {
		return newMockIterator(r.entries, entry.children, dirEntry.Depth + 1), nil
	}
	children := []string{}
	for path := range r.entries {
		if len(path) > len(dirEntry.Path) && strings.HasPrefix(path, dirEntry.Path) && strings.LastIndex(path, "/") == len(dirEntry.Path) {
			children = append(children, path)
		}
	}
	slices.Sort(children)
	return newMockIterator(r.entries, children, dirEntry.Depth + 1), nil
}

func (r *mockReader) ReadRootEntry(name string) (base.DirEntry, error) {
	entry, ok := r.entries[name]
	if !ok {
		return base.DirEntry{}, errors.New("path does not exist")
	}
	if entry.Err != nil {
		return base.DirEntry{}, entry.Err
	}
	var size int64
	if entry.Content != nil {
		size = int64(len(*entry.Content))
	}
	return base.DirEntry{Path: name, Depth: 0, IsDir: entry.Content == nil, Size: size, ModTime: entry.ModTime}, nil
}

type mockIterator struct {
	entries MockEntries
	children []string
	depth int
	position int
	value base.DirEntry
	err error
}

func newMockIterator(entries MockEntries, children []string, depth int) base.Iterator[base.DirEntry] {
	return &mockIterator{entries, children, depth, -1, base.DirEntry{}, nil}
}

func (i *mockIterator) Next() bool {
	if i.err != nil || i.children == nil || i.position >= len(i.children) - 1 {
		return false
	}
	i.position++
	path := i.children[i.position]
	entry := i.entries[path]
	i.err = entry.Err
	if i.err != nil {
		return false
	}
	var size int64
	if entry.Content != nil {
		size = int64(len(*entry.Content))
	}
	i.value = base.DirEntry{Path: path, Depth: i.depth, IsDir: entry.Content == nil, Size: size, ModTime: entry.ModTime}
	return true
}

func (i *mockIterator) Value() base.DirEntry {
	return i.value
}

func (i *mockIterator) Err() error {
	return nil
}