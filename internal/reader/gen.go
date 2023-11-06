package reader

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var textLines = []string{
	"\"The only way to do great work is to love what you do.\" - Steve Jobs",
	"\"In the end, we will remember not the words of our enemies, but the silence of our friends.\" - Martin Luther King Jr.",
	"\"The best way to predict your future is to create it.\" - Abraham Lincoln",
	"\"Be the change you wish to see in the world.\" - Mahatma Gandhi",
	"\"Believe you can and you're halfway there.\" - Theodore Roosevelt",
	"\"Success is not final, failure is not fatal: It is the courage to continue that counts.\" - Winston Churchill",
	"\"Happiness is not something ready-made. It comes from your own actions.\" - Dalai Lama",
	"\"The greatest glory in living lies not in never falling, but in rising every time we fall.\" - Nelson Mandela",
	"\"The only true wisdom is in knowing you know nothing.\" - Socrates",
	"\"The purpose of our lives is to be happy.\" - Dalai Lama",
}

var entryNames = []string{
	"bicycle",
	"sunflower",
	"notebook",
	"guitar",
	"pineapple",
	"camera",
	"elephant",
	"backpack",
	"balloon",
	"umbrella",
}

type gen struct {
	rnd *rand.Rand
	maxLines int          // max lines in file, min is 0
	maxDepth int          // max depth of a path, min is 0
	maxDirs int           // max child dirs in a parent dir, min is 0
	maxFiles int          // max child files in a parent dir, min is 0
	modTimeBase time.Time // base of the mod time range. see maxHours
	maxHours int64        // max hours range of a mod time of an entry. range: modTimeBase +/- maxHours/2 
}

func NewEntriesGen(seed int64, maxLines, maxDepth, maxDirs, maxFiles int, modTimeBase time.Time, maxHours int64) *gen {
	return &gen{rand.New(rand.NewSource(seed)), maxLines, maxDepth, maxDirs, maxFiles, modTimeBase, maxHours}
}

func (g *gen) Generate() (MockEntries, string, []string) {
	entries := make(MockEntries)
	rootName := entryNames[g.rnd.Intn(len(entryNames))]
	entries[rootName] = MockEntry{ModTime: g.modTime()}
	contents := []string{}
	contents = g.dirChildren(entries, rootName, contents, 0)
	return entries, rootName, contents
}

func (g *gen) dirChildren(entries MockEntries, dirName string, contents []string, depth int) []string {
	depth += 1
	if depth > g.maxDepth {
		parent := entries[dirName]
		parent.children = []string{}
		entries[dirName] = parent
		return contents
	}
	dirsCount := 0
	if g.maxDirs > 0 {
		dirsCount = g.rnd.Intn(g.maxDirs)
	}
	filesCount := 0
	if g.maxFiles > 0 {
		filesCount = g.rnd.Intn(g.maxFiles)
	}
	children := make([]string, dirsCount + filesCount)
	for i := 0; i < filesCount; i++ {
		filePath := g.entryName(dirName, i, depth)
		content := g.fileContent()
		contents = append(contents, content)
		entries[filePath] = MockEntry{ModTime: g.modTime(), Content: &content}
		children[i] = filePath
	}
	for i := 0; i < dirsCount; i++ {
		dirPath := g.entryName(dirName, i + filesCount, depth)
		entries[dirPath] = MockEntry{ModTime: g.modTime()}
		contents = g.dirChildren(entries, dirPath, contents, depth)
		children[i + filesCount] = dirPath
	}
	parent := entries[dirName]
	parent.children = children
	entries[dirName] = parent
	return contents
}

func (g *gen) entryName(dirName string, index, depth int) string {
	count := len(entryNames)
	n := ""
	if index >= count {
		n = strconv.Itoa(index / count)
	}
	if depth % 2 == 0 {
		return dirName + "/" + entryNames[count - 1 - (index % count)] + n
	}
	return dirName + "/" + entryNames[index % count] + n
}

func (g *gen) fileContent() string {
	linesCount := 0
	if g.maxLines > 0 {
		linesCount = g.rnd.Intn(g.maxLines)
	}
	lines := make([]string, linesCount)
	for i := 0; i < int(linesCount); i++ {
		lines[i] = textLines[g.rnd.Intn(len(textLines))]
	}
	return strings.Join(lines, "\n")
}

func (g *gen) modTime() time.Time {
	modTimeBase := g.modTimeBase
	if g.maxHours <= 0 {
		return modTimeBase
	}
	return modTimeBase.Add(time.Hour * time.Duration(g.rnd.Int63n(g.maxHours) - (g.maxHours / 2)))
}