package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessRunning(t *testing.T) {
	r := makeRunner("echo", "a")
	ch := make(chan error)
	r.run(ch)
	err := <-ch
	assert.Nil(t, err)
}

func TestProcessKill(t *testing.T) {
	r := makeRunner("sleep", "20")
	ch := make(chan error)
	r.run(ch)
	assert.Nil(t, r.kill())
	<-ch // wait and seee that kills is actually done
}

func TestProcessRunningFails(t *testing.T) {
	r := makeRunner("foobarbaz")
	ch := make(chan error)
	r.run(ch)
	err := <-ch
	assert.NotNil(t, err)
}
