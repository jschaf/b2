package runner

import (
	"bytes"
	"fmt"
	"github.com/jschaf/b2/pkg/errs"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"go.uber.org/zap"
)

const largeString = `long long long long long long long long long long long ` +
	`long long long long long long long long long long long long long long ` +
	`long long long long long long long long long long long long long long `

const (
	behaviorEnvVar            = "PROCESS_TEST_BEHAVIOR"
	behaviorWriteStdoutStderr = "write_stdout_stderr"
	behaviorWriteLargeStdout  = "write_large_stdout"
	behaviorChangeWorkingDir  = "change_working_dir"
	behaviorWaitForSIGINT     = "wait_for_sigint"
	behaviorWaitForSIGKILL    = "wait_for_sigkill"
	behaviorExitCode          = "exit_code"
	behaviorIgnored           = "ignored"
	largeOutputSample         = "header\nalpha\nbravo\n" + largeString + "\ncharlie"
	stdoutSample              = "alpha"
	stderrSample              = "bravo"
	exitCodeSample            = 88
)

// go test uses TestMain if it exists. We use it to control behavior of individual
// tests with an env var. The way this works is:
//
// 1. go test builds a binary for this test file.
// 2. Since TestMain is defined, the test binary runs TestMain.
// 3. behaviorEnvVar is unset, so the binary runs the default case in TestMain
//    which runs test cases in the file.
// 4. Individual tests cases re-run the test binary with os.Args[0] for this
//    test file but set behaviorEnvVar.
// 5. The test binary runs TestMain but since behaviorEnvVar is set, we don't
//    re-run all tests and instead manipulate stdout, stderr, or signals to
//    test behavior.
//
// See https://golang.org/pkg/testing/#hdr-Main
func TestMain(m *testing.M) {
	switch os.Getenv(behaviorEnvVar) {
	case behaviorWriteLargeStdout:
		for _, s := range strings.Split(largeOutputSample, "\n") {
			fmt.Printf(s + "\n")
			time.Sleep(time.Millisecond * 5)
		}

	case behaviorWaitForSIGINT:
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		recv := <-signals
		fmt.Printf("Received signal %v\n", recv)

	case behaviorWaitForSIGKILL:
		sigints := make(chan os.Signal, 1)
		signal.Notify(sigints, syscall.SIGINT)
		recvInt := <-sigints
		fmt.Printf("Received signal %v\n", recvInt)
		time.Sleep(time.Second)

	case behaviorWriteStdoutStderr:
		fmt.Fprintf(os.Stdout, stdoutSample)
		fmt.Fprintf(os.Stderr, stderrSample)
		// write twice to expose any buffering issues.
		fmt.Fprintf(os.Stdout, stdoutSample)
		fmt.Fprintf(os.Stderr, stderrSample)

	case behaviorExitCode:
		os.Exit(exitCodeSample)

	case behaviorChangeWorkingDir:
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(os.Stdout, filepath.Base(dir))
		os.Exit(0)

	default:
		os.Exit(m.Run())
	}
}

func TestProcess_StdoutStderr(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	p := NewTestProc(
		ProcessConfig{Stdout: stdout, Stderr: stderr},
		behaviorWriteStdoutStderr)

	if err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if s, err := ioutil.ReadAll(stdout); err != nil {
		t.Error(err)
	} else if string(s) != stdoutSample+stdoutSample {
		t.Errorf("Expected Stdout() %s, got %s", stdoutSample+stdoutSample, string(s))
	}

	if s, err := ioutil.ReadAll(stderr); err != nil {
		t.Error(err)
	} else if string(s) != stderrSample+stderrSample {
		t.Errorf("Expected Stderr() %s, got %s", stderrSample+stderrSample, string(s))
	}

	assertProcessWasKilled(t, p)
}

func TestProcess_ExitCodeError(t *testing.T) {
	p := NewTestProc(ProcessConfig{}, behaviorExitCode)

	if err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if p.WaitStatus().ExitStatus() != exitCodeSample {
		t.Errorf("expected error code %d, got %d", exitCodeSample, p.WaitStatus().ExitStatus())
	}

	err := p.ExitCodeError()
	if err == nil {
		t.Fatal("expected ExitCodeError() to be non-nil error; got nil")
	}

	if !strings.Contains(err.Error(), strconv.Itoa(exitCodeSample)) {
		t.Errorf("expected ExitCodeError() to contain error code %d, got %s", exitCodeSample, err)
	}

	assertProcessWasKilled(t, p)
}

func TestProcess_RunRunErrors(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	p := NewTestProc(
		ProcessConfig{Stdout: stdout, Stderr: stderr},
		behaviorWriteStdoutStderr)

	if err := p.Run(); err != nil {
		t.Error(err)
	}
	if err := p.Run(); err == nil {
		t.Error("expected 2nd Run() call to error")
	}
}

func TestProcess_RunRunPreservesFirstError(t *testing.T) {
	p := NewTestProc(ProcessConfig{Path: "<not_real_binary>"}, behaviorIgnored)

	if err := p.Run(); err == nil {
		t.Fatal("expected process to fail because binary doesn't exist")
	}
	exitErr1 := p.ExitCodeError().Error()
	if !strings.Contains(exitErr1, p.Config().Path) {
		t.Fatalf("expected ExitCodeError() to contain %s; got %s",
			p.Config().Path, exitErr1)
	}

	if err := p.Run(); err == nil {
		t.Error("expected 2nd Run() call to error")
	}

	exitErr2 := p.ExitCodeError().Error()
	if !strings.Contains(exitErr2, p.Config().Path) {
		t.Fatalf("expected ExitCodeError() to contain %s; got %s",
			p.Config().Path, exitErr2)
	}

	if !strings.Contains(exitErr2, "attempted to run") {
		t.Fatalf("expected ExitCodeError() to contain 'attempted to run'; got %s",
			exitErr2)
	}
}

func TestProcess_SignalsStartAndFinished(t *testing.T) {
	p := NewTestProc(ProcessConfig{}, behaviorWriteLargeStdout)

	var wasStarted, wasFinished bool

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-p.Started()
		wasStarted = true
		<-p.Finished()
		wasFinished = true
	}()

	if err := p.Run(); err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	if !wasStarted {
		t.Error("Expected started channel to be closed")
	}
	if !wasFinished {
		t.Error("Expected finished channel to be closed")
	}
	assertProcessWasKilled(t, p)
}

func NewTestProc(c ProcessConfig, behavior string) Process {
	if c.Path == "" {
		c.Path = os.Args[0]
	}
	if c.Stdout == nil {
		c.Stdout = ioutil.Discard
	}
	if c.Stderr == nil {
		c.Stderr = ioutil.Discard
	}
	if behavior != "" {
		c.Env = append(c.Env, behaviorEnvVar+"="+behavior)
	}
	return NewProcess(c, zap.NewNop())
}

func TestProcess_SignalGroup_WaitsForSIGINT(t *testing.T) {
	p := NewTestProc(ProcessConfig{}, behaviorWaitForSIGINT)

	var signalError error

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-p.Started()
		signalError = p.SignalGroup(syscall.SIGINT)
		<-p.Finished()
	}()

	if err := p.Run(); err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	if signalError != nil {
		t.Errorf("expected no error from SignalGroup(); got %s", signalError)
	}

	assertProcessWasKilled(t, p)
}

func TestProcess_Cancel_WaitsForSIGKILL(t *testing.T) {
	p := NewTestProc(ProcessConfig{}, behaviorWaitForSIGKILL)

	var cancelErr error

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-p.Started()
		cancelErr = p.Cancel(time.Millisecond * 5)
		<-p.Finished()
	}()

	if err := p.Run(); err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	if cancelErr != nil {
		t.Errorf("expected no error from Cancel(); got %s", cancelErr)
	}

	if p.ExitCodeError() == nil {
		t.Errorf("expected exit code error but was nil")
	}

	assertProcessWasKilled(t, p)
}

func TestProcess_SignalsStartAndFinishedOnError(t *testing.T) {
	p := NewProcess(ProcessConfig{Path: "<not_a_real_binary>"}, zap.NewNop())

	var wasStarted, wasFinished bool
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-p.Started():
			wasStarted = true
		case <-time.After(time.Millisecond * 20):
			t.Error("failed to read started channel in time")
		}

		select {
		case <-p.Finished():
			wasFinished = true
		case <-time.After(time.Millisecond * 20):
			t.Error("failed to read finished channel in time")
		}
	}()

	if err := p.Run(); err == nil {
		t.Fatal("expected Run() to fail because the binary doesn't exist")
	}

	wg.Wait()

	if !wasStarted {
		t.Fatalf("expected started channel to be closed")
	}
	if !wasFinished {
		t.Fatalf("expected finished channel to be closed")
	}
	assertProcessWasKilled(t, p)
}

func TestProcess_ChangesWorkingDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "go_test_process_current_dir_")
	if err != nil {
		t.Fatal(err)
	}
	defer errs.TestCapturingErr(t, func() error { return os.RemoveAll(dir) }, "remove temp dir")
	// Use basename to skip resolving symlinks.
	basename := filepath.Base(dir)

	stdout := &bytes.Buffer{}
	p := NewTestProc(ProcessConfig{Dir: dir, Stdout: stdout}, behaviorChangeWorkingDir)

	if err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if s, err := ioutil.ReadAll(stdout); err != nil {
		t.Error(err)
	} else if string(s) != basename {
		t.Errorf("Expected Stdout() %s, got %s", basename, string(s))
	}

	assertProcessWasKilled(t, p)
}

func assertProcessWasKilled(t *testing.T, p Process) {
	t.Helper()

	proc, err := os.FindProcess(p.PID())
	if err != nil {
		return
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		t.Fatalf("process %d exists and is running but should have been killed", p.PID())
	}
}
