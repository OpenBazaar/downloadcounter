package connect

import (
	"database/sql"
	"os"

	"github.com/gocraft/health"

	// Import MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

// Connections contains the connections used for the package
type Connections struct {
	Config
	DB     *sql.DB
	Stream *health.Stream
}

// NewConnections returns a `*Connections` for the given `*Config`
func NewConnections(conf Config) (*Connections, error) {
	conns := &Connections{Config: conf}

	InitHealth(conns, conf)

	err := InitMySQL(conns, conf)
	if err != nil {
		return nil, err
	}

	return conns, err
}

// NewConnectionsFromEnv returns a `*Connections` using the environment as a data source
func NewConnectionsFromEnv() (*Connections, error) {
	return NewConnections(NewConfigFromEnv())
}

// CloseAll terminates all connections gracefully
func (c *Connections) CloseAll() error {
	err := c.DB.Close()
	return err
}

// InitHealth initializes a `*health.Stream` for instrumentation
func InitHealth(conns *Connections, conf Config) {
	conns.Stream = health.NewStream()
	conns.Stream.AddSink(&health.WriterSink{os.Stdout})
}

// InitMySQL initializes a db connection
func InitMySQL(conns *Connections, conf Config) error {
	var err error
	conns.DB, err = sql.Open("mysql", conf.MySQLDSN)
	if err != nil {
		return err
	}

	return conns.DB.Ping()
}
