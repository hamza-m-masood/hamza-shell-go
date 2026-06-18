package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
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
		command = strings.TrimSpace(command)
		tokens := strings.Split(command, " ")
		if command == "exit" {
			break
		} else if tokens[0] == "echo" {
			for i := 1; i < len(tokens); i++ {
				fmt.Print(tokens[i])
				if i != len(tokens)-1 {
					fmt.Print(" ")
				}
				if i == len(tokens)-1 {
					fmt.Println()
				}
			}
		} else {

			fmt.Println(command + ": command not found")
		}
	}
}
