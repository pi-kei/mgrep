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
				continue;
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
	rootDirEntry, err := l.reader.ReadRootEntry(rootPath)
	if err != nil {
		return err
	}
	if rootDirEntry.IsDir {
		var scanDir func(base.DirEntry, func(base.DirEntry) error) error
		scanDir = func(dirEntry base.DirEntry, callback func(base.DirEntry) error) error {
			if l.filter.SkipDirEntry(dirEntry) {
				return nil
			}
			entries, err := l.reader.ReadDir(dirEntry)
			for _, entry := range entries {
				if entry.IsDir {
					err := scanDir(entry, callback)
					if err != nil {
						return err
					}
				} else {
					if l.filter.SkipFileEntry(entry) {
						continue
					}
					err = callback(entry)
					if err != nil {
						return err
					}
				}
			}
			if err != nil {
				return err
			}
			return nil
		}
		return scanDir(rootDirEntry, callback)
	}
	if l.filter.SkipFileEntry(rootDirEntry) {
		return nil
	}
	return callback(rootDirEntry);
}