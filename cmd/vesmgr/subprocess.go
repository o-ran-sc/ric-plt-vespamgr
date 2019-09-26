package main

import (
	"os"
	"os/exec"
)

type cmdRunner struct {
	exe  string
	args []string
	cmd  *exec.Cmd
}

func (r *cmdRunner) run(result chan error) {
	r.cmd = exec.Command(r.exe, r.args...)
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
	err := r.cmd.Start()
	go func() {
		if err != nil {
			result <- err
		} else {
			result <- r.cmd.Wait()
		}
	}()
}

func (r *cmdRunner) kill() error {
	return r.cmd.Process.Kill()
}

func makeRunner(exe string, arg ...string) cmdRunner {
	r := cmdRunner{exe: exe, args: arg}
	return r
}
