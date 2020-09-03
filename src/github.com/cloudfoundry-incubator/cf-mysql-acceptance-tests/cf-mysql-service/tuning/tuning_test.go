package tuning_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MySQL Server Tuning Configuration", func() {
	var (
		db *sql.DB
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
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Correctly sets MySQL internal variables based on values in the manifest", func() {
		rows, err := db.Query("SHOW VARIABLES;")
		Expect(err).ToNot(HaveOccurred())

		mysqlVariables := map[string]string{}

		defer rows.Close()
		for rows.Next() {
			var name, value string
			err = rows.Scan(&name, &value)
			Expect(err).ToNot(HaveOccurred())
			mysqlVariables[name] = value
		}

		err = rows.Err()
		Expect(err).ToNot(HaveOccurred())

		buf, err := ioutil.ReadFile(helpers.TestConfig.Tuning.ExpectationFilePath)
		Expect(err).ToNot(HaveOccurred())

		compareConfig := map[string]string{}

		if err := json.Unmarshal(buf, &compareConfig); err != nil {
			Expect(err).ToNot(HaveOccurred())
		}

		var mismatchErrors []string

		for k := range compareConfig {
			if mysqlVariables[k] != compareConfig[k] {
				mismatchErrors = append(mismatchErrors, fmt.Sprintf("%s: \n\t(Expected):\t%s \n\t(Actual):\t%s",
					k, compareConfig[k], mysqlVariables[k]))

			}
		}

		Expect(len(mismatchErrors)).To(Equal(0), strings.Join(mismatchErrors, "\n"))
	})
})
