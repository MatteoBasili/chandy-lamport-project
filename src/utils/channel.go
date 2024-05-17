package utils

type MarkChannel struct {
	SendCh chan AppMessage // node <-- SendMark --> snap
	RecvCh chan AppMessage // node --- mark|msg --> snap
}

type AppMsgChannel struct {
	SendCh chan RespMessage // node <--    msg   --- app
	RecvCh chan AppMessage  // node ---    msg   --> app
}
