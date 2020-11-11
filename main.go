package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type snippet struct {
	Memory   []byte
	Vars     int
	AllChars []byte
}

func main() {
	tmpServerGoFile, _ := ioutil.TempFile("./", "serverExec_*.go")
	tmpWriteVar := generateServer()
	files, _ := ioutil.ReadDir("./")
	imports := [][]byte{
		[]byte("encoding/json"),
		[]byte("fmt"),
		[]byte("io/ioutil"),
		[]byte("net"),
		[]byte("net/http"),
		[]byte("strconv"),
		[]byte("time")}
	for _, file := range files {
		if file.Name() != filepath.Base(os.Args[0]) && file.Name() != tmpServerGoFile.Name() {
			contentBytes, _ := ioutil.ReadFile(file.Name())
			contentString := string(contentBytes)
			contentString = strings.Replace(contentString, "`", "\\`", -1)
			var memories []snippet
			var isAntislash bool = false
			var isGo bool = false
			var isEcho bool = false
			var echoVar []byte = []byte{}
			var isImport bool = false
			var isImportStr bool = false
			var importVar []byte = []byte{}
			var isString bool = false
			var skip int = 0
			var tmpSnippet snippet = snippet{[]byte{}, 0, []byte{}}
			for i, char := range contentBytes {
				if skip > 0 {
					tmpSnippet.AllChars = append(tmpSnippet.AllChars, char)
					skip--
					continue
				}
				if isEcho == true {
					tmpSnippet.AllChars = append(tmpSnippet.AllChars, char)
					if char != ' ' {
						echoVar = append(echoVar, char)
					} else {
						tmpSnippet.Vars++
						tmpSnippet.Memory = append(tmpSnippet.Memory, append([]byte("snippetVar"+strconv.Itoa(tmpSnippet.Vars)+" = "), echoVar...)...)
						echoVar = []byte{}
						isEcho = false
					}
				} else if isImport == true {
					tmpSnippet.AllChars = append(tmpSnippet.AllChars, char)
					if char == '"' {
						if isImportStr == false {
							isImportStr = true
						} else {
							isAlreadyInUse := false
							for _, imp := range imports {
								if string(imp) == strings.Trim(string(importVar), " ") {
									isAlreadyInUse = true
									break
								}
							}
							if isAlreadyInUse == false {
								imports = append(imports, []byte(strings.Trim(string(importVar), " ")))
							}
							importVar = []byte{}
							isImport = false
							isImportStr = false
						}
					} else {
						importVar = append(importVar, char)
					}
				} else if isAntislash == true {
					tmpSnippet.AllChars = append(tmpSnippet.AllChars, char)
					isAntislash = false
					tmpSnippet.Memory = append(tmpSnippet.Memory, char)
				} else if char == '\\' {
					tmpSnippet.AllChars = append(tmpSnippet.AllChars, char)
					isAntislash = true
				} else if char == '"' {
					tmpSnippet.AllChars = append(tmpSnippet.AllChars, char)
					if isString == true {
						isString = false
					} else {
						isString = true
					}
					tmpSnippet.Memory = append(tmpSnippet.Memory, char)
				} else if char == '<' && contentBytes[i+1] == 'g' && contentBytes[i+2] == 'o' && isString == false {
					tmpSnippet.AllChars = append(tmpSnippet.AllChars, char)
					skip = 2
					isGo = true
				} else if char == 'g' && contentBytes[i+1] == 'o' && contentBytes[i+2] == '>' && isString == false {
					tmpSnippet.AllChars = append(tmpSnippet.AllChars, append([]byte{char}, []byte("o>")...)...)
					skip = 2
					if isGo == true {
						isGo = false
						memories = append(memories, tmpSnippet)
						tmpSnippet = snippet{[]byte{}, 0, []byte{}}
					} else {
						isGo = true
					}
				} else if char == 'e' && contentBytes[i+1] == 'c' && contentBytes[i+2] == 'h' && contentBytes[i+3] == 'o' && isString == false {
					tmpSnippet.AllChars = append(tmpSnippet.AllChars, char)
					skip = 3
					isEcho = true
				} else if char == 'i' && contentBytes[i+1] == 'm' && contentBytes[i+2] == 'p' && contentBytes[i+3] == 'o' && contentBytes[i+4] == 'r' && contentBytes[i+5] == 't' && isString == false {
					tmpSnippet.AllChars = append(tmpSnippet.AllChars, char)
					skip = 5
					isImport = true
				} else {
					if isGo == true {
						tmpSnippet.AllChars = append(tmpSnippet.AllChars, char)
						tmpSnippet.Memory = append(tmpSnippet.Memory, char)
					}
				}
			}
			sName := strings.Split(file.Name(), ".")
			tmpWriteVar += "\n case \"" + file.Name() + "\", \"" + strings.Join(sName[:len(sName)-1], "") + "\":"
			tmpVars := ""
			for _, snippet := range memories {
				for vi := 0; vi <= snippet.Vars; vi++ {
					if vi != 0 {
						tmpWriteVar += "\nvar snippetVar" + strconv.Itoa(vi) + " string = \"\""
						tmpVars += "`+snippetVar" + strconv.Itoa(vi) + "+`"
					}
				}
				tmpWriteVar += "\n" + string(snippet.Memory)
				contentString = strings.Replace(contentString, string(snippet.AllChars), tmpVars, -1)
			}
			tmpWriteVar += "\nw.Write([]byte(`" + contentString + "`))"
		}
	}
	allimps := ""
	if len(imports) > 0 {
		for _, imp := range imports {
			allimps += "\"" + string(imp) + "\"\n"
		}
		allimps = strings.TrimRight(allimps, "\n")
	}
	tmpWriteVar = strings.Replace(tmpWriteVar, "----importpackages----", allimps, 1)
	tmpServerGoFile.Write([]byte(tmpWriteVar + "\n}\n}"))
	tmpServerGoFile.Close()
	if runtime.GOOS == "windows" {
		cmd := exec.Command("go", "build", "-o", "./server.exe")
		err := cmd.Run()
		if err != nil {
			fmt.Println("An error occured while compiling")
			fmt.Println(err)
		}
	} else {
		cmd := exec.Command("go", "build", "-o", "./server")
		err := cmd.Run()
		if err != nil {
			fmt.Println("An error occured while compiling")
			fmt.Println(err)
		}
	}
	os.Remove("./" + tmpServerGoFile.Name())
}
