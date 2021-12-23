package nfcptl

import (
	"errors"
	"fmt"
	"testing"
)

const (
	testTokenAbsent  StateType = "testTokenAbsent"
	testTokenPresent StateType = "testTokenPresent"

	testTokenPlaced  EventType = "testTokenPlaced"
	testTokenRemoved EventType = "testTokenRemoved"
)

type testTokenPlacedAction struct{}

func (a testTokenPlacedAction) Execute(_ *Driver) EventType {
	fmt.Println("testTokenPlacedAction: The LED has been switched ON")
	return OK
}

type testTokenRemovedAction struct{}

func (a testTokenRemovedAction) Execute(_ *Driver) EventType {
	fmt.Println("testTokenRemovedAction: The LED has been switched OFF")
	return OK
}

func TestStateMachineErrNoStateMapping(t *testing.T) {
	sm, err := NewStateMachine(nil)

	if !errors.Is(err, ErrNoStateMapping) {
		t.Errorf("Expected error %s, got %s", ErrNoStateMapping, err)
	}

	if sm != nil {
		t.Errorf("Expected nil, got %v", sm)
	}

	sm, err = NewStateMachine(States{})

	if !errors.Is(err, ErrNoStateMapping) {
		t.Errorf("Expected error %s, got %s", ErrNoStateMapping, err)
	}

	if sm != nil {
		t.Errorf("Expected nil, got %v", sm)
	}
}

func TestStateMachineErrNoDefaultState(t *testing.T) {
	sm, err := NewStateMachine(States{
		TokenAbsent: State{
			Action: &testTokenPlacedAction{},
			Events: Events{
				testTokenPlaced: testTokenPresent,
			},
		},
	})

	if !errors.Is(err, ErrNoDefaultState) {
		t.Errorf("Expected error %s, got: %s", ErrNoDefaultState, err)
	}

	if sm != nil {
		t.Errorf("Expected nil, got %v", sm)
	}
}

func TestStateMachineErrDefaultAction(t *testing.T) {
	sm, err := NewStateMachine(States{
		Default: State{
			Action: &testTokenPlacedAction{},
			Events: Events{
				testTokenPlaced: testTokenPresent,
			},
		},
	})

	if !errors.Is(err, ErrDefaultAction) {
		t.Errorf("Expected error %s, got: %s", ErrDefaultAction, err)
	}

	if sm != nil {
		t.Errorf("Expected nil, got %v", sm)
	}
}

func TestStateMachineErrDefaultEvent(t *testing.T) {
	sm, err := NewStateMachine(States{
		Default: State{
			Events: nil,
		},
	})

	if !errors.Is(err, ErrDefaultEvent) {
		t.Errorf("Expected error %s, got: %s", ErrDefaultEvent, err)
	}

	if sm != nil {
		t.Errorf("Expected nil, got %v", sm)
	}

	sm, err = NewStateMachine(States{
		Default: State{
			Events: Events{},
		},
	})

	if !errors.Is(err, ErrDefaultEvent) {
		t.Errorf("Expected error %s, got: %s", ErrDefaultEvent, err)
	}

	if sm != nil {
		t.Errorf("Expected nil, got %v", sm)
	}

	sm, err = NewStateMachine(States{
		Default: State{
			Events: Events{
				testTokenPlaced:  testTokenPresent,
				testTokenRemoved: testTokenAbsent,
			},
		},
	})

	if !errors.Is(err, ErrDefaultEvent) {
		t.Errorf("Expected error %s, got: %s", ErrDefaultEvent, err)
	}

	if sm != nil {
		t.Errorf("Expected nil, got %v", sm)
	}

	sm, err = NewStateMachine(States{
		Default: State{
			Events: Events{
				testTokenPlaced: testTokenPresent,
			},
		},
	})

	if errors.Is(err, ErrDefaultEvent) {
		t.Errorf("Expected nil, got: %s", err)
	}

	if sm == nil {
		t.Errorf("Expected StateMachine , got %v", sm)
	}
}

func TestStateMachineErrNoAction(t *testing.T) {
	sm, err := NewStateMachine(States{
		Default: State{
			Events: Events{
				testTokenRemoved: TokenAbsent,
			},
		},
		TokenAbsent: State{
			Events: Events{
				testTokenPlaced: testTokenPresent,
			},
		},
	})

	if !errors.Is(err, ErrNoAction) {
		t.Errorf("Expected error %s, got: %s", ErrNoAction, err)
	}

	if sm != nil {
		t.Errorf("Expected nil, got %v", sm)
	}
}

func TestStateMachine(t *testing.T) {
	sm, err := NewStateMachine(States{
		Default: State{
			Events: Events{
				testTokenRemoved: testTokenAbsent,
			},
		},
		testTokenAbsent: State{
			Action: &testTokenRemovedAction{},
			Events: Events{
				testTokenPlaced: testTokenPresent,
			},
		},
		testTokenPresent: State{
			Action: &testTokenPlacedAction{},
			Events: Events{
				testTokenRemoved: testTokenAbsent,
			},
		},
	})

	if err != nil {
		t.Fatalf("Expected nil, got err: %v", err)
	}

	err = sm.Init(nil)
	if err != nil {
		t.Errorf("Expected nil, got err: %v", err)
	}

	err = sm.SendEvent(testTokenRemoved, nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	err = sm.SendEvent(testTokenPlaced, nil)
	if err != nil {
		t.Errorf("Couldn't switch the LED on, err: %v", err)
	}

	err = sm.SendEvent(testTokenPlaced, nil)
	if err != ErrEventRejected {
		t.Error("Expected the event rejected error, got nil")
	}

	err = sm.SendEvent(testTokenRemoved, nil)
	if err != nil {
		t.Errorf("Couldn't switch the LED off, err: %v", err)
	}
}
