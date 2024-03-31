package state

import (
	"time"
)

// GIUSTO // Stato da salvare durante l'algoritmo di Chandy-Lamport
type Snapshot struct {
	ID        int       // rappresenta l'ID del processo
	Balance   int       // rappresenta il saldo del processo al momento dello snapshot
	Timestamp time.Time // indica il momento in cui lo snapshot è stato preso
}
