package ssh

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"syscall"
	"unsafe"

	"github.com/chzyer/readline"
	"github.com/golang/glog"
	"github.com/tcolgate/hugot"
	"golang.org/x/crypto/ssh"
)

type sshAdpt struct {
	listener net.Listener
	config   *ssh.ServerConfig

	running sync.Once

	rch chan *hugot.Message
}

func (a *sshAdpt) runOnce() {
	a.running.Do(func() {
		a.run()
	})
}

// New creates a new SSH Adapter
func New(l net.Listener, cfg *ssh.ServerConfig) *sshAdpt {
	return &sshAdpt{
		l,
		cfg,
		sync.Once{},

		make(chan *hugot.Message),
	}
}

func (a *sshAdpt) Receive() <-chan *hugot.Message {
	a.run()

	return nil
}

func (a *sshAdpt) Send(*hugot.Message) {
	a.run()

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
			log.Printf("Failed to handshake (%s)", err)
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
		log.Printf("Could not accept channel (%s)", err)
		return
	}

	user := sshConn.User()
	session := string(sshConn.SessionID())

	// Prepare teardown function
	close := func() {
		connection.Close()
	}

	rl, err := readline.NewEx(&readline.Config{
		UniqueEditLine: true,
	})
	if err != nil {
		glog.Error(err.Error())
		return
	}
	defer rl.Close()

	rl.ResetHistory()
	rl.SetPrompt("> ")

	done := make(chan struct{})
	go func() {
		for {
			select {
			case m := <-s.sch:
				log.Printf("%s: %s", s.nick, m.Text)
			case <-done:
				break
			}
		}
		done <- struct{}{}
	}()

	for {
		ln, err := rl.Readline()
		if err != nil {
			break
		}

		s.rch <- &hugot.Message{Text: ln, ToBot: true, From: user, UserID: user, Channel: session}
	}

	rl.Clean()
	done <- struct{}{}

	//pipe session to bash and visa-versa
	var once sync.Once
	go func() {
		io.Copy(connection, rl)
		once.Do(close)
	}()
	go func() {
		io.Copy(rl, connection)
		once.Do(close)
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

/*
	rl, err := readline.NewEx(&readline.Config{
		UniqueEditLine: true,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	rl.ResetHistory()
	log.SetOutput(rl.Stderr())

	rl.SetPrompt(s.user + "> ")

	done := make(chan struct{})
	go func() {
		for {
			select {
			case m := <-s.sch:
				log.Printf("%s: %s", s.nick, m.Text)
			case <-done:
				break
			}
		}
		done <- struct{}{}
	}()

	for {
		ln, err := rl.Readline()
		if err != nil {
			break
		}

		u, err := user.Current()
		if err != nil {
			glog.Errorf("Could not get current user")
			continue
		}

		s.rch <- &hugot.Message{Text: ln, ToBot: true, From: s.user, UserID: u.Uid}
	}

	rl.Clean()
	done <- struct{}{}
*/
