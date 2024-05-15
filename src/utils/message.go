package utils

type AppMessage struct {
	ID   string
	Body int
	From int
	To   int
}

type AppMsgWithResp struct {
	Msg    AppMessage
	RespCh chan AppMessage
}

func NewAppMsg(id string, body int, from int, to int) AppMessage {
	return AppMessage{ID: id, Body: body, From: from, To: to}
}
