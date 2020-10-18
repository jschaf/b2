package pg

import (
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jschaf/b2/pkg/chans"
	"github.com/jschaf/b2/pkg/logs"
	"github.com/jschaf/b2/pkg/nets"
	"github.com/jschaf/b2/runner"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	pgDevDir = "/d/pgdev"
)

type RunConf struct {
	// The location of the Postgres install to use.
	InstallDir string
	// The path to data directory. If empty, uses a temp dir.
	DataDir string
}

const (
	cancelGracePeriod = time.Second * 5
	defaultPort       = 38541
	pgReadyLogMsg     = "database system is ready to accept connections"
)

type NewClusterConf struct {
	InstallDir string
	DataDir    string
}

// Cluster is a Postgres database cluster.
type Cluster struct {
	InstallDir string
	DataDir    string
	Superuser  string
	Host       string
	Port       string         // use string because we always convert to a string
	Process    runner.Process // backing Postgres process after Start returns
	l          *zap.Logger
}

func NewCluster(c NewClusterConf, l *zap.Logger) *Cluster {
	return &Cluster{
		InstallDir: c.InstallDir,
		DataDir:    c.DataDir,
		Superuser:  "postgres",
		Host:       "localhost",
		Port:       "",
		Process:    nil,
		l:          l,
	}
}

// InitClusterConf contains options to pass to initdb that aren't general
// enough to store in Cluster.
type InitClusterConf struct {
	Encoding string   // default encoding for the cluster, defaults to UTF-8
	Locale   string   // default locale for the cluster, defaults to C
	Args     []string // additional args to pass to initdb
}

func (pgc *Cluster) InitCluster(conf InitClusterConf) error {
	if conf.Encoding == "" {
		conf.Encoding = "UTF-8"
	}
	if conf.Locale == "" {
		conf.Locale = "C"
	}
	if pgc.DataDir == "" {
		dataDir, err := ioutil.TempDir(os.TempDir(), "pgembed-data-dir-")
		if err != nil {
			return fmt.Errorf("create pgembed temp data dir: %w", err)
		}
		pgc.DataDir = dataDir
	}
	pgc.l.Sugar().Infof("pgembed data dir: %s", pgc.DataDir)

	args := []string{
		"--pgdata=" + pgc.DataDir,
		"--username=" + pgc.Superuser,
		"--encoding=" + conf.Encoding,
		"--locale=" + conf.Locale,
	}
	for _, arg := range conf.Args {
		args = append(args, arg)
	}
	initdbProc := runner.NewProcess(runner.ProcessConfig{
		Path:        filepath.Join(pgc.InstallDir, "bin", "initdb"),
		Args:        args,
		Env:         nil,
		Dir:         pgc.DataDir,
		Description: "pgembed-initdb",
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}, pgc.l)
	if err := runProc(initdbProc, time.Second*2); err != nil {
		return fmt.Errorf("run initdb process: %w", err)
	}

	return nil
}

// RemoveAllData deletes the Postgres data dir.
func (pgc *Cluster) RemoveAllData() error {
	return os.RemoveAll(pgc.DataDir)
}

// Start starts the Postgres process, blocking until the Postgres process is
// healthy by examining the log output.
func (pgc *Cluster) Start() error {
	port := defaultPort
	if isOpen, err := nets.IsPortOpen(port); err != nil {
		return fmt.Errorf("IsPortOpen() error: %w", err)
	} else if !isOpen {
		newPort, err := nets.FindAvailablePort()
		if err != nil {
			return fmt.Errorf("find available port: %w", err)
		}
		pgc.l.Info("default port unavailable", zap.Int("port", newPort))
		port = newPort
	}
	pgc.Port = strconv.Itoa(port)

	triggerW := logs.NewTriggerWriter(pgReadyLogMsg)

	pgc.Process = runner.NewProcess(runner.ProcessConfig{
		Path: filepath.Join(pgc.InstallDir, "bin", "postgres"),
		Args: []string{
			// "-d", "1", // debug level [1-5]
			"-D", pgc.DataDir,
			"-p", strconv.Itoa(port),
		},
		Dir:         pgc.DataDir,
		Description: "pgembed-postgres-postmaster",
		Stdout:      os.Stdout,
		Stderr:      io.MultiWriter(os.Stderr, triggerW),
	}, pgc.l)

	go func() {
		if err := runProc(pgc.Process, time.Hour*1); err != nil {
			pgc.l.Error("timeout for postgres process", zap.Error(err))
		}
	}()

	pgc.l.Info("waiting for postgres ready log line")
	startDeadline := time.Second * 5
	if err := triggerW.Wait(startDeadline); err != nil {
		return fmt.Errorf("postgres start failed; ready message %q not found", pgReadyLogMsg)
	}
	pgc.l.Info("found postgres ready log line")

	return nil
}

// ConnConf is the connection options to use to create a new Postgres
// connection to a Postgres Cluster.
type ConnConf struct {
	Database string // name of the database to connect to
}

// NewConn establishes a connection with the Postgres cluster.
func (pgc *Cluster) NewConn(conf ConnConf) (*pgx.Conn, error) {
	port, err := strconv.Atoi(pgc.Port)
	if err != nil {
		return nil, fmt.Errorf("bad port for new connection: %w", err)
	}

	conn, err := pgx.Connect(pgx.ConnConfig{
		Host:     pgc.Host,
		Port:     uint16(port),
		Database: conf.Database,
		User:     pgc.Superuser,
	})
	if err != nil {
		return nil, fmt.Errorf("new cluster connection: %w", err)
	}
	return conn, nil
}

type CreateDBConf struct {
	Database string   // name of the database to create
	Args     []string // extra args to pass to createdb
}

// CreateDB runs Postgres createdb to create a new database in the Postgres
// cluster.
func (pgc *Cluster) CreateDB(c CreateDBConf) error {
	args := make([]string, 0, len(c.Args)+5)
	args = append(args, "--host="+pgc.Host)
	args = append(args, "--port="+pgc.Port)
	args = append(args, "--username="+pgc.Superuser)
	for _, arg := range c.Args {
		args = append(args, arg)
	}
	args = append(args, c.Database)

	createdDBProc := runner.NewProcess(runner.ProcessConfig{
		Path:        filepath.Join(pgc.InstallDir, "bin", "createdb"),
		Args:        args,
		Dir:         pgc.DataDir,
		Description: "pgembed-createdb",
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}, pgc.l)

	if err := runProc(createdDBProc, time.Second*5); err != nil {
		return fmt.Errorf("creatdb: %w", err)
	}
	return nil
}

type BenchConf struct {
	// The name of the already-created database to test in.
	DBName string
	// Additional args to pass to pgbench.
	Args []string
}

func (pgc *Cluster) Bench(conf BenchConf, l *zap.Logger) error {
	args := make([]string, len(conf.Args)+1)
	copy(args, conf.Args)
	args[len(args)-1] = conf.DBName

	pgBenchProc := runner.NewProcess(runner.ProcessConfig{
		Path:        filepath.Join(pgc.InstallDir, "bin", "pgbench"),
		Args:        args,
		Dir:         pgc.DataDir,
		Description: "pgembed-pgbench",
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}, l)
	if err := runProc(pgBenchProc, time.Minute*15); err != nil {
		return fmt.Errorf("run pgbench: %w", err)
	}
	return nil
}

// runProc is a helper function to run a process and cancel it if takes longer
// than timeout.
func runProc(p runner.Process, timeout time.Duration) error {
	go func() { _ = p.Run() }() // error checked in ExitCodeError below
	if timeoutErr := chans.Wait(p.Finished(), timeout); timeoutErr != nil {
		if err := p.Cancel(cancelGracePeriod); err != nil {
			return fmt.Errorf("cancel error after running process exceeded timeout %s: %w",
				p.Config().Description, err)
		} else {
			return fmt.Errorf("successfully canceled process after exceeding timeout %s: %w",
				p.Config().Description, timeoutErr)
		}
	}
	return p.ExitCodeError()
}
