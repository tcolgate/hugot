package ssh

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/golang/glog"
	"github.com/tcolgate/hugot"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type sshAdpt struct {
	nick string

	listener net.Listener
	config   *ssh.ServerConfig

	running sync.Once

	rch chan *hugot.Message

	sync.RWMutex
	schs map[string]chan *hugot.Message
}

func (a *sshAdpt) runOnce() {
	a.running.Do(func() {
		a.run()
	})
}

// New creates a new SSH Adapter
func New(nick string, l net.Listener, cfg *ssh.ServerConfig) *sshAdpt {
	return &sshAdpt{
		nick,
		l,
		cfg,
		sync.Once{},

		make(chan *hugot.Message),

		sync.RWMutex{},
		make(map[string]chan *hugot.Message),
	}
}

func (a *sshAdpt) Receive() <-chan *hugot.Message {
	go a.run()

	return a.rch
}

func (a *sshAdpt) Send(ctx context.Context, m *hugot.Message) {
	go a.run()

	a.RLock()
	sch, ok := a.schs[m.Channel]
	a.RUnlock()

	if !ok {
		glog.Errorf("ssh command to unknown session")
		return
	}

	sch <- m
}

func (a *sshAdpt) run() {
	for {
		tcpConn, err := a.listener.Accept()
		if err != nil {
			glog.Errorf("Failed to accept incoming connection (%s)", err)
			continue
		}
		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, a.config)
		if err != nil {
			glog.Errorf("Failed to handshake (%s)", err)
			continue
		}

		glog.Infof("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		// Discard all global out-of-band Requests
		go ssh.DiscardRequests(reqs)
		// Accept all channels
		go a.handleChannels(chans, sshConn)
	}
}

func (a *sshAdpt) handleChannels(chans <-chan ssh.NewChannel, sshConn ssh.Conn) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go a.handleChannel(newChannel, sshConn)
	}
}

func (a *sshAdpt) handleChannel(newChannel ssh.NewChannel, sshConn ssh.Conn) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	connection, requests, err := newChannel.Accept()
	if err != nil {
		glog.Errorf("Could not accept channel (%s)", err)
		return
	}

	user := sshConn.User()
	session := base64.StdEncoding.EncodeToString(sshConn.SessionID())
	a.Lock()
	sch := make(chan *hugot.Message)
	a.schs[session] = sch
	a.Unlock()

	// Prepare teardown function
	closeconn := func() {
		connection.Close()
	}

	t := terminal.NewTerminal(connection, user+"> ")

	done := make(chan struct{})
	go func() {
		for {
			select {
			case m := <-sch:
				for _, l := range strings.Split(m.Text, "\n") {
					fmt.Fprintf(t, "%s: %s\r\n", a.nick, l)
				}
			case <-done:
				break
			}
		}
		done <- struct{}{}
	}()

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				// We only accept the default shell
				// (i.e. no command in the Payload)
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req":
				//termLen := req.Payload[3]
				//w, h := parseDims(req.Payload[termLen+4:])
				//should do something with the terminal info
				req.Reply(true, nil)
			case "window-change":
				//w, h := parseDims(req.Payload)
				//setWinsize(bashf.Fd(), w, h)
			}
		}
	}()

	for {
		ln, err := t.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		a.rch <- &hugot.Message{Text: string(ln), ToBot: true, From: user, UserID: user, Channel: session}
	}

	a.Lock()
	defer a.Unlock()
	delete(a.schs, session)

	done <- struct{}{}
	closeconn()
}

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

// Winsize stores the Height and Width of a terminal.
type winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

// setWinsize sets the size of the given pty.
func setWinsize(fd uintptr, w, h uint32) {
	ws := &winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
