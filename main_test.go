package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Run(t *testing.T) {
	streams, in, out, _ := NewTestIOStreams()
	in.WriteString("y\n")
	run(streams)
	assert.Contains(t, out.String(), "returning y", "wrong output message")
	assert.Contains(t, out.String(), "input y/n:", "wrong prompt message")
}
