package nfcptl

import (
	"context"
	"fmt"
	"testing"
)

const (
	testTokenAbsent  StateType = "testTokenAbsent"
	testTokenPresent StateType = "testTokenPresent"

	testTokenPlaced  StateEventType = "testTokenPlaced"
	testTokenRemoved StateEventType = "testTokenRemoved"
)

type testTokenPlacedAction struct{}

func (a testTokenPlacedAction) execute(_ context.Context) StateEventType {
	fmt.Println("The LED has been switched ON")
	return NoOp
}

type testTokenRemovedAction struct{}

func (a testTokenRemovedAction) execute(_ context.Context) StateEventType {
	fmt.Println("The LED has been switched OFF")
	return NoOp
}

func newPortalState() (*stateMachine, error) {
	return NewStateMachine(testTokenAbsent, States{
		testTokenAbsent: State{
			action: &testTokenPlacedAction{},
			events: Events{
				testTokenPlaced: testTokenPresent,
			},
		},
		testTokenPresent: State{
			action: &testTokenRemovedAction{},
			events: Events{
				testTokenRemoved: testTokenAbsent,
			},
		},
	})
}

func TestStateMachine(t *testing.T) {
	portalState, _ := newPortalState()

	err := portalState.SendEvent(testTokenRemoved, nil)
	if err != ErrEventRejected {
		t.Errorf("Expected the event rejected error, got nil")
	}

	err = portalState.SendEvent(testTokenPlaced, nil)
	if err != nil {
		t.Errorf("Couldn't switch the LED on, err: %v", err)
	}

	err = portalState.SendEvent(testTokenPlaced, nil)
	if err != ErrEventRejected {
		t.Errorf("Expected the event rejected error, got nil")
	}

	err = portalState.SendEvent(testTokenRemoved, nil)
	if err != nil {
		t.Errorf("Couldn't switch the LED off, err: %v", err)
	}
}
