// Copyright (c) 2016 - 2019 Sqreen. All Rights Reserved.
// Please refer to our terms for more information:
// https://www.sqreen.io/terms.html

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var globalFlags instrumentationToolFlagSet

func main() {
	args := os.Args[1:]
	cmd, cmdArgPos, _ := parseCommand(&globalFlags, args)

	// Hide instrumentation tool arguments
	if cmdArgPos != -1 {
		args = args[cmdArgPos:]
	}

	if cmd != nil {
		newArgs, _ := cmd()
		if newArgs != nil {
			// Args are replaced
			args = newArgs
		}
	}
	forwardCommand(args)
	os.Exit(0)
}

// forwardCommand runs the given command's argument list and exits the process
// with the exit code that was returned.
func forwardCommand(args []string) error {
	path := args[0]
	args = args[1:]
	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	//quotedArgs := fmt.Sprintf("%+q", args)
	//log.Printf("forwarding command `%s %s`", path, quotedArgs[1:len(quotedArgs)-1])
	return cmd.Run()
}

type parseCommandFunc func([]string) (commandExecutionFunc, error)
type commandExecutionFunc func() (newArgs []string, err error)

var commandParserMap = map[string]parseCommandFunc{
	"compile": parseCompileCommand,
}

// getCommand returns the command and arguments. The command is expectedFlags to be
// the first argument.
func parseCommand(instrToolFlagSet *instrumentationToolFlagSet, args []string) (commandExecutionFunc, int, error) {
	cmdIdPos := parseFlagsUntilFirstNonOptionArg(instrToolFlagSet, args)
	if cmdIdPos == -1 {
		return nil, cmdIdPos, errors.New("unexpected arguments")
	}
	cmdId := args[cmdIdPos]
	args = args[cmdIdPos:]

	cmdId, err := parseCommandID(cmdId)
	if err != nil {
		return nil, cmdIdPos, err
	}
	//fmt.Printf("CommandId is %v and cmdIdPos is %v.\n", cmdId, cmdIdPos)
	if commandParser, exists := commandParserMap[cmdId]; exists {
		//fmt.Printf("CommandParser is %v, Args are %v.\n", commandParser, args)
		cmd, err := commandParser(args)
		return cmd, cmdIdPos, err
	} else {
		return nil, cmdIdPos, nil
	}
}

func parseCommandID(cmd string) (string, error) {
	// It mustn't be empty
	if cmd == "" {
		return "", errors.New("unexpected empty command name")
	}

	// Take the base of the absolute path of the go tool
	cmd = filepath.Base(cmd)
	// Remove the file extension if any
	if ext := filepath.Ext(cmd); ext != "" {
		cmd = strings.TrimSuffix(cmd, ext)
	}
	return cmd, nil
}
