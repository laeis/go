package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

//Lines is
type Lines struct {
	withSiblings rune
	corner       rune
	delimiter    rune
	connnection  string
}

// Set default symbols for lines
func getLineSymbols() Lines {
	return Lines{'├', '└', '│', "───"}
}

//filterFileStructer remove files if its do not need in output
func filterFileStructer(dirInfo *[]os.FileInfo) {
	filterDirectory := make([]os.FileInfo, 0)
	for _, valueDir := range *dirInfo {
		if !valueDir.IsDir() {
			continue
		}
		filterDirectory = append(filterDirectory, valueDir)
	}
	*dirInfo = filterDirectory
}

//createEntityPrefix create string prefix for der tree features
func createEntityPrefix(beforePrefix string, isLast bool) string {
	lines := getLineSymbols()
	currentLeveldelimiter := beforePrefix
	if isLast {
		currentLeveldelimiter += string(lines.corner)
	} else {
		currentLeveldelimiter += string(lines.withSiblings)
	}
	currentLeveldelimiter += lines.connnection
	return currentLeveldelimiter
}

//createParentPrefix create prefix string for children node
func createParentPrefix(before string, isDir bool, isLast bool) string {
	lines := getLineSymbols()
	nexLevelFirstDeliment := before
	if isDir && !isLast {
		nexLevelFirstDeliment += string(lines.delimiter)
	}
	nexLevelFirstDeliment += "\t"
	return nexLevelFirstDeliment
}

//createEntitySuffix create suffix with files size for output string
func createEntitySuffix(size int64) string {
	var sizeValue string
	if size == 0 {
		sizeValue = "empty"
	} else {
		sizeValue = strconv.FormatInt(size, 10) + "b"
	}
	return " (" + sizeValue + ")"
}

//printTree function for recursive go-round tree  and put tree node in output
func printTree(out io.Writer, path string, printFiles bool, before string) error {

	file, err := os.Open(path) // For read access.
	defer func() {
		file.Close()
	}()

	dirInfo, err := file.Readdir(-1)

	if err != nil {
		return err
	}
	//Clear slice of files(if neccesary)
	if !printFiles {
		filterFileStructer(&dirInfo)
	}

	if len(dirInfo) == 0 {
		return nil
	}
	//Sort dorectory List by name
	sort.SliceStable(dirInfo, func(i, j int) bool {
		iEelem := dirInfo[i].Name()
		jElem := dirInfo[j].Name()
		return iEelem < jElem
	})

	for i, value := range dirInfo {
		//Create string for output
		currentLeveldelimiter := createEntityPrefix(before, len(dirInfo)-1 == i)
		nexLevelFirstDeliment := createParentPrefix(before, value.IsDir(), len(dirInfo)-1 == i)
		pathString := currentLeveldelimiter + value.Name()
		if !value.IsDir() {
			pathString += createEntitySuffix(value.Size())
		}
		//Write to output
		fmt.Fprintln(out, pathString)
		//If is directory, start all process againe
		if value.IsDir() {
			//Create path for scan child direcoty
			childPath := path + string(os.PathSeparator) + value.Name()
			//Run recursive alghoritm
			err := printTree(out, childPath, printFiles, nexLevelFirstDeliment)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	err := printTree(out, path, printFiles, "")
	return err
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
