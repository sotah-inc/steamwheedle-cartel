package state

type State string

const (
	Erroneous State = "erroneous"
	Complete  State = "complete"
)
