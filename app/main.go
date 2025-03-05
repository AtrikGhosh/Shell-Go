package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	// "path"
	"path/filepath"
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
			case "type":
				arg := cmd.args[0]
				if arg == "echo" || arg == "exit" || arg == "type" || arg == "pwd"{
					fmt.Println(cmd.args[0] + " is a shell builtin")
				} else if path,err := exec.LookPath(arg);err == nil{
					fmt.Println(arg + " is " + path)
				}else {
					fmt.Println(arg + ": not found")
				}
			case "pwd":
				pwd, err := os.Getwd()
				if err != nil {
					fmt.Printf("Error printing directory: %s\n", err)
				} else {
					fmt.Println(pwd)
				}
			case "cd":
				if len(cmd.args) > 0{
					location,_ := filepath.Abs(strings.Join(cmd.args, "\\"))

					// if location_arr[0] == "~" {
					// 	location := os.Getenv("HOME")+strings.Join(location_arr[1:], "/")
					// }else if location_arr[0] == "."

					if err := os.Chdir(location); err != nil {
						fmt.Printf("%s: No such file or directory\n", location)
					}
				} else{
					fmt.Println("Invalid Argument: No file or directory specified")
				}	
			default:	
				command := exec.Command(cmd.name, cmd.args...)
				command.Stdout = os.Stdout
				command.Stderr = os.Stderr

				err := command.Run()
				if err!=nil{
					fmt.Println(cmd.name + ": command not found")
			}
		}

	}
}
