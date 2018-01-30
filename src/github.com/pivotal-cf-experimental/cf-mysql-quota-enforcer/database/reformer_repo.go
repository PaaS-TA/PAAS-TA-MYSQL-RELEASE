package database

import (
	"database/sql"
	"fmt"

	"github.com/pivotal-golang/lager"
)

// LEFT JOIN is required so that dropping all tables will restore write access
const reformersQueryPattern = `
SELECT reformers.name AS db
FROM (
	SELECT violator_dbs.name AS name, tables.data_length, tables.index_length
	FROM   (
		SELECT DISTINCT Db AS name from mysql.db
		WHERE  (Insert_priv = 'N' OR Update_priv = 'N' OR Create_priv = 'N')
	) AS violator_dbs
	JOIN        %s.service_instances AS instances ON name = instances.db_name COLLATE utf8_general_ci
	LEFT JOIN   information_schema.tables AS tables ON tables.table_schema = name
	GROUP  BY   name
	HAVING ROUND(SUM(COALESCE(tables.data_length + tables.index_length,0) / 1024 / 1024), 1) < MAX(instances.max_storage_mb)
) AS reformers
`

func NewReformerRepo(brokerDBName string, db *sql.DB, logger lager.Logger) Repo {
	query := fmt.Sprintf(reformersQueryPattern, brokerDBName)
	return newRepo(query, db, logger, "quota reformer")
}
