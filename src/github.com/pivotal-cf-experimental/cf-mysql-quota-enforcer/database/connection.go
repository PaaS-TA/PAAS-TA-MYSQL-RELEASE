package database

import (
	"database/sql"
	"fmt"

	"github.com/pivotal-cf-experimental/cf-mysql-quota-enforcer/config"
)

func NewConnection(dbConfig config.Config) (*sql.DB, error) {

	var userPass string
	if dbConfig.Password != "" {
		userPass = fmt.Sprintf("%s:%s", dbConfig.User, dbConfig.Password)
	} else {
		userPass = dbConfig.User
	}

	return sql.Open("mysql", fmt.Sprintf(
		"%s@tcp(%s:%d)/%s",
		userPass,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	))
}
