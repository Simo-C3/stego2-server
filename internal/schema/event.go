package schema

type Type string

const (
	TypeFinCurrentSeq         Type = "FinCurrentSeq"
	TypeTypingKey             Type = "TypingKey"
	TypeNextSeq               Type = "NextSeq"
	TypeAttack                Type = "Attack"
	TypeChangeRoom            Type = "ChangeRoomState"
	TypeChangeOtherUserState  Type = "ChangeOtherUserState"
	TypeChangeOtherUsersState Type = "ChangeOtherUsersState"
	TypeStartGame             Type = "StartGame"
	TypeChangeWordDifficult   Type = "ChangeWordDifficult"
	TypeResult                Type = "Result"
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
	Type  string `json:"type"`
	Level int    `json:"level"`
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
		InputSeq string `json:"inputSeq"`
	} `json:"payload"`
}

type ChangeRoomState struct {
	Type    Type                   `json:"type"`
	Payload ChangeRoomStatePayload `json:"payload"`
}

type ChangeWordDifficult struct {
	Difficult int    `json:"difficult"`
	Cause     string `json:"cause"`
}

type ChangeRoomStatePayload struct {
	UserNum    int    `json:"userNum"`
	Status     string `json:"status"`
	StartedAt  *int64 `json:"startedAt"`
	StartDelay int    `json:"startDelay"`
	MaxUserNum int    `json:"maxUserNum"`
	OwnerID    string `json:"ownerId"`
}

type ChangeOtherUserState struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Life     int    `json:"life"`
	Seq      string `json:"seq"`
	InputSeq string `json:"inputSeq"`
	Rank     int    `json:"rank"`
}

type Result struct {
	UserID      string `json:"userId"`
	Rank        int    `json:"rank"`
	DisplayName string `json:"displayName"`
}

func NewResult(userID string, rank int, displayName string) *Result {
	return &Result{
		UserID:      userID,
		Rank:        rank,
		DisplayName: displayName,
	}
}
