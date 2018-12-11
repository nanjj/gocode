package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const (
	stamblerre = "stamblerre"
	mdempsky   = "mdempsky"
)

var (
	self, _   = os.Executable()
	needExit  = false
	needDebug = false
)

func Go111ModuleOn() (on bool) {
	output := strings.TrimSpace(OutputCommand("go", "env", "GOMOD"))
	on = strings.HasSuffix(output, "go.mod")
	return
}

func ProcStarted(command string) bool {
	cmd := exec.Command("killall", "-0", command)
	if err := cmd.Run(); err == nil {
		return true
	}
	return false
}

func OutputCommand(command string, args ...string) (output string) {
	cmd := exec.Command(command, args...)
	if b, err := cmd.Output(); err == nil {
		output = string(b)
	}
	return
}

func RunCommand(command string, args ...string) {
	if needDebug {
		log.Printf("Run command %s %v", command, args)
	}
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func StartProc(command string, args ...string) {
	if needDebug {
		logname := fmt.Sprintf("/tmp/%s.log", path.Base(command))
		args = append([]string{command}, args...)
		args = append(args, "&>", logname)
		command = "/bin/bash"
		args = []string{"-c", strings.Join(args, " ")}
	}
	args = append([]string{command}, args...)
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var stdin, stdout, stderr *os.File
	stdin, err = os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}
	stdout, err = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}
	stderr = stdout
	attrs := os.ProcAttr{
		Dir:   cwd,
		Env:   os.Environ(),
		Files: []*os.File{stdin, stdout, stderr},
	}
	p, err := os.StartProcess(command, args, &attrs)
	if err != nil {
		panic(err)
	}
	if err = p.Release(); err != nil {
		panic(err)
	}
}

func FileExists(p string) bool {
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		return true
	}
	return false
}

func main() {
	now := time.Now()
	defer func() {
		if needDebug {
			log.Printf("Elapsed duration: %v", time.Since(now))
		}
	}()
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "exit" {
			needExit = true
		} else if arg == "-debug" {
			needDebug = true
		}
	}
	// check and install
	for _, name := range []string{mdempsky, stamblerre} {
		output := fmt.Sprintf("%s.%s", self, name)
		pkg := fmt.Sprintf("github.com/%s/gocode", name)
		if !FileExists(output) {
			OutputCommand("go", "get", "-d", pkg)
			OutputCommand("go", "build", "-o", output, pkg)
		}
		if needExit {
			OutputCommand(fmt.Sprintf("%s.%s", self, name), "exit")
		} else {
			if !ProcStarted(fmt.Sprintf("gocode.%s", name)) {
				sArgs := []string{"-s", "-sock", "unix", "-cache"}
				if needDebug {
					sArgs = append(sArgs, "-debug")
				}
				StartProc(output, sArgs...)
			}
		}
	}
	if needExit {
		return
	}
	mod := Go111ModuleOn()
	current := mdempsky
	if mod {
		current = stamblerre
	}
	RunCommand(fmt.Sprintf("%s.%s", self, current), args...)
}
