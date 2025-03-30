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
	"sort"
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

func readInput(ioReader io.Reader) (input string){
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error enabling raw mode: ", err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	currPos := -1
	reader := bufio.NewReader(ioReader)
	tab_flag := false
	loop: 
		for {
			char ,_,err := reader.ReadRune()
			if err != nil {
				fmt.Println(err)
				break loop
			}

			switch char{
				case '\x03': // Ctrl+C
					os.Exit(0)

				case '\r', '\n': // Enter
					fmt.Print("\r\n")
					break loop
				
				case '\b', 127: //Backspace
					if currPos >= 0 {
						input = input[:currPos] + input[currPos+1:]
						currPos-=1
						fmt.Print("\b \b")
					}

				case '\t': 
					suffixes:= autocomplete(input)
					if len(suffixes) == 1 {
						suffix := suffixes[0]
						if suffix != "" {
							input += suffix + " "
							fmt.Print(suffix+" ")
							currPos = len(input)-1
						}
					} else if len(suffixes) > 1 {
						if tab_flag {
							sort.Strings(suffixes)
							for i,suffix := range(suffixes){
								suffixes[i] = input+suffix
							}
							fmt.Print("\r\n"+strings.Join(suffixes,"  ") + "\r\n$ "+input)
							tab_flag = false
						} else {
							sort.Strings(suffixes)
							first_suffix := suffixes[0]
							last_suffix := suffixes[len(suffixes)-1]
							min_len := min(len(first_suffix),len(last_suffix))
							common_suffix_index := 0
							for common_suffix_index<min_len && first_suffix[common_suffix_index] == last_suffix[common_suffix_index] {
								common_suffix_index += 1
							}

							input += first_suffix[:common_suffix_index]
							fmt.Print(first_suffix[:common_suffix_index])
							currPos = len(input)-1
							tab_flag = true
						}
					} else {
						fmt.Print("\a")
					}

				case 27:
					// todo : arrow
				
				default:
					input += string(char)
					fmt.Print(string(char))
					currPos += 1

			}
		}
		
		return input
}

func autocomplete(prefix string) (suffixes []string) {
	if prefix == "" {
		return
	}
	for _, v := range builtinCMDs {
		after, found := strings.CutPrefix(v, prefix)
		if found && !slices.Contains(suffixes,after){
			suffixes = append(suffixes, after)
		}
	}
	path := os.Getenv("PATH")
	directories := strings.Split(path, ":")
	for _, directory := range directories {
		files, err := os.ReadDir(directory)
		if err == nil {
			for _, file := range files {
				if file.IsDir() {
					continue
				} else{
					after, found := strings.CutPrefix(file.Name(), prefix)
					if found && !slices.Contains(suffixes,after){
						suffixes = append(suffixes, after)
					}
				}
			}
		}
	}
	return suffixes
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
				if slices.Contains(builtinCMDs, arg){
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
