package mysql

import (
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Secret struct {
	Username             string `json:"username"`
	Password             string `json:"password"`
	Engine               string `json:"engine"`
	Host                 string `json:"host"`
	Port                 int    `json:"port"`
	DBName               string `json:"dbname"`
	DBInstanceIdentifier string `json:"dbInstanceIdentifier"`
	MasterARN            string `json:"masterarn"`
}

func (src Secret) Unmarshal(dest *mysql.Config) {
	dest.User = src.Username
	dest.Passwd = src.Password
	dest.Addr = fmt.Sprintf("%s:%d", src.Host, src.Port)
	dest.DBName = src.DBName
}

func NewConfig() *mysql.Config {
	c := mysql.NewConfig()
	c.Net = "tcp"
	c.ParseTime = true
	c.Loc = time.Local
	c.Collation = "utf8mb4_general_ci"

	return c
}
