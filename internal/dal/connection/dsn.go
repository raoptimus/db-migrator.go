package connection

import (
	"fmt"
	"net/url"
	"path"
)

func NormalizeClickhouseDSN(dsn string) (string, error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return dsn, err
	}

	password, _ := dsnURL.User.Password()
	hostWithPort := dsnURL.Host
	if dsnURL.Port() == "" {
		hostWithPort += ":9000"
	}

	return fmt.Sprintf("tcp://%s?username=%s&password=%s&database=%s&%s",
		hostWithPort,
		dsnURL.User.Username(),
		password,
		path.Base(dsnURL.Path),
		dsnURL.RawQuery,
	), nil
}
