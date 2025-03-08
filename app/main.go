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
	if len(parts) > 1 {
		arg_str := parts[1]
		var curr_sub_str strings.Builder
		in_single_quotes := false
		in_double_qoutes := false
		escaped := 0 //0 = unescaped 1 = semi escaped 2 = escaped
		
		for _,char := range arg_str {
			switch {
				case escaped == 2 || (escaped == 1  && (char == '"' || char == '$' || char =='\n' || char == '\\')):
					curr_sub_str.WriteRune(char)
					escaped = 0
				case escaped == 1:
					curr_sub_str.WriteRune('\\')
					curr_sub_str.WriteRune(char)
					escaped = 0
				case char == '\\' && !in_double_qoutes && !in_single_quotes:
					escaped = 2
				case char == '\\' && in_double_qoutes && !in_single_quotes:
					escaped = 1
				case char == '"'  && !in_single_quotes:
					in_double_qoutes = !in_double_qoutes
				case char == '\'' && !in_double_qoutes:
					in_single_quotes = !in_single_quotes // Toggle quote state
				case char == ' ' && !in_single_quotes && !in_double_qoutes:
					if curr_sub_str.Len() > 0 {
						args_list = append(args_list, curr_sub_str.String())
						curr_sub_str.Reset()
					}
				default:
					curr_sub_str.WriteRune(char)
			}
		}

		if curr_sub_str.Len() > 0 {
			args_list = append(args_list, curr_sub_str.String())
		}
	}



	// if len(parts)>1{
	// 	arg_str := parts[1]+" "
	// 	len_arg_str := len(arg_str)
	// 	curr_sub_str := ""
	// 	curr_idx := 0
	// 	for curr_idx < len_arg_str{
			
	// 		if(arg_str[curr_idx] == '\''){
	// 			end := curr_idx + 1 + strings.Index(arg_str[curr_idx+1:],"'")
	// 			curr_sub_str += arg_str[curr_idx+1:end]
	// 			curr_idx = end
	// 		} else if arg_str[curr_idx] == ' '{
	// 			if len(curr_sub_str)>0{
	// 				args_list = append(args_list, curr_sub_str)
	// 				curr_sub_str = ""
	// 			}
				
	// 		} else{
	// 			curr_sub_str += string(arg_str[curr_idx])
	// 		}

	// 		curr_idx += 1
	// 	}
	// 	if len(curr_sub_str) > 0{
	// 		args_list = append(args_list, curr_sub_str)
	// 	}


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
