package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		command, err := reader.ReadString('\n')
		command = strings.TrimSpace(command)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		cmd_list := strings.Split(command," ")

		cmd := cmd_list[0]
		args := strings.Join(cmd_list[1:]," ")
		switch cmd{
			case "exit":
				fallthrough
			case "exit 0":
				os.Exit(0)
			case "echo":
				fmt.Println(args)
			case "type":
				if args == "echo" || args == "exit" || args == "type" {
					fmt.Println(args + " is a shell builtin")
				} else {
					fmt.Println(args + ": not found")
				}
			default:
				fmt.Printf(command + ": command not found\n")
		}
	}
}
