package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
)

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
		builtin := []string{"exit", "echo", "type", "pwd"}
		switch tokens[0] {
		case "exit":
			if len(tokens) > 1 {
				fmt.Fprintln(os.Stderr, "exit: too many arguments")
				continue
			}
			return
		case "echo":
			fmt.Println(strings.Join(tokens[1:], " "))
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
