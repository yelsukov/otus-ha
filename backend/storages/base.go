package storages

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
)

type QueryStmt struct {
	Conditions []string
	Params     []interface{}
}

type ExecStmt struct {
	Set []string
	QueryStmt
}

func closeRows(rows *sql.Rows) {
	if rows == nil {
		return
	}
	if err := rows.Close(); err != nil {
		log.WithError(err).Warn("failed to close rows")
	}
}

func closeStmt(stmt *sql.Stmt) {
	if stmt == nil {
		return
	}
	if err := stmt.Close(); err != nil {
		log.WithError(err).Warn("failed to close statement")
	}
}
