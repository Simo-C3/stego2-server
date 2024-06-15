package schema

type Type string

const (
	TypeFinCurrentSeq Type = "FinCurrentSeq"
	TypeTypingKey     Type = "TypingKey"
	TypeNextSeq       Type = "NextSeq"
	TypeAttack        Type = "Attack"
	TypeChangeRoom    Type = "ChangeRoomState"
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
	UserNum   int    `json:"userNum"`
	Status    string `json:"status"`
	StartedAt int64  `json:"startedAt"`
	OwnerID   string `json:"ownerId"`
}

type ChangeOtherUserState struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Life     int    `json:"life"`
	Seq      string `json:"seq"`
	InputSeq string `json:"inputSeq"`
	Rank     int    `json:"rank"`
}
