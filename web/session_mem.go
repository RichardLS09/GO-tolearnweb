package web

import (
	"sync"
	"time"
)

type MemSessionPool struct {
	pool       *sync.Map
	duration   time.Duration
	gcInterval time.Duration
}

func (this *Server) UseMemSession(duration time.Duration, gcInterval time.Duration) {
	memSessionPool := &MemSessionPool{
		pool:       new(sync.Map),
		duration:   duration,
		gcInterval: gcInterval,
	}
	memSessionPool.startGC()
	this.UseSession(memSessionPool)
}

func (this *MemSessionPool) startGC() {
	ticker := time.NewTicker(this.gcInterval)
	go func() {
		for t := range ticker.C {
			this.pool.Range(func(token, session interface{}) bool {
				if session.(*Session).IsExpired(t) {
					this.Del(token.(string))
				}
				return true
			})
		}
	}()
}

func (this *MemSessionPool) Del(token string) {
	this.pool.Delete(token)
}

func (this *MemSessionPool) Get(token string) *Session {
	session, ok := this.pool.Load(token)
	if ok {
		session.(*Session).Refresh()
		return session.(*Session)
	}
	return this.add(token)
}

func (this *MemSessionPool) Set(token string, session *Session) {
	this.pool.Store(token, session)
}

func (this *MemSessionPool) add(token string) *Session {
	session := NewSession(token, this.duration)
	this.Set(token, session)
	return session
}
