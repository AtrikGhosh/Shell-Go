package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint


type Command struct {
	name string
	args []string
}


func parseCmd(str string) Command {
	parts := strings.Fields(str)
	name := parts[0]
	args := parts[1:]
	return Command{name, args}
}



func main() {

	// Wait for user input
	reader := bufio.NewReader(os.Stdin)

	for{
		fmt.Fprint(os.Stdout, "$ ")

		input,_ := reader.ReadString('\n')
		cmd := parseCmd(strings.TrimSpace(input))

		switch cmd.name{
			case "exit":
				os.Exit(0)
			case "echo":
				fmt.Println(strings.Join(cmd.args," "))
			default:	
				fmt.Println(cmd.name + ": command not found")
		}

	}
}
