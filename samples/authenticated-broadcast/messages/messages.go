package messages

type baseMessage struct {
	Data string
}

func newBaseMessage(data string) baseMessage {
	return baseMessage{
		Data: data,
	}
}

type StartMessage struct {
	baseMessage
}

func NewStartMessage(data string) *StartMessage {
	return &StartMessage{
		newBaseMessage(data),
	}
}

func (m *StartMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

type EchoMessage struct {
	baseMessage
}

func NewEchoMessage(data string) *EchoMessage {
	return &EchoMessage{
		newBaseMessage(data),
	}
}

func (m *EchoMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}
