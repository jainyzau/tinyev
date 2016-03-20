package socket

import (
	"fmt"
)

type EventLoop struct {
	queue      chan interface{}
	handler    func(event interface{})
	exitSignal chan int
}

func NewEventLoop(handler func(event interface{})) *EventLoop {
	return &EventLoop{
		queue:   make(chan interface{}, 2048),
		handler: handler,
	}
}

func (this *EventLoop) PostEvent(event interface{}) {
	this.queue <- event
}

func (this *EventLoop) Run() {
	for event := range this.queue {
		fmt.Println("EventLoop::Run")
		this.handler(event)
	}
}

func (this *EventLoop) Stop(result int) {
	this.exitSignal <- result
}

func (this *EventLoop) Wait() int {
	return <-this.exitSignal
}
