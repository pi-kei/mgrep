package scanner

import (
	"bufio"
	"errors"
	"regexp"

	"github.com/pi-kei/mgrep/internal/base"
)

type Line struct {
	reader base.Reader
}

func NewLineScanner(reader base.Reader) base.Scanner {
	return &Line{reader}
}

func (l *Line) ScanFile(fileEntry base.DirEntry, searchRegexp *regexp.Regexp, callback func(base.SearchResult) error) error {
	file, err := l.reader.OpenFile(fileEntry)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		if slice := searchRegexp.FindIndex(scanner.Bytes()); slice != nil {
			err := callback(base.SearchResult{Path: fileEntry.Path, LineNumber: lineNumber, StartIndex: slice[0], EndIndex: slice[1], Line: scanner.Text()})
			if err != nil {
				if errors.Is(err, base.ErrSkipItem) {
					continue
				}
				if errors.Is(err, base.ErrSkipAll) {
					return nil
				}
				return err
			}
		}
	}
	return nil
}

func (l *Line) ScanDirs(rootPath string, depth int, callback func(base.DirEntry) error) error {
	rootDirEntry, rootErr := l.reader.ReadRootEntry(rootPath, depth)
	if rootErr != nil {
		return rootErr
	}

	if !rootDirEntry.IsDir {
		rootErr := callback(rootDirEntry)
		if rootErr != nil {
			if errors.Is(rootErr, base.ErrSkipItem) {
				return nil
			}
			if errors.Is(rootErr, base.ErrSkipAll) {
				return nil
			}
			return rootErr
		}
		return nil
	}

	var scanDir func(base.DirEntry, func(base.DirEntry) error) error
	scanDir = func(dirEntry base.DirEntry, callback func(base.DirEntry) error) error {
		iter, err := l.reader.ReadDir(dirEntry)
		if iter == nil {
			return err
		}
		for iter.Next() {
			entry := iter.Value()
			var loopErr error
			if entry.IsDir {
				loopErr = callback(entry)
				if loopErr != nil {
					if errors.Is(loopErr, base.ErrSkipItem) {
						continue
					}
					return loopErr
				}
				loopErr = scanDir(entry, callback)
				if loopErr != nil {
					return loopErr
				}
			} else {
				loopErr = callback(entry)
				if loopErr != nil {
					if errors.Is(loopErr, base.ErrSkipItem) {
						continue
					}
					return loopErr
				}
			}
		}
		if iter.Err() != nil {
			return iter.Err()
		}
		return err
	}

	rootErr = callback(rootDirEntry)
	if rootErr != nil {
		if errors.Is(rootErr, base.ErrSkipItem) {
			return nil
		}
		if errors.Is(rootErr, base.ErrSkipAll) {
			return nil
		}
		return rootErr
	}
	rootErr = scanDir(rootDirEntry, callback)
	if errors.Is(rootErr, base.ErrSkipAll) {
		return nil
	}
	return rootErr
}
