package scanner

import (
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/pi-kei/mgrep/internal/base"
	"github.com/pi-kei/mgrep/internal/reader"
)

func TestLineScanner_ScanDirs(t *testing.T) {
	now := time.Now().UTC()
	content := "hello\nsecond line hhhhh\nthird line"
	testEntries := reader.MockEntries{
		"aaa": {ModTime: now, Content: nil},
		"aaa/bbb": {ModTime: now, Content: nil},
		"aaa/bbb/ccc": {ModTime: now, Content: nil},
		"aaa/bbb/ccc/ddd": {ModTime: now, Content: nil},
		"aaa/bbb/ccc/ddd/hhh": {ModTime: now, Content: &content},
		"aaa/bbb/ccc/eee": {ModTime: now, Content: nil},
		"aaa/bbb/ccc/fff": {ModTime: now, Content: nil},
		"aaa/bbb/ggg": {ModTime: now, Content: nil},
	}
	reader := reader.NewMockReader(testEntries)
	scanner := NewLineScanner(reader)
	
	// Walk through whole tree scructure
	callbacks := []base.DirEntry{
		{Path: "aaa", Depth: 0, IsDir: true, Size: 0, ModTime: testEntries["aaa"].ModTime},
		{Path: "aaa/bbb", Depth: 1, IsDir: true, Size: 0, ModTime: testEntries["aaa/bbb"].ModTime},
		{Path: "aaa/bbb/ccc", Depth: 2, IsDir: true, Size: 0, ModTime: testEntries["aaa/bbb/ccc"].ModTime},
		{Path: "aaa/bbb/ccc/ddd", Depth: 3, IsDir: true, Size: 0, ModTime: testEntries["aaa/bbb/ccc/ddd"].ModTime},
		{Path: "aaa/bbb/ccc/ddd/hhh", Depth: 4, IsDir: false, Size: int64(len(content)), ModTime: testEntries["aaa/bbb/ccc/ddd/hhh"].ModTime},
		{Path: "aaa/bbb/ccc/eee", Depth: 3, IsDir: true, Size: 0, ModTime: testEntries["aaa/bbb/ccc/eee"].ModTime},
		{Path: "aaa/bbb/ccc/fff", Depth: 3, IsDir: true, Size: 0, ModTime: testEntries["aaa/bbb/ccc/fff"].ModTime},
		{Path: "aaa/bbb/ggg", Depth: 2, IsDir: true, Size: 0, ModTime: testEntries["aaa/bbb/ggg"].ModTime},
	}
	calledTimes := 0
	err := scanner.ScanDirs("aaa", 0, func(entry base.DirEntry) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v expected %v", entry, callbacks[calledTimes])
		}
		calledTimes++
		return nil
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err != nil {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// Nonexisting root path
	callbacks = []base.DirEntry{}
	calledTimes = 0
	err = scanner.ScanDirs("nonexisting", 0, func(entry base.DirEntry) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v expected %v", entry, callbacks[calledTimes])
		}
		calledTimes++
		return nil
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err == nil {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// File root path
	callbacks = []base.DirEntry{
		{Path: "aaa/bbb/ccc/ddd/hhh", Depth: 4, IsDir: false, Size: int64(len(content)), ModTime: testEntries["aaa/bbb/ccc/ddd/hhh"].ModTime},
	}
	calledTimes = 0
	err = scanner.ScanDirs("aaa/bbb/ccc/ddd/hhh", 4, func(entry base.DirEntry) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v expected %v", entry, callbacks[calledTimes])
		}
		calledTimes++
		return nil
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err != nil {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// File root path, skip item
	callbacks = []base.DirEntry{
		{Path: "aaa/bbb/ccc/ddd/hhh", Depth: 4, IsDir: false, Size: int64(len(content)), ModTime: testEntries["aaa/bbb/ccc/ddd/hhh"].ModTime},
	}
	calledTimes = 0
	err = scanner.ScanDirs("aaa/bbb/ccc/ddd/hhh", 4, func(entry base.DirEntry) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v expected %v", entry, callbacks[calledTimes])
		}
		calledTimes++
		return scanner.GetSkipItem()
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err != nil {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// File root path, skip all
	callbacks = []base.DirEntry{
		{Path: "aaa/bbb/ccc/ddd/hhh", Depth: 4, IsDir: false, Size: int64(len(content)), ModTime: testEntries["aaa/bbb/ccc/ddd/hhh"].ModTime},
	}
	calledTimes = 0
	err = scanner.ScanDirs("aaa/bbb/ccc/ddd/hhh", 4, func(entry base.DirEntry) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v expected %v", entry, callbacks[calledTimes])
		}
		calledTimes++
		return scanner.GetSkipAll()
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err != nil {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// File root path, error from callback
	callbacks = []base.DirEntry{
		{Path: "aaa/bbb/ccc/ddd/hhh", Depth: 4, IsDir: false, Size: int64(len(content)), ModTime: testEntries["aaa/bbb/ccc/ddd/hhh"].ModTime},
	}
	calledTimes = 0
	testError := errors.New("test")
	err = scanner.ScanDirs("aaa/bbb/ccc/ddd/hhh", 4, func(entry base.DirEntry) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v expected %v", entry, callbacks[calledTimes])
		}
		calledTimes++
		return testError
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err == nil || !errors.Is(err, testError) {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// Dir root path, skip item
	callbacks = []base.DirEntry{
		{Path: "aaa", Depth: 0, IsDir: true, Size: 0, ModTime: testEntries["aaa"].ModTime},
	}
	calledTimes = 0
	err = scanner.ScanDirs("aaa", 0, func(entry base.DirEntry) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v expected %v", entry, callbacks[calledTimes])
		}
		calledTimes++
		return scanner.GetSkipItem()
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err != nil {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// Dir root path, skip all
	callbacks = []base.DirEntry{
		{Path: "aaa", Depth: 0, IsDir: true, Size: 0, ModTime: testEntries["aaa"].ModTime},
	}
	calledTimes = 0
	err = scanner.ScanDirs("aaa", 0, func(entry base.DirEntry) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v expected %v", entry, callbacks[calledTimes])
		}
		calledTimes++
		return scanner.GetSkipAll()
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err != nil {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// Dir root path, error from callback
	callbacks = []base.DirEntry{
		{Path: "aaa", Depth: 0, IsDir: true, Size: 0, ModTime: testEntries["aaa"].ModTime},
	}
	calledTimes = 0
	testError = errors.New("test")
	err = scanner.ScanDirs("aaa", 0, func(entry base.DirEntry) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v expected %v", entry, callbacks[calledTimes])
		}
		calledTimes++
		return testError
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err == nil || !errors.Is(err, testError) {
		t.Errorf("ScanDirs returned error %v", err)
	}
}

func TestLineScanner_ScanFile(t *testing.T) {
	now := time.Now().UTC()
	content := "hello\nsecond line hhhhh\nthird line"
	testEntries := reader.MockEntries{
		"aaa": {ModTime: now, Content: nil},
		"aaa/bbb": {ModTime: now, Content: nil},
		"aaa/bbb/ccc": {ModTime: now, Content: nil},
		"aaa/bbb/ccc/ddd": {ModTime: now, Content: nil},
		"aaa/bbb/ccc/ddd/hhh": {ModTime: now, Content: &content},
		"aaa/bbb/ccc/eee": {ModTime: now, Content: nil},
		"aaa/bbb/ccc/fff": {ModTime: now, Content: nil},
		"aaa/bbb/ggg": {ModTime: now, Content: nil},
	}
	reader := reader.NewMockReader(testEntries)
	scanner := NewLineScanner(reader)

	// Scan lines
	fileEntry := base.DirEntry{Path: "aaa/bbb/ccc/ddd/hhh", Depth: 4, IsDir: false, Size: int64(len(content)), ModTime: testEntries["aaa/bbb/ccc/ddd/hhh"].ModTime}
	callbacks := []base.SearchResult{
		{Path: fileEntry.Path, LineNumber: 1, StartIndex: 0, EndIndex: 5, Line: "hello"},
		{Path: fileEntry.Path, LineNumber: 2, StartIndex: 12, EndIndex: 17, Line: "second line hhhhh"},
	}
	calledTimes := 0
	err := scanner.ScanFile(fileEntry, regexp.MustCompile(`h\w{4}`), func(entry base.SearchResult) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v", entry)
		}
		calledTimes++
		return nil
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err != nil {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// Not a file
	fileEntry = base.DirEntry{Path: "aaa/bbb/ccc/ddd", Depth: 3, IsDir: true, Size: 0, ModTime: testEntries["aaa/bbb/ccc/ddd"].ModTime}
	callbacks = []base.SearchResult{}
	calledTimes = 0
	err = scanner.ScanFile(fileEntry, regexp.MustCompile(`h\w{4}`), func(entry base.SearchResult) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v", entry)
		}
		calledTimes++
		return nil
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err == nil {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// Skip item
	fileEntry = base.DirEntry{Path: "aaa/bbb/ccc/ddd/hhh", Depth: 4, IsDir: false, Size: int64(len(content)), ModTime: testEntries["aaa/bbb/ccc/ddd/hhh"].ModTime}
	callbacks = []base.SearchResult{
		{Path: fileEntry.Path, LineNumber: 1, StartIndex: 0, EndIndex: 5, Line: "hello"},
		{Path: fileEntry.Path, LineNumber: 2, StartIndex: 12, EndIndex: 17, Line: "second line hhhhh"},
	}
	calledTimes = 0
	err = scanner.ScanFile(fileEntry, regexp.MustCompile(`h\w{4}`), func(entry base.SearchResult) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v", entry)
		}
		calledTimes++
		return scanner.GetSkipItem()
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err != nil {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// Skip all
	fileEntry = base.DirEntry{Path: "aaa/bbb/ccc/ddd/hhh", Depth: 4, IsDir: false, Size: int64(len(content)), ModTime: testEntries["aaa/bbb/ccc/ddd/hhh"].ModTime}
	callbacks = []base.SearchResult{
		{Path: fileEntry.Path, LineNumber: 1, StartIndex: 0, EndIndex: 5, Line: "hello"},
	}
	calledTimes = 0
	err = scanner.ScanFile(fileEntry, regexp.MustCompile(`h\w{4}`), func(entry base.SearchResult) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v", entry)
		}
		calledTimes++
		return scanner.GetSkipAll()
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err != nil {
		t.Errorf("ScanDirs returned error %v", err)
	}

	// Return error from callback
	fileEntry = base.DirEntry{Path: "aaa/bbb/ccc/ddd/hhh", Depth: 4, IsDir: false, Size: int64(len(content)), ModTime: testEntries["aaa/bbb/ccc/ddd/hhh"].ModTime}
	callbacks = []base.SearchResult{
		{Path: fileEntry.Path, LineNumber: 1, StartIndex: 0, EndIndex: 5, Line: "hello"},
	}
	calledTimes = 0
	testError := errors.New("test")
	err = scanner.ScanFile(fileEntry, regexp.MustCompile(`h\w{4}`), func(entry base.SearchResult) error {
		if !reflect.DeepEqual(entry, callbacks[calledTimes]) {
			t.Errorf("Callback called with %v", entry)
		}
		calledTimes++
		return testError
	})
	if calledTimes != len(callbacks) {
		t.Errorf("Callback called %v times", calledTimes)
	}
	if err == nil || !errors.Is(err, testError) {
		t.Errorf("ScanDirs returned error %v", err)
	}
}