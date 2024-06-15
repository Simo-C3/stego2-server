package schema

type Type string

const (
	TypeFinCurrentSeq Type = "FinCurrentSeq"
	TypeTypingKey     Type = "TypingKey"
	TypeNextSeq       Type = "NextSeq"
)

type Base struct {
	Type    Type        `json:"type"`
	Payload interface{} `json:"payload"`
}

type AttackEvent struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Damage int    `json:"damage"`
}

type NextSeqEvent struct {
	Value string `json:"value"`
	Ruby  string `json:"ruby"`
	Type  string `json:"type"`
}

type FinCurrentSeq struct {
	Type    Type `json:"type"`
	Payload struct {
		Cause string `json:"cause"`
	} `json:"payload"`
}

type TypingKey struct {
	Type    Type `json:"type"`
	Payload struct {
		Key rune `json:"inputKey"`
	} `json:"payload"`
}

type ChangeRoomState struct {
	Type      Type `json:"type"`
	UserNum   int
	Status    string
	StartedAt int64
	OwnerID   string
}
