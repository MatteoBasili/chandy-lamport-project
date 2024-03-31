package node

import (
	"fmt"
	"time"
	"chandy_lamport/src/utils"
	"chandy_lamport/src/state"
)

var Processes []*Process

// Rappresentazione di un processo nel sistema distribuito
type Process struct {
	ID          int
	Balance     int
	Snapshots   []state.Snapshot     // memorizza tutti gli snapshot fatti dal processo
	OutChannels []int          // Canali di uscita
	InChannels  []chan utils.Message // Canali di ingresso
	Recording   bool           // Variabile di stato per indicare se il processo sta registrando sui canali in entrata
	MarkerSeen  bool           // Variabile di stato per indicare se il processo ha ricevuto un marker
	//mutex          sync.Mutex     // mutex di sincronizzazione utilizzato per garantire l'accesso concorrente sicuro alle risorse condivise all'interno del processo
}

// GIUSTO // Funzione per registrare lo stato di un processo
func (p *Process) TakeSnapshot() {
	//p.mutex.Lock()
	//defer p.mutex.Unlock()

	if !p.Recording {
		snapshot := state.Snapshot{
			ID:        p.ID,
			Balance:   p.Balance,
			Timestamp: time.Now(),
		}
		p.Snapshots = append(p.Snapshots, snapshot)
		fmt.Printf("Snapshot taken for Process %d at %v. Balance: $%d\n", p.ID, snapshot.Timestamp, snapshot.Balance)

		p.MarkerSeen = true
	}
}

// Ricezione di un messaggio
func (p *Process) ReceiveMessage(msg utils.Message) {
	// Controlla se il processo sta registrando, quindi elabora il messaggio solo se lo è
	if p.IsRecording() {
		p.InChannels[msg.ReceiverID] <- msg
	}
}

// Invio di un marker
func (p *Process) SendMarker() {
	// Invia il messaggio marker su tutti i canali di uscita tranne il canale locale
	markerMessage := utils.MarkerMessage{
		SenderID: p.ID,
	}
	for _, outChannel := range p.OutChannels {
		if outChannel != 0 {
			go p.sendMarkerTo(outChannel, markerMessage)
		}
	}
}

// Invio del marker a un canale specifico
func (p *Process) sendMarkerTo(outChannel int, marker utils.MarkerMessage) {
	fmt.Printf("Process %d is sending marker to Process %d...\n", p.ID, outChannel)
	Processes[outChannel-1].ReceiveMarker(marker)
}

// Ricezione di un marker
func (p *Process) ReceiveMarker(marker utils.MarkerMessage) {
	fmt.Printf("Process %d: marker received from Process %d\n", p.ID, marker.SenderID)
	//p.mutex.Lock()
	//defer p.mutex.Unlock()

	if !p.MarkerSeen {
		// Se il processo non ha mai ricevuto un marker, registra il proprio stato
		p.TakeSnapshot()

		fmt.Printf("Process %d marked the channel C_%d%d as empty\n", p.ID, marker.SenderID, p.ID)

		// Invia un marker su tutti i canali di uscita
		p.SendMarker()

		// Comincia a registrare su tutti i canali di ingresso
		p.StartRecording()
	} else {
		// Se il processo ha già visto un marker, smette di registrare su questo canale
		p.StopRecording(marker.SenderID)
	}
}

// Avvio della registrazione su tutti i canali di ingresso
func (p *Process) StartRecording() {
	//p.mutex.Lock()
	//defer p.mutex.Unlock()
	p.Recording = true
	fmt.Printf("Process %d starts to record messages on its input channels\n", p.ID)
}

// Sospensione della registrazione
func (p *Process) StopRecording(inChannel int) {
	//p.mutex.Lock()
	//defer p.mutex.Unlock()
	p.Recording = false
	fmt.Printf("Process %d stops to record messages on its input channel C_%d%d\n", p.ID, inChannel, p.ID)
}

// Verifica se il processo sta registrando
func (p *Process) IsRecording() bool {
	//p.mutex.Lock()
	//defer p.mutex.Unlock()
	return p.Recording
}

// Funzione che viene chiamata quando un processo inizia l'algoritmo di Chandy-Lamport
func (p *Process) InitiateSnapshot() {
	p.TakeSnapshot() // Registra il proprio stato
	p.Recording = true
	p.SendMarker()     // Invia il messaggio marker su tutti i canali di uscita
	p.StartRecording() // Comincia a registrare i messaggi che riceve su tutti i suoi canali di entrata
}