package reader

import (
	"errors"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/pi-kei/mgrep/internal/base"
)

// This structure is not preventing you from a file that have children entries.
// Just keep it in mind and try to avoid it
type Entries = map[string]struct{
	ModTime time.Time   // modification time
	Content *string     // files must have this not equal to nil, dirs must have this equal to nil
	Err error           // error reading this entry or nil
}

type mockReader struct {
	entries Entries
}

func NewMockReader(entries Entries) base.Reader {
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
	children := []base.DirEntry{}
	for path, entry := range r.entries {
		if len(path) > len(dirEntry.Path) && strings.HasPrefix(path, dirEntry.Path) && strings.LastIndex(path, "/") == len(dirEntry.Path) {
			var size int64
			if entry.Content != nil {
				size = int64(len(*entry.Content))
			}
			children = append(children, base.DirEntry{Path: path, Depth: dirEntry.Depth + 1, IsDir: entry.Content == nil, Size: size, ModTime: entry.ModTime})
		}
	}
	sort.Sort(byPath(children))
	return newMockIterator(r.entries, children), nil
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
	entries Entries
	children []base.DirEntry
	position int
	value base.DirEntry
	err error
}

func newMockIterator(entries Entries, children []base.DirEntry) base.Iterator[base.DirEntry] {
	return &mockIterator{entries, children, -1, base.DirEntry{}, nil}
}

func (i *mockIterator) Next() bool {
	if i.err != nil || i.children ==nil || i.position >= len(i.children) - 1 {
		return false
	}
	i.position++
	i.err = i.entries[i.children[i.position].Path].Err
	if i.err != nil {
		return false
	}
	i.value = i.children[i.position]
	return true
}

func (i *mockIterator) Value() base.DirEntry {
	return i.value
}

func (i *mockIterator) Err() error {
	return nil
}

type byPath []base.DirEntry
func (e byPath) Len() int { return len(e) }
func (e byPath) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e byPath) Less(i, j int) bool { return e[i].Path < e[j].Path }