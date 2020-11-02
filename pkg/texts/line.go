package texts

import (
	"bufio"
	"strings"
)

func FirstLine(s string) string {
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func Join(s ...string) string {
	return strings.Join(s, "")
}

func JoinSpace(s ...string) string {
	return strings.Join(s, " ")
}
