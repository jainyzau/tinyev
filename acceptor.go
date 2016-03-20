package socket

import (
	"net"
)

type Acceptor struct {
	eventLoop         *EventLoop
	address           string
	sessionManager    *sessionMgr
	listener          net.Listener
	running           bool
	newSessionHandler func(session *Session)
}

func (this *Acceptor) Start() {
	ln, err := net.Listen("tcp", this.address)
	this.listener = ln
	if err != nil {
		this.eventLoop.PostEvent(NewEvent(Event_AcceptorInitFailed, this))
		return
	}
	this.running = true
	this.eventLoop.PostEvent(NewEvent(Event_AcceptorInitDone, this))

	go func() {
		for this.running {
			conn, err := ln.Accept()

			if err != nil {
				this.eventLoop.PostEvent(NewEvent(Event_AcceptorAcceptFailed, this))
				break
			}

			ses := NewSession(conn)
			this.sessionManager.Add(*ses)
			this.newSessionHandler(ses)
			ses.Start(this.eventLoop)
			this.eventLoop.PostEvent(NewEvent(Event_SessionAccepted, AcceptorData{this, ses}))
		}
	}()

	this.eventLoop.PostEvent(NewEvent(Event_AcceptorStart, this))
}

func (this *Acceptor) Stop() {
	if !this.running {
		return
	}

	this.eventLoop.PostEvent(NewEvent(Event_AcceptorStop, this))
	this.running = false
	this.listener.Close()
}

func NewAcceptor(el *EventLoop, address string, newSessionHandler func(session *Session)) *Acceptor {

	this := &Acceptor{
		address:           address,
		eventLoop:         el,
		sessionManager:    newSessionManager(),
		newSessionHandler: newSessionHandler,
	}
	this.eventLoop.PostEvent(NewEvent(Event_AcceptorInit, nil))

	return this
}
