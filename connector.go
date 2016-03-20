package socket

import (
	"net"
	"time"
)

type Connector struct {
	eventLoop           *EventLoop
	address             string
	autoReconnectSec    int // 重连间隔时间, 0为不重连
	autoReconnectMaxTry int
	reconnectCount      int
	closeSignal         chan bool
	working             bool // 重入锁
	session             *Session
	newSessionHandler   func(*Session)
}

// 自动重连间隔=0不重连
func (self *Connector) SetAutoReconnectSec(sec int) {
	self.autoReconnectSec = sec
}

func (self *Connector) SetAutoReconnectMaxTry(times int) {
	self.autoReconnectMaxTry = times
}

func (self *Connector) Start() {
	if self.working {
		return
	}

	go self.connect(self.address)
	self.eventLoop.PostEvent(NewEvent(Event_ConnectorStart, self))
}

func (self *Connector) retryAllowed() bool {
	return self.reconnectCount < self.autoReconnectMaxTry || self.autoReconnectMaxTry == -1
}

func (self *Connector) connect(address string) {
	self.working = true

	for {
		// 开始连接
		cn, err := net.Dial("tcp", address)
		// 连不上
		if err != nil {
			self.eventLoop.PostEvent(NewEvent(Event_ConnectorConnectFailed, ConnectorData{self, err}))

			// 没重连就退出
			if self.autoReconnectSec == 0 {
				break
			}

			// 有重连就等待
			time.Sleep(time.Duration(self.autoReconnectSec) * time.Second)
			if self.retryAllowed() {
				self.eventLoop.PostEvent(NewEvent(Event_ConnectorTryReconnecting, self))
				self.reconnectCount += 1
				continue
			}
			break
		}

		// 创建Session
		self.session = NewSession(cn)
		self.newSessionHandler(self.session)
		self.session.Start(self.eventLoop)
		self.eventLoop.PostEvent(NewEvent(Event_ConnectorConnected, self))
		self.reconnectCount = 0

		// 内部断开回调
		self.session.OnClose = func() {
			self.closeSignal <- true
		}

		if <-self.closeSignal {
			// 没重连就退出
			if self.autoReconnectSec == 0 {
				break
			}

			// 有重连就等待
			time.Sleep(time.Duration(self.autoReconnectSec) * time.Second)
			if self.retryAllowed() {
				self.eventLoop.PostEvent(NewEvent(Event_ConnectorTryReconnecting, self))
				self.reconnectCount += 1
				continue
			}
			break
		}
	}

	self.working = false
}

func (self *Connector) Stop() {
	if self.session.conn != nil {
		self.eventLoop.PostEvent(NewEvent(Event_ConnectorStop, self))
		self.session.conn.Close()
	}
}

func (self *Connector) Session() *Session {
	return self.session
}

func NewConnector(eventLoop *EventLoop, address string, newSessionHandler func(session *Session)) *Connector {
	self := &Connector{
		eventLoop:           eventLoop,
		address:             address,
		closeSignal:         make(chan bool),
		newSessionHandler:   newSessionHandler,
		autoReconnectMaxTry: -1,
	}
	self.eventLoop.PostEvent(NewEvent(Event_ConnectorInit, self))
	return self
}
