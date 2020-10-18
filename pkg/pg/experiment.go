package pg

import (
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jschaf/b2/pkg/api"
	"github.com/jschaf/b2/pkg/fake"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"time"
)

type Experiment struct {
	// The DDL (data definition language) necessary to create tables and
	// indexes to run the experiment.
	Cluster *Cluster
	DDL     io.Reader
	l       *zap.Logger
}

type ExperimentConf struct {
	Cluster *Cluster
	DDL     io.Reader
}

func NewExperiment(conf ExperimentConf, l *zap.Logger) *Experiment {
	ex := Experiment{
		Cluster: conf.Cluster,
		DDL:     conf.DDL,
		l:       l,
	}
	return &ex
}

func (ex *Experiment) Init() error {
	ex.l.Info("init experiment")
	conn, err := ex.Cluster.NewConn(ConnConf{})
	if err != nil {
		return err
	}
	ddl, err := ioutil.ReadAll(ex.DDL)
	if err != nil {
		return fmt.Errorf("read DDL: %w", err)
	}
	if _, err := conn.Exec(string(ddl)); err != nil {
		return fmt.Errorf("exec DDL: %w", err)
	}
	return nil
}

func (ex *Experiment) Populate() error {
	ex.l.Info("populate experiment")
	conn, err := ex.Cluster.NewConn(ConnConf{})
	if err != nil {
		return err
	}
	startFaker := time.Now()
	evs := make([]api.Event, 1<<14)
	faker := fake.NewEventFaker()
	if err := faker.WriteEvents(evs); err != nil {
		return fmt.Errorf("populate fake events: %w", err)
	}
	ex.l.Info("experiment created data", zap.Duration("duration", time.Since(startFaker)))

	evSrc := NewEventSource(evs)

	start := time.Now()
	n, err := conn.CopyFrom(pgx.Identifier{"event"}, evSrc.Rows(), evSrc)
	if err != nil {
		return fmt.Errorf("copy from event source: %w", err)
	} else if n != len(evs) {
		return fmt.Errorf(
			"copy from event source short read - inserted %d rows; expected %d total rows: %w",
			n, len(evs), err)
	}
	ex.l.Info("experiment populated data", zap.Duration("duration", time.Since(start)))
	return nil
}

func (ex *Experiment) Run() error {
	ex.l.Info("run experiment")
	start := time.Now()
	conn, err := ex.Cluster.NewConn(ConnConf{})
	if err != nil {
		return fmt.Errorf("new conn: %w", err)
	}
	limit := 1 << 1
	for i := 0; i < limit; i++ {
		_, err = conn.Exec("SELECT count(*) FROM event where time between 1 and 1000000000")
		if err != nil {
			return fmt.Errorf("exec sum query: %w", err)
		}
	}
	ex.l.Info("experiment ran", zap.Int("numTimes", limit), zap.Duration("duration", time.Since(start)))
	return nil
}
