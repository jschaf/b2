package nets

import (
	"github.com/jschaf/b2/pkg/errs"
	"net"
	"strconv"
	"testing"
)

func TestFindAvailablePort(t *testing.T) {
	port, err := FindAvailablePort()
	if err != nil {
		t.Error(err)
	}
	if port == 0 {
		t.Error("port == 0")
	}
	l, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		t.Error(err)
	}
	defer errs.CapturingT(t, l.Close, "")
}
