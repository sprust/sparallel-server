package sparallel_server

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"io"
	"os"
	"os/exec"
	"sparallel_server/pkg/foundation/errs"
	"strings"
)

type Process struct {
	Uuid   string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

type ProcessResponse struct {
	Data  string
	Error error
}

func CreateProcess(ctx context.Context, command string) (*Process, error) {
	parts := strings.Fields(command)

	// Остальные элементы - аргументы
	var args []string

	if len(parts) > 1 {
		args = parts[1:]
	}

	cmd := exec.CommandContext(ctx, parts[0], args...)

	cmd.Cancel = func() error {
		err := cmd.Process.Signal(os.Interrupt)

		if err != nil && !errors.Is(err, os.ErrProcessDone) {
			err = cmd.Process.Kill()
		}

		if err != nil {
			return errs.Err(err)
		}

		return nil
	}

	stderr := new(strings.Builder)

	cmd.Stderr = stderr

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, errs.Err(err)
	}

	stdin, err := cmd.StdinPipe()

	if err != nil {
		return nil, errs.Err(err)
	}

	if err = cmd.Start(); err != nil {
		return nil, errs.Err(err)
	}

	go func(cmd *exec.Cmd) {
		_ = cmd.Wait()
	}(cmd)

	return &Process{
		Uuid:   uuid.New().String(),
		cmd:    cmd,
		stdout: stdout,
		stdin:  stdin,
	}, nil
}

func (p *Process) IsRunning() bool {
	if p.cmd.ProcessState == nil {
		return true
	}

	return !p.cmd.ProcessState.Exited()
}

func (p *Process) Read() *ProcessResponse {
	buffer := make([]byte, 4096)
	n, err := p.stdout.Read(buffer)

	if err != nil {
		if err == io.EOF {
			return nil
		}
		return &ProcessResponse{Error: errs.Err(err)}
	}

	if n > 0 {
		return &ProcessResponse{Data: string(buffer[:n])}
	}

	return nil
}

func (p *Process) Close() {
	_ = p.cmd.Process.Kill()
}
