package shell

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
	"syscall"
	"text/template"
)

const (
	maxRenderIterations = 10
	bufferSize          = 64 * 1024
)

type CommandBuilder struct {
	cmdTmpl  []*template.Template
	envVars  []string
	tmplVars map[string]any
}

func NewCommandBuilder(command []string, envVars map[string]string, tmplVars map[string]any) (*CommandBuilder, error) {
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

	return &CommandBuilder{
		cmdTmpl:  cmdTmpl,
		envVars:  environ,
		tmplVars: tmplVars,
	}, nil
}

func (c *CommandBuilder) WithTemplateVars(templateVars map[string]any) *CommandBuilder {
	clone := &CommandBuilder{
		cmdTmpl:  c.cmdTmpl,
		envVars:  c.envVars,
		tmplVars: make(map[string]any),
	}

	if c.tmplVars != nil {
		for k, v := range c.tmplVars {
			clone.tmplVars[k] = v
		}
	}

	for k, v := range templateVars {
		clone.tmplVars[k] = v
	}

	return clone
}

func (c *CommandBuilder) Stream(ctx context.Context, w io.Writer) (int64, error) {
	commandParts, err := c.renderCommand(c.tmplVars)
	if err != nil {
		return 0, err
	}

	run := exec.Command(commandParts[0], commandParts[1:]...)

	go func() {
		<-ctx.Done()
		logging.Debug(ctx, "context canceled, stopping shell command")
		if run.Process != nil {
			run.Process.Signal(syscall.SIGINT)
		}
	}()
	defer run.Wait()

	run.Env = c.envVars
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
			logging.Debug(ctx, "command output", "msg", scanner.Text())
		}
	}()

	buf := make([]byte, bufferSize)
	bytesWritten := int64(0)

	for {
		if ctx.Err() != nil {
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

func (c *CommandBuilder) renderCommand(tmplVars map[string]any) ([]string, error) {
	cmdLen := len(c.cmdTmpl)

	if cmdLen == 1 {
		result, err := renderTemplate(c.cmdTmpl[0], tmplVars)
		if err != nil {
			return nil, err
		}
		return []string{"sh", "-c", result}, nil
	}

	command := make([]string, cmdLen)
	for i, tmpl := range c.cmdTmpl {
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
