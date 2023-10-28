package scanner

import (
	"bufio"
	"regexp"

	"github.com/pi-kei/mgrep/internal/base"
)

type Line struct{
	reader base.Reader
	filter base.Filter
}

func NewLineScanner(reader base.Reader, filter base.Filter) base.Scanner {
	return &Line{reader, filter}
}

func (l *Line) GetReader() base.Reader {
	return l.reader
}

func (l *Line) GetFilter() base.Filter {
	return l.filter
}

func (l *Line) ScanFile(fileEntry base.DirEntry, searchRegexp *regexp.Regexp, callback func(base.SearchResult) error) error {
	file, err := l.GetReader().OpenFile(fileEntry)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		if slice := searchRegexp.FindStringIndex(line); slice != nil {
			result := base.SearchResult{Path: fileEntry.Path, LineNumber: lineNumber, StartIndex: slice[0], EndIndex: slice[1], Line: line}
			if l.GetFilter().SkipSearchResult(result) {
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
	rootDirEntry, rootErr := l.GetReader().ReadRootEntry(rootPath)
	if rootErr != nil {
		return rootErr
	}
	
	if !rootDirEntry.IsDir {
		if l.GetFilter().SkipFileEntry(rootDirEntry) {
			return nil
		}
		return callback(rootDirEntry)
	}
	
	var scanDir func(base.DirEntry, func(base.DirEntry) error) error
	scanDir = func(dirEntry base.DirEntry, callback func(base.DirEntry) error) error {
		entriesIter, err := l.GetReader().ReadDir(dirEntry)
		for entriesIter.Next() {
			entry := entriesIter.Value()
			var loopErr error
			if entry.IsDir {
				if l.GetFilter().SkipDirEntry(entry) {
					continue
				}
				loopErr = callback(entry)
				if loopErr != nil {
					return loopErr
				}
				loopErr = scanDir(entry, callback)
			} else {
				if l.GetFilter().SkipFileEntry(entry) {
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
	
	if l.GetFilter().SkipDirEntry(rootDirEntry) {
		return nil
	}
	rootErr = callback(rootDirEntry)
	if rootErr != nil {
		return rootErr
	}
	return scanDir(rootDirEntry, callback)
}