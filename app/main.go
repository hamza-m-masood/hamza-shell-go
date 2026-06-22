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

func processTokens(tokens []string) (string, error) {
	builtin := []string{"exit", "echo", "type", "pwd", "cd"}
	switch tokens[0] {
	case "echo":
		echoOutput := fmt.Sprint(strings.Join(tokens[1:], " "))
		return echoOutput, nil
	case "cat":
		files := tokens[1:]
		content := []string{}
		for _, file := range files {
			contentBytes, err := os.ReadFile(file)
			if err != nil {
				catErrorOutput := fmt.Sprintf("error reading file: %v: %v\n", file, err)
				catError := errors.New(catErrorOutput)
				return "", catError
			}
			content = append(content, strings.TrimSpace(string(contentBytes)))
		}
		return strings.Join(content, ""), nil
	case "pwd":
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return wd, nil
	case "type":
		for i := 1; i < len(tokens); i++ {
			if slices.Contains(builtin, tokens[i]) {
				return fmt.Sprintf("%v is a shell builtin", tokens[i]), nil
			} else {
				path, err := exec.LookPath(tokens[i])
				if err != nil {
					typeErrorOutput := fmt.Sprint(tokens[i] + ": not found")
					typeError := errors.New(typeErrorOutput)
					return "", typeError
				} else {
					return fmt.Sprintf("%v is %v", tokens[i], path), nil
				}
			}
		}
	case "cd":
		var err error
		if tokens[1] == "~" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				cdErrorOutput := fmt.Sprintf("error geting home directory of user: %v", err)
				cdError := errors.New(cdErrorOutput)
				return "", cdError
			}
			err = os.Chdir(homeDir)
		} else {
			err = os.Chdir(tokens[1])
		}
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				cdErrorOutput := fmt.Sprintf("cd: %v: No such file or directory\n", tokens[1])
				cdError := errors.New(cdErrorOutput)
				return "", cdError
			} else {
				return fmt.Sprintf("error changing directory: %v\n", err), nil
			}
		}
	default:
		_, err := exec.LookPath(tokens[0])
		if err != nil {
			return fmt.Sprint(tokens[0] + ": not found"), nil
		} else {
			cmd := exec.Command(tokens[0], tokens[1:]...)
			cmd.Stdin = os.Stdin
			output, err := cmd.Output()
			if err != nil {
				defaultErrorOutput := fmt.Sprintf("couldn't run command: %v: %v", tokens[0], err)
				defaultError := errors.New(defaultErrorOutput)
				return "", defaultError
			}
			return strings.TrimSpace(string(output)), nil
		}
	}
	return "", nil
}

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
		if tokens[0] == "exit" {
			if len(tokens) > 1 {
				fmt.Fprintln(os.Stderr, "exit: too many arguments")
				continue
			}
			return
		}
		if slices.Contains(tokens, ">") || slices.Contains(tokens, "1>") {
			var index int
			if slices.Contains(tokens, ">") {
				index = slices.Index(tokens, ">")
			} else if slices.Contains(tokens, "1>") {
				index = slices.Index(tokens, "1>")
			}

			if index == -1 {
				fmt.Println("Could not redirect output")
			}

			p := tokens[:index]
			output, err := processTokens(p)
			if err != nil {
				fmt.Print(err)
			}
			//only one file on right side of ">" operator
			if len(tokens[index+1:]) != 1 {
				fmt.Println("Please enter only one file name")
				continue

			}
			fileName := tokens[index+1:][0]
			b := []byte(output)
			err = os.WriteFile(fileName, b, 0644)
			if err != nil {
				fmt.Println("error writing to file:", err)
			}
		} else {
			output, err := processTokens(tokens)
			if err != nil {
				fmt.Println(err)
			}
			//TODO: fix fmt.Print to fmt.Println
			fmt.Println(output)
		}

	}

}
