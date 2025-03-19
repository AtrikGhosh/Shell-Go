// TO DO - OPTIMIZEEEEE

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

type Command struct {
	name string
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
				in_single_quotes = !in_single_quotes
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
	args := args_list[1:]
	return Command{name, args}
}

func handleRedirection(cmd Command, redirectIdx int, redirectType string, writeMode int) {
	finalArgs, destFile := cmd.args[:redirectIdx], cmd.args[redirectIdx+1]
	writeFlags := os.O_WRONLY|os.O_CREATE|writeMode
	command := exec.Command(cmd.name, finalArgs...)
	if redirectType == ">" || redirectType == "1>" {
		file, err := os.OpenFile(destFile, writeFlags, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error opening destination file:", err)
			return
		}
		defer file.Close()
		command.Stdout = file
		command.Stderr = os.Stderr
	} else if redirectType == "2>" {
		file, err := os.OpenFile(destFile, writeFlags, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error opening destination file:", err)
			return
		}
		defer file.Close()
		command.Stdout = os.Stdout
		command.Stderr = file
	} else if redirectType == ">>" || redirectType == "1>>" {
		file, err := os.OpenFile(destFile, writeFlags, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error opening destination file:", err)
			return
		}
		defer file.Close()
		command.Stdout = file
		command.Stderr = os.Stderr
	} else if redirectType == "2>>" {
		file, err := os.OpenFile(destFile, writeFlags, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error opening destination file:", err)
			return
		}
		defer file.Close()
		command.Stdout = os.Stdout
		command.Stderr = file
	}
	command.Run()
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
					file,err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
					if err != nil {
						fmt.Println("Error opening destination file:", err)
					}
					_,err = file.WriteString(strings.Trim(strings.Join(text," "),"\r\n")+"\n")
					if err != nil {
						fmt.Println("Error:", err)
					}
					file.Close()

				} else if idx := slices.Index(cmd.args,"2>"); idx != -1 {
					text,filepath := cmd.args[:idx],cmd.args[idx+1]
					file,err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
					if err != nil {
						fmt.Println("Error opening destination file:", err)
					}
					_,err = fmt.Println(strings.Join(text," "))
					if err != nil {
						file.WriteString(err.Error())
					}
					file.Close()

				} else if idx := slices.IndexFunc(cmd.args,func(s string) bool {return s == ">>" || s == "1>>"}); idx != -1 {
					text,filepath := cmd.args[:idx],cmd.args[idx+1]
					file,err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
					if err != nil {
						fmt.Println("Error opening destination file:", err)
					}
					_,err = file.WriteString(strings.Trim(strings.Join(text," "),"\r\n")+"\n")
					if err != nil {
						fmt.Println("Error:", err)
					}
					file.Close()

				} else if idx := slices.Index(cmd.args,"2>>"); idx != -1 {
					text,filepath := cmd.args[:idx],cmd.args[idx+1]
					file,err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
					if err != nil {
						fmt.Println("Error opening destination file:", err)
					}
					_,err = fmt.Println(strings.Join(text," "))
					if err != nil {
						file.WriteString(err.Error())
					}
					file.Close()

				} else {		
					fmt.Println(strings.Join(cmd.args," "))
				}
			case "type":
				arg := cmd.args[0]
				if slices.Contains([]string{"echo", "exit", "type", "pwd", "cd"}, arg){
					fmt.Println(arg + " is a shell builtin")
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
				if idx := slices.IndexFunc(cmd.args, func(s string) bool { return s == ">" || s == "1>"}); idx != -1 {
					handleRedirection(cmd, idx, cmd.args[idx], os.O_TRUNC)
				} else if idx := slices.Index(cmd.args,"2>"); idx != -1 {
					handleRedirection(cmd, idx, cmd.args[idx], os.O_TRUNC)
				} else if idx := slices.IndexFunc(cmd.args, func(s string) bool { return s== ">>" || s=="1>>"}); idx != -1 {
					handleRedirection(cmd, idx, cmd.args[idx], os.O_APPEND)
				} else if idx := slices.Index(cmd.args,"2>>"); idx != -1 {
					handleRedirection(cmd, idx, cmd.args[idx], os.O_APPEND)
				} else {
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
}
