package video

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iptv-gateway/internal/logging"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

const bufferSize = 1 * 1024 * 1024

type StreamerConfig struct {
	Command      []string          `yaml:"command"`
	EnvVars      map[string]string `yaml:"env_vars"`
	TemplateVars map[string]string `yaml:"template_vars"`
}

type Streamer struct {
	config StreamerConfig
}

func NewStreamer(config StreamerConfig) *Streamer {
	return &Streamer{
		config: config,
	}
}

func (s *Streamer) Stream(ctx context.Context, w io.Writer) (int64, error) {
	streamCtx, cancelStream := context.WithCancel(ctx)
	defer cancelStream()

	go func() {
		<-ctx.Done()
		logging.Debug(ctx, "context canceled, stopping stream")
		cancelStream()
	}()

	commandParts, err := s.prepareCommand()
	if err != nil {
		return 0, err
	}

	run := exec.CommandContext(streamCtx, commandParts[0], commandParts[1:]...)
	logging.Debug(ctx, "executing command", "cmd", strings.Join(run.Args, " "))

	run.Env = os.Environ()
	for key, value := range s.config.EnvVars {
		run.Env = append(run.Env, key+"="+value)
	}

	stdout, err := run.StdoutPipe()
	if err != nil {
		return 0, err
	}

	stderr, stderrErr := run.StderrPipe()
	if stderrErr != nil {
		return 0, stderrErr
	}

	if startErr := run.Start(); startErr != nil {
		return 0, startErr
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			logging.Debug(streamCtx, "command output", "msg", scanner.Text())
		}
	}()

	buf := make([]byte, bufferSize)
	bytesWritten := int64(0)

	defer run.Process.Wait()

	for {
		if streamCtx.Err() != nil {
			return bytesWritten, nil
		}

		n, err := stdout.Read(buf)
		if err != nil {
			return bytesWritten, nil
		}

		if n > 0 {
			bytesWritten += int64(n)
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return bytesWritten, nil
			}
		}
	}
}

func (s *Streamer) ContentType() string {
	return "video/mp2t"
}

// TODO: move to subscription initialization to avoid processing in runtime
func (s *Streamer) prepareCommand() ([]string, error) {
	if len(s.config.Command) == 0 {
		return nil, errors.New("command cannot be empty")
	}

	command := make([]string, len(s.config.Command))
	for i, part := range s.config.Command {
		result, err := s.renderCommandPart(part)
		if err != nil {
			return nil, err
		}
		command[i] = result
	}

	if len(command) == 1 {
		command = []string{"sh", "-c", command[0]}
	}

	return command, nil
}

func (s *Streamer) renderCommandPart(tmplStr string) (string, error) {
	var prevResult string
	currentTmplStr := tmplStr
	buf := &bytes.Buffer{}

	for {
		buf.Reset()

		tmpl, err := template.New("command-part").Parse(currentTmplStr)
		if err != nil {
			return "", fmt.Errorf("parse template: %w", err)
		}

		if err := tmpl.Execute(buf, s.config.TemplateVars); err != nil {
			return "", fmt.Errorf("render: %w", err)
		}

		newResult := buf.String()
		if prevResult == newResult {
			break
		}

		prevResult = newResult
		currentTmplStr = newResult
	}

	return buf.String(), nil
}
