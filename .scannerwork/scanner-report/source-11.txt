package utils

type MarkChannels struct {
	SendCh chan AppMessage // process <-- SendMark --> snap
	RecvCh chan AppMessage // process --- mark|msg --> snap
}

type AppMsgChannels struct {
	SendToProcCh chan RespMessage // app ---    msg   --> process
	RecvCh       chan AppMessage  // process ---    msg   --> app
	SendToSnapCh chan AppMessage  // process ---    msg   --> snap
}

type StatesChannels struct {
	SaveCh chan FullState // process --- state --> snap
	CurrCh chan FullState // snap --- state --> process
	RecvCh chan FullState // process --- state --> snap
}
