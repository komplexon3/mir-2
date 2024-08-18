package messages

type StartMessage struct {
	Data string
}

func (m *StartMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewStartMessage(data string) *StartMessage {
	return &StartMessage{
		Data: data,
	}
}

type EchoMessage struct {
	Data string
}

func (m *EchoMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewEchoMessage(data string) *EchoMessage {
	return &EchoMessage{
		Data: data,
	}
}

type ReadyMessage struct {
	Data string
}

func (m *ReadyMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewReadyMessage(data string) *ReadyMessage {
	return &ReadyMessage{
		Data: data,
	}
}
