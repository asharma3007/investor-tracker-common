package common

import (
	"database/sql"
	"fmt"
	"os"
)

func ConnectDb() *sql.DB {
	Log("Connecting to db")

	dbUrl := os.Getenv(EnvDatabaseUrl)
	port := os.Getenv(EnvDatabasePort)
	dbname := os.Getenv(EnvDatabaseName)

	user := GetSecret("")
	password := GetSecret("")

	Log("Connection string")
	// connection string
	mysqlconn := user + ":" + password + "@tcp(" + dbUrl + ":" + port +")/" + dbname

	Log("Opening db")
	// open database
	db, err := sql.Open("mysql", mysqlconn)
	CheckError(err)

	Log("Doing ping")
	// check db
	err = db.Ping()
	CheckError(err)

	fmt.Println("Connected!")

	return db
}
