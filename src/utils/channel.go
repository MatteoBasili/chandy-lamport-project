package utils

type MarkChannels struct {
	SendCh chan AppMessage // snap --- Marker --> process
	RecvCh chan AppMessage // process --- mark|msg --> snap
}

type AppMsgChannels struct {
	SendToProcCh chan RespMessage // app     ---    msg   --> process
	RecvCh       chan AppMessage  // process ---    msg   --> app
	SendToSnapCh chan AppMessage  // process ---    msg   --> snap
}

type StatesChannels struct {
	SaveCh chan FullState // process --- state --> snap
	CurrCh chan FullState // snap    --- state --> process
	RecvCh chan FullState // process --- state --> snap
}

func CreateMarkChans(size int) MarkChannels {
	markChans := MarkChannels{
		SendCh: make(chan AppMessage, size),
		RecvCh: make(chan AppMessage, size),
	}
	return markChans
}

func CreateAppMsgChans(size int) AppMsgChannels {
	appMsgChans := AppMsgChannels{
		SendToProcCh: make(chan RespMessage, size),
		RecvCh:       make(chan AppMessage, size),
		SendToSnapCh: make(chan AppMessage, size),
	}
	return appMsgChans
}

func CreateStatesChans(size int) StatesChannels {
	statesChans := StatesChannels{
		SaveCh: make(chan FullState, size),
		CurrCh: make(chan FullState, size),
		RecvCh: make(chan FullState, size),
	}
	return statesChans
}
