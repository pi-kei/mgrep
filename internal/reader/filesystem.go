package reader

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pi-kei/mgrep/internal/base"
)

type FileSystem struct{}

func NewFileSystemReader() base.Reader {
	return &FileSystem{}
}

func (fs *FileSystem) OpenFile(fileEntry base.DirEntry) (io.ReadCloser, error) {
	return os.Open(fileEntry.Path)
}

func (fs *FileSystem) ReadDir(dirEntry base.DirEntry) (base.Iterator[base.DirEntry], error) {
	fsDirEntries, err := os.ReadDir(dirEntry.Path)
	return newIterator(dirEntry, fsDirEntries), err
}

func (fs *FileSystem) ReadRootEntry(name string) (base.DirEntry, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return base.DirEntry{}, err
	}
	return base.DirEntry{Path: name, Depth: 0, IsDir: info.IsDir(), Size: info.Size(), ModTime: info.ModTime()}, nil
}

type iterator struct {
	parent base.DirEntry
	entries []fs.DirEntry
	position int
	value base.DirEntry
	err error
}

func newIterator(parent base.DirEntry, entries []fs.DirEntry) base.Iterator[base.DirEntry] {
	return &iterator{parent, entries, -1, base.DirEntry{}, nil}
}

func (i *iterator) Next() bool {
	if i.err != nil || i.position >= len(i.entries) - 1 {
		return false
	}
	i.position++
	info, err := i.entries[i.position].Info()
	if err != nil {
		i.err = err
		return false
	}
	i.value = base.DirEntry{
		Path: filepath.Join(i.parent.Path, info.Name()),
		Depth: i.parent.Depth + 1,
		IsDir: info.IsDir(),
		Size: info.Size(),
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