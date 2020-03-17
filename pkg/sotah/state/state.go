package state

type State string

const (
	None      State = "none"
	Erroneous State = "erroneous"
	Complete  State = "complete"
)
