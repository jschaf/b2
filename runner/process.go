// Package process manages operating system processes.
package runner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type Process interface {
	// Run starts the process and blocks until it finishes.
	//
	// NOTE: Run only errors on setup errors like if the process was already
	// started or if the process path was empty. If the underlying process
	// has a non-zero exit code, Run does not error. Use ExitCodeError to get
	// an error on a non-zero exit code.
	Run() error

	// Config returns the settings used to create this process.
	Config() ProcessConfig

	// Started closes once the process is started. Implementations always close
	// this channel and always close it before Finished. Once a process is
	// started, the process was either started with a PID or failed to run.
	Started() <-chan struct{}

	// Finished closes once the process finishes either normally or is killed.
	// Implementations always close this channel and always close it after
	// Started.
	Finished() <-chan struct{}

	// WaitStatus returns the status of exec.Cmd.Wait.
	WaitStatus() syscall.WaitStatus

	// WaitResult returns the raw error of exec.Cmd.Wait.
	WaitResult() error

	// PID returns the PID of a started process or 0 if the process hasn't been
	// run yet.
	PID() int

	// SignalGroup sends the signal s to this process and any processes within the
	// same process group as this process, meaning any child processes.
	SignalGroup(s syscall.Signal) error

	// Cancel sends the signal SIGINT to the process. If the process is not
	// finished after gracePeriod, Cancel sends SIGKILL.
	//
	// - If the process is not yet Run, Cancel ensure this process will not Run.
	//   Any subsequent calls to Run will error.
	// - If the process is already Finished, Cancel does nothing.
	Cancel(gracePeriod time.Duration) error

	// ExitCodeError returns an error if the process fails to run or if the
	// process finished with a non-zero exit code.
	ExitCodeError() error
}

// process is an OS process.
type process struct {
	config            ProcessConfig
	waitResult        error
	cmd               *exec.Cmd
	runErr            error
	pid               int
	logger            *zap.Logger
	started, finished chan struct{}
	mu                sync.Mutex
	status            syscall.WaitStatus
	isRun             bool
}

// Config contains the parameters used to start and run an OS process.
type ProcessConfig struct {
	Path        string
	Args        []string
	Env         []string
	Dir         string
	Description string
	Stdout      io.Writer
	Stderr      io.Writer
}

// New creates a new process from a Config.
func NewProcess(c ProcessConfig, l *zap.Logger) Process {
	descLogger := l
	if c.Description != "" {
		descLogger = l.With(zap.String("description", c.Description))
	}

	p := process{
		logger:   descLogger,
		config:   c,
		started:  make(chan struct{}),
		finished: make(chan struct{}),
	}
	return &p
}

func (p *process) Config() ProcessConfig {
	return p.config
}

func (p *process) WaitStatus() syscall.WaitStatus {
	return p.status
}

func (p *process) WaitResult() error {
	return p.waitResult
}

func (p *process) PID() int {
	return p.pid
}

// Run executes the process and blocks until it finishes. Run only returns an
// error if it failed to start a process or if it was called more than once.
// The error from the process is stored in waitResult and status. Run always
// closes both the Started and Finished channels in order.
func (p *process) Run() (mErr error) {
	isFirstRun := true // updated in mutex below before checked in defer
	defer func() {
		if isFirstRun {
			p.runErr = mErr
			close(p.finished)
		} else {
			// mErr is "attempted to run already started process". Preserve the original
			// run err with error wrapping.
			p.runErr = fmt.Errorf(mErr.Error()+" - original error: %w", p.runErr)
		}
	}()

	// Ensure we only start this process once.
	p.mu.Lock()
	if p.isRun {
		isFirstRun = false
		return errors.New("attempted to run already started process")
	}
	p.isRun = true
	p.mu.Unlock()

	p.cmd = exec.Command(p.config.Path, p.config.Args...)

	// Start a process group so we can kill any child processes that this
	// process creates.
	p.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	// Wire up input and output.
	p.cmd.Stdout = p.config.Stdout
	p.cmd.Stderr = p.config.Stderr
	p.cmd.Stdin = nil

	// Copy the host env to get env vars like PATH. Copy the config path last
	// to overwrite the host env with the config env.
	p.cmd.Env = append(os.Environ(), p.config.Env...)

	p.cmd.Dir = p.config.Dir

	err := p.cmd.Start()
	close(p.started) // signal to consumers that the process is started
	if err != nil {
		return err
	}
	p.pid = p.cmd.Process.Pid
	pidLogger := p.logger.With(zap.Int("pid", p.pid))
	pidLogger.Info("process running")

	// Wait for the command to complete. The returned error is nil only if:
	// - The process had a exit code of 0.
	// - All IO finished to stdout and stderr.
	p.waitResult = p.cmd.Wait()
	pidLogger.Info("process finished")

	// Get wait status from the wait result.
	if p.waitResult != nil {
		if err, ok := p.waitResult.(*exec.ExitError); ok {
			if s, ok := err.Sys().(syscall.WaitStatus); ok {
				p.status = s
			} else {
				return errors.New("unable to get status from wait result")
			}
		}
	}

	return nil
}

// Started returns a channel that is closed once the process is started.
func (p *process) Started() <-chan struct{} {
	return p.started
}

// Finished returns a channel that is closed once the process is done.
func (p *process) Finished() <-chan struct{} {
	return p.finished
}

func (p *process) ExitCodeError() error {
	if p.runErr != nil {
		return fmt.Errorf("run process: %w", p.runErr)
	}
	if p.status.ExitStatus() == 0 {
		return nil
	}

	return fmt.Errorf("process %s failed exitCode=%d", p.config.Description, p.status.ExitStatus())
}

func (p *process) SignalGroup(s syscall.Signal) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Prevent killing PID 1 (usually the login daemon or shell). PID of 0 or -1
	// means uninitialized which should only happen if Run fails to start the
	// process.
	if p.pid < 2 {
		return fmt.Errorf("cannot cancel PID %d", p.pid)
	}
	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("cancel process with no cmd description=%s", p.Config().Description)
	}

	p.logger.Info("Sending signal to process", zap.Int("signal", int(s)), zap.Int("PID", p.pid))
	// The negative PID kills all processes within the process group started by the PID.
	// https://stackoverflow.com/a/11000554/30900
	return syscall.Kill(-p.pid, s)
}

func (p *process) Cancel(gracePeriod time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.Started():
		break
	default:
		p.isRun = true // prevent running the process after Cancel
		return nil     // process not yet started, so do nothing
	}

	select {
	case <-p.Finished():
		return nil // process already finished, nothing to cancel
	default:
		break
	}

	// Prevent killing PID 1 (usually the login daemon or shell). PID of 0 or -1
	// means uninitialized which should only happen if Run fails to start the
	// process.
	if p.pid < 2 {
		return fmt.Errorf("cannot cancel PID %d", p.pid)
	}
	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("cancel process with no cmd; description=%s", p.Config().Description)
	}

	// If we get here, the process is still running. Try SIGINT first.
	// Not using SignalGroup because it locks the mutex we already locked.
	if err := syscall.Kill(-p.pid, syscall.SIGINT); err != nil {
		return fmt.Errorf("send signal %s to process %s: %w",
			syscall.SIGINT, p.Config().Description, err)
	}

	select {
	case <-p.Finished():
		return nil
	case <-time.After(gracePeriod):
		p.logger.Info("process not canceled in time, terminating")
		if err := syscall.Kill(-p.pid, syscall.SIGKILL); err != nil {
			return fmt.Errorf("send signal %s to process %s: %w",
				syscall.SIGKILL, p.Config().Description, err)
		}
		return nil
	}

}
