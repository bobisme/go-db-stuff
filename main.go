package main

import (
	"database/sql"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/bobisme/oczeroexporter"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"go.opencensus.io/trace"
)

func panicIf(err error) {
	if err != nil {
		log.Panic().Err(err).Msg("")
	}
}

type PersonRecord struct {
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string
}

func initdb() *sqlx.DB {
	driverName, err := ocsql.Register("pgx", ocsql.WithAllTraceOptions())
	panicIf(err)
	db, err := sql.Open(driverName, "postgres://postgres@localhost:5432/postgres")
	panicIf(err)
	// connConfig, err := pgx.ParseURI("postgres://postgres@localhost:5432/postgres")
	// panicIf(err)
	// connPool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
	// 	ConnConfig:     connConfig,
	// 	AfterConnect:   nil,
	// 	MaxConnections: 20,
	// 	AcquireTimeout: 30 * time.Second,
	// })
	// panicIf(err)
	// db, err := stdlib.OpenFromConnPool(connPool)
	// if err != nil {
	// 	connPool.Close()
	// 	panicIf(err)
	// }
	return sqlx.NewDb(db, "pgx")
}

const schema = `
	CREATE TABLE IF NOT EXISTS person (
		first_name text,
		last_name text,
		email text
	)`

func instrument() (flush func()) {
	e := new(oczeroexporter.Exporter)
	trace.RegisterExporter(e)
	je, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     "localhost:6831",
		CollectorEndpoint: "http://localhost:14268/api/traces",
		ServiceName:       "demo",
	})
	panicIf(err)
	trace.RegisterExporter(je)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	return func() {
		je.Flush()
	}
}

func main() {
	defer instrument()()
	db := initdb()
	db.MustExec(schema)
	tx, err := db.Begin()
	panicIf(err)
	tx.Exec("INSERT INTO person (first_name, last_name, email) VALUES ($1, $2, $3)", "Bob", "–", "bob@example.com")
	tx.Commit()
	bob := new(PersonRecord)
	for i := 0; i < 10; i++ {
		err = db.Get(bob, "SELECT * FROM person WHERE email=$1", "bob@example.com")
		panicIf(err)
	}
	log.Print(bob)
}
