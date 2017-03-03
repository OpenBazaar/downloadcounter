package connect

import "os"

// Config contains the information needed manipulate runtime behavior
type Config struct {
	MySQLDSN string
	Table    string
	Repo     string
}

// NewConfigFromEnv returns a new `*Config` using the environment as a data source
func NewConfigFromEnv() Config {
	return Config{
		MySQLDSN: envString("DOWNLOAD_COUNTER_MYSQL_DSN", "root@unix(/tmp/mysql.sock)/network"),
		Table:    envString("DOWNLOAD_COUNTER_TABLE", "download_counts"),
		Repo:     envString("DOWNLOAD_COUNTER_REPO", "OpenBazaar/OpenBazaar-Installer"),
	}
}

func envString(name string, defaultVal string) string {
	val := os.Getenv(name)
	if val == "" {
		return defaultVal
	}
	return val
}
