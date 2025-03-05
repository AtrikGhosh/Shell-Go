package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {

	// Wait for user input
	reader := bufio.NewReader(os.Stdin)

	for{
		fmt.Fprint(os.Stdout, "$ ")
		
		input,_ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		fmt.Println(input + ": command not found\n")
	}
}
