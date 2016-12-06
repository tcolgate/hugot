package leadership

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/docker/leadership"
	"github.com/docker/libkv/store"
	"github.com/golang/glog"
	"github.com/tcolgate/hugot"
)

type leaderHndler struct {
	next  hugot.RawHandler
	store store.Store
	key   string
	name  string

	*sync.Mutex
	isLeader bool
}

// New creates a new handler that uses an external consistent store to
// elect a leader among competing bots. Until the handler is elected leader,
// all messages are ignored. Messaes arriving via WebHooks will still arrive.
func New(next hugot.RawHandler, store store.Store, key, name string) hugot.RawHandler {
	lh := &leaderHndler{
		next,
		store,
		key,
		name,

		&sync.Mutex{},
		false,
	}

	go lh.elect()

	return lh
}

func (lh *leaderHndler) Describe() (string, string) {
	return "leadership", "HA support for electing a leader bot"
}

func (lh *leaderHndler) ProcessMessage(ctx context.Context, w hugot.ResponseWriter, m *hugot.Message) error {
	lh.Lock()
	defer lh.Unlock()
	if lh.isLeader {
		return lh.next.ProcessMessage(ctx, w, m)
	}
	return nil
}

func (lh *leaderHndler) elect() {
	c := leadership.NewCandidate(lh.store, lh.key, lh.name, 5*time.Second)
	electedCh, _ := c.RunForElection()

	// We'll resign if we get a signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {

		case <-sigs:
			func() {
				lh.Lock()
				defer lh.Unlock()

				if lh.isLeader {
					c.Resign()
				}
			}()

		case isElected := <-electedCh:
			func() {
				lh.Lock()
				defer lh.Unlock()

				if isElected {
					lh.isLeader = true
					glog.Infof("I won the election. I'm now the leader")
				} else {
					lh.isLeader = false
					glog.Infof("I lost the election")
				}
			}()
		}
	}
}
