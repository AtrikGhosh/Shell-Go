package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint


type Command struct {
	name string
	flags []string
	args []string
}


func parseCmd(arg_str string) Command {

	var args_list []string 
	var curr_sub_str strings.Builder
	in_single_quotes := false
	in_double_qoutes := false
	escaped := false
	
	for _,char := range arg_str {
		switch {
			case escaped:
				if (in_double_qoutes && (char != '"' && char != '$' && char != '\\' && char != '\n')){
					curr_sub_str.WriteRune('\\')
				}
				curr_sub_str.WriteRune(char)
				escaped = false
			case char == '\\':
				if(in_single_quotes){
					curr_sub_str.WriteRune(char)
				}else{
					escaped = true
				}
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
	name := args_list[0]
	i := 1
	n := len(args_list)
	for (i<n && args_list[i][0] == '-'){
		i++
	}
	flags := args_list[1:i]
	args := args_list[i:]
	return Command{name,flags, args}
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
				if idx := slices.IndexFunc(cmd.args,func(s string) bool {return s == ">" || s == "1>"}); idx != -1 {
					text,filepath := cmd.args[:idx],cmd.args[idx+1]
					err := os.WriteFile(filepath, []byte(strings.Join(text," ")), 0o777)
					if err != nil {
						fmt.Println("Error:", err)
					}

				}else{		
					fmt.Println(strings.Join(cmd.args," "))
				}
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
				if idx := slices.IndexFunc(cmd.args,func(s string) bool {return s == ">" || s == "1>"}); idx != -1 {
						src_files,dest_file := cmd.args[:idx],cmd.args[idx+1]
						var validFiles []string
						for _, file := range src_files {
							if _, err := os.Stat(file); os.IsNotExist(err) {
								fmt.Fprintf(os.Stderr, "%s: %s: No such file or directory\n", cmd.name, file)
							} else {
								validFiles = append(validFiles, file)
							}
						}

						final_args := append(cmd.flags, validFiles...)
						command := exec.Command(cmd.name, final_args...)
						output,err := command.CombinedOutput()
						if err!=nil{
							fmt.Println(err)
						}
						err = os.WriteFile(dest_file, []byte(output), 0o777)
						if err != nil {
							fmt.Println("Error:", err)
						}

				} else if idx := slices.Index(cmd.args,"2>"); idx != -1 {
					src_files,dest_file := cmd.args[:idx],cmd.args[idx+1]
					var validFiles []string
					for _, file := range src_files {
						if _, err := os.Stat(file); os.IsNotExist(err) {
							err = os.WriteFile(dest_file, []byte(fmt.Sprintf("%s: %s: No such file or directory\n", cmd.name, file)), 0o777)
							if err != nil {
								fmt.Println("Error:", err)
							}
						} else {
							validFiles = append(validFiles, file)
						}
					}

					final_args := append(cmd.flags, validFiles...)
					command := exec.Command(cmd.name, final_args...)
					err := command.Run()
					if err!=nil{
						err = os.WriteFile(dest_file, []byte(err.Error()), 0o777)
						if err != nil {
							fmt.Println("Error:", err)
						}
					}
					

				} else {
					command := exec.Command(cmd.name, cmd.args...)
					err := command.Run()
					command.Stdout = os.Stdout
					command.Stderr = os.Stderr
					if err!=nil{
						fmt.Println(cmd.name + ": command not found")
					// }else{
					// 	stdOut := strings.Trim(string(output),"\r\n")
					// 	fmt.Fprintln(os.Stdout,stdOut)
					}
				}
		}

	}
}
