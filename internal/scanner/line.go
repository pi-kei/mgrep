package scanner

import (
	"bufio"
	"errors"
	"regexp"

	"github.com/pi-kei/mgrep/internal/base"
)

type Line struct{
	reader base.Reader
	skipItem error
	skipAll error
}

func NewLineScanner(reader base.Reader) base.Scanner {
	return &Line{reader, errors.New("skip item"), errors.New("skip all")}
}

func (l *Line) GetReader() base.Reader {
	return l.reader
}

func (l *Line) GetSkipItem() error {
	return l.skipItem
}

func (l *Line) GetSkipAll() error {
	return l.skipAll
}

func (l *Line) ScanFile(fileEntry base.DirEntry, searchRegexp *regexp.Regexp, callback func(base.SearchResult) error) error {
	file, err := l.GetReader().OpenFile(fileEntry)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		if slice := searchRegexp.FindIndex(scanner.Bytes()); slice != nil { 
			err := callback(base.SearchResult{Path: fileEntry.Path, LineNumber: lineNumber, StartIndex: slice[0], EndIndex: slice[1], Line: scanner.Text()})
			if err != nil {
				if errors.Is(err, l.GetSkipItem()) {
					continue
				}
				if errors.Is(err, l.GetSkipAll()) {
					return nil
				}
				return err
			}
		}
	}
	return nil
}

func (l *Line) ScanDirs(rootPath string, depth int, callback func(base.DirEntry) error) error {
	rootDirEntry, rootErr := l.GetReader().ReadRootEntry(rootPath, depth)
	if rootErr != nil {
		return rootErr
	}
	
	if !rootDirEntry.IsDir {
		rootErr := callback(rootDirEntry)
		if rootErr != nil {
			if errors.Is(rootErr, l.GetSkipItem()) {
				return nil
			}
			if errors.Is(rootErr, l.GetSkipAll()) {
				return nil
			}
			return rootErr
		}
		return nil
	}
	
	var scanDir func(base.DirEntry, func(base.DirEntry) error) error
	scanDir = func(dirEntry base.DirEntry, callback func(base.DirEntry) error) error {
		iter, err := l.GetReader().ReadDir(dirEntry)
		if iter == nil {
			return err
		}
		for iter.Next() {
			entry := iter.Value()
			var loopErr error
			if entry.IsDir {
				loopErr = callback(entry)
				if loopErr != nil {
					if errors.Is(loopErr, l.GetSkipItem()) {
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
					if errors.Is(loopErr, l.GetSkipItem()) {
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
		if errors.Is(rootErr, l.GetSkipItem()) {
			return nil
		}
		if errors.Is(rootErr, l.GetSkipAll()) {
			return nil
		}
		return rootErr
	}
	rootErr = scanDir(rootDirEntry, callback)
	if errors.Is(rootErr, l.GetSkipAll()) {
		return nil
	}
	return rootErr
}