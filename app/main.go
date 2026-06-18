package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")
		command, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				break
			}
			fmt.Fprintln(os.Stderr, "Error reading input", err)
			os.Exit(1)
		}
		// to remove the carriage return at the end
		command = strings.TrimSpace(command)
		// to get every space separated word (also known as a token) and return a slice of all the words entered
		tokens := strings.Fields(command)
		builtin := []string{"exit", "echo", "type"}
		switch tokens[0] {
		case "exit":
			if len(tokens) > 1 {
				fmt.Fprintln(os.Stderr, "exit: too many arguments")
				continue
			}
			return
		case "echo":
			fmt.Println(strings.Join(tokens[1:], " "))
		case "type":
			for i := 1; i < len(tokens); i++ {
				if slices.Contains(builtin, tokens[i]) {
					fmt.Printf("%v is a shell builtin\n", tokens[i])
				} else {
					pathDirs := strings.Split(os.Getenv("PATHS"), ":")
					for _, dir := range pathDirs {
						files, err := os.ReadDir(dir)
						if err != nil {
							fmt.Printf("error reading path %v: %v", dir, err)
						}
						for _, file := range files {
							fullPath := filepath.Join(dir, file.Name())
							fileInfo, err := os.Lstat(fullPath)
							if err != nil {
								fmt.Printf("error getting file info %v: %v", fullPath, err)
							}
							if !fileInfo.IsDir() {
								permission := fileInfo.Mode().Perm()
								if permission&0o111 != 0 && file.Name() == tokens[i] {
									fmt.Printf("%v is %v\n", tokens[i], fullPath)
								}

							}
						}
					}
					fmt.Println(tokens[i] + ": not found")
				}
			}
		default:
			fmt.Println(command + ": command not found")
		}
	}
}
