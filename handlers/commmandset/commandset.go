package commandset

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"sort"
	"strings"
)

var (
	// ErrUnknownCommand is returned by a command mux if the command did
	// not match any of it's registered handlers.
	ErrUnknownCommand = errors.New("unknown command")
)

// CommandPathFromContext returns the path used to get to
// this command handler
func CommandPathFromContext(ctx context.Context) []string {
	iv := ctx.Value(cmdPathKey)

	if iv == nil {
		return []string{}
	}

	v := iv.([]string)
	return v
}

// CommandSet assists with supporting command handlers with sub-commands.
type CommandSet map[string]command.Command

// NewCommandSet creates an empty commands set.
func NewCommandSet() *CommandSet {
	cs := make(CommandSet)
	return &cs
}

// AddCommandHandler adds a CommandHandler to a CommandSet
func (cs *CommandSet) AddCommandHandler(c CommandHandler) {
	n, _ := c.Describe()

	(*cs)[n] = c
}

type byAlpha struct {
	ns  []string
	ds  []string
	chs []CommandHandler
}

func (b *byAlpha) Len() int           { return len(b.ns) }
func (b *byAlpha) Less(i, j int) bool { return b.ns[i] < b.ns[j] }
func (b *byAlpha) Swap(i, j int) {
	b.ns[i], b.ns[j] = b.ns[j], b.ns[i]
	b.ds[i], b.ds[j] = b.ns[j], b.ds[i]
	b.chs[i], b.chs[j] = b.chs[j], b.chs[i]
}

// List returns the names and usage of the subcommands of
// a CommandSet.
func (cs *CommandSet) List() ([]string, []string, []CommandHandler) {
	cmds := []string{}
	descs := []string{}
	chs := []CommandHandler{}
	hasHelp := false

	for _, ch := range *cs {
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
		hh := (*cs)["help"]
		_, hd := hh.Describe()
		sorted.ns = append([]string{"help"}, sorted.ns...)
		sorted.ds = append([]string{hd}, sorted.ds...)
		sorted.chs = append([]CommandHandler{hh}, sorted.chs...)
	}

	return sorted.ns, sorted.ds, sorted.chs
}

// NextCommand picks the next commands to run from this command set based on the content
// of the message
func (cs *CommandSet) NextCommand(m *Message) (CommandHandler, error) {
	// This is repeated from RunCommandHandler, probably something wrong there
	if len(m.Args()) == 0 {
		cmds, _, _ := cs.List()
		return nil, fmt.Errorf("required sub-command missing: %s", strings.Join(cmds, ", "))
	}

	matches := []CommandHandler{}
	matchesns := []string{}
	ematches := []CommandHandler{}
	for name, cmd := range *cs {
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

func (cs *CommandSet) ProcessMessage(ctx context.Context, w ResponseWriter, m *Message) error {
	ch, err := cs.NextCommand(m)
	if err != nil {
		return err
	}
	hn, _ := ch.Describe()
	m.FlagSet = flag.NewFlagSet(hn, flag.ContinueOnError)
	m.flagOut = &bytes.Buffer{}
	m.FlagSet.SetOutput(m.flagOut)
	err = ch.ProcessMessage(ctx, w, m)
	if err == flag.ErrHelp {
		fmt.Fprint(w, cmdUsage(ch, hn, nil).Error())
		return ErrSkipHears
	}

	return err
}
