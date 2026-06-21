package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
)

func tokenize(command string) []string {
	var tokens []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escape := false
	space := false

	for _, ch := range command {
		if escape {
			current.WriteRune(ch)
			escape = false
			continue
		}
		if ch == ' ' && space && !inSingleQuote && !inDoubleQuote {
			continue
		} else {
			space = false
		}
		switch {
		// handling for single and double quotes
		case ch == '\'' && !inSingleQuote && !inDoubleQuote:
			inSingleQuote = true
		case ch == '\'' && inSingleQuote:
			inSingleQuote = false
		case ch == '"' && !inSingleQuote && !inDoubleQuote:
			inDoubleQuote = true
		case ch == '"' && inDoubleQuote:
			inDoubleQuote = false
		case ch == ' ' && !inSingleQuote && !inDoubleQuote:
			tokens = append(tokens, current.String())
			fmt.Println("found space!")
			fmt.Println("token:", tokens)
			fmt.Println("current:", current)
			current.Reset()
			space = true
		case ch == '\\' && !inSingleQuote:
			escape = true
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

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
		command = strings.TrimSpace(command)
		tokens := tokenize(command)
		builtin := []string{"exit", "echo", "type", "pwd", "cd"}
		switch tokens[0] {
		case "exit":
			if len(tokens) > 1 {
				fmt.Fprintln(os.Stderr, "exit: too many arguments")
				continue
			}
			return
		case "echo":
			fmt.Println(strings.Join(tokens[1:], " "))
		case "cat":
			files := tokens[1:]
			content := []string{}
			for _, file := range files {
				contentBytes, err := os.ReadFile(file)
				if err != nil {
					fmt.Printf("error reading file: %v: %v\n", file, err)
				}
				content = append(content, strings.TrimSpace(string(contentBytes)))
			}
			fmt.Println(strings.Join(content, ""))
		case "pwd":
			wd, err := os.Getwd()
			if err != nil {
				fmt.Println("Couldn't get current working directory: %v", err)
			}
			fmt.Printf("%v\n", wd)
		case "type":
			for i := 1; i < len(tokens); i++ {
				if slices.Contains(builtin, tokens[i]) {
					fmt.Printf("%v is a shell builtin\n", tokens[i])
				} else {
					path, err := exec.LookPath(tokens[i])
					if err != nil {
						fmt.Println(tokens[i] + ": not found")
					} else {
						fmt.Printf("%v is %v\n", tokens[i], path)
					}
				}
			}
		case "cd":
			var err error
			if tokens[1] == "~" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					fmt.Printf("error geting home directory of user: %v", err)
				}
				err = os.Chdir(homeDir)
			} else {
				err = os.Chdir(tokens[1])
			}
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					fmt.Printf("cd: %v: No such file or directory\n", tokens[1])
				} else {
					fmt.Println("error changing directory: %v\n", err)
				}
			}
		default:
			_, err := exec.LookPath(tokens[0])
			if err != nil {
				fmt.Println(tokens[0] + ": not found")
			} else {
				cmd := exec.Command(tokens[0], tokens[1:]...)
				// cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				err := cmd.Run()
				if err != nil {
					fmt.Printf("couldn't run command: %v: %v", tokens[0], err)
				}
			}
		}
	}
}
