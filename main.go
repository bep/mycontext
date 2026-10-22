// Copyright 2026 Bjørn Erik Pedersen
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/bep/simplecobra"
	"github.com/spf13/cobra"
)

func main() {
	if err := run(context.Background(), os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer, args []string) error {
	x, err := simplecobra.New(newRootCommand(out))
	if err != nil {
		return err
	}
	_, err = x.Execute(ctx, args)
	return err
}

type rootCommand struct {
	s        *store
	commands []simplecobra.Commander
}

func newRootCommand(out io.Writer) *rootCommand {
	r := &rootCommand{}
	var create, local, dryRun bool

	r.commands = []simplecobra.Commander{
		&simpleCommand{
			name:  "path",
			short: "Print the context directory for the current branch",
			args:  cobra.NoArgs,
			withc: func(cmd *cobra.Command) {
				cmd.Flags().BoolVar(&create, "create", false, "create the context if it does not exist")
				cmd.Flags().BoolVar(&local, "local", false, "print the machine-local (not committed) subdirectory")
			},
			run: func(args []string) error {
				dir, err := r.s.path(create, local)
				if err != nil {
					return err
				}
				fmt.Fprintln(out, dir)
				return nil
			},
		},
		&simpleCommand{
			name:  "ls",
			short: "List the contexts of the current repository",
			args:  cobra.NoArgs,
			run: func(args []string) error {
				return r.s.list(out)
			},
		},
		&simpleCommand{
			name:  "rm",
			use:   "rm [branch]",
			short: "Remove a context (defaults to the current branch); the Git branch is left alone",
			args:  cobra.MaximumNArgs(1),
			run: func(args []string) error {
				if len(args) == 1 {
					return r.s.remove(args[0])
				}
				branch, err := r.s.currentBranch()
				if err != nil {
					return err
				}
				return r.s.remove(branch)
			},
		},
		&simpleCommand{
			name:  "mv",
			use:   "mv <old> <new>",
			short: "Rename a context",
			args:  cobra.ExactArgs(2),
			run: func(args []string) error {
				return r.s.move(args[0], args[1])
			},
		},
		&simpleCommand{
			name:  "prune",
			short: "Remove contexts for local branches that no longer exist",
			args:  cobra.NoArgs,
			withc: func(cmd *cobra.Command) {
				cmd.Flags().BoolVar(&dryRun, "dry-run", false, "only print what would be removed")
			},
			run: func(args []string) error {
				return r.s.prune(dryRun, out)
			},
		},
	}
	return r
}

func (r *rootCommand) Name() string { return "mycontext" }

func (r *rootCommand) Init(cd *simplecobra.Commandeer) error {
	cd.CobraCommand.Short = "Manage per repository/branch contexts in ../.mycontext"
	return nil
}

func (r *rootCommand) PreRun(this, runner *simplecobra.Commandeer) error {
	if this == runner {
		return nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	r.s, err = newStore(wd)
	return err
}

func (r *rootCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	return cd.CobraCommand.Help()
}

func (r *rootCommand) Commands() []simplecobra.Commander { return r.commands }

type simpleCommand struct {
	name  string
	use   string
	short string
	args  cobra.PositionalArgs
	withc func(cmd *cobra.Command)
	run   func(args []string) error
}

func (c *simpleCommand) Name() string { return c.name }

func (c *simpleCommand) Init(cd *simplecobra.Commandeer) error {
	cmd := cd.CobraCommand
	if c.use != "" {
		cmd.Use = c.use
	}
	cmd.Short = c.short
	cmd.Args = c.args
	if c.withc != nil {
		c.withc(cmd)
	}
	return nil
}

func (c *simpleCommand) PreRun(this, runner *simplecobra.Commandeer) error { return nil }

func (c *simpleCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	return c.run(args)
}

func (c *simpleCommand) Commands() []simplecobra.Commander { return nil }
