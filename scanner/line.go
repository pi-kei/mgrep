package scanner

import (
	"bufio"
	"regexp"

	"github.com/pi-kei/mgrep/base"
)

type Line struct{
	reader base.Reader
}

func NewLineScanner(reader base.Reader) base.Scanner {
	return &Line{reader}
}

func (l *Line) ScanFile(fileEntry base.DirEntry, searchRegexp *regexp.Regexp, options base.SearchOptions, callback func(base.SearchResult) error) error {
	file, err := l.reader.OpenFile(fileEntry)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		if slice := searchRegexp.FindStringIndex(line); slice != nil {
			err := callback(base.SearchResult{Path: fileEntry.Path, LineNumber: lineNumber, StartIndex: slice[0], EndIndex: slice[1], Line: line})
			if err != nil && err != base.SkipSearchResult {
				return err
			}
		}
	}
	return nil
}

func (l *Line) ScanDir(rootPath string, options base.SearchOptions, callback func(base.DirEntry) error) error {
	rootDirEntry, rootErr := l.reader.ReadRootEntry(rootPath)
	if rootErr != nil {
		return rootErr
	}
	rootErr = callback(rootDirEntry)
	if rootErr != nil {
		if rootErr == base.SkipDirEntry || rootErr == base.SkipFileEntry {
			return nil;
		}
		return rootErr
	}
	if rootDirEntry.IsDir {
		var scanDir func(base.DirEntry, base.SearchOptions, func(base.DirEntry) error) error
		scanDir = func(dirEntry base.DirEntry, options base.SearchOptions, callback func(base.DirEntry) error) error {
			entries, err := l.reader.ReadDir(dirEntry)
			for _, entry := range entries {
				var loopErr error
				if entry.IsDir {
					loopErr = callback(dirEntry)
					if loopErr == base.SkipDirEntry {
						continue
					}
					loopErr = scanDir(entry, options, callback)
				} else {
					loopErr = callback(entry)
					if loopErr == base.SkipFileEntry {
						continue
					}
				}
				if loopErr != nil {
					return loopErr
				}
			}
			if err != nil {
				return err
			}
			return nil
		}
		return scanDir(rootDirEntry, options, callback)
	}
	return nil;
}