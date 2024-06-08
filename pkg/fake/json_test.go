package fake

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type entryOpt = func(encoder *UnsafeJSONEncoder)

func strEntry(key, val string) entryOpt {
	return func(enc *UnsafeJSONEncoder) {
		enc.WriteStringEntry(key, val)
	}
}

func intEntry(key string, val int) entryOpt {
	return func(enc *UnsafeJSONEncoder) {
		enc.WriteIntEntry(key, val)
	}
}

func TestNewUnsafeJSONEncoder(t *testing.T) {
	tests := []struct {
		name    string
		entries []entryOpt
		want    string
	}{
		{"one string", []entryOpt{strEntry("foo", "bar")}, `{"foo":"bar"}`},
		{"one number", []entryOpt{intEntry("foo", 1234)}, `{"foo":1234}`},
		{"negative number", []entryOpt{intEntry("foo", -10000)}, `{"foo":-10000}`},
		{
			"two strings",
			[]entryOpt{strEntry("foo", "bar"), strEntry("qux", "baz")},
			`{"foo":"bar","qux":"baz"}`,
		},
		{
			"string-number",
			[]entryOpt{strEntry("foo", "bar"), intEntry("qux", -999)},
			`{"foo":"bar","qux":-999}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := NewUnsafeJSONEncoder(EncoderConfig{})
			enc.StartObject()
			for _, entry := range tt.entries {
				entry(enc)
			}
			got := string(enc.EndObject())
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("UnsafeJSONEncoder mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewUnsafeJSONEncoder_WriteInt(t *testing.T) {
	for i := 0; i < 10000; i++ {
		is := strconv.Itoa(i)
		keepGoing := true
		enc := NewUnsafeJSONEncoder(EncoderConfig{})
		enc.StartObject()
		enc.WriteIntEntry("m", i)
		enc.WriteIntEntry("n", -i)
		got := string(enc.EndObject())
		neg := "-" + is
		if i == 0 {
			neg = is
		}
		want := `{"m":` + is + `,"n":` + neg + "}"
		if diff := cmp.Diff(want, got); diff != "" {
			keepGoing = false
			t.Fatalf("WriteIntEntry mismatch (-want +got):\n%s", diff)
		}
		if !keepGoing {
			// Avoid spamming 10k tests if the implementation is broken.
			break
		}
	}
}

func TestNewUnsafeJSONEncoder_WriteInt_strangeNums(t *testing.T) {
	type step struct {
		lo  int
		inc func(int) int
	}
	steps := []step{
		{1, func(i int) int { return i * 10 }},
		{9, func(i int) int { return i*10 + 9 }},
		{7, func(i int) int { return i * 13 }},
	}

	for _, step := range steps {
		name := fmt.Sprintf("lo=%d next=%d", step.lo, step.inc(step.lo))
		t.Run(name, func(t *testing.T) {
			for i := step.lo; i > 0; i = step.inc(i) {
				is := strconv.Itoa(i)
				enc := NewUnsafeJSONEncoder(EncoderConfig{})
				enc.StartObject()
				enc.WriteIntEntry("m", i)
				enc.WriteIntEntry("n", -i)
				got := string(enc.EndObject())
				want := `{"m":` + is + `,"n":-` + is + "}"
				if diff := cmp.Diff(want, got); diff != "" {
					t.Fatalf("WriteIntEntry mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func BenchmarkUnsafeJSONEncoder_WriteStringEntry(b *testing.B) {
	for i := 0; i < b.N; i++ {
		enc := NewUnsafeJSONEncoder(EncoderConfig{})
		enc.StartObject()
		for j := 0; j < 32; j++ {
			enc.WriteStringEntry("foo-bar", "alpha-bravo-charlie")
		}
		enc.EndObject()
	}
}

func BenchmarkUnsafeJSONEncoder_WriteIntEntry(b *testing.B) {
	for i := 0; i < b.N; i++ {
		enc := NewUnsafeJSONEncoder(EncoderConfig{})
		enc.StartObject()
		for j := 0; j < 32; j++ {
			enc.WriteIntEntry("foo", (j*31)<<12)
		}
		enc.EndObject()
	}
}
