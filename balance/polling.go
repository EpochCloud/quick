package balance

import (
	"sync"
)

type Polling struct {
	curIndex int
}

var lock sync.Mutex

func (p *Polling) Balance(service []string) (addr string, err error) {
	if len(service) == 0 {
		return
	}
	lock.Lock()
	lens := len(service)
	lock.Unlock()
	if p.curIndex >= lens {
		lock.Lock()
		p.curIndex = 0
		lock.Unlock()
	}
	lock.Lock()
	addr = service[p.curIndex]
	p.curIndex = (p.curIndex + 1) % lens
	lock.Unlock()
	return
}
