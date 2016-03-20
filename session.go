package socket

import (
	"io"
	"net"
	"sync"
)

type closeWritePacket struct {
}

type Session struct {
	writeChan       chan interface{}
	conn            net.Conn
	id              int64
	endSync         sync.WaitGroup
	needNotifyWrite bool
	reader          func(io.Reader) ([]byte, error)
	context         interface{}
	OnClose         func()
}

func (this *Session) ID() int64 {
	return this.id
}

func (this *Session) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *Session) Close() {
	this.writeChan <- closeWritePacket{}
}

func (this *Session) Send(data []byte) {
	this.writeChan <- data
}

func (this *Session) Context() interface{} {
	return this.context
}

func (this *Session) Start(loop *EventLoop) {
	go this.exitThread()
	go this.recvThread(loop)
	go this.sendThread()
}

func (this *Session) sendThread() {
	for {
		switch data := (<-this.writeChan).(type) {
		case closeWritePacket:
			goto exitSendLoop
		case []byte:
			n, err := this.conn.Write(data)
			if err != nil || n != len(data) {
				goto exitSendLoop
			}
		}
	}
exitSendLoop:
	this.needNotifyWrite = false
	this.conn.Close()
	this.endSync.Done()
}

func (this *Session) recvThread(loop *EventLoop) {
	for {
		buffer, err := this.reader(this.conn)
		if err != nil {
			loop.PostEvent(NewEvent(Event_SessionClosed, this))
			break
		}

		loop.PostEvent(NewEvent(Event_SessionData, SessionData{this, buffer}))
	}

	if this.needNotifyWrite {
		this.writeChan <- closeWritePacket{}
	}

	this.endSync.Done()
}

func (this *Session) exitThread() {
	this.endSync.Add(2)
	this.endSync.Wait()
	if this.OnClose != nil {
		this.OnClose()
	}
}

func (this *Session) SetReader(reader func(io.Reader) ([]byte, error)) {
	this.reader = reader
}

func NewSession(conn net.Conn) *Session {
	this := &Session{
		writeChan:       make(chan interface{}),
		conn:            conn,
		needNotifyWrite: true,
	}

	return this
}
