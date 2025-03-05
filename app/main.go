package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"path/filepath"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint


type Command struct {
	name string
	args []string
}


func parseCmd(str string) Command {
	parts := strings.SplitN(str," ",2)
	name := parts[0]
	var args_list []string 
	if len(parts)>1{
		arg_str := parts[1]

		// var start int
		// var end int
		// var prev_end int
		// var curr_start int

		// for start<len(arg_str) {
		// 	curr_start = strings.Index(arg_str[start:], "'")
		// 	if curr_start == -1 {
		// 		for arg := range strings.SplitSeq(arg_str[prev_end:], " "){
		// 			trimmed_arg := strings.TrimSpace(arg)
		// 			if len(trimmed_arg)>0 {
		// 				args_list = append(args_list, trimmed_arg)
		// 			}
		// 		}
		// 		break
		// 	}
		// 	start += curr_start
		// 	for arg := range strings.SplitSeq(arg_str[prev_end:start], " "){
		// 		trimmed_arg := strings.TrimSpace(arg)
		// 		if len(trimmed_arg)>0 {
		// 			args_list = append(args_list, trimmed_arg)
		// 		}
		// 	}
		// 	end = start + 1 + strings.Index(arg_str[start+1:],"'")
		// 	args_list = append(args_list, arg_str[start+1:min(len(arg_str)-1,end)])
		// 	start = end+1
		// 	prev_end = end+1
		// 	fmt.Println(args_list)
		// }
	
		for {
			start := strings.Index(arg_str, "'")
			if start == -1 {
				for arg := range strings.SplitSeq(arg_str, " "){
					trimmed_arg := strings.TrimSpace(arg)
					if len(trimmed_arg)>0 {
						args_list = append(args_list, trimmed_arg)
					}
				}	
				break
			}
			for arg := range strings.SplitSeq(arg_str[:start], " "){
				trimmed_arg := strings.TrimSpace(arg)
				if len(trimmed_arg)>0 {
					args_list = append(args_list, trimmed_arg)
				}
			
			}
			arg_str = arg_str[start+1:]
			end := strings.Index(arg_str, "'")
			qouted_arg := arg_str[:end]
			args_list = append(args_list, qouted_arg)
			arg_str = arg_str[end+1:]
		}
	}
	return Command{name, args_list}
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
					location_arr := strings.Split(cmd.args[0], "/")
					var location string
					if cmd.args[0][0] == '~' {
						location = os.Getenv("HOME")+"/"+filepath.Clean(strings.Join(location_arr[1:], "/"))
					}else{
						location, _ = filepath.Abs(cmd.args[0])
					}

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
