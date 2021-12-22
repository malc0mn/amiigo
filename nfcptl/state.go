package nfcptl

// Derived from Venil Noronha's simple state machine framework.

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ErrEventRejected is the error returned when the state machine cannot process an event in the
// state that it is in.
var ErrEventRejected = errors.New("invalid State for event")

const (
	// NoOp represents a no-op event.
	NoOp StateEventType = "NoOp"
)

// StateType represents a state type in the state machine. E.g. 'off'.
type StateType string

// StateEventType represents an event type in the state machine. E.g. 'switchOff'.
type StateEventType string

// Action represents the action to be executed in a given state. E.g. 'switchOffAction'.
type Action interface {
	Execute(ctx context.Context) StateEventType
}

// Events represents a mapping of StateEventTypes and StateTypes. E.g. 'switchOff: off' can be read
// as: the 'switchOff' event will bring the machine in state 'off'.
type Events map[StateEventType]StateType

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

// NewStateMachine builds a new state machine set to the given initial state.
func NewStateMachine(initial StateType, states States) (*StateMachine, error) {
	if states == nil {
		return nil, errors.New("states cannot be nil")
	}

	var foundInitial bool
	for st, s := range states {
		if st == initial {
			foundInitial = true
		}
		if s.Action == nil {
			return nil, fmt.Errorf("%s has no Action", st)
		}
	}

	if !foundInitial {
		return nil, fmt.Errorf("initial state %s not found in states", initial)
	}

	return &StateMachine{
		curr: initial,
		states: states,
	}, nil
}

// getNextState returns the next state for the event based on the current state, or an error if the
// event cannot be handled in the current state.
func (sm *StateMachine) getNextState(event StateEventType) (StateType, error) {
	if s, ok := sm.states[sm.curr]; ok {
		if s.Events != nil {
			if next, ok := s.Events[event]; ok {
				return next, nil
			}
		}
	}

	return "", ErrEventRejected
}

// SendEvent sends an event to the state machine.
func (sm *StateMachine) SendEvent(event StateEventType, ctx context.Context) error {
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

		nextEvent := s.Action.Execute(ctx)
		if nextEvent == NoOp {
			return nil
		}
		event = nextEvent
	}
}