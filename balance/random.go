package balance

import (
	"math/rand"
	"time"
)

type Random struct{}

func (p *Random) Balance(service []string) (addr string, err error) {
	rand.Seed(time.Now().UnixNano())
	lens := len(service)
	index := rand.Intn(lens)
	addr = service[index]
	return
}
