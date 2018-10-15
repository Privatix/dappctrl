// +build linux darwin

package exec

import (
	"io"
	"os/exec"
	"syscall"
)

// Process is a process, which allows to kill itself together with its children.
type Process struct {
	Stderr io.ReadCloser
	cmd    *exec.Cmd
}

// StartProcess starts a new process with a given name and arguments.
func StartProcess(name string, arg ...string) (*Process, error) {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Process{stderr, cmd}, nil
}

// Close closes the instance without killing the underlying process.
func (p *Process) Close() {
	p.Stderr.Close()
}

// Wait waits for the process to exit.
func (p *Process) Wait() error {
	return p.cmd.Wait()
}

// Kill kills the process.
func (p *Process) Kill() error {
	return syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
}
