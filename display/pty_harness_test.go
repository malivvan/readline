//go:build unix

package display_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/hinshun/vt10x"

	"github.com/malivvan/readline"
)

const (
	childEnvVar        = "READLINE_PTY_CHILD"
	promptEnvVar       = "READLINE_PTY_PROMPT"
	prefillEnvVar      = "READLINE_PTY_PREFILL"
	noProbeEnvVar      = "READLINE_PTY_NOPROBE"
	hintProviderEnvVar = "READLINE_PTY_HINTPROVIDER"
	transientEnvVar    = "READLINE_PTY_TRANSIENT"
	asyncMSEnvVar      = "READLINE_PTY_ASYNC_MS"
	asyncCompEnvVar    = "READLINE_PTY_ASYNCCOMP"
	autocompleteEnvVar = "READLINE_PTY_AUTOCOMPLETE"
)

func TestMain(m *testing.M) {
	if os.Getenv(childEnvVar) == "1" {
		invokePTYChild()
		return
	}

	os.Exit(m.Run())
}

func invokePTYChild() {

	if n, err := strconv.Atoi(os.Getenv(prefillEnvVar)); err == nil && n > 0 {
		_, _ = fmt.Fprint(os.Stdout, strings.Repeat("\r\n", n))
	}

	rl := readline.NewShell()

	if os.Getenv(noProbeEnvVar) == "1" {
		_ = rl.Config.Set("cursor-position-probe", false)
	}

	if os.Getenv(hintProviderEnvVar) == "1" {
		rl.Hint.SetProvider(func(line []rune, _ int) []rune {
			if len(line) == 0 {
				return nil
			}

			return []rune("HINT:" + string(line))
		})
	}

	if msg := os.Getenv(transientEnvVar); msg != "" {
		rl.Hint.SetTransient(msg)
	}

	if os.Getenv(asyncCompEnvVar) == "1" {
		var extra int32

		rl.Completer = func(_ []rune, _ int) readline.Completions {
			values := []string{"alpha", "alpine"}
			if atomic.LoadInt32(&extra) == 1 {
				values = append(values, "alpaca")
			}

			return readline.CompleteValues(values...)
		}

		if ms, err := strconv.Atoi(os.Getenv(asyncMSEnvVar)); err == nil && ms > 0 {
			go func() {
				time.Sleep(time.Duration(ms) * time.Millisecond)
				atomic.StoreInt32(&extra, 1)
				rl.RefreshCompletions()
			}()
		}
	} else if ms, err := strconv.Atoi(os.Getenv(asyncMSEnvVar)); err == nil && ms > 0 {

		repeat := 1
		if n, err := strconv.Atoi(os.Getenv("READLINE_PTY_ASYNC_REPEAT")); err == nil && n > 0 {
			repeat = n
		}

		go func() {
			for i := range repeat {
				time.Sleep(time.Duration(ms) * time.Millisecond)
				rl.Hint.SetTransient(fmt.Sprintf("ASYNCPING-%d", i))
			}
		}()
	}

	if os.Getenv(autocompleteEnvVar) == "1" {
		_ = rl.Config.Set("autocomplete", true)

		if rl.Completer == nil {
			rl.Completer = func(_ []rune, _ int) readline.Completions {
				return readline.CompleteValues("alpha", "alef", "alpine", "almond").Usage("pick a word")
			}
		}
	}

	prompt := os.Getenv(promptEnvVar)
	if prompt == "" {
		prompt = "> "
	}

	rl.Prompt.Primary(func() string { return prompt })

	if rp := os.Getenv("READLINE_PTY_RIGHTPROMPT"); rp != "" {
		rl.Prompt.Right(func() string { return rp })
	}

	line, err := rl.Readline()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "\r\n[ERR:%s]\r\n", err)
		os.Exit(0)
	}

	_, _ = fmt.Fprintf(os.Stdout, "\r\n[LINE:%s]\r\n", line)
	os.Exit(0)
}

type console struct {
	t    *testing.T
	cmd  *exec.Cmd
	ptmx *os.File
	term vt10x.Terminal

	mu         sync.Mutex
	done       chan struct{}
	probeCount int

	probeReply func(cur vt10x.Cursor) string
}

type consoleConfig struct {
	prompt     string
	cols, rows int

	prefill int

	noProbe bool

	hintProvider bool

	transient string

	asyncMS int

	asyncComp bool

	asyncRepeat int

	rightPrompt string

	autocomplete bool

	probeReply func(vt10x.Cursor) string
}

func craftConsole(t *testing.T, prompt string, cols, rows int) *console {
	t.Helper()

	return beginConsole(t, consoleConfig{prompt: prompt, cols: cols, rows: rows})
}

func beginConsole(t *testing.T, cfg consoleConfig) *console {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), os.Args[0])
	cmd.Env = append(os.Environ(),
		childEnvVar+"=1",
		promptEnvVar+"="+cfg.prompt,
		fmt.Sprintf("%s=%d", prefillEnvVar, cfg.prefill),
		"INPUTRC=/dev/null",
		"TERM=xterm-256color",
	)

	if cfg.noProbe {
		cmd.Env = append(cmd.Env, noProbeEnvVar+"=1")
	}

	if cfg.hintProvider {
		cmd.Env = append(cmd.Env, hintProviderEnvVar+"=1")
	}

	if cfg.transient != "" {
		cmd.Env = append(cmd.Env, transientEnvVar+"="+cfg.transient)
	}

	if cfg.asyncMS > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%d", asyncMSEnvVar, cfg.asyncMS))
	}

	if cfg.asyncComp {
		cmd.Env = append(cmd.Env, asyncCompEnvVar+"=1")
	}

	if cfg.autocomplete {
		cmd.Env = append(cmd.Env, autocompleteEnvVar+"=1")
	}

	if cfg.asyncRepeat > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("READLINE_PTY_ASYNC_REPEAT=%d", cfg.asyncRepeat))
	}

	if cfg.rightPrompt != "" {
		cmd.Env = append(cmd.Env, "READLINE_PTY_RIGHTPROMPT="+cfg.rightPrompt)
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(cfg.rows), Cols: uint16(cfg.cols)})
	if err != nil {
		t.Fatalf("start child under pty: %v", err)
	}

	c := &console{
		t:          t,
		cmd:        cmd,
		ptmx:       ptmx,
		term:       vt10x.New(vt10x.WithSize(cfg.cols, cfg.rows)),
		done:       make(chan struct{}),
		probeReply: cfg.probeReply,
	}

	go c.readLoop()

	t.Cleanup(c.close)

	return c
}

func (c *console) readLoop() {
	defer close(c.done)

	buf := make([]byte, 4096)

	for {
		n, err := c.ptmx.Read(buf)
		if n > 0 {
			chunk := buf[:n]

			c.mu.Lock()
			_, _ = c.term.Write(chunk)
			c.mu.Unlock()

			if bytes.Contains(chunk, []byte("\x1b[6n")) {
				c.mu.Lock()
				c.probeCount++
				c.mu.Unlock()

				c.replyCursorPos()
			}
		}

		if err != nil {
			return
		}
	}
}

func (c *console) replyCursorPos() {
	c.mu.Lock()
	cur := c.term.Cursor()
	fn := c.probeReply
	c.mu.Unlock()

	reply := fmt.Sprintf("\x1b[%d;%dR", cur.Y+1, cur.X+1)
	if fn != nil {
		reply = fn(cur)
	}

	if reply != "" {
		_, _ = c.ptmx.WriteString(reply)
	}
}

func (c *console) send(s string) {
	c.t.Helper()

	if _, err := c.ptmx.WriteString(s); err != nil {
		c.t.Fatalf("send %q: %v", s, err)
	}
}

func (c *console) probeQueries() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.probeCount
}

func (c *console) screen() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.term.String()
}

const screenWaitTimeout = 15 * time.Second

func (c *console) waitForScreen(substr string) string {
	c.t.Helper()

	deadline := time.Now().Add(screenWaitTimeout)

	for {
		s := c.screen()
		if strings.Contains(s, substr) {
			return s
		}

		if time.Now().After(deadline) {
			c.t.Fatalf("timed out waiting for %q on screen; got:\n%s", substr, s)
			return s
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func (c *console) waitUntil(cond func(screen string) bool) string {
	c.t.Helper()

	deadline := time.Now().Add(screenWaitTimeout)

	for {
		s := c.screen()
		if cond(s) {
			return s
		}

		if time.Now().After(deadline) {
			c.t.Fatalf("timed out waiting for screen condition; got:\n%s", s)
			return s
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func (c *console) close() {
	_ = c.ptmx.Close()

	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}

	_ = c.cmd.Wait()
	<-c.done
}
