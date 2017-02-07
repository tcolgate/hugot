package command

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/tcolgate/hugot"
)

// Set assists with supporting command handlers with sub-commands.
// A set can also be used as the initial ToBot handler of a Mux to
// provide users an set of interactive CLI style commands.
type Set map[string]Commander

// NewSet creates an empty commands set.
func NewSet(ics ...Commander) Set {
	cs := map[string]Commander{}
	Set(cs).MustAdd(ics...)
	return cs
}

// Describe implmenets the describer methods, this is not
// really needed.
func (s Set) Describe() (string, string) {
	return "", ""
}

// Add adds a Command to a Set
func (s Set) Add(cs ...Commander) error {
	for _, c := range cs {
		if c == nil {
			panic(errors.New("Attempt to add null handler to set"))
		}
		n, _ := c.Describe()
		if _, ok := s[n]; ok {
			return fmt.Errorf("Duplicate command %s", n)
		}
		s[n] = c
	}
	return nil
}

// MustAdd adds Commands to a set, panics if any fail
func (s Set) MustAdd(cs ...Commander) error {
	if err := s.Add(cs...); err != nil {
		panic(err)
	}
	return nil
}

type byAlpha struct {
	ns  []string
	ds  []string
	chs []Commander
}

func (b *byAlpha) Len() int           { return len(b.ns) }
func (b *byAlpha) Less(i, j int) bool { return b.ns[i] < b.ns[j] }
func (b *byAlpha) Swap(i, j int) {
	b.ns[i], b.ns[j] = b.ns[j], b.ns[i]
	b.ds[i], b.ds[j] = b.ds[j], b.ds[i]
	b.chs[i], b.chs[j] = b.chs[j], b.chs[i]
}

// List returns the names and usage of the subcommands of
// a Set.
func (s Set) List() ([]string, []string, []Commander) {
	cmds := []string{}
	descs := []string{}
	chs := []Commander{}
	hasHelp := false

	for _, ch := range s {
		n, d := ch.Describe()
		if n == "help" {
			hasHelp = true
			continue
		}
		cmds = append(cmds, n)
		descs = append(descs, d)
		chs = append(chs, ch)
	}

	sorted := &byAlpha{cmds, descs, chs}
	sort.Sort(sorted)
	if hasHelp {
		hh := s["help"]
		_, hd := hh.Describe()
		sorted.ns = append([]string{"help"}, sorted.ns...)
		sorted.ds = append([]string{hd}, sorted.ds...)
		sorted.chs = append([]Commander{hh}, sorted.chs...)
	}

	return sorted.ns, sorted.ds, sorted.chs
}

// NextCommand picks the next commands to run from this command set based on the content
// of the message
func (s Set) NextCommand(m *Message) (Commander, error) {
	// This is repeated from RunCommandHandler, probably something wrong there
	if len(m.Args()) == 0 {
		cmds, _, _ := s.List()
		return nil, fmt.Errorf("required sub-command missing: %s", strings.Join(cmds, ", "))
	}

	matches := []Commander{}
	matchesns := []string{}
	ematches := []Commander{}
	for name, cmd := range s {
		if strings.HasPrefix(name, m.args[0]) {
			matches = append(matches, cmd)
			matchesns = append(matchesns, name)
		}
		if name == m.args[0] {
			ematches = append(ematches, cmd)
		}
	}

	switch {
	case len(matches) == 0 && len(ematches) == 0:
		return nil, ErrUnknownCommand
	case len(ematches) > 1:
		return nil, fmt.Errorf("multiple exact matches for %s", m.args[0])

	case len(ematches) == 1:
		return ematches[0], nil
	case len(matches) == 1:
		return matches[0], nil

	default:
		return nil, fmt.Errorf("ambigious command, %s: %s", m.args[0], strings.Join(matchesns, ", "))
	}
}

// Command executes the next command in the command set.
func (s Set) Command(ctx context.Context, w hugot.ResponseWriter, m *Message) error {
	ch, err := s.NextCommand(m)
	if err != nil {
		return err
	}
	hn, _ := ch.Describe()
	m.FlagSet = flag.NewFlagSet(hn, flag.ContinueOnError)
	m.FlagOut = &bytes.Buffer{}
	m.FlagSet.SetOutput(m.FlagOut)
	return ch.Command(ctx, w, m)
}

//ProcessMessage allows a CommandSet to be used as a basic hugot.handler
func (s Set) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, hm *hugot.Message) error {
	m := &Message{Message: hm}

	ch, err := s.NextCommand(m)
	if err != nil {
		return err
	}
	hn, _ := ch.Describe()
	m.FlagSet = flag.NewFlagSet(hn, flag.ContinueOnError)
	m.FlagOut = &bytes.Buffer{}
	m.FlagSet.SetOutput(m.FlagOut)

	err = ch.Command(ctx, w, m)
	if len(m.FlagOut.Bytes()) > 0 {
		return ErrUsage(string(m.FlagOut.String()))
	}
	return err
}

// Helper defines an interface for handlers that wish to be able to
// provide help via the help command (see handlers/help)
type Helper interface {
	Help(ctx context.Context, w io.Writer, m *Message) error
}

// Help passes the help request on to the next suitable command in the Set.
// The argument should be a unique prefix of one of the Set's commands.
func (s Set) Help(ctx context.Context, w io.Writer, m *Message) error {
	if len(m.Args()) != 0 {
		ch, err := s.NextCommand(m)
		if err != nil {
			return err
		}
		hn, _ := ch.Describe()
		m.FlagSet = flag.NewFlagSet(hn, flag.ContinueOnError)
		m.FlagOut = &bytes.Buffer{}
		m.FlagSet.SetOutput(m.FlagOut)
		m.SetArgs(append(m.Args(), "-h"))
		rw := hugot.NewNullResponseWriter(*m.Message)
		ch.Command(ctx, rw, m)
		return nil
	}

	hns, descs, _ := s.List()
	tw := new(tabwriter.Writer)
	tw.Init(w, 0, 8, 1, '\t', 0)

	if len(hns) > 0 {
		fmt.Fprintln(tw, "Commands:")
		for i := range hns {
			fmt.Fprintf(tw, "  %s\t - %s\n", hns[i], descs[i])
		}
		tw.Flush()
	}
	return nil
}
