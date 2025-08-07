package video

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"io"
	"iptv-gateway/internal/logging"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

const (
	bufferSize          = 1 * 1024 * 1024
	maxRenderIterations = 10
)

type Streamer struct {
	cmdTmpl  []*template.Template
	envVars  []string
	tmplVars map[string]any
}

func NewStreamer(command []string, envVars map[string]string, tmplVars map[string]any) (*Streamer, error) {
	cmdTmpl := make([]*template.Template, 0, len(command))

	for _, cmdPart := range command {
		tmpl, err := template.
			New("").
			Funcs(sprig.FuncMap()).
			Parse(cmdPart)

		if err != nil {
			return nil, fmt.Errorf("parse template: %w", err)
		}
		cmdTmpl = append(cmdTmpl, tmpl)
	}

	environ := os.Environ()
	for key, value := range envVars {
		environ = append(environ, key+"="+value)
	}

	return &Streamer{
		cmdTmpl:  cmdTmpl,
		envVars:  environ,
		tmplVars: tmplVars,
	}, nil
}

func (s *Streamer) WithTemplateVars(templateVars map[string]any) *Streamer {
	clone := &Streamer{
		cmdTmpl:  s.cmdTmpl,
		envVars:  s.envVars,
		tmplVars: make(map[string]any),
	}

	if s.tmplVars != nil {
		for k, v := range s.tmplVars {
			clone.tmplVars[k] = v
		}
	}

	for k, v := range templateVars {
		clone.tmplVars[k] = v
	}

	return clone
}

func (s *Streamer) Stream(ctx context.Context, w io.Writer) (int64, error) {
	streamCtx, cancelStream := context.WithCancel(ctx)
	defer cancelStream()

	go func() {
		<-ctx.Done()
		logging.Debug(ctx, "context canceled, stopping stream")
		cancelStream()
	}()

	commandParts, err := s.renderCommand(s.tmplVars)
	if err != nil {
		return 0, err
	}

	run := exec.CommandContext(streamCtx, commandParts[0], commandParts[1:]...)
	logging.Debug(ctx, "executing command", "cmd", strings.Join(run.Args, " "))

	run.Env = s.envVars
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

func (s *Streamer) renderCommand(tmplVars map[string]any) ([]string, error) {
	cmdLen := len(s.cmdTmpl)

	if cmdLen == 1 {
		result, err := renderTemplate(s.cmdTmpl[0], tmplVars)
		if err != nil {
			return nil, err
		}
		return []string{"sh", "-c", result}, nil
	}

	command := make([]string, cmdLen)
	for i, tmpl := range s.cmdTmpl {
		result, err := renderTemplate(tmpl, tmplVars)
		if err != nil {
			return nil, err
		}
		command[i] = result
	}

	return command, nil
}

func renderTemplate(tmpl *template.Template, tmplVars map[string]any) (string, error) {
	buf := &bytes.Buffer{}
	var prevResult string

	iter := 0
	for iter < maxRenderIterations {
		buf.Reset()
		if err := tmpl.Execute(buf, tmplVars); err != nil {
			return "", fmt.Errorf("render: %w", err)
		}
		newResult := buf.String()
		if prevResult == newResult {
			break
		}
		prevResult = newResult
		iter++
	}

	return prevResult, nil
}
