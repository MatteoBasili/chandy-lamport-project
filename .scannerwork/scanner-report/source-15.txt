package utils

type Message struct {
	ID   string
	Body int
}

type AppMessage struct {
	Msg      Message
	IsMarker bool
	From     int
	To       int
}

type RespMessage struct {
	AppMsg AppMessage
	RespCh chan AppMessage
}

func NewAppMsg(id string, body int, from int, to int) AppMessage {
	msg := AppMessage{
		Msg:      Message{ID: id, Body: body},
		IsMarker: false,
		From:     from,
		To:       to,
	}
	return msg
}

func NewMarkMsg(from int) AppMessage {
	return AppMessage{Msg: Message{}, IsMarker: true, From: from, To: -1}
}
