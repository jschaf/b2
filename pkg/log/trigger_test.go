package log

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

func newStringReaders(ss ...string) io.Reader {
	rs := make([]io.Reader, len(ss))
	for i, s := range ss {
		rs[i] = strings.NewReader(s)
	}
	return io.MultiReader(rs...)
}

func TestTriggerWriter_Wait(t *testing.T) {
	tests := []struct {
		trigger string
		reader  io.Reader
		wantErr error
	}{
		{"foo", newStringReaders("foo\n"), nil},
		{"foo", newStringReaders("foo\nqux"), nil},
		{"foo", newStringReaders("qux\nfoo\nqux"), nil},
		{"foo", newStringReaders("qux\nbaz\n"), TriggerNotFoundErr},
	}
	for _, tt := range tests {
		t.Run(tt.trigger, func(t1 *testing.T) {
			tw := NewTriggerWriter(tt.trigger)
			var copyErr error
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, copyErr = io.Copy(tw, tt.reader)
				if err := tw.Close(); err != nil {
					t.Error(err)
				}
			}()
			if err := tw.Wait(time.Second); err != tt.wantErr {
				t.Errorf("Wait() call 1 - expected error %s; got %s", tt.wantErr, err)
			}
			// Wait should return the same thing if called again.
			if err := tw.Wait(time.Second); err != tt.wantErr {
				t.Errorf("Wait() call 2 - expected error %s; got %s", tt.wantErr, err)
			}
			wg.Wait()
			if copyErr != nil {
				t.Error(copyErr)
			}
		})
	}
}

func TestTriggerWriter_DoesntBlock(t *testing.T) {
	tests := []struct {
		trigger string
		reader  io.Reader
	}{
		{"foo", newStringReaders("qux\nfoo\nbar\n")},
		{"foo", newStringReaders(
			"qux\nfoo\n", strings.Repeat("qux-bar-baz-qux\n", 1024))},
		{
			"foo",
			newStringReaders(
				"qux\nbar\n",
				strings.Repeat("qux-bar-baz-qux\n", 1024),
				"qux\nfoo\n",
				"end",
				strings.Repeat("qux-bar-baz-qux\n", 1024),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.trigger, func(t1 *testing.T) {
			tw := NewTriggerWriter(tt.trigger)
			buf := bytes.NewBuffer(make([]byte, 10))
			mw := io.MultiWriter(tw, buf)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, copyErr := io.Copy(mw, tt.reader); copyErr != nil {
					t.Error(copyErr)
				}
				if err := tw.Close(); err != nil {
					t.Error(err)
				}
			}()
			if err := tw.Wait(time.Second); err != nil {
				t.Error(err)
			}
			wg.Wait()
			// If tw blocks on writes, we'll get stuck trying to io.ReadAll
			// from the other writer.
			if _, err := io.ReadAll(buf); err != nil {
				t.Error(err)
			}
		})
	}
}
