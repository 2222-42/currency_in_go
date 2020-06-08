package card

//go:generate stringer -type=Suit

type Suit int

const (
	Spade Suit = iota
	Club
	Diamon
	Heart
)
