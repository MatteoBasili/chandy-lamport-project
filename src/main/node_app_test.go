package main

import (
	"sdccProject/src/utils"
	"testing"
)

func TestMakeSnapshot(t *testing.T) {
	// Crea un'istanza del tuo NodeApp
	nodeApp := NewNodeApp(0) // Modifica il numero di indice se necessario

	// Simula una chiamata RPC al metodo MakeSnapshot
	var resp *utils.GlobalState
	err := nodeApp.MakeSnapshot(nil, resp)

	// Verifica se non ci sono errori
	if err != nil {
		t.Errorf("Errore durante la chiamata a MakeSnapshot: %v", err)
	}

	// Verifica se la risposta è stata ottenuta correttamente
	if resp == nil {
		t.Error("La risposta da MakeSnapshot è nulla")
	}
}

func TestSendAppMsg(t *testing.T) {
	// Crea un'istanza del tuo NodeApp
	nodeApp := NewNodeApp(0) // Modifica il numero di indice se necessario

	// Simula una chiamata RPC al metodo SendAppMsg
	msg := &utils.AppMessage{} // Assicurati di fornire dati validi per il messaggio
	var resp utils.Result
	err := nodeApp.SendAppMsg(msg, &resp)

	// Verifica se non ci sono errori
	if err != nil {
		t.Errorf("Errore durante la chiamata a SendAppMsg: %v", err)
	}

	// Potresti aggiungere ulteriori verifiche qui in base alla logica del tuo codice
}

// Aggiungi altri test per il codice rimanente secondo necessità
