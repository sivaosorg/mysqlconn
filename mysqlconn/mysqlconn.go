package mysqlconn

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/sivaosorg/govm/common"
	"github.com/sivaosorg/govm/dbx"
	"github.com/sivaosorg/govm/logger"
	"github.com/sivaosorg/govm/mysql"

	_ "github.com/go-sql-driver/mysql"
)

var (
	instance *MySql
	_logger  = logger.NewLogger()
)

func NewMySql() *MySql {
	m := &MySql{}
	return m
}

func (m *MySql) SetConn(value *sql.DB) *MySql {
	m.conn = value
	return m
}

func NewClient(config mysql.MysqlConfig) (*MySql, dbx.Dbx) {
	s := dbx.NewDbx().SetDatabase(config.Database)
	if !config.IsEnabled {
		s.SetConnected(false).
			SetMessage("Mysql unavailable").
			SetError(fmt.Errorf(s.Message))
		return &MySql{}, *s
	}
	if instance != nil {
		s.SetConnected(true)
		return instance, *s
	}
	client, err := sql.Open(common.EntryKeyMysql, Dsn(config))
	if err != nil {
		s.SetConnected(false).SetError(err).SetMessage(err.Error())
		return &MySql{}, *s
	}
	if config.MaxOpenConn <= 0 {
		config.MaxOpenConn = 10
	}
	if config.MaxIdleConn <= 0 {
		config.MaxIdleConn = 5
	}
	if config.MaxLifeTimeMinutesConn <= 0 {
		config.MaxLifeTimeMinutesConn = 5
	}
	client.SetMaxOpenConns(config.MaxOpenConn)
	client.SetMaxIdleConns(config.MaxIdleConn)
	client.SetConnMaxLifetime(time.Duration(time.Duration(config.MaxLifeTimeMinutesConn).Minutes()))
	err = client.PingContext(context.Background())
	if err != nil {
		s.SetConnected(false).SetError(err).SetMessage(err.Error())
		return &MySql{}, *s
	}
	if config.DebugMode {
		_logger.Info(fmt.Sprintf("Mysql client connection:: %s", config.Json()))
		_logger.Info(fmt.Sprintf("Connected successfully to mysql:: %s (database: %s)", Dsn(config), config.Database))
	}
	instance = NewMySql().SetConn(client)
	pid := os.Getpid()
	s.SetConnected(true).SetMessage("Connection established").SetPid(pid).SetNewInstance(true)
	return instance, *s
}

func Dsn(config mysql.MysqlConfig) string {
	hostname := fmt.Sprintf("%s:%d", config.Host, config.Port)
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", config.Username, config.Password, hostname, config.Database)
}
