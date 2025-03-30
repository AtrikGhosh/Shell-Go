// TO DO - OPTIMIZEEEEE

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/term"
)

var builtinCMDs = []string{
	"exit",
	"echo",
	"type",
	"pwd",
	"cd",
}

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
	file, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|writeMode, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening destination file:", err)
		return
	}
	defer file.Close()
	command := exec.Command(cmd.name, finalArgs...)
	if redirectType == "stdout" {
		command.Stdout = file
		command.Stderr = os.Stderr
	} else if redirectType == "stderr" {
		command.Stdout = os.Stdout
		command.Stderr = file
	}
	command.Run()
}

func handleEchoRedirection(args []string,  redirectIdx int, redirectType string, writeMode int){
	text,filepath := args[:redirectIdx],args[redirectIdx+1]
	file,err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|writeMode, 0644)
	if err != nil {
		fmt.Println("Error opening destination file:", err)
	}
	defer file.Close()
	if redirectType == "stdout" {
		_,err = file.WriteString(strings.Trim(strings.Join(text," "),"\r\n")+"\n")
		if err != nil {
			fmt.Println("Error:", err)
		}
	} else if redirectType == "stderr" {
		_,err = fmt.Println(strings.Join(text," "))
		if err != nil {
			file.WriteString(err.Error())
		}
		file.Close()
	}
}

func readInput(rd io.Reader) (input string) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	r := bufio.NewReader(rd)
	
loop:
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			fmt.Println(err)
			continue
		}
		switch c {
		case '\x03': // Ctrl+C
			os.Exit(0)
		case '\r', '\n': // Enter
			fmt.Fprint(os.Stdout, "\r\n")
			break loop
		case '\x7F': // Backspace
			if length := len(input); length > 0 {
				input = input[:length-1]
				fmt.Fprint(os.Stdout, "\b \b")
			}
		case '\t': // Tab
			suffix := autocomplete(input)
			if suffix != "" {
				input += suffix + " "
				fmt.Fprint(os.Stdout, suffix+" ")
			}
		default:
			input += string(c)
			fmt.Fprint(os.Stdout, string(c))
		}
	}
	return
}

func autocomplete(prefix string) (suffix string) {
	if prefix == "" {
		return
	}
	suffixes := []string{}
	for _, v := range builtinCMDs {
		after, found := strings.CutPrefix(v, prefix)
		if found {
			suffixes = append(suffixes, after)
		}
	}
	if len(suffixes) == 1 {
		return suffixes[0]
	}
	return
}

func main() {
	

	for{
		
		fmt.Print("$ ")
		input := readInput(os.Stdin)
		cmd := parseCmd(strings.TrimSpace(input))
		
		switch cmd.name{
			case "exit":
				os.Exit(0)
			case "echo":
				if idx := slices.IndexFunc(cmd.args,func(s string) bool {return s == ">" || s == "1>"}); idx != -1 {
					handleEchoRedirection(cmd.args,idx,"stdout",os.O_TRUNC)
				} else if idx := slices.Index(cmd.args,"2>"); idx != -1 {
					handleEchoRedirection(cmd.args,idx,"stderr",os.O_TRUNC)
				} else if idx := slices.IndexFunc(cmd.args,func(s string) bool {return s == ">>" || s == "1>>"}); idx != -1 {
					handleEchoRedirection(cmd.args,idx,"stdout",os.O_APPEND)
				} else if idx := slices.Index(cmd.args,"2>>"); idx != -1 {
					handleEchoRedirection(cmd.args,idx,"stderr",os.O_APPEND)
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
					handleRedirection(cmd, idx, "stdout", os.O_TRUNC)
				} else if idx := slices.Index(cmd.args,"2>"); idx != -1 {
					handleRedirection(cmd, idx, "stderr", os.O_TRUNC)
				} else if idx := slices.IndexFunc(cmd.args, func(s string) bool { return s== ">>" || s=="1>>"}); idx != -1 {
					handleRedirection(cmd, idx, "stdout", os.O_APPEND)
				} else if idx := slices.Index(cmd.args,"2>>"); idx != -1 {
					handleRedirection(cmd, idx, "stderr", os.O_APPEND)
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
