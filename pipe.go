package socket

import (
	"io"
	"os/exec"
	"sync"
)

type Pipe struct {
	eventLoop       *EventLoop
	writeChan       chan interface{}
	name            string
	cmd             *exec.Cmd
	endSync         sync.WaitGroup
	needNotifyWrite bool
	in              io.WriteCloser
	out             io.ReadCloser
	reader          func(io.Reader) ([]byte, error)
	context         interface{}
	OnClose         func()
}

func (this *Pipe) Start() {
	go this.exitThread()
	go this.recvThread(this.eventLoop)
	go this.sendThread()
	this.cmd.Start()
}

func (this *Pipe) Close() {
	this.writeChan <- closeWritePacket{}
}

func (this *Pipe) Send(data []byte) {
	this.writeChan <- data
}

func (this *Pipe) SetContext(c interface{}) {
	this.context = c
}

func (this *Pipe) Context() interface{} {
	return this.context
}

func (this *Pipe) sendThread() {
	for {
		switch data := (<-this.writeChan).(type) {
		case closeWritePacket:
			goto exitSendLoop
		case []byte:
			n, err := this.in.Write(data)
			if err != nil || n != len(data) {
				goto exitSendLoop
			}
		}
	}
exitSendLoop:
	//TODO: Send close signal to cmd
	this.needNotifyWrite = false
	this.in.Close()
	this.out.Close()
	this.endSync.Done()
}

func (this *Pipe) recvThread(loop *EventLoop) {
	for {
		buffer, err := this.reader(this.out)
		if err != nil {
			loop.PostEvent(NewEvent(Event_PipeClosed, this))
			break
		}

		loop.PostEvent(NewEvent(Event_PipeData, PipeData{this, buffer}))
	}

	if this.needNotifyWrite {
		this.writeChan <- closeWritePacket{}
	}

	this.endSync.Done()
}

func (this *Pipe) exitThread() {
	this.endSync.Add(2)
	this.endSync.Wait()
	this.cmd.Wait()
	if this.OnClose != nil {
		this.OnClose()
	}
}

func NewPipe(eventLoop *EventLoop, cmd *exec.Cmd, reader func(io.Reader) ([]byte, error)) (*Pipe, error) {
	pipe := &Pipe{
		eventLoop: eventLoop,
		writeChan: make(chan interface{}),
		reader:    reader,
		cmd:       cmd,
	}

	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	pipe.in = in

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	pipe.out = out

	return pipe, nil
}
