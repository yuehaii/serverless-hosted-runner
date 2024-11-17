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
	conn                                  string
	db_handler                            *sql.DB
	sslmode                               string
	idx_id, msg_id, status_id, comment_id string
	status_fail                           string
	rows                                  *sql.Rows
}

func CreatePostgresDB() IPostgresDB {
	return &PostgresDB{DB{"postgres", os.Getenv("ALLEN_DB_HOST"), os.Getenv("ALLEN_DB_PORT"),
		os.Getenv("ALLEN_DB_USR"), os.Getenv("ALLEN_DB_PWD"), os.Getenv("ALLEN_DB_DBNAME"),
		os.Getenv("ALLEN_DB_TABLE")}, "", nil, "disable",
		"id", "request_message", "state", "status_message", "Failed", nil}
}

func CreateMySqlDB(host, port, usr, pwd, db, table string) IDB {
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
	pg.db_handler = h

	err = pg.db_handler.Ping()
	if err != nil {
		logrus.Errorf("Connect ping postgres failure: %s", err)
		pg.Disconnect()
		return err
	}
	fmt.Println("Connect connected.")
	return nil
}

func (pg PostgresDB) Disconnect() {
	pg.db_handler.Close()
}

func (pg *PostgresDB) ListRow(nin string) (*sql.Rows, error) {
	sub_where_clause := `" where "` + pg.status_id + `"!='` + pg.status_fail + `'`
	if nin != "" {
		sub_where_clause += ` and "` + pg.idx_id + `" not in (` + nin + ")"
	}
	list_query := `SELECT "` + pg.idx_id + `", "` + pg.msg_id + `", "` +
		pg.status_id + `", "` + pg.comment_id + `" FROM "` + pg.table + sub_where_clause
	items, err := pg.db_handler.Query(list_query)
	logrus.Infof("ListRow sql: %s", list_query)
	if err != nil {
		logrus.Errorf("ListRow failure: %s", err)
	}
	pg.rows = items
	return items, err
}

func (pg PostgresDB) UpdateStatus(idx string, stat string, comm string) error {
	err_stat := pg.UpdateField(pg.status_id, stat, idx)
	err_comm := pg.UpdateField(pg.comment_id, comm, idx)
	if err_stat != nil {
		return err_stat
	} else {
		return err_comm
	}
}

func (pg PostgresDB) UpdateField(k string, v string, id string) error {
	up_sql := `update "` + pg.table + `" set "` + k + `"='` + v + `' where "` + pg.idx_id + `"=` + id
	_, err := pg.db_handler.Exec(up_sql)
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
	pg.rows.Close()
	return false, nil
}
