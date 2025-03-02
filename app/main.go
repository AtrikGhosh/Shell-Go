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
		switch cmd_list[0]{
			case "exit":
				fallthrough
			case "exit 0":
				os.Exit(0)
			case "echo":
				fmt.Println(strings.Join(cmd_list[1:]," "))
			default:
				fmt.Printf(command + ": command not found\n")
		}
	}
}
