package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/kballard/go-shellquote"
)

var (
	// isWindows determines if running on Windows or *nix
	isWindows = runtime.GOOS == "windows"
)

// Cmd represents an OS command
type Cmd struct {
	// Name is the name of the command
	Name string
	// Args are any args to pass to the command
	Args []string
}

// Copy copies a string to the OS clipboard
func Copy(data string) error {
	var cpCmd string
	if isWindows {
		cpCmd = "clip"
	} else {
		cpCmd = "pbcopy"
	}

	echo := New("echo").WithArgs(data)
	copy := New(cpCmd)
	_, _, err := Pipeline(echo, copy)
	return err
}

// Open opens a location
// OSX only
func Open(location string) error {
	return New("open").WithArgs(location).Run()
}

// String prints the command as a string
func (cmd *Cmd) String() string {
	return fmt.Sprintf("%s %s", cmd.Name, strings.Join(cmd.Args, " "))
}

// WithArgs adds arguments to the current command
func (cmd *Cmd) WithArgs(args ...string) *Cmd {
	for _, arg := range args {
		cmd.Args = append(cmd.Args, arg)
	}
	return cmd
}

// CombinedOutput runs the command and returns its combined standard output and standard error.
func (cmd *Cmd) CombinedOutput() (string, error) {
	output, err := exec.Command(cmd.Name, cmd.Args...).CombinedOutput()
	return string(output), err
}

// Run runs command with `Exec` on platforms except Windows
// which only supports `Spawn`
func (cmd *Cmd) Run() error {
	if isWindows {
		return cmd.Spawn()
	}
	return cmd.Exec()
}

// Spawn runs command with spawn(3)
func (cmd *Cmd) Spawn() error {
	c := exec.Command(cmd.Name, cmd.Args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

// Exec runs command with exec(3)
// Note that Windows doesn't support exec(3): http://golang.org/src/pkg/syscall/exec_windows.go#L339
func (cmd *Cmd) Exec() error {
	binary, err := exec.LookPath(cmd.Name)
	if err != nil {
		return fmt.Errorf("command not found: %s", cmd.Name)
	}

	args := []string{binary}
	args = append(args, cmd.Args...)

	return syscall.Exec(binary, args, os.Environ())
}

// Pipeline runs a series of commands, passing the output of one command to the
// input of the next
func Pipeline(list ...*Cmd) (string, string, error) {
	var output bytes.Buffer
	var stderr bytes.Buffer
	var err error

	// Require at least one command
	if len(list) < 1 {
		return "", "", nil
	}

	var newCmd *exec.Cmd
	cmds := make([]*exec.Cmd, 0, 4)

	// Convert into an exec.Cmd
	for _, cmd := range list {
		newCmd = exec.Command(cmd.Name, cmd.Args...)
		cmds = append(cmds, newCmd)
	}

	// Collect the output from the command(s)
	last := len(cmds) - 1
	for i, cmd := range cmds[:last] {
		// Connect each command's stdin to the previous command's stdout
		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			return "", "", err
		}
		// Connect each command's stderr to a buffer
		cmd.Stderr = &stderr
	}

	// Connect the output and error for the last command
	cmds[last].Stdout, cmds[last].Stderr = &output, &stderr

	// Start each command
	for _, cmd := range cmds {
		if err = cmd.Start(); err != nil {
			return output.String(), stderr.String(), err
		}
	}

	// Wait for each command to complete
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return output.String(), stderr.String(), err
		}
	}

	// Return the pipeline output and the collected standard error
	return output.String(), stderr.String(), nil
}

// New creates a new Cmd instance
func New(cmd string) *Cmd {
	var args []string
	cmds, err := shellquote.Split(cmd)
	if err != nil {
		panic(err)
	}

	name := cmds[0]
	for _, arg := range cmds[1:] {
		args = append(args, arg)
	}
	return &Cmd{Name: name, Args: args}
}
