package enforcer_test

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	_ "github.com/go-sql-driver/mysql"

	"github.com/pivotal-cf-experimental/cf-mysql-quota-enforcer/config"
	"github.com/pivotal-cf-experimental/cf-mysql-quota-enforcer/database"
)

var _ = Describe("Enforcer Integration", func() {

	var exec = func(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
		GinkgoWriter.Write([]byte(fmt.Sprintf("EXEC SQL: %s\n", query)))
		return db.Exec(query, args...)
	}

	var tableSizeMB = func(dbName, tableName string, db *sql.DB) float64 {
		var sizeMB float64
		row := db.QueryRow(`SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 1)
            FROM information_schema.TABLES
            WHERE table_schema = ? AND table_name = ?`, dbName, tableName)
		err := row.Scan(&sizeMB)
		Expect(err).ToNot(HaveOccurred())
		return sizeMB
	}

	var createSizedTable = func(numRows int, dbName, tableName string, db *sql.DB) {
		_, err := exec(db, fmt.Sprintf(
			`CREATE TABLE %s 
			(id MEDIUMINT AUTO_INCREMENT, data LONGBLOB, PRIMARY KEY (id))
			ENGINE = INNODB`,
			tableName,
		))
		Expect(err).NotTo(HaveOccurred())

		data := make([]byte, 1024*1024)
		for row := 0; row < numRows; row++ {
			_, err = exec(db, fmt.Sprintf("INSERT INTO %s (data) VALUES (?)", tableName), data)
			Expect(err).NotTo(HaveOccurred())
		}

		// Optimizing forces the size metadata to update (normally happens every few seconds)
		_, err = exec(db, fmt.Sprintf("OPTIMIZE TABLE %s", tableName))
		Expect(err).ToNot(HaveOccurred())

		// Check that the table size matches our expectations
		// If this doesn't work, then the quota enforcer isn't going to work either...
		Expect(tableSizeMB(dbName, tableName, db)).To(BeNumerically(">=", numRows))
	}

	var userConfigs []config.Config
	var dbNames []string

	BeforeEach(func() {
		// MySQL mandates usernames are <= 16 chars
		user0 := uuidWithUnderscores("user")[:16]
		user1 := uuidWithUnderscores("user")[:16]
		user2 := uuidWithUnderscores("user")[:16]

		dbNames = []string{
			uuidWithUnderscores("cf"),
			uuidWithUnderscores("cf"),
		}

		user0Config := config.Config{
			Host:     rootConfig.Host,
			Port:     rootConfig.Port,
			User:     user0,
			Password: uuidWithUnderscores("password"),
			DBName:   dbNames[0],
		}
		err := user0Config.Validate()
		Expect(err).ToNot(HaveOccurred())

		user1Config := config.Config{
			Host:     rootConfig.Host,
			Port:     rootConfig.Port,
			User:     user1,
			Password: uuidWithUnderscores("password"),
			DBName:   dbNames[0],
		}
		err = user1Config.Validate()
		Expect(err).ToNot(HaveOccurred())

		user2Config := config.Config{
			Host:     rootConfig.Host,
			Port:     rootConfig.Port,
			User:     user2,
			Password: uuidWithUnderscores("password"),
			DBName:   dbNames[1],
		}
		err = user2Config.Validate()
		Expect(err).ToNot(HaveOccurred())

		userConfigs = []config.Config{
			user0Config,
			user1Config,
			user2Config,
		}
	})

	Describe("Writing pid file", func() {
		Context("when the quota enforcer is running continuously", func() {
			var (
				session     *gexec.Session
				pidFile     string
				pidFileFlag string
			)

			Context("when the pid file location is valid", func() {
				BeforeEach(func() {
					pidFile = fmt.Sprintf("%s/enforcer.pid", tempDir)
					pidFileFlag = fmt.Sprintf("-pidFile=%s", pidFile)
				})

				It("writes its pid to the provided file", func() {
					Expect(fileExists(pidFile)).To(BeFalse())
					session = runEnforcerContinuously(pidFileFlag)
					Expect(fileExists(pidFile)).To(BeTrue())
				})

				AfterEach(func() {
					session.Kill()

					// Once signalled, the session should shut down relatively quickly
					session.Wait(5 * time.Second)

					// We don't care what the exit code is
					Eventually(session).Should(gexec.Exit())
				})
			})

			Context("when the pid file location is invalid", func() {
				BeforeEach(func() {
					pidFile = "/invalid_path/enforcer.pid"
					pidFileFlag = fmt.Sprintf("-pidFile=%s", pidFile)
				})

				It("exits with error", func() {
					session = runEnforcerContinuously(pidFileFlag)

					Eventually(session.Err).Should(gbytes.Say(pidFile))
					Eventually(session).Should(gexec.Exit())
					Expect(session.ExitCode()).ToNot(Equal(0))
				})
			})
		})
	})

	Describe("Signal handling", func() {
		Context("when the quota enforcer is running continuously", func() {
			var session *gexec.Session

			BeforeEach(func() {
				session = runEnforcerContinuously()
			})

			It("shuts down on any signal", func() {
				session.Kill()

				// Once signalled, the session should shut down relatively quickly
				session.Wait(5 * time.Second)

				// We don't care what the exit code is
				Eventually(session).Should(gexec.Exit())
			})
		})
	})

	Describe("Quota enforcement", func() {
		var (
			plan                string
			maxStorageMB        int
			dataTableName       string
			tempTableName       string
			unimpactedTableName string
		)

		BeforeEach(func() {
			plan = uuidWithUnderscores("plan")
			maxStorageMB = 10
			dataTableName = uuidWithUnderscores("data")
			tempTableName = uuidWithUnderscores("temp")
			unimpactedTableName = uuidWithUnderscores("unimpacted")
		})

		Context("when multiple databases exist with multiple users", func() {
			var (
				user0Connection, user1Connection, user2Connection *sql.DB
			)

			BeforeEach(func() {
				db, err := database.NewConnection(rootConfig)
				Expect(err).NotTo(HaveOccurred())
				defer db.Close()

				for _, dbName := range dbNames {
					_, err = exec(db, fmt.Sprintf(
						"CREATE DATABASE IF NOT EXISTS %s", dbName))
					Expect(err).NotTo(HaveOccurred())

					_, err = exec(db,
						"INSERT INTO service_instances (guid,plan_guid,max_storage_mb,db_name) VALUES(?,?,?,?)", dbName, plan, maxStorageMB, dbName)
					Expect(err).NotTo(HaveOccurred())
				}

				for _, userConfig := range userConfigs {
					_, err = exec(db, fmt.Sprintf(
						"CREATE USER %s IDENTIFIED BY '%s'", userConfig.User, userConfig.Password))
					Expect(err).NotTo(HaveOccurred())

					_, err = exec(db, fmt.Sprintf(
						"GRANT ALL PRIVILEGES ON %s.* TO %s", userConfig.DBName, userConfig.User))
					Expect(err).NotTo(HaveOccurred())

					_, err = exec(db, "FLUSH PRIVILEGES")
					Expect(err).NotTo(HaveOccurred())
				}

				user0Connection, err = database.NewConnection(userConfigs[0])
				Expect(err).NotTo(HaveOccurred())

				user1Connection, err = database.NewConnection(userConfigs[1])
				Expect(err).NotTo(HaveOccurred())

				user2Connection, err = database.NewConnection(userConfigs[2])
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				db, err := database.NewConnection(rootConfig)
				Expect(err).NotTo(HaveOccurred())
				defer db.Close()

				for _, dbName := range dbNames {
					_, err = exec(db, fmt.Sprintf(
						"DROP DATABASE IF EXISTS %s", dbName))
					Expect(err).NotTo(HaveOccurred())
				}

				for _, userConfig := range userConfigs {
					_, err = exec(db, fmt.Sprintf(
						"DROP TABLE IF EXISTS %s.%s", userConfig.DBName, dataTableName))
					Expect(err).NotTo(HaveOccurred())

					_, err = exec(db, fmt.Sprintf(
						"DROP TABLE IF EXISTS %s.%s", userConfig.DBName, tempTableName))
					Expect(err).NotTo(HaveOccurred())

					_, err = exec(db, fmt.Sprintf(
						"REVOKE ALL PRIVILEGES, GRANT OPTION FROM %s", userConfig.User))
					Expect(err).NotTo(HaveOccurred())

					_, err = exec(db, fmt.Sprintf(
						"DROP USER %s", userConfig.User))
					Expect(err).NotTo(HaveOccurred())

					_, err = exec(db,
						"FLUSH PRIVILEGES")
					Expect(err).NotTo(HaveOccurred())
				}

				defer user0Connection.Close()
				defer user1Connection.Close()
				defer user2Connection.Close()
			})

			It("Enforces the quota for all users on the same database and does not impact other users on other databases", func() {
				By("Revoking write access when over the quota", func() {
					createSizedTable(maxStorageMB/2, userConfigs[0].DBName, dataTableName, user0Connection)
					createSizedTable(maxStorageMB/2, userConfigs[0].DBName, tempTableName, user0Connection)

					createSizedTable(maxStorageMB/2, userConfigs[2].DBName, unimpactedTableName, user2Connection)

					runEnforcerOnce()

					// Users 0 and 1 cannot write to db 0
					_, err := user0Connection.Exec(fmt.Sprintf(
						"INSERT INTO %s (data) VALUES (?)", dataTableName), []byte{'1'})
					Expect(err).To(HaveOccurred())

					_, err = user1Connection.Exec(fmt.Sprintf(
						"INSERT INTO %s (data) VALUES (?)", dataTableName), []byte{'1'})
					Expect(err).To(HaveOccurred())

					// User 2 can still write to db 1
					_, err = user2Connection.Exec(fmt.Sprintf(
						"INSERT INTO %s (data) VALUES (?)", unimpactedTableName), []byte{'1'})
					Expect(err).NotTo(HaveOccurred())
				})

				By("Re-enabling write access when back under the quota", func() {
					_, err := user0Connection.Exec(fmt.Sprintf("DROP TABLE %s", tempTableName))
					Expect(err).NotTo(HaveOccurred())

					runEnforcerOnce()

					// Users 0 and 1 can now write to db 0
					_, err = user0Connection.Exec(fmt.Sprintf(
						"INSERT INTO %s (data) VALUES (?)", dataTableName), []byte{'1'})
					Expect(err).NotTo(HaveOccurred())

					_, err = user1Connection.Exec(fmt.Sprintf(
						"INSERT INTO %s (data) VALUES (?)", dataTableName), []byte{'1'})
					Expect(err).NotTo(HaveOccurred())

					// User 2 can still write to db 1
					_, err = user2Connection.Exec(fmt.Sprintf(
						"INSERT INTO %s (data) VALUES (?)", unimpactedTableName), []byte{'1'})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			It("restores write access after dropping all tables", func() {
				db, err := database.NewConnection(userConfigs[0])
				Expect(err).NotTo(HaveOccurred())
				defer db.Close()

				By("Revoking write access when over quota", func() {
					createSizedTable(maxStorageMB, userConfigs[0].DBName, dataTableName, db)

					runEnforcerOnce()

					_, err = exec(db, fmt.Sprintf(
						"INSERT INTO %s (data) VALUES (?)", dataTableName), []byte{'1'})
					Expect(err).To(HaveOccurred())
				})

				By("Re-enabling write access when back under the quota", func() {
					_, err := exec(db, fmt.Sprintf(
						"DROP TABLE %s", dataTableName))
					Expect(err).NotTo(HaveOccurred())

					runEnforcerOnce()

					createSizedTable(maxStorageMB/2, userConfigs[0].DBName, dataTableName, db)
					_, err = exec(db, fmt.Sprintf(
						"INSERT INTO %s (data) VALUES (?)", dataTableName), []byte{'1'})
					Expect(err).NotTo(HaveOccurred())
				})

			})
		})
	})
})

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
