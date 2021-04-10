package clickhouseMigration

import (
	"fmt"
	"net/url"
	"path"
)

func NormalizeDSN(dsn string) (string, error) {
	dsnUrl, err := url.Parse(dsn)
	if err != nil {
		return dsn, err
	}

	password, _ := dsnUrl.User.Password()
	hostWithPort := dsnUrl.Host
	if dsnUrl.Port() == "" {
		hostWithPort += ":9000"
	}

	return fmt.Sprintf("tcp://%s?username=%s&password=%s&database=%s&%s",
		hostWithPort,
		dsnUrl.User.Username(),
		password,
		path.Base(dsnUrl.Path),
		dsnUrl.RawQuery,
	), nil
}
