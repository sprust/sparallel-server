package processes

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sparallel_server/pkg/foundation/errs"
	"strconv"
	"strings"
)

var lenOfHeaderLen = 20

type Process struct {
	Uuid   string
	Cmd    *exec.Cmd
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
}

type Response struct {
	Data  string
	Error error
}

type FinishedHandler func(processUuid string, cmd *exec.Cmd)

func CreateProcess(ctx context.Context, command string, handler FinishedHandler) (*Process, error) {
	parts := strings.Fields(command)

	var args []string

	if len(parts) > 1 {
		args = parts[1:]
	}

	processUuid := uuid.New().String()

	cmd := exec.CommandContext(ctx, parts[0], args...)

	cmd.Cancel = func() error {
		slog.Debug("Canceling process [" + processUuid + "] [" + strconv.Itoa(cmd.Process.Pid) + "]")

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

	go func(_ context.Context, cmd *exec.Cmd, handler FinishedHandler, processUuid string) {
		_ = cmd.Wait()

		handler(processUuid, cmd)
	}(ctx, cmd, handler, processUuid)

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
	slog.Debug("Write data with len [" + fmt.Sprint(len(data)) + "] to process: [" + p.Uuid + "]")

	dataLength := fmt.Sprintf("%0*d", lenOfHeaderLen, len(data))

	_, err := p.Stdin.Write([]byte(dataLength + data))

	return errs.Err(err)
}

func (p *Process) Read() *Response {
	headerBytes := make([]byte, lenOfHeaderLen)

	_, err := io.ReadFull(p.Stdout, headerBytes)

	if err != nil {
		return &Response{
			Error: errs.Err(err),
		}
	}

	dataLen := 0

	lengthHeader := string(headerBytes)

	_, err = fmt.Sscanf(lengthHeader, "%d", &dataLen)

	if err != nil {
		return &Response{
			Error: errors.New(lengthHeader + p.readOutput()),
		}
	}

	dataBytes := make([]byte, dataLen)

	_, err = io.ReadFull(p.Stdout, dataBytes)

	if err != nil {
		return &Response{
			Error: errs.Err(err),
		}
	}

	return &Response{
		Data: string(dataBytes),
	}
}

func (p *Process) Close() error {
	err := p.Cmd.Process.Kill()

	return errs.Err(err)
}

func (p *Process) readOutput() string {
	buffer := make([]byte, 1024)
	n, err := p.Stdout.Read(buffer)

	if err != nil {
		if err == io.EOF {
			return ""
		}

		if p.Cmd.ProcessState != nil && p.Cmd.ProcessState.Exited() {
			return "worker down: " + p.Cmd.ProcessState.String()
		}

		return err.Error()
	}

	if n > 0 {
		return string(buffer[:n])
	}

	return ""
}
