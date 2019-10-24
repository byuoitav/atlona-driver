package atlona

import "context"

type VideoSwitcher5x1 struct {
	Username string
	Password string
}

func (v *VideoSwitcher5x1) GetInputByOutput(ctx context.Context, addr, output string) (string, error) {
	return "", nil
}

func (v *VideoSwitcher5x1) SetInputByOutput(ctx context.Context, addr, output string) error {
	return nil
}
