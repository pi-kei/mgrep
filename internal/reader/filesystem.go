package reader

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pi-kei/mgrep/internal/base"
)

type FileSystem struct{}

func NewFileSystem() base.Reader {
	return &FileSystem{}
}

func (fs *FileSystem) OpenFile(fileEntry base.DirEntry) (io.ReadCloser, error) {
	return os.Open(fileEntry.Path)
}

func (fs *FileSystem) ReadDir(dirEntry base.DirEntry) (base.Iterator[base.DirEntry], error) {
	fsDirEntries, err := os.ReadDir(dirEntry.Path)
	return newIterator(dirEntry.Path, dirEntry.Depth+1, fsDirEntries), err
}

func (fs *FileSystem) ReadRootEntry(name string, depth int) (base.DirEntry, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return base.DirEntry{}, err
	}
	return base.DirEntry{Path: name, Depth: depth, IsDir: info.IsDir(), Size: info.Size(), ModTime: info.ModTime()}, nil
}

type iterator struct {
	parentPath string
	depth      int
	entries    []fs.DirEntry
	position   int
	value      base.DirEntry
	err        error
}

func newIterator(parentPath string, depth int, entries []fs.DirEntry) base.Iterator[base.DirEntry] {
	return &iterator{parentPath, depth, entries, -1, base.DirEntry{}, nil}
}

func (i *iterator) Next() bool {
	if i.err != nil || i.entries == nil || i.position >= len(i.entries)-1 {
		return false
	}
	i.position++
	info, err := i.entries[i.position].Info()
	if err != nil {
		i.err = err
		return false
	}
	i.value = base.DirEntry{
		Path:    filepath.Join(i.parentPath, info.Name()),
		Depth:   i.depth,
		IsDir:   info.IsDir(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}
	return true
}

func (i *iterator) Value() base.DirEntry {
	return i.value
}

func (i *iterator) Err() error {
	return i.err
}
