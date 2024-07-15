package messages

type StartMessage struct {
	Data []byte
}

func (m *StartMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewStartMessage(data []byte) *StartMessage {
	return &StartMessage{
		Data: data,
	}
}

type EchoMessage struct {
	Data []byte
}

func (m *EchoMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewEchoMessage(data []byte) *EchoMessage {
	return &EchoMessage{
		Data: data,
	}
}

type ReadyMessage struct {
	Data []byte
}

func (m *ReadyMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewReadyMessage(data []byte) *ReadyMessage {
	return &ReadyMessage{
		Data: data,
	}
}
