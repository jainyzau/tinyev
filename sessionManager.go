package socket

import (
	"sync"
	"sync/atomic"
)

type sessionMgr struct {
	sesMap      map[int64]Session
	sesMapGuard sync.RWMutex
	sesIDAcc    int64
}

func (self *sessionMgr) Add(ses Session) {
	self.sesMapGuard.Lock()
	defer self.sesMapGuard.Unlock()

	ses.id = atomic.AddInt64(&self.sesIDAcc, 1)
	self.sesMap[ses.id] = ses
}

func (self *sessionMgr) Remove(ses Session) {
	self.sesMapGuard.Lock()
	delete(self.sesMap, ses.ID())
	self.sesMapGuard.Unlock()
}

func (self *sessionMgr) GetSession(id int64) (Session, bool) {
	self.sesMapGuard.RLock()
	defer self.sesMapGuard.RUnlock()
	v, ok := self.sesMap[id]
	return v, ok
}

func (self *sessionMgr) IterateSession(callback func(Session) bool) {
	self.sesMapGuard.RLock()
	defer self.sesMapGuard.RUnlock()

	for _, ses := range self.sesMap {
		if !callback(ses) {
			break
		}
	}
}

func (self *sessionMgr) SessionCount() int {
	self.sesMapGuard.Lock()
	defer self.sesMapGuard.Unlock()

	return len(self.sesMap)
}

func newSessionManager() *sessionMgr {
	return &sessionMgr{
		sesMap: make(map[int64]Session),
	}
}
