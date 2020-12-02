package common

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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
		ConnectionName: GetSecret(EnvSecretDatabaseConnectionName),
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
	//Access denied for user 'trackerapp'@'cloudsqlproxy~107.178.231.18' (using password: YES)
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
		socketDir = "cloudsql"
	}

	Log(fmt.Sprintf("user %v socket %v conx %v db %v", options.User, socketDir, options.ConnectionName, options.DbName))

	var dbURI string
	dbURI = fmt.Sprintf("%s:%s@unix(/%s/%s)/%s?parseTime=true", options.User, options.Password, socketDir, options.ConnectionName, options.DbName)

	// dbPool is the pool of database connections.
	//_, err := sql.Open("mysql", dbURI)
	//CheckError(fmt.Errorf("sql.Open: %v", err))

	return dbURI
}


func ConnectDbMongo() (*mongo.Client, *mongo.Database) {
	dbClient, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	CheckError(err)

	context, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = dbClient.Connect(context)
	CheckError(err)

	err = dbClient.Ping(context, readpref.Primary())
	CheckError(err)

	databases, err := dbClient.ListDatabaseNames(context, bson.M{})
	CheckError(err)

	fmt.Println(databases)

	database := dbClient.Database("globetrotter")
	//collectionUserData = database.Collection("UserData")

	return dbClient, database
}
