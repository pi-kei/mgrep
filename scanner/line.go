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
	if fileEntry.Size == 0 {
		// nothing to search
		return nil
	}
	if fileEntry.Size > options.MaxSize {
		// skip file because of options
		return nil
	}
	if options.Include != nil && !options.Include.MatchString(fileEntry.Path) {
		// skip file because of options
		return nil
	}
	if options.Exclude != nil && options.Exclude.MatchString(fileEntry.Path) {
		// skip file because of options
		return nil
	}
	file, err := l.reader.OpenFile(fileEntry)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		if len(line) > options.MaxLength {
			// skip line because of options
			continue
		}
		if slice := searchRegexp.FindStringIndex(line); slice != nil {
			err := callback(base.SearchResult{Path: fileEntry.Path, LineNumber: lineNumber, StartIndex: slice[0], EndIndex: slice[1], Line: line})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Line) ScanDir(rootPath string, options base.SearchOptions, callback func(base.DirEntry) error) error {
	rootDirEntry, err := l.reader.ReadRootEntry(rootPath)
	if err != nil {
		return err
	}
	if rootDirEntry.IsDir {
		var scanDir func(base.DirEntry, base.SearchOptions, func(base.DirEntry) error) error
		scanDir = func(dirEntry base.DirEntry, options base.SearchOptions, callback func(base.DirEntry) error) error {
			if dirEntry.Depth > options.MaxDepth {
				return nil
			}
			entries, err := l.reader.ReadDir(dirEntry)
			for _, entry := range entries {
				if entry.IsDir {
					err := scanDir(entry, options, callback)
					if err != nil {
						return err
					}
				} else {
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
		return scanDir(rootDirEntry, options, callback)
	}
	return callback(rootDirEntry);
}