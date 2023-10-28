package reader

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/pi-kei/mgrep/base"
)

func TestFileSystemReader_OpenFile(t *testing.T) {
	fsr := NewFileSystemReader()
	
	_, file, _, _ := runtime.Caller(0)
	testFilePath := filepath.Join(filepath.Dir(file), "thisisfortest", "thisisfortest.txt")
	info, err := os.Lstat(testFilePath)
	if err != nil {
		t.Errorf("Lstat error: %v", err)
	}
	if info.IsDir() {
		t.Errorf("IsDir returned true")
	}
	reader, err := fsr.OpenFile(base.DirEntry{Path: testFilePath, Depth: 0, IsDir: info.IsDir(), Size: info.Size(), ModTime: info.ModTime()})
	if err != nil {
		t.Errorf("OpenFile error: %v", err)
	}
	err = reader.Close()
	if err != nil {
		t.Errorf("Close error: %v", err)
	}
}

func TestFileSystemReader_ReadDir(t *testing.T) {
	fsr := NewFileSystemReader()
	
	_, file, _, _ := runtime.Caller(0)
	testDirPath := filepath.Join(filepath.Dir(file), "thisisfortest")
	info, err := os.Lstat(testDirPath)
	if err != nil {
		t.Errorf("Lstat error: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("IsDir returned false")
	}
	iterator, err := fsr.ReadDir(base.DirEntry{Path: testDirPath, Depth: 0, IsDir: info.IsDir(), Size: info.Size(), ModTime: info.ModTime()})
	if err != nil {
		t.Errorf("ReadDir error: %v", err)
	}
	next := iterator.Next()
	if !next {
		t.Errorf("First call to Next() returned %v", next)
	}
	value := iterator.Value()
	if value.Path != filepath.Join(testDirPath, "thisisfortest.txt") || value.Depth != 1 || value.IsDir {
		t.Errorf("First call to Value() retured %v", value)
	}
	next = iterator.Next()
	if next {
		t.Errorf("Second call to Next() returned %v", next)
	}
}

func TestFileSystemReader_ReadRootEntry(t *testing.T) {
	fsr := NewFileSystemReader()
	
	_, file, _, _ := runtime.Caller(0)
	testDirPath := filepath.Join(filepath.Dir(file), "thisisfortest")
	info, err := os.Lstat(testDirPath)
	if err != nil {
		t.Errorf("Lstat error: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("IsDir returned false")
	}
	entry, err := fsr.ReadRootEntry(testDirPath)
	if err != nil {
		t.Errorf("ReadRootEntry error: %v", err)
	}
	if !reflect.DeepEqual(entry, base.DirEntry{Path: testDirPath, Depth: 0, IsDir: info.IsDir(), Size: info.Size(), ModTime: info.ModTime()}) {
		t.Errorf("ReadRootEntry returned: %v", entry)
	}
	testDirPath = filepath.Join(filepath.Dir(file), "thisisfortest_nonexisting")
	entry, err = fsr.ReadRootEntry(testDirPath)
	if err == nil || !reflect.DeepEqual(entry, base.DirEntry{}) {
		t.Errorf("ReadRootEntry returned no error and entry %v", entry)
	}
}

func TestIterator_Empty(t *testing.T) {
	it := newIterator(base.DirEntry{
		Path: "a",
		Depth: 1,
	}, []fs.DirEntry{})

	err := it.Err()
	if err != nil {
		t.Errorf("First call to Err() returned %v", err)
	}
	value := it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{}) {
		t.Errorf("First call to Value() returned %v", value)
	}
	next := it.Next()
	if next {
		t.Errorf("First call to Next() returned %v", next)
	}
	err = it.Err()
	if err != nil {
		t.Errorf("Second call to Err() returned %v", err)
	}
	value = it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{}) {
		t.Errorf("Second call to Value() returned %v", value)
	}
	next = it.Next()
	if next {
		t.Errorf("Second call to Next() returned %v", next)
	}
	err = it.Err()
	if err != nil {
		t.Errorf("Third call to Err() returned %v", err)
	}
	value = it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{}) {
		t.Errorf("Third call to Value() returned %v", value)
	}
}
func TestIterator_ErrorInTheMiddle(t *testing.T) {
	t1 := time.Now().UTC()
	t2 := time.Now().UTC()
	it := newIterator(base.DirEntry{
		Path: "a",
		Depth: 1,
	}, []fs.DirEntry{
		&mockFsDirEntry{infoReturn: &mockFsFileInfo{sizeReturn: 100, modTimeReturn: t1, nameReturn: "b", isDirReturn: false}},
		&mockFsDirEntry{infoReturn: &mockFsFileInfo{sizeReturn: 256, modTimeReturn: t2, nameReturn: "c", isDirReturn: true}},
		&mockFsDirEntry{infoError: errors.New("error"), nameReturn: "d", isDirReturn: true},
		&mockFsDirEntry{infoReturn: &mockFsFileInfo{sizeReturn: 1024, modTimeReturn: t2, nameReturn: "e", isDirReturn: true}},
	})

	err := it.Err()
	if err != nil {
		t.Errorf("First call to Err() returned %v", err)
	}
	value := it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{}) {
		t.Errorf("First call to Value() returned %v", value)
	}
	next := it.Next()
	if !next {
		t.Errorf("First call to Next() returned %v", next)
	}
	err = it.Err()
	if err != nil {
		t.Errorf("Second call to Err() returned %v", err)
	}
	value = it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{Path: filepath.Join("a", "b"), Depth: 2, IsDir: false, Size: 100, ModTime: t1}) {
		t.Errorf("Second call to Value() returned %v", value)
	}
	next = it.Next()
	if !next {
		t.Errorf("Second call to Next() returned %v", next)
	}
	err = it.Err()
	if err != nil {
		t.Errorf("Third call to Err() returned %v", err)
	}
	value = it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Errorf("Third call to Value() returned %v", value)
	}
	next = it.Next()
	if next {
		t.Errorf("Third call to Next() returned %v", next)
	}
	err = it.Err()
	if err == nil || err.Error() != "error" {
		t.Errorf("Fourth call to Err() returned %v", err)
	}
	value = it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Errorf("Fourth call to Value() returned %v", value)
	}
	next = it.Next()
	if next {
		t.Errorf("Fourth call to Next() returned %v", next)
	}
	err = it.Err()
	if err == nil || err.Error() != "error" {
		t.Errorf("Fifth call to Err() returned %v", err)
	}
	value = it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Errorf("Fifth call to Value() returned %v", value)
	}
}

func TestIterator_NoErrorsTilTheEnd(t *testing.T) {
	t1 := time.Now().UTC()
	t2 := time.Now().UTC()
	it := newIterator(base.DirEntry{
		Path: "a",
		Depth: 1,
	}, []fs.DirEntry{
		&mockFsDirEntry{infoReturn: &mockFsFileInfo{sizeReturn: 100, modTimeReturn: t1, nameReturn: "b", isDirReturn: false}},
		&mockFsDirEntry{infoReturn: &mockFsFileInfo{sizeReturn: 256, modTimeReturn: t2, nameReturn: "c", isDirReturn: true}},
	})

	err := it.Err()
	if err != nil {
		t.Errorf("First call to Err() returned %v", err)
	}
	value := it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{}) {
		t.Errorf("First call to Value() returned %v", value)
	}
	next := it.Next()
	if !next {
		t.Errorf("First call to Next() returned %v", next)
	}
	err = it.Err()
	if err != nil {
		t.Errorf("Second call to Err() returned %v", err)
	}
	value = it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{Path: filepath.Join("a", "b"), Depth: 2, IsDir: false, Size: 100, ModTime: t1}) {
		t.Errorf("Second call to Value() returned %v", value)
	}
	next = it.Next()
	if !next {
		t.Errorf("Second call to Next() returned %v", next)
	}
	err = it.Err()
	if err != nil {
		t.Errorf("Third call to Err() returned %v", err)
	}
	value = it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Errorf("Third call to Value() returned %v", value)
	}
	next = it.Next()
	if next {
		t.Errorf("Third call to Next() returned %v", next)
	}
	err = it.Err()
	if err != nil {
		t.Errorf("Fourth call to Err() returned %v", err)
	}
	value = it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Errorf("Fourth call to Value() returned %v", value)
	}
	next = it.Next()
	if next {
		t.Errorf("Fourth call to Next() returned %v", next)
	}
	err = it.Err()
	if err != nil {
		t.Errorf("Fifth call to Err() returned %v", err)
	}
	value = it.Value()
	if !reflect.DeepEqual(value, base.DirEntry{Path: filepath.Join("a", "c"), Depth: 2, IsDir: true, Size: 256, ModTime: t2}) {
		t.Errorf("Fifth call to Value() returned %v", value)
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