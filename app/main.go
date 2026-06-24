package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/chzyer/readline"
)

type Output struct {
	Content    string
	IsStdError bool
}

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

func processTokens(tokens []string) []Output {
	builtin := []string{"exit", "echo", "type", "pwd", "cd"}
	var outputs []Output
	switch tokens[0] {
	case "echo":
		outputs = []Output{
			{
				Content:    fmt.Sprint(strings.Join(tokens[1:], " ") + "\n"),
				IsStdError: false,
			},
		}
		return outputs
	case "cat":
		files := tokens[1:]
		for _, file := range files {
			contentBytes, err := os.ReadFile(file)
			if err != nil {
				outputs = append(outputs, Output{Content: fmt.Sprintf("%v: nonexistent: No such file or directory", tokens[0]) + "\n", IsStdError: true})
				continue
			}
			outputs = append(outputs, Output{Content: string(contentBytes), IsStdError: false})
		}
		return outputs
	case "pwd":
		wd, err := os.Getwd()

		if err != nil {
			outputs = append(outputs, Output{Content: string(err.Error()) + "\n", IsStdError: true})
			return outputs
		}
		outputs = append(outputs, Output{Content: wd + "\n", IsStdError: true})
		return outputs
	case "type":
		for i := 1; i < len(tokens); i++ {
			if slices.Contains(builtin, tokens[i]) {
				outputs = append(outputs, Output{Content: fmt.Sprintf("%v is a shell builtin", tokens[i]) + "\n", IsStdError: false})
			} else {
				path, err := exec.LookPath(tokens[i])
				if err != nil {
					outputs = append(outputs, Output{Content: fmt.Sprint(tokens[i] + ": not found" + "\n"), IsStdError: true})
				} else {
					outputs = append(outputs, Output{Content: fmt.Sprintf("%v is %v", tokens[i], path) + "\n", IsStdError: false})
				}
			}
		}
		return outputs
	case "cd":
		var err error
		if tokens[1] == "~" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				outputs = append(outputs, Output{Content: fmt.Sprintf("error geting home directory of user: %v", err) + "\n", IsStdError: true})
				return outputs
			}
			err = os.Chdir(homeDir)
		} else {
			err = os.Chdir(tokens[1])
		}
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				outputs = append(outputs, Output{Content: fmt.Sprintf("cd: %v: No such file or directory", tokens[1]) + "\n", IsStdError: true})
				return outputs
			} else {
				outputs = append(outputs, Output{Content: fmt.Sprintf("error changing directory: %v", err) + "\n", IsStdError: true})
				return outputs
			}
		}
	default:
		_, err := exec.LookPath(tokens[0])
		outputs := []Output{}
		if err != nil {
			outputs = append(outputs, Output{Content: fmt.Sprint(tokens[0]+": not found") + "\n", IsStdError: true})
			return outputs
		} else {
			cmd := exec.Command(tokens[0], tokens[1:]...)
			cmd.Stdin = os.Stdin
			output, err := cmd.Output()
			if err != nil {
				outputs = append(outputs, Output{Content: fmt.Sprintf("%v: nonexistent: No such file or directory", tokens[0]) + "\n", IsStdError: true})
				return outputs
			}
			outputs = append(outputs, Output{Content: strings.TrimSpace(string(output)) + "\n", IsStdError: false})
			return outputs
		}
	}
	return nil
}

// find the first occurence of the index and value of any element in r that is present in s
func containsAny(s, r []string) (int, string) {
	var element string
	var index int
	for _, el := range r {
		if slices.Contains(s, el) {
			element = el
		}
	}
	if len(element) > 0 {
		index = slices.Index(s, element)
		return index, element
	}

	return 0, ""
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("echo"),
	readline.PcItem("exit"),
)

func main() {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		HistoryFile:     "/tmp/readline.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    completer,

		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	l.CaptureExitSignal()
	// reader := bufio.NewReader(os.Stdin)
	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		tokens := tokenize(line)
		if tokens[0] == "exit" {
			if len(tokens) > 1 {
				fmt.Fprintln(os.Stderr, "exit: too many arguments")
				continue
			}
			return
		}
		redirects := []string{">", "1>", "2>", ">>", "1>>", "2>>"}
		index, redirect := containsAny(tokens, redirects)
		if len(redirect) > 0 {
			p := tokens[:index]
			output := processTokens(p)
			//only one file on right side of ">" operator
			if len(tokens[index+1:]) != 1 {
				fmt.Println("Please enter only one file name")
				continue
			}
			fileName := tokens[index+1:][0]
			stdOutput := strings.Builder{}
			stdErrOutput := strings.Builder{}
			for _, o := range output {
				if o.IsStdError {
					stdErrOutput.WriteString(o.Content)
					continue
				}
				stdOutput.WriteString(o.Content)
			}
			stdOutputB := []byte(stdOutput.String())
			stdErrOutputB := []byte(stdErrOutput.String())
			if redirect == ">" || redirect == "1>" {
				_ = os.WriteFile(fileName, stdOutputB, 0644)
			} else if redirect == "2>" {
				_ = os.WriteFile(fileName, stdErrOutputB, 0644)
			} else if redirect == ">>" || redirect == "1>>" {
				f, _ := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
				defer f.Close()
				f.WriteString(stdOutput.String())
			} else if redirect == "2>>" {
				f, _ := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
				defer f.Close()
				f.WriteString(stdErrOutput.String())
			}

			if redirect != "2>" && redirect != "2>>" {
				if len(stdErrOutput.String()) > 0 {
					fmt.Print(stdErrOutput.String())
				}
			} else {
				if len(stdOutput.String()) > 0 {
					fmt.Print(stdOutput.String())
				}
			}
		} else {
			output := processTokens(tokens)
			for _, o := range output {
				fmt.Print(o.Content)
			}
		}
	}

}
