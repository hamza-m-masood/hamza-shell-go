package main

import (
	"bufio"
	"fmt"
	"os"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	// TODO: Uncomment the code below to pass the first stage
	fmt.Print("$ ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	fmt.Printf("%v: command not found", scanner.Text())
}
