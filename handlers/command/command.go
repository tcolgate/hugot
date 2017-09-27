// Package command provides CLI style interactive commands.
package command

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	shellwords "github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tcolgate/hugot"
)

type ctxKey int

const (
	ctxPathKey ctxKey = iota
)

func init() {
	cobra.EnablePrefixMatching = true
}

// Handler is responsible for executing a user command. Before the command
// is run the Setupper is called to build the new command. The command
// is the executed. This allows the setup stage to capture flag information
// that is needed when the command itself is execture
type Handler struct {
	Setupper
}

// New creates a handler from any CommnadSetupper
func New(f Setupper) *Handler {
	return &Handler{f}
}

// NewFunc creates a Handler from a SetupFunc
func NewFunc(f func(cmd *Command) error) *Handler {
	return &Handler{SetupFunc(f)}
}

// Setupper describes a Handler that can set up the context for a user
// executing a function. It should set up the provided function as required,
// and ensure the Run functions capture any required flag arguments
type Setupper interface {
	CommandSetup(*Command) error
}

// RunFunc is a specification for a function to be run in response to a user
// executing a command. msg is the original message, args is the set of arguments.
// And flags required should be setup during Setup dunction for the handler
type RunFunc func(ctx context.Context, w hugot.ResponseWriter, msg *hugot.Message, args []string) error
type cobraFuncE func(*cobra.Command, []string) error

// Set is a collection of command to be run by a mux.Mux
type Set map[string]*Handler

// Describe implements the hugot.Handler interface for a Set
func (cs Set) Describe() (string, string) {
	return "commands", "blah"
}

// ProcessMessage implements the hugot.Handler interface for a Set
func (cs Set) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	var err error
	var args []string

	if args, err = shellwords.Parse(m.Text); err != nil {
		args = strings.Split(m.Text, " ")
	}
	if len(args) == 0 {
		fmt.Fprintf(w, "What are you asking of me?")
		return nil
	}

	var names []string
	var matches []*Handler
	for n, h := range cs {
		if strings.HasPrefix(n, args[0]) {
			names = append(names, n)
			matches = append(matches, h)
		}
	}

	if len(names) == 0 {
		return ErrUnknownCommand
	}

	if len(names) > 1 {
		return fmt.Errorf("Ambigious commands, pick: %v", strings.Join(names, ", "))
	}

	m.Text = strings.Join(args[1:], " ")
	return matches[0].ProcessMessage(ctx, w, m)
}

// MustAdd adds a command to a Set
func (cs Set) MustAdd(c Setupper) {
	root := &Command{}
	c.CommandSetup(root)
	cob := root.cmdToCobra(context.TODO(), nil, nil)
	cs[cob.Name()] = &Handler{c}
}

// Help implements mux.Helper for the command.CommandSet
func (cs Set) Help(w io.Writer) error {
	out := &bytes.Buffer{}

	tw := new(tabwriter.Writer)
	tw.Init(out, 0, 8, 1, '\t', 0)

	if len(cs) == 0 {
		return nil
	}

	var cns []string

	for n := range cs {
		cns = append(cns, n)
	}
	sort.Strings(cns)

	fmt.Fprintf(out, "Commands:\n")
	for _, cn := range cns {
		root := &Command{}
		err := cs[cn].CommandSetup(root)
		if err != nil {
			continue
		}
		cob := root.cmdToCobra(context.TODO(), nil, nil)

		fmt.Fprintf(tw, "  %s\t - %s\n", cob.Name(), cob.Short)
	}
	tw.Flush()

	io.Copy(w, out)
	fmt.Fprintln(w)

	return nil
}

// Command describe a CLI type command that can be executed by a user.
// It a wrapper around an github.com/spf13/cobra command
type Command struct {
	Use     string
	Short   string
	Long    string
	Example string

	PersistentPreRun  RunFunc
	PreRun            RunFunc
	Run               RunFunc
	PostRun           RunFunc
	PersistentPostRun RunFunc
	SilenceErrors     bool
	SilenceUsage      bool

	flags  *pflag.FlagSet
	pflags *pflag.FlagSet

	cob         *cobra.Command
	subcommands []*Command
}

// Flags returns the active FlagSet for this command
func (cmd *Command) Flags() *pflag.FlagSet {
	if cmd.flags == nil {
		cob := &cobra.Command{Use: cmd.Use}
		cmd.flags = pflag.NewFlagSet(cob.Name(), pflag.ContinueOnError)
	}
	return cmd.flags
}

// PersistentFlags returns the persistent  FlagSet for this command
func (cmd *Command) PersistentFlags() *pflag.FlagSet {
	if cmd.pflags == nil {
		cob := &cobra.Command{Use: cmd.Use}
		cmd.pflags = pflag.NewFlagSet(cob.Name(), pflag.ContinueOnError)
	}
	return cmd.pflags
}

// AddCommand adds a new subcommand to this command
func (cmd *Command) AddCommand(scmd *Command) {
	cmd.subcommands = append(cmd.subcommands, scmd)
}

// SetupFunc takes a Command and is expected to configure it
// to provide some command functionality
type SetupFunc func(*Command) error

// CommandSetup you to use a CommandSetupFunc directly as a Handler
func (f SetupFunc) CommandSetup(cmd *Command) error {
	return f(cmd)
}

// Describe implements hugot.Handler for the command handler
func (h *Handler) Describe() (string, string) {
	root := &Command{}
	h.CommandSetup(root)
	cob := root.cmdToCobra(context.TODO(), nil, nil)

	return cob.Name(), cob.Short
}

// ProcessMessage implements hugot.Handler for the command handler
func (h *Handler) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	var err error

	root := &Command{cob: &cobra.Command{}}
	h.CommandSetup(root)

	args, err := shellwords.Parse(m.Text)
	if err != nil {
		return ErrBadCLI
	}

	cob := root.cmdToCobra(ctx, w, m)
	cob.SetOutput(w)
	cob.SetArgs(args)

	for _, c := range root.subcommands {
		cob.AddCommand(c.cmdToCobra(ctx, w, m))
	}

	return cob.Execute()
}

// Help implements mux.Helper for the command.Handler
func (h *Handler) Help(w io.Writer) error {
	root := &Command{}
	h.CommandSetup(root)

	cob := root.cmdToCobra(context.TODO(), nil, nil)
	cob.SetOutput(w)

	return root.cob.Help()
}

func (cmd *Command) cmdToCobra(ctx context.Context, w hugot.ResponseWriter, msg *hugot.Message) *cobra.Command {
	cob := &cobra.Command{
		Use:           cmd.Use,
		Short:         cmd.Short,
		Long:          cmd.Long,
		Example:       cmd.Example,
		SilenceErrors: cmd.SilenceErrors,
		SilenceUsage:  cmd.SilenceUsage,
	}
	cob.PersistentPreRunE = cmd.PersistentPreRun.makeCobraRunEFunc(ctx, cmd, w, msg)
	cob.PreRunE = cmd.PreRun.makeCobraRunEFunc(ctx, cmd, w, msg)
	cob.RunE = cmd.Run.makeCobraRunEFunc(ctx, cmd, w, msg)
	cob.PostRunE = cmd.PostRun.makeCobraRunEFunc(ctx, cmd, w, msg)
	cob.PersistentPostRunE = cmd.PersistentPostRun.makeCobraRunEFunc(ctx, cmd, w, msg)
	cob.Flags().AddFlagSet(cmd.flags)
	cob.PersistentFlags().AddFlagSet(cmd.pflags)

	return cob
}

type cmdContextKey string

// FromContext retrieves the Command the caused the calling function be be called
func FromContext(ctx context.Context) *Command {
	return ctx.Value(cmdContextKey("command")).(*Command)
}

func (cf RunFunc) makeCobraRunEFunc(ctx context.Context, cmd *Command, w hugot.ResponseWriter, msg *hugot.Message) cobraFuncE {
	if cf == nil {
		return nil
	}
	ctx = context.WithValue(ctx, cmdContextKey("command"), cmd)
	return func(cob *cobra.Command, args []string) error {
		cob.SetOutput(w)
		return cf(ctx, w, msg, args)
	}
}
