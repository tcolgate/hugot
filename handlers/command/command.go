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

type CmdRunFunc func(cmd *Command, w hugot.ResponseWriter, msg *hugot.Message, args []string) error
type cobraFuncE func(*cobra.Command, []string) error

type Command struct {
	Use     string
	Short   string
	Long    string
	Example string

	PersistentPreRun  CmdRunFunc
	PreRun            CmdRunFunc
	Run               CmdRunFunc
	PostRun           CmdRunFunc
	PersistentPostRun CmdRunFunc
	SilenceErrors     bool
	SilenceUsage      bool

	flags  *pflag.FlagSet
	pflags *pflag.FlagSet

	cob         *cobra.Command
	subcommands []*Command
}

func (cmd *Command) Flags() *pflag.FlagSet {
	if cmd.flags == nil {
		cob := &cobra.Command{Use: cmd.Use}
		cmd.flags = pflag.NewFlagSet(cob.Name(), pflag.ContinueOnError)
	}
	return cmd.flags
}

func (cmd *Command) PersistentFlags() *pflag.FlagSet {
	if cmd.pflags == nil {
		cob := &cobra.Command{Use: cmd.Use}
		cmd.pflags = pflag.NewFlagSet(cob.Name(), pflag.ContinueOnError)
	}
	return cmd.pflags
}

func (cmd *Command) AddCommand(scmd *Command) {
	cmd.subcommands = append(cmd.subcommands, scmd)
}

type Handler struct {
	CommandSetupper
}

type CommandSetupper interface {
	CommandSetup(*Command) error
}

type CommandSet map[string]*Handler

func (h CommandSet) Describe() (string, string) {
	return "commands", "blah"
}

func (h CommandSet) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
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
	for n, h := range h {
		if strings.HasPrefix(n, args[0]) {
			names = append(names, n)
			matches = append(matches, h)
		}
	}

	if len(names) == 0 {
		return ErrUnknownCommand
	}

	if len(names) > 1 {
		return fmt.Errorf("Ambigious commands, pick: ", strings.Join(names, ", "))
	}

	m.Text = strings.Join(args[1:], " ")
	return matches[0].ProcessMessage(ctx, w, m)
}

func (cs CommandSet) MustAdd(c CommandSetupper) {
	root := Command{}
	c.CommandSetup(&root)
	cob := cmdToCobra(&root, nil, nil)
	cs[cob.Name()] = &Handler{c}
}

func (cs CommandSet) Help(w io.Writer) error {
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
		root := Command{}
		err := cs[cn].CommandSetup(&root)
		if err != nil {
			continue
		}
		cob := cmdToCobra(&root, nil, nil)

		fmt.Fprintf(tw, "  %s\t - %s\n", cob.Name(), cob.Short)
	}
	tw.Flush()

	io.Copy(w, out)
	fmt.Fprintln(w)

	return nil
}

func (h *Handler) Describe() (string, string) {
	root := Command{}
	h.CommandSetup(&root)
	cob := cmdToCobra(&root, nil, nil)

	return cob.Name(), cob.Short
}

func New(f CommandSetupper) *Handler {
	return &Handler{f}
}

func NewFunc(f func(cmd *Command) error) *Handler {
	return &Handler{CommandSetupFunc(f)}
}

type CommandSetupFunc func(*Command) error

func (f CommandSetupFunc) CommandSetup(cmd *Command) error {
	return f(cmd)
}

func (h *Handler) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	var err error

	root := Command{cob: &cobra.Command{}}
	h.CommandSetup(&root)

	args, err := shellwords.Parse(m.Text)
	if err != nil {
		return ErrBadCLI
	}

	cob := cmdToCobra(&root, w, m)
	cob.SetOutput(w)
	cob.SetArgs(args)

	for _, c := range root.subcommands {
		cob.AddCommand(cmdToCobra(c, w, m))
	}

	return cob.Execute()
}

func (h *Handler) Help(w io.Writer) error {
	root := Command{}
	h.CommandSetup(&root)

	cob := cmdToCobra(&root, nil, nil)
	cob.SetOutput(w)

	return root.cob.Help()
}

func cmdToCobra(cmd *Command, w hugot.ResponseWriter, msg *hugot.Message) *cobra.Command {
	cob := &cobra.Command{
		Use:           cmd.Use,
		Short:         cmd.Short,
		Long:          cmd.Long,
		Example:       cmd.Example,
		SilenceErrors: cmd.SilenceErrors,
		SilenceUsage:  cmd.SilenceUsage,
	}
	cob.PersistentPreRunE = cmd.PersistentPreRun.makeCobraRunEFunc(cmd, w, msg)
	cob.PreRunE = cmd.PreRun.makeCobraRunEFunc(cmd, w, msg)
	cob.RunE = cmd.Run.makeCobraRunEFunc(cmd, w, msg)
	cob.PostRunE = cmd.PostRun.makeCobraRunEFunc(cmd, w, msg)
	cob.PersistentPostRunE = cmd.PersistentPostRun.makeCobraRunEFunc(cmd, w, msg)
	cob.Flags().AddFlagSet(cmd.flags)
	cob.PersistentFlags().AddFlagSet(cmd.pflags)

	return cob
}

func (cf CmdRunFunc) makeCobraRunEFunc(cmd *Command, w hugot.ResponseWriter, msg *hugot.Message) cobraFuncE {
	if cf == nil {
		return nil
	}
	return func(cob *cobra.Command, args []string) error {
		cob.SetOutput(w)
		return cf(cmd, w, msg, args)
	}
}
