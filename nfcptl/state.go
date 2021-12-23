package nfcptl

// Derived from Venil Noronha's simple state machine framework.

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrNoStateMapping is the error reported when you have supplied nil or an empty States struct
	// to NewStateMachine.
	ErrNoStateMapping = errors.New("states cannot be empty")
	// ErrNoDefaultState is the error reported when you have not defined the Default state in the
	// States struct.
	ErrNoDefaultState = errors.New("no default state")
	// ErrDefaultAction is the error reported by NewStateMachine when you have attached an action
	// to the Default state.
	ErrDefaultAction = errors.New("the default state cannot have an action")
	// ErrDefaultEvent is the error reported by NewStateMachine when you have supplied an incorrect
	// amount of events for the Default state.
	ErrDefaultEvent = errors.New("the default state must have a single event")
	// ErrNoAction is the error reported by NewStateMachine when you have supplied a state
	// transition without an Action.
	ErrNoAction = errors.New("no action")

	// ErrInitFailed is the error returned by Init if the call to init failed.
	ErrInitFailed = errors.New("state machine init failed")
	// ErrEventRejected is the error returned when the state machine cannot process an event in the
	// state that it is in.
	ErrEventRejected = errors.New("invalid State for event")
)

const (
	// Default represents the initial state of the state machine.
	Default StateType = ""
	// TokenAbsent represents the state of where there is no token on the NFC portal.
	TokenAbsent StateType = "TokenAbsent"
	// TokenPresent represents the state of where there is a token on the NFC portal.
	TokenPresent StateType = "TokenPresent"
)

// StateType represents a state type in the state machine. E.g. 'off'.
type StateType string

// Action represents the action to be executed in a given state. E.g. 'switchOffAction'.
type Action interface {
	Execute(d Driver) EventType
}

// Events represents a mapping of StateEventTypes and StateTypes. E.g. 'switchOff: off' can be read
// as: the 'switchOff' event will bring the machine in state 'off'.
type Events map[EventType]StateType

// State binds a state with an action and a set of events it can handle.
type State struct {
	Action Action
	Events Events
}

// States represents a mapping of StateTypes and their implementations.
type States map[StateType]State

// StateMachine represents the State machine.
type StateMachine struct {
	// prev represents the previous state.
	prev StateType

	// curr represents the current state.
	curr StateType

	// states holds the configuration of states and events handled by the state machine.
	states States

	// mu ensures that only 1 event is processed by the state machine at any given time.
	mu sync.Mutex
}

// NewStateMachine builds a new state machine. It performs basic validation on your configured
// states. It will still be possible to pass in inconsistent mappings so take care.
func NewStateMachine(states States) (*StateMachine, error) {
	if states == nil || len(states) == 0 {
		return nil, ErrNoStateMapping
	}

	if _, ok := states[Default]; !ok {
		return nil, ErrNoDefaultState
	}

	for st, s := range states {
		if st == Default {
			if s.Action != nil {
				return nil, ErrDefaultAction
			}
			if s.Events == nil || len(s.Events) == 0 || len(s.Events) > 1 {
				return nil, ErrDefaultEvent
			}
		} else if s.Action == nil {
			return nil, fmt.Errorf("%s: %w", st, ErrNoAction)
		}
	}

	return &StateMachine{states: states}, nil
}

// Init will initialise the state machine by sending the event set for the Default state.
func (sm *StateMachine) Init(d Driver) error {
	if sm.curr == Default {
		if s, ok := sm.states[Default]; ok {
			for e, _ := range s.Events {
				return sm.SendEvent(e, d)
			}
		}
	}

	return ErrInitFailed
}

// getNextState returns the next state for the event based on the current state, or an error if the
// event cannot be handled in the current state.
func (sm *StateMachine) getNextState(event EventType) (StateType, error) {
	if s, ok := sm.states[sm.curr]; ok {
		if s.Events != nil {
			if next, ok := s.Events[event]; ok {
				return next, nil
			}
		}
	}

	return Default, ErrEventRejected
}

// SendEvent sends an event to the state machine.
func (sm *StateMachine) SendEvent(event EventType, d Driver) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for {
		nextState, err := sm.getNextState(event)
		if err != nil {
			return ErrEventRejected
		}

		s, ok := sm.states[nextState]
		if !ok || s.Action == nil {
			panic(fmt.Sprintf("%s not found or has no Action", nextState))
		}

		sm.prev = sm.curr
		sm.curr = nextState

		nextEvent := s.Action.Execute(d)
		if nextEvent == OK {
			return nil
		}
		event = nextEvent
	}
}
