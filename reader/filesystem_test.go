package reader

import (
	"errors"
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/pi-kei/mgrep/base"
)


func TestIterator_Empty(t *testing.T) {
	it := newIterator(base.DirEntry{
		Path: "a",
		Depth: 1,
	}, []fs.DirEntry{})

	// Initial state

	if it.Err() != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{}) {
		t.Fail()
	}

	// No next, still initial state

	if it.Next() {
		t.Fail()
	}

	if it.Err() != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{}) {
		t.Fail()
	}

	// No next, still initial state again

	if it.Next() {
		t.Fail()
	}

	if it.Err() != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{}) {
		t.Fail()
	}
}
func TestIterator_ErrorInTheMiddle(t *testing.T) {
	t1 := time.Now().UTC()
	t2 := time.Now().UTC()
	it := newIterator(base.DirEntry{
		Path: "a",
		Depth: 1,
	}, []fs.DirEntry{
		&mockFsDirEntry{nameReturn: "b", isDirReturn: false, infoReturn: &mockFsFileInfo{sizeReturn: 100, modTimeReturn: t1}},
		&mockFsDirEntry{nameReturn: "c", isDirReturn: true, infoReturn: &mockFsFileInfo{sizeReturn: 256, modTimeReturn: t2}},
		&mockFsDirEntry{nameReturn: "d", isDirReturn: true, infoError: errors.New("error")},
		&mockFsDirEntry{nameReturn: "e", isDirReturn: true, infoReturn: &mockFsFileInfo{sizeReturn: 1024, modTimeReturn: t2}},
	})

	// Initial state

	if it.Err() != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{}) {
		t.Fail()
	}

	// First element

	if !it.Next() {
		t.Fail()
	}

	if it.Err() != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{Path: filepath.Join("a", "b"), Depth: 2, IsDir: false, Size: 100, ModTime: t1}) {
		t.Fail()
	}

	// Second element

	if !it.Next() {
		t.Fail()
	}

	if it.Err() != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Fail()
	}

	// Third element

	if it.Next() {
		t.Fail()
	}

	if it.Err() == nil || it.Err().Error() != "error" {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Fail()
	}

	// Still third element

	if it.Next() {
		t.Fail()
	}

	if it.Err() == nil || it.Err().Error() != "error" {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Fail()
	}
}

func TestIterator_NoErrorsTilTheEnd(t *testing.T) {
	t1 := time.Now().UTC()
	t2 := time.Now().UTC()
	it := newIterator(base.DirEntry{
		Path: "a",
		Depth: 1,
	}, []fs.DirEntry{
		&mockFsDirEntry{nameReturn: "b", isDirReturn: false, infoReturn: &mockFsFileInfo{sizeReturn: 100, modTimeReturn: t1}},
		&mockFsDirEntry{nameReturn: "c", isDirReturn: true, infoReturn: &mockFsFileInfo{sizeReturn: 256, modTimeReturn: t2}},
	})

	// Initial state

	if it.Err() != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{}) {
		t.Fail()
	}

	// First element

	if !it.Next() {
		t.Fail()
	}

	if it.Err() != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{Path: filepath.Join("a", "b"), Depth: 2, IsDir: false, Size: 100, ModTime: t1}) {
		t.Fail()
	}

	// Second element

	if !it.Next() {
		t.Fail()
	}

	if it.Err() != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Fail()
	}

	// Still second element but no more next

	if it.Next() {
		t.Fail()
	}

	if it.Err() != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Fail()
	}

	// Still second element but no more next again

	if it.Next() {
		t.Fail()
	}

	if it.Err() != nil {
		t.Fail()
	}

	if !reflect.DeepEqual(it.Value(), base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Fail()
	}
}

type mockFsDirEntry struct {
	nameReturn string
	isDirReturn bool
	infoReturn fs.FileInfo
	infoError error
	typeReturn fs.FileMode
}

func (m *mockFsDirEntry) Name() string {
	return m.nameReturn
}

func (m *mockFsDirEntry) IsDir() bool {
	return m.isDirReturn
}

func (m *mockFsDirEntry) Info() (fs.FileInfo, error) {
	return m.infoReturn, m.infoError
}

func (m *mockFsDirEntry) Type() fs.FileMode {
	return m.typeReturn
}

type mockFsFileInfo struct {
	nameReturn string
	sizeReturn int64
	modTimeReturn time.Time
	isDirReturn bool
	modeReturn fs.FileMode
	sysReturn any
}

func (m *mockFsFileInfo) Name() string {
	return m.nameReturn
}

func (m *mockFsFileInfo) Size() int64 {
	return m.sizeReturn
}

func (m *mockFsFileInfo) ModTime() time.Time {
	return m.modTimeReturn
}

func (m *mockFsFileInfo) IsDir() bool {
	return m.isDirReturn
}

func (m *mockFsFileInfo) Mode() fs.FileMode {
	return m.modeReturn
}

func (m *mockFsFileInfo) Sys() any {
	return m.sysReturn
}