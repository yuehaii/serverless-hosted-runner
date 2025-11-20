package common

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type IDB interface {
	InitConnection()
	Connect() error
	Disconnect()
	ListRow(nin string) (*sql.Rows, error)
	UpdateField(k string, v string, w string) error
}
type IPostgresDB interface {
	IDB
	UpdateStatus(idx string, stat string, comm string) error
	IterateRows(idx *string, msg *string, status *string, comment *string) (bool, error)
}

type DB struct {
	dbtype, host, port, usr, pwd, db, table string
}

type PostgresDB struct {
	DB
	conn                              string
	dbHandler                         *sql.DB
	sslmode                           string
	idxID, msgID, statusID, commentID string
	statusFail                        string
	rows                              *sql.Rows
}

func CreatePostgresDB() IPostgresDB {
	return &PostgresDB{DB{"postgres", os.Getenv("ALLEN_DB_HOST"), os.Getenv("ALLEN_DB_PORT"),
		os.Getenv("ALLEN_DB_USR"), os.Getenv("ALLEN_DB_PWD"), os.Getenv("ALLEN_DB_DBNAME"),
		os.Getenv("ALLEN_DB_TABLE")}, "", nil, "disable",
		"id", "request_message", "state", "status_message", "Failed", nil}
}

func CreateMySQLDB(host, port, usr, pwd, db, table string) IDB {
	return nil
}

func (pg *PostgresDB) InitConnection() {
	pg.conn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s ",
		pg.host, pg.port, pg.usr, pg.pwd, pg.db, pg.sslmode)
}

func (pg *PostgresDB) Connect() error {
	h, err := sql.Open(pg.dbtype, pg.conn)
	if err != nil {
		logrus.Errorf("Connect open postgres failure: %s", err)
		return err
	}
	pg.dbHandler = h

	err = pg.dbHandler.Ping()
	if err != nil {
		logrus.Errorf("Connect ping postgres failure: %s", err)
		pg.Disconnect()
		return err
	}
	fmt.Println("Connect connected.")
	return nil
}

func (pg PostgresDB) Disconnect() {
	if err := pg.dbHandler.Close(); err != nil {
		logrus.Errorf("Disconnect db handler Close failure: %v", err)
	}
}

func (pg *PostgresDB) ListRow(nin string) (*sql.Rows, error) {
	subWhereClause := `" where "` + pg.statusID + `"!='` + pg.statusFail + `'`
	if nin != "" {
		subWhereClause += ` and "` + pg.idxID + `" not in (` + nin + ")"
	}
	listQuery := `SELECT "` + pg.idxID + `", "` + pg.msgID + `", "` +
		pg.statusID + `", "` + pg.commentID + `" FROM "` + pg.table + subWhereClause
	items, err := pg.dbHandler.Query(listQuery)
	logrus.Infof("ListRow sql: %s", listQuery)
	if err != nil {
		logrus.Errorf("ListRow failure: %s", err)
	}
	pg.rows = items
	return items, err
}

func (pg PostgresDB) UpdateStatus(idx string, stat string, comm string) error {
	errStat := pg.UpdateField(pg.statusID, stat, idx)
	errComm := pg.UpdateField(pg.commentID, comm, idx)
	if errStat != nil {
		return errStat
	} else {
		return errComm
	}
}

func (pg PostgresDB) UpdateField(k string, v string, id string) error {
	upSQL := `update "` + pg.table + `" set "` + k + `"='` + v + `' where "` + pg.idxID + `"=` + id
	_, err := pg.dbHandler.Exec(upSQL)
	if err != nil {
		logrus.Errorf("UpdateField Update failure: %s", err)
	}
	return err
}

func (pg *PostgresDB) IterateRows(idx *string, msg *string, status *string, comment *string) (bool, error) {
	if pg.rows.Next() {
		err := pg.rows.Scan(idx, msg, status, comment)
		if err != nil {
			logrus.Errorf("IterateRows Update failure: %s", err)
			return true, err
		}
		fmt.Println("IterateRows parse one row: ", *idx, *msg, *status, *comment)
		return true, nil
	}
	fmt.Println("IterateRows finish parsing rows")
	if err := pg.rows.Close(); err != nil {
		logrus.Errorf("IterateRows Close failure: %v", err)
	}
	return false, nil
}
