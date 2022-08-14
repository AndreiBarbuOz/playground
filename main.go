package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

type IOStreams struct {
	// In provider of the io.Read method. Usually os.Stdin
	In io.Reader
	// Out provider of the io.Write method. Usually os.Stdout
	Out io.Writer
	// Err provider of the io.Write method. Usually os.Stderr
	Err io.Writer
}

// NewTestIOStreams returns a valid IOStreams and in, out, err buffers for unit tests
func NewTestIOStreams() (IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}

	return IOStreams{
		In:  in,
		Out: out,
		Err: err,
	}, in, out, err
}

func NewIOCommand(ioStreams IOStreams) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "io",
		Short: "Debug IO interactions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ioStreams)
		},
		SilenceUsage: true,
	}

	return cmd
}

func run(ioStreams IOStreams) error {
	ioStreams.Out.Write([]byte("input y/n: "))

	d, _ := bufio.NewReader(ioStreams.In).ReadString('\n')

	userInput := strings.TrimSpace(d)

	switch strings.ToLower(userInput) {
	case "y":
		fmt.Fprintln(ioStreams.Out, "returning y")
	case "n":
		fmt.Fprintln(ioStreams.Out, "returning n")
	}

	return nil
}

func main() {
	var root *cobra.Command

	ioStreams := IOStreams{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}

	root = NewIOCommand(ioStreams)
	err := root.Execute()
	if err != nil {
		os.Exit(1)
	}

}
