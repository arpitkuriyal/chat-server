package common

type Message struct {
	From   string `json:"from"`
	Text   string `json:"text"`
	Type   string `json:"type"` // "join", "join-accept", "join-reject", "chat", "system"
	IsHost bool   `json:"isHost"`
}
