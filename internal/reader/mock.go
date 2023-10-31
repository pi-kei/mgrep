package reader

import (
	"errors"
	"io"
	"strings"

	"github.com/pi-kei/mgrep/internal/base"
)

type TreeNode struct {
	Entry base.DirEntry
	Children []TreeNode
	Content *string
}

type mockReader struct {
	tree *TreeNode
}

func NewMockReader(tree *TreeNode) base.Reader {
	return &mockReader{tree}
}

func (r *mockReader) getNode(path string) *TreeNode {
	if path == "" {
		return nil
	}
	parts := strings.Split(path, "/")
	subPath := parts[0]
	if r.tree.Entry.Path != subPath {
		return nil
	}
	node := r.tree
	for i := 1; i < len(parts); i++ {
		if node.Children == nil {
			return nil
		}
		subPath = subPath + "/" + parts[i]
		found := -1
		for j := 0; j < len(node.Children); j++ {
			if subPath == node.Children[j].Entry.Path {
				found = j
				break
			}
		}
		if found >= 0 {
			node = &node.Children[found]
		} else {
			return nil
		}
	}
	return node
}

func (r *mockReader) OpenFile(fileEntry base.DirEntry) (io.ReadCloser, error) {
	node := r.getNode(fileEntry.Path)
	if node == nil {
		return nil, errors.New("path does not exist")
	}
	if node.Content == nil || node.Children != nil || node.Entry.IsDir {
		return nil, errors.New("path is not a file")
	}
	reader := strings.NewReader(*node.Content)
	return io.NopCloser(reader), nil
}

func (r *mockReader) ReadDir(dirEntry base.DirEntry) (base.Iterator[base.DirEntry], error) {
	node := r.getNode(dirEntry.Path)
	if node == nil {
		return nil, errors.New("path does not exist")
	}
	if node.Children == nil || node.Content != nil || !node.Entry.IsDir {
		return nil, errors.New("path is not a directory")
	}
	return newMockIterator(node.Children), nil
}

func (r *mockReader) ReadRootEntry(name string) (base.DirEntry, error) {
	node := r.getNode(name)
	if node == nil {
		return base.DirEntry{}, errors.New("path does not exist")
	}
	return node.Entry, nil
}

type mockIterator struct {
	children []TreeNode
	position int
	value base.DirEntry
}

func newMockIterator(children []TreeNode) base.Iterator[base.DirEntry] {
	return &mockIterator{children, -1, base.DirEntry{}}
}

func (i *mockIterator) Next() bool {
	if i.children ==nil || i.position >= len(i.children) - 1 {
		return false
	}
	i.position++
	i.value = i.children[i.position].Entry
	return true
}

func (i *mockIterator) Value() base.DirEntry {
	return i.value
}

func (i *mockIterator) Err() error {
	return nil
}