package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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
	version := OutputCommand("go", "version")
	if !strings.Contains(version, "go1.11") {
		return
	}
	if mod := os.Getenv("GO111MODULE"); mod == "on" {
		on = true
		return
	} else if mod == "off" {
		return
	}
	goPath := os.Getenv("GOPATH")
	pathes := strings.Split(goPath, ":")
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	for _, p := range pathes {
		if strings.HasPrefix(wd, p) {
			return
		}
	}
	on = true
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
	args = append([]string{command}, args...)
	cmd := exec.Command("nohup", args...)
	cmd.Start()
	cmd.Process.Release()
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
				StartProc(output, "-s", "-sock", "unix", "-cache")
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
