package atlona

import (
	"context"
	"strings"
)

type VideoSwitcher2x1 struct {
	Username string
	Password string
}

// GetInputByOutput .
func (v *VideoSwitcher2x1) GetInputByOutput(ctx context.Context, addr, output string) (string, error) {
	output = strings.Replace(output, "AUDIO", "", 1)

	return "", nil
}

// SetInputByOutput .
func (v *VideoSwitcher2x1) SetInputByOutput(ctx context.Context, addr, output string) error {
	output = strings.Replace(output, "AUDIO", "", 1)

	return nil
}
