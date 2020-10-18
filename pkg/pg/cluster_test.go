package pg

import (
	"github.com/jschaf/b2/pkg/chans"
	"github.com/jschaf/b2/pkg/errs"
	"go.uber.org/zap/zaptest"
	"io/ioutil"
	"testing"
	"time"
)

func TestCluster_Start(t *testing.T) {
	dir, err := ioutil.TempDir("/dev/shm", "test-cluster-")
	if err != nil {
		t.Fatal(err)
	}
	pgc := NewCluster(NewClusterConf{
		InstallDir: pgDevDir,
		DataDir:    dir,
	}, zaptest.NewLogger(t))
	defer errs.TestCapturingErr(t, pgc.RemoveAllData, "remove all data")

	if err := pgc.InitCluster(InitClusterConf{}); err != nil {
		t.Fatalf("init db: %s", err)
	}

	if err := pgc.Start(); err != nil {
		t.Fatalf("start db: %s", err)
	}

	createDBConf := CreateDBConf{Database: "test_db", Args: []string{}}
	if err := pgc.CreateDB(createDBConf); err != nil {
		t.Fatalf("create db: %s", err)
	}

	conn, err := pgc.NewConn(ConnConf{Database: createDBConf.Database})
	if err != nil {
		t.Fatalf("connect to postgres: %s", err)
	}
	var n int32
	if err := conn.QueryRow("SELECT 42").Scan(&n); err != nil {
		t.Fatalf("query postgres: %s", err)
	}
	if n != 42 {
		t.Errorf("expected SELECT to return 42; got %d", n)
	}

	if err := pgc.Process.Cancel(time.Second * 10); err != nil {
		t.Fatalf("stop postgres process: %s", err)
	}

	if err := chans.Wait(pgc.Process.Finished(), time.Second*10); err != nil {
		t.Fatalf("timeout waiting for postgres process to finish: %s", err)
	}
}
