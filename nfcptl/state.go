package nfcptl

type State byte

const (
	tokenPlaced = iota
	tokenNotPlaced
	tokenReading
	tokenWriting
)

type CmdState struct {
	cmd   DriverCommand
	state State
}

type Transition func(s *State)

var StateTransitions = map[CmdState]Transition{
	//
}