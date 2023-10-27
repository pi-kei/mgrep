package reader

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pi-kei/mgrep/base"
)

type FileSystem struct{}

func NewFileSystemReader() base.Reader {
	return &FileSystem{}
}

func (fs *FileSystem) OpenFile(fileEntry base.DirEntry) (interface {
	io.Reader
	io.Closer
}, error) {
	return os.Open(fileEntry.Path)
}

func (fs *FileSystem) ReadDir(dirEntry base.DirEntry) ([]base.DirEntry, error) {
	fsDirEntries, err := os.ReadDir(dirEntry.Path)
	entries := []base.DirEntry{}
	for _, fsDirEntry := range fsDirEntries {
		path := filepath.Join(dirEntry.Path, fsDirEntry.Name())
		depth := dirEntry.Depth + 1
		isDir := fsDirEntry.IsDir()
		info, err := fsDirEntry.Info()
		if err != nil {
			return entries, err
		}
		size := info.Size()
		entries = append(entries, base.DirEntry{Path: path, Depth: depth, IsDir: isDir, Size: size})
	}
	return entries, err
}

func (fs *FileSystem) ReadRootEntry(name string) (base.DirEntry, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return base.DirEntry{}, err
	}
	return base.DirEntry{Path: name, Depth: 0, IsDir: info.IsDir(), Size: info.Size()}, nil
}