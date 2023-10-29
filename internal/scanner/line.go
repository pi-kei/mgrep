package scanner

import (
	"bufio"
	"regexp"

	"github.com/pi-kei/mgrep/internal/base"
)

type Line struct{
	reader base.Reader
}

func NewLineScanner(reader base.Reader) base.Scanner {
	return &Line{reader}
}

func (l *Line) GetReader() base.Reader {
	return l.reader
}

func (l *Line) ScanFile(fileEntry base.DirEntry, searchRegexp *regexp.Regexp, callback func(base.SearchResult) (bool, error)) error {
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
			_, err := callback(result)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Line) ScanDirs(rootPath string, callback func(base.DirEntry) (bool, error)) error {
	rootDirEntry, rootErr := l.GetReader().ReadRootEntry(rootPath)
	if rootErr != nil {
		return rootErr
	}
	
	if !rootDirEntry.IsDir {
		_, rootErr := callback(rootDirEntry)
		return rootErr
	}
	
	var scanDir func(base.DirEntry, func(base.DirEntry) (bool, error)) error
	scanDir = func(dirEntry base.DirEntry, callback func(base.DirEntry) (bool, error)) error {
		entriesIter, err := l.GetReader().ReadDir(dirEntry)
		for entriesIter.Next() {
			entry := entriesIter.Value()
			var loopErr error
			var skipped bool
			if entry.IsDir {
				skipped, loopErr = callback(entry)
				if loopErr != nil {
					return loopErr
				}
				if skipped {
					continue
				}
				loopErr = scanDir(entry, callback)
				if loopErr != nil {
					return loopErr
				}
			} else {
				_, loopErr = callback(entry)
				if loopErr != nil {
					return loopErr
				}
			}
		}
		if entriesIter.Err() != nil {
			return entriesIter.Err()
		}
		return err
	}
	
	skipped, rootErr := callback(rootDirEntry)
	if rootErr != nil {
		return rootErr
	}
	if skipped {
		return nil
	}
	return scanDir(rootDirEntry, callback)
}