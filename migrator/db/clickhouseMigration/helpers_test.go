package clickhouseMigration

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNormalizeDSNSuccessfully(t *testing.T) {
	var err error
	var actualDSN, originDSN, expectedDSN string

	for _, data := range getDataProvider() {
		originDSN, expectedDSN = data[0], data[1]
		actualDSN, err = NormalizeDSN(originDSN)
		assert.NoError(t, err)
		assert.Equal(t, expectedDSN, actualDSN)
	}
}

func getDataProvider() [][]string {
	return [][]string{
		{
			"clickhouse://default:@clickhouse:9000/docker?sslmode=disable&compress=true&debug=false&cluster=test_cluster",
			"tcp://clickhouse:9000?username=default&password=&database=docker&sslmode=disable&compress=true&debug=false&cluster=test_cluster",
		},
		{
			"clickhouse://user:pass@clickhouse/docker?sslmode=disable&compress=true&debug=false&cluster=test_cluster",
			"tcp://clickhouse:9000?username=user&password=pass&database=docker&sslmode=disable&compress=true&debug=false&cluster=test_cluster",
		},
	}
}
