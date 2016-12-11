package main

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/igormartire/gorfiv/models"
	"github.com/igormartire/gorfiv/server"
	"github.com/spf13/viper"
)

type Config struct {
	database map[string]string
	api      map[string]string
}

func main() {
	var config Config
	if err := config.load(); err != nil {
		panic(err)
	}

	db, err := connectDb(config.database)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	mysqlRepo := models.NewSQLRepo(db)

	server.Run(server.NewEnv(mysqlRepo), config.api["token"])
}

func (c *Config) load() (err error) {
	viper.AddConfigPath("config")
	viper.SetConfigName("app")

	err = viper.ReadInConfig()
	if err != nil {
		return err
	} else {
		c.database = viper.GetStringMapString("database")
		c.api = viper.GetStringMapString("api")
	}

	return nil
}

func connectDb(params map[string]string) (db *sql.DB, err error) {
	db, err = sql.Open("mysql", (&mysql.Config{
		User:      params["user"],
		Passwd:    params["password"],
		DBName:    params["name"],
		Collation: "utf8_general_ci",
		ParseTime: true,
	}).FormatDSN())

	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
