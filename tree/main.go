package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
)

type ByAlphabet []fs.DirEntry

func (a ByAlphabet) Len() int           { return len(a) }
func (a ByAlphabet) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAlphabet) Less(i, j int) bool { return a[i].Name() < a[j].Name() }

func filterFiles(dir []fs.DirEntry) []fs.DirEntry {
	var dirs []fs.DirEntry
	for _, d := range dir {
		if d.IsDir() {
			dirs = append(dirs, d)
		}
	}
	return dirs
}

func deepDirTree(prefix string, isParentLast bool, out io.Writer, path string, printFiles bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	dir, err := file.ReadDir(0)
	if err != nil {
		return err
	}
	if !printFiles {
		dir = filterFiles(dir)
	}
	sort.Sort(ByAlphabet(dir))
	countFiles := len(dir)
	if isParentLast {
		prefix += "	"
	} else {
		prefix += "│	"
	}
	for i, d := range dir {
		finfo, err := d.Info()
		if err != nil {
			return err
		}
		size := finfo.Size()
		bytesCount := ""
		if !d.IsDir() {
			if size == 0 {
				bytesCount = " (empty)"
			} else {
				bytesCount = fmt.Sprintf(" (%vb)", size)
			}
		}
		outLine := ""
		if i == countFiles-1 {
			outLine = prefix + "└───" + d.Name() + bytesCount + "\n"
		} else {
			outLine = prefix + "├───" + d.Name() + bytesCount + "\n"
		}
		out.Write([]byte(outLine))
		if d.IsDir() {
			isParentLast := i == countFiles-1
			deepDirTree(prefix, isParentLast, out, path+"/"+d.Name(), printFiles)
		}
	}

	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	dir, err := file.ReadDir(0)
	if err != nil {
		return err
	}
	if !printFiles {
		dir = filterFiles(dir)
	}
	sort.Sort(ByAlphabet(dir))
	countFiles := len(dir)
	for i, d := range dir {
		finfo, err := d.Info()
		if err != nil {
			return err
		}
		size := finfo.Size()
		bytesCount := ""
		if !d.IsDir() {
			if size == 0 {
				bytesCount = " (empty)"
			} else {
				bytesCount = fmt.Sprintf(" (%vb)", size)
			}
		}
		outLine := ""
		if i == countFiles-1 {
			outLine = "└───" + d.Name() + bytesCount + "\n"
		} else {
			outLine = "├───" + d.Name() + bytesCount + "\n"
		}
		out.Write([]byte(outLine))
		if d.IsDir() {
			isParentLast := i == countFiles-1
			deepDirTree("", isParentLast, out, path+"/"+d.Name(), printFiles)
		}
	}

	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
