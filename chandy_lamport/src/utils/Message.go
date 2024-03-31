package utils

// Messaggio generico
type Message struct {
	SenderID   int
	ReceiverID int
	Amount     int
}

// Messaggio marker
type MarkerMessage struct {
	SenderID int
}