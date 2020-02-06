package markdown

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"h1 > p", "<h1>hello world</h1>\n<p>para</p>\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := Render(); err != nil {
				t.Errorf("Render() error: %s", err)
			} else if !reflect.DeepEqual(got.String(), tt.want) {
				t.Errorf("Render() = %v, want %v", got, tt.want)
			}
		})
	}
}
