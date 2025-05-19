package sparallel_server

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sparallel_server/pkg/foundation/errs"
	"strings"
)

type Process struct {
	Uuid   string
	Cmd    *exec.Cmd
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
}

type ProcessResponse struct {
	Data  string
	Error error
}

type ProcessFinishedHandler func(processUuid string)

func CreateProcess(ctx context.Context, command string, handler ProcessFinishedHandler) (*Process, error) {
	parts := strings.Fields(command)

	var args []string

	if len(parts) > 1 {
		args = parts[1:]
	}

	cmd := exec.CommandContext(ctx, parts[0], args...)

	cmd.Cancel = func() error {
		slog.Debug("Cancel process: " + command)

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

	processUuid := uuid.New().String()

	go func(cmd *exec.Cmd, handler ProcessFinishedHandler, processUuid string) {
		_ = cmd.Wait()

		handler(processUuid)
	}(cmd, handler, processUuid)

	return &Process{
		Uuid:   processUuid,
		Cmd:    cmd,
		Stdout: stdout,
		Stdin:  stdin,
	}, nil
}

func (p *Process) IsRunning() bool {
	if p.Cmd.ProcessState == nil {
		return true
	}

	return !p.Cmd.ProcessState.Exited()
}

func (p *Process) Write(data string) error {
	slog.Debug("Write: [" + data + "] to process: [" + p.Uuid + "]")

	_, err := p.Stdin.Write([]byte(data))

	return errs.Err(err)
}

func (p *Process) Read() *ProcessResponse {
	buffer := make([]byte, 4096)
	n, err := p.Stdout.Read(buffer)

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

func (p *Process) Close() error {
	err := p.Cmd.Process.Kill()

	return errs.Err(err)
}
