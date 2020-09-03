package standalone_test

import (
	"database/sql"
	"fmt"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MySQL Standalone Deployment", func() {
	var (
		db     *sql.DB
		dbName string
	)

	BeforeEach(func() {
		standalone := helpers.TestConfig.Standalone
		connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
			standalone.MySQLUsername,
			standalone.MySQLPassword,
			standalone.Host,
			standalone.Port)

		var err error
		db, err = sql.Open("mysql", connectionString)
		Expect(err).ToNot(HaveOccurred())

		err = db.Ping()
		Expect(err).ToNot(HaveOccurred())

		dbName = uuidWithUnderscores("testDb")
		_, err = db.Query(fmt.Sprintf("CREATE DATABASE %s", dbName))
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		_, err := db.Query(fmt.Sprintf("DROP DATABASE %s", dbName))
		Expect(err).ToNot(HaveOccurred())
		db.Close()
	})

	It("writes data to test DB and reads it back", func() {

		_, err := db.Query(fmt.Sprintf(
			"CREATE TABLE %s.testTable (id int,name varchar(250), PRIMARY KEY (id))", dbName))
		Expect(err).ToNot(HaveOccurred())

		expectedId := 42
		expectedName := "test"
		_, err = db.Query(fmt.Sprintf("INSERT INTO %s.testTable VALUES (%d, '%s')", dbName, expectedId, expectedName))
		Expect(err).ToNot(HaveOccurred())

		row := db.QueryRow(fmt.Sprintf("SELECT * FROM %s.testTable", dbName))

		var actualId int
		var actualName string
		err = row.Scan(&actualId, &actualName)
		Expect(err).ToNot(HaveOccurred())

		Expect(actualId).To(Equal(expectedId), "Actual ID did not match expected")
		Expect(actualName).To(Equal(expectedName), "Actual name did not match expected")
	})
})
