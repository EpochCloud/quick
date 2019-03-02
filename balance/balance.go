package balance

type Balancing interface {
	Balance([]string) (addr string, err error)
}

func NewRandom() *Random {
	return &Random{}
}

func NewPolling() *Polling {
	return &Polling{}
}
