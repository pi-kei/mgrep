package scanner

import (
	"bufio"
	"regexp"

	"github.com/pi-kei/mgrep/base"
)

type Line struct{
	reader base.Reader
	filter base.Filter
}

func NewLineScanner(reader base.Reader, filter base.Filter) base.Scanner {
	return &Line{reader, filter}
}

func (l *Line) ScanFile(fileEntry base.DirEntry, searchRegexp *regexp.Regexp, callback func(base.SearchResult) error) error {
	file, err := l.reader.OpenFile(fileEntry)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		if slice := searchRegexp.FindStringIndex(line); slice != nil {
			result := base.SearchResult{Path: fileEntry.Path, LineNumber: lineNumber, StartIndex: slice[0], EndIndex: slice[1], Line: line}
			if l.filter.SkipSearchResult(result) {
				continue
			}
			err := callback(result)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Line) ScanDirs(rootPath string, callback func(base.DirEntry) error) error {
	rootDirEntry, rootErr := l.reader.ReadRootEntry(rootPath)
	if rootErr != nil {
		return rootErr
	}
	
	if !rootDirEntry.IsDir {
		if l.filter.SkipFileEntry(rootDirEntry) {
			return nil
		}
		return callback(rootDirEntry)
	}
	
	var scanDir func(base.DirEntry, func(base.DirEntry) error) error
	scanDir = func(dirEntry base.DirEntry, callback func(base.DirEntry) error) error {
		entriesIter, err := l.reader.ReadDir(dirEntry)
		for entriesIter.Next() {
			entry := entriesIter.Value()
			var loopErr error
			if entry.IsDir {
				if l.filter.SkipDirEntry(entry) {
					continue
				}
				loopErr = scanDir(entry, callback)
			} else {
				if l.filter.SkipFileEntry(entry) {
					continue
				}
				loopErr = callback(entry)
			}
			if loopErr != nil {
				return loopErr
			}
		}
		if entriesIter.Err() != nil {
			return entriesIter.Err()
		}
		return err
	}
	
	if l.filter.SkipDirEntry(rootDirEntry) {
		return nil
	}
	return scanDir(rootDirEntry, callback)
}