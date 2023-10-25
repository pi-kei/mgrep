package main

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/fatih/color"
)

func scanFile(fileEntry DirEntry, searchRegexp *regexp.Regexp, options SearchOptions, callback func(SearchResult) error) error {
	info, err := fileEntry.dirEntry.Info()
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		// nothing to search
		return nil
	}
	if info.Size() > options.maxSize {
		// skip file because of options
		return nil
	}
	if options.include != nil && !options.include.MatchString(fileEntry.path) {
		// skip file because of options
		return nil
	}
	if options.exclude != nil && options.exclude.MatchString(fileEntry.path) {
		// skip file because of options
		return nil
	}
	file, err := os.Open(fileEntry.path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		if len(line) > options.maxLength {
			// skip line because of options
			continue
		}
		if slice := searchRegexp.FindStringIndex(line); slice != nil {
			err := callback(SearchResult{fileEntry.path, lineNumber, slice[0], slice[1], line})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func scanDir(dirEntry DirEntry, options SearchOptions, callback func(DirEntry) error) error {
	if dirEntry.depth > options.maxDepth {
		return nil
	}
	osDirEntries, err := os.ReadDir(dirEntry.path)
	for _, osDirEntry := range osDirEntries {
		newDirEntry := DirEntry{filepath.Join(dirEntry.path, osDirEntry.Name()), dirEntry.depth + 1, osDirEntry}
		if osDirEntry.IsDir() {
			err := scanDir(newDirEntry, options, callback)
			if err != nil {
				return err
			}
		} else {
			err := callback(newDirEntry)
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

var highlight = color.New(color.Bold, color.FgHiYellow).SprintFunc()

func printResult(result SearchResult) {
	startPart := result.line[0:result.startIndex]
	resultPart := highlight(result.line[result.startIndex:result.endIndex])
	endPart := result.line[result.endIndex:]
	fmt.Printf("%s[%v,%v]:%s%s%s\n", result.path, result.lineNumber, result.startIndex+1, startPart, resultPart, endPart)
}

func search(rootPath string, searchRegexp *regexp.Regexp, options SearchOptions, ctx context.Context) {
	fileInfo, err := os.Lstat(rootPath)
	if err != nil {
		fmt.Println("Error scanning dir", err)
		return
	}
	rootDirEntry := DirEntry{rootPath, 0, fs.FileInfoToDirEntry(fileInfo)}
	
	filesChannel := make(chan DirEntry, options.bufferSize)
	resultsChannel := make(chan SearchResult, options.bufferSize)
	
	var resultsWG sync.WaitGroup

	go func() {
		defer close(filesChannel)
		if rootDirEntry.dirEntry.IsDir() {
			err := scanDir(rootDirEntry, options, func(fileEntry DirEntry) error {
				select {
				case filesChannel <- fileEntry:
					return nil
				case <-ctx.Done():
					return filepath.SkipAll
				}
			})
			if err != nil {
				fmt.Println("Error scanning dir", err)
			}
		} else {
			select {
			case filesChannel <- rootDirEntry:
				return
			case <-ctx.Done():
				fmt.Println("Error scanning dir", filepath.SkipAll)
				return
			}
		}
	}()

	for i := 0; i < options.concurrency; i++ {
		resultsWG.Add(1)
		go func() {
			defer resultsWG.Done()
			for fileEntry := range filesChannel {
				err := scanFile(fileEntry, searchRegexp, options, func(result SearchResult) error {
					select {
					case resultsChannel <- result:
						return nil
					case <-ctx.Done():
						return filepath.SkipAll
					}
				})
				if err != nil {
					fmt.Println("Error scanning file", err)
				}
			}
		}()
	}

	go func() {
		defer close(resultsChannel)
		resultsWG.Wait()
	}()

	for result := range resultsChannel {
		printResult(result)
	}
}