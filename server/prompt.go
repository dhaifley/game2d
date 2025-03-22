package server

import (
	"context"
	"time"
)

// Prompter values are able to send prompts to AI services.
type Prompter interface {
	Prompt(ctx context.Context,
		prompt string,
		state []byte,
	) (string, []byte, error)
}

// mockPrompter is a mock implementation of the Prompter interface.
type mockPrompter struct {
	res   string
	state []byte
	delay time.Duration
}

// NewMockPrompter creates a new mock prompter with the given response, state
// and delay.
func NewMockPrompter(res string,
	state []byte,
	delay time.Duration,
) *mockPrompter {
	return &mockPrompter{
		res:   res,
		state: state,
		delay: delay,
	}
}

// Prompt sends a prompt to the mock prompter and returns the response and
// state.
func (m *mockPrompter) Prompt(ctx context.Context,
	prompt string,
	state []byte,
) (string, []byte, error) {
	res := prompt
	rState := state

	if m.res != "" {
		res = m.res
	}

	if m.state != nil {
		rState = m.state
	}

	time.Sleep(m.delay)

	return res, rState, nil
}
