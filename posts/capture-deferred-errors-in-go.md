+++
slug = "capture-deferred-errors-in-go"
date = 2025-01-02
visibility = "published"
+++

# Capture deferred errors in Go

I rarely ignore errors in Go as I usually regret when I do ignore them. Go is
lauded and criticzed for it's straightforward approach to error handling,
leaning on `if err != nil { return fmt.Errrorf(...) }`. That's all well and
good, but one of the more tedious types of errors to capture are errors in a
deferred function. For example, when opening a file:

```go
func populateFile() error
	f, err := os.Open("foo.txt", os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close() // BAD: ignored error
	writeInterestingData(f)
	return nil
}
```

The problem is that we've ignored the error from closing the file. The usual way
is to defer an immediately invoked anonymous function and then overwrite the
named return argument.

```go
func populateFile() (err error)
	f, err := os.Open("foo.txt", os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); err != nil {
			// BAD: overwrites existing error
			err = fmt.Errorf("close file: %w", closeErr) 
		}
	}()
	writeInterestingData(f)
	return nil
}
```

## Close by Thanos

Inspired by Thanos' [coding style guide] we can simplify the somewhat unwieldy
anonymous function. Thanos defines [`CloseWithErrCapture`], which calls Close
and joins the error with an existing error, using a pointer to an error. The
function is:

[coding style guide]: https://thanos.io/tip/contributing/coding-style-guide.md/#defers-dont-forget-to-check-returned-errors
[`CloseWithErrCapture`]: https://github.com/thanos-io/thanos/blob/ca40906c83d94cfcbe4bcc181a286663aeb268d5/pkg/runutil/runutil.go#L156,

```go
// CloseWithErrCapture closes closer, wraps any error with message from fmt and args, and stores this in err.
func CloseWithErrCapture(err *error, closer io.Closer, format string, a ...interface{}) {
	merr := errutil.MultiError{}

	merr.Add(*err)
	merr.Add(errors.Wrapf(closer.Close(), format, a...))

	*err = merr.Err()
}
```

Armed with our helper, we simplify populateFile to:

```go
func populateFile() (err error)
	f, err := os.Open("foo.txt", os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer CloseWithErrCapture(&err, f, "close file")
	writeInterestingData(f)
	return nil
}
```

## An err of simplicity

I happily used Thanos' approach but was troubled that I could use it for similar
cleanup functions like Flush, Stop, Shutdown. Additionally, Go 1.20 introduced
[￼`errors.Join`￼](https://pkg.go.dev/errors#Join), meaning we can skip defining
a custom `MultiError` type.

```go
package errs

import (
	"errors"
	"fmt"
)

// Capture runs errFunc and assigns the error, if any, to *errPtr. Preserves the
// original error by wrapping with errors.Join if errFunc returns a non-nil
// error.
func Capture(errPtr *error, errFunc func() error, msg string) {
	err := errFunc()
	if err == nil {
		return
	}
	*errPtr = errors.Join(*errPtr, fmt.Errorf("%s: %w", msg, err))
}
```

At the call site, we pass the function method instead of an interface type
implementing `io.Closer`.

```go
func populateFile() (err error)
	f, err := os.Open("foo.txt", os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer errs.Capture(&err, f.Close, "close file")
	writeInterestingData(f)
	return nil
}
```

I think `errs.Capture` improves on `CloseWithErrCapture` by capturing any error
returning function, and is less wordy to boot.
