package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"launchpad.net/gnuflag"
	"launchpad.net/juju/go/log"
	"os"
	"path/filepath"
)

// Context represents the run context of a Command. Command implementations
// should interpret file names relative to Dir (see AbsPath below), and print
// output and errors to Stdout and Stderr respectively.
type Context struct {
	Dir    string
	Stdout io.Writer
	Stderr io.Writer
}

// DefaultContext returns a Context suitable for use in non-hosted situations.
func DefaultContext() *Context {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	return &Context{abs, os.Stdout, os.Stderr}
}

// AbsPath returns an absolute representation of path, with relative paths
// interpreted as relative to ctx.Dir.
func (ctx *Context) AbsPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(ctx.Dir, path)
}

// Main runs the given Command in the supplied Context with the given
// arguments, which should not include the command name. It returns a code
// suitable for passing to os.Exit.
func Main(c Command, ctx *Context, args []string) int {
	var err error
	printErr := func() { fmt.Fprintf(ctx.Stderr, "ERROR: %v\n", err) }

	f := gnuflag.NewFlagSet(c.Info().Name, gnuflag.ContinueOnError)
	f.Usage = func() {}
	f.SetOutput(ioutil.Discard)
	printHelp := func() { c.Info().printHelp(ctx.Stderr, f) }

	switch err = c.Init(f, args); err {
	case nil:
		if err = c.Run(ctx); err != nil {
			log.Debugf("%s command failed: %s\n", c.Info().Name, err)
			printErr()
			return 1
		}
	case gnuflag.ErrHelp:
		printHelp()
	default:
		printErr()
		printHelp()
		return 2
	}
	return 0
}
