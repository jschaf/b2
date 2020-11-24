// +build linux

package pg

import (
	"github.com/jschaf/b2/pkg/chans"
	"github.com/jschaf/b2/pkg/errs"
	"github.com/jschaf/b2/pkg/texts"
	"go.uber.org/zap/zaptest"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestExperiment_Run(t *testing.T) {
	dir, err := ioutil.TempDir("/dev/shm", "test-experiment-")
	if err != nil {
		t.Fatal(err)
	}
	pgc := NewCluster(NewClusterConf{
		InstallDir: pgDevDir,
		DataDir:    dir,
	}, zaptest.NewLogger(t))
	defer errs.CapturingT(t, pgc.RemoveAllData, "remove all data")

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

	ddl := texts.Dedent(`
		create table event (
			event_id bigint not null,
			user_id bigint not null,
			time bigint not null,
			session_id bigint not null,
			data jsonb not null
		);
	`)
	experiment := NewExperiment(ExperimentConf{
		Cluster: pgc,
		DDL:     strings.NewReader(ddl),
	}, zaptest.NewLogger(t))

	if err := experiment.Init(); err != nil {
		t.Fatal(err)
	}
	if err := experiment.Populate(); err != nil {
		t.Fatal(err)
	}
	if err := experiment.Run(); err != nil {
		t.Fatal(err)
	}

	if err := pgc.Process.Cancel(time.Second * 5); err != nil {
		t.Fatalf("stop postgres process: %s", err)
	}

	if err := chans.Wait(pgc.Process.Finished(), time.Second*10); err != nil {
		t.Fatalf("timeout waiting for postgres process to finish: %s", err)
	}
}
