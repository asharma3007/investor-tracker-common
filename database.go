package common

import (
	"database/sql"
	"fmt"
	"os"
)

type dbOptions struct {
	Url       string
	User      string
	Password  string
	Port      string
	DbName    string
	PrivateIp string
	ConnectionName string
}

func getDbOptions() dbOptions {
	return dbOptions{
		Url:      os.Getenv(EnvDatabaseUrl),
		User:     GetSecret(EnvSecretDbUser),
		DbName:   os.Getenv(EnvDatabaseName),
		Password: GetSecret(EnvSecretDbPassword),
		Port:     os.Getenv(EnvDatabasePort),
		PrivateIp:	os.Getenv(EnvDatabasePrivateIp),
		ConnectionName: os.Getenv(EnvSecretDatabaseConnectionName),
	}
}

func ConnectDb() *sql.DB {
	Log("Connecting to db")

	options := getDbOptions()

	Log("Connection string")
	// connection string
	connectString := getConnectionString(options)

	Log("Opening db")
	// open database
	db, err := sql.Open("mysql", connectString)
	CheckError(err)

	Log("Doing ping")
	// check db
	err = db.Ping()
	CheckError(err)

	fmt.Println("Connected!")

	return db
}

func getConnectionString(options dbOptions) string {
	if options.PrivateIp == "1" {
		return getConnectionStringDirect(options)
	} else {
		return getConnectionStringSocket(options)
	}
}

func getConnectionStringDirect(options dbOptions) string {
	return  options.User + ":" + options.Password + "@tcp(" + options.Url + ":" + options.Port +")/" + options.DbName
}

func getConnectionStringSocket(options dbOptions) string {
	socketDir, isSet := os.LookupEnv("DB_SOCKET_DIR")
	if !isSet {
		socketDir = "/cloudsql"
	}

	var dbURI string
	dbURI = fmt.Sprintf("%s:%s@unix(/%s/%s)/%s?parseTime=true", options.User, options.Password, socketDir, options.ConnectionName, options.DbName)

	// dbPool is the pool of database connections.
	_, err := sql.Open("mysql", dbURI)
	CheckError(fmt.Errorf("sql.Open: %v", err))

	return dbURI
}
