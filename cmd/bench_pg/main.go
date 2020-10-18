package main

import (
	"fmt"
	"github.com/jschaf/b2/pkg/errs"
	"github.com/jschaf/b2/pkg/logs"
	"github.com/jschaf/b2/pkg/pg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"time"
)

const (
	pgDevDir = "/d/pgdev"
)

func run(l *zap.Logger) (mErr error) {
	pgc := pg.NewCluster(pg.NewClusterConf{
		InstallDir: pgDevDir,
	}, l)

	defer errs.CapturingErr(&mErr,
		func() error { return os.RemoveAll(pgc.DataDir) },
		"remove temp postgres data dir")

	if err := pgc.InitCluster(pg.InitClusterConf{}); err != nil {
		return fmt.Errorf("postgres init DB: %w", err)
	}

	if err := pgc.Start(); err != nil {
		return fmt.Errorf("postgres start: %w", err)
	}

	createDBConf := pg.CreateDBConf{Database: "bench"}
	if err := pgc.CreateDB(createDBConf); err != nil {
		return err
	}

	err := pgc.Bench(pg.BenchConf{
		DBName: createDBConf.Database,
		Args: []string{
			"--initialize",
			"--port=" + pgc.Port,
			"--username=" + pgc.Superuser,
		},
	}, l)
	if err != nil {
		return err
	}

	err = pgc.Bench(pg.BenchConf{
		DBName: createDBConf.Database,
		Args: []string{
			"--client=10",
			"--port=" + pgc.Port,
			"--username=" + pgc.Superuser,
		},
	}, l)
	if err != nil {
		return err
	}
	if err := pgc.Process.Cancel(time.Minute); err != nil {
		return err
	}

	<-pgc.Process.Finished()
	return nil
}

func main() {
	l, err := logs.NewShortDevLogger(zapcore.DebugLevel)
	if err != nil {
		log.Fatalf("create zap logger: %s", err.Error())
	}

	if err := run(l); err != nil {
		l.Fatal("run bench_pg", zap.Error(err))
	}
}
