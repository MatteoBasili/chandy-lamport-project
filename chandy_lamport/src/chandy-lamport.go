package main

import (
	"fmt"
	"time"
	"chandy_lamport/src/node"
	"chandy_lamport/src/utils"
)

func main() {
	// Creazione dei processi
	numProcesses := 5
	node.Processes = make([]*node.Process, numProcesses)
	for i := 0; i < numProcesses; i++ {
		node.Processes[i] = &node.Process{
			ID:          i + 1,
			Balance:     1000,
			OutChannels: make([]int, numProcesses),          // Inizializza un slice vuoto
			InChannels:  make([]chan utils.Message, numProcesses), // Inizializza un slice vuoto
			Recording:   false,
			MarkerSeen:  false,
		}
	}
	// Creazione dei Canali
	for i := 0; i < numProcesses; i++ {
		for j := 0; j < numProcesses; j++ {
			if j != i {
				node.Processes[i].OutChannels[j] = node.Processes[j].ID
				node.Processes[i].InChannels[j] = make(chan utils.Message)
			}
		}
		// Stampare il contenuto di OutChannels
		fmt.Printf("Process %d - OutChannels: %v\n", node.Processes[i].ID, node.Processes[i].OutChannels)
	}

	// FATTO // Passo 1 C-L

	/*// Avvio del timer che chiama TakeSnapshot ogni secondo per il processo con ID 1
	  go func() {
	      for {
	          node.Processes[0].TakeSnapshot() // Processo con ID 1 (l'indice è 0 perché gli indici partono da 0 in Go)
	          time.Sleep(time.Second)     // Attendi un secondo
	      }
	  }()

	  // Mantieni il programma in esecuzione
	  select {}*/

	// FATTO // Passo 2 C-L

	// Avvio della procedura InitiateSnapshot per il processo con ID 1
	go func() {
		// Attendi 5 secondi prima di avviare il processo InitiateSnapshot per consentire il completamento delle inizializzazioni
		time.Sleep(time.Second)
		node.Processes[0].InitiateSnapshot() // Processo con ID 1 (l'indice è 0 perché gli indici partono da 0 in Go)
	}()

	time.Sleep(10 * time.Second)

	// FATTO // Passo 3 C-L

	//  // Passo 4 C-L

	fmt.Println("Simulation completed.")
}
