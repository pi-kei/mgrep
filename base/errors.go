package base

import "errors"

var SkipAll = errors.New("skip everything and stop the walk")
var SkipDirEntry = errors.New("skip directory")
var SkipFileEntry = errors.New("skip file")
var SkipSearchResult = errors.New("skip search result")