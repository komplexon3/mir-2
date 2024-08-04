package idledetection

type IdleNotification struct {
	Ack           chan struct{}
	NoLongerIdleC chan NoLongerIdleNotification
}

func NewIdleNotification() IdleNotification {
	return IdleNotification{
		Ack:           make(chan struct{}),
		NoLongerIdleC: make(chan NoLongerIdleNotification),
	}
}

type NoLongerIdleNotification struct {
	Ack chan struct{}
}

func NewNoLongerIdleNotification() NoLongerIdleNotification {
	return NoLongerIdleNotification{
		Ack: make(chan struct{}),
	}
}
