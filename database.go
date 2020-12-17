package common

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type dbOptions struct {
	Url            string
	User           string
	Password       string
	Port           string
	DbName         string
	PrivateIp      string
	ConnectionName string
}

func getDbOptions() dbOptions {
	return dbOptions{
		Url:            os.Getenv(EnvDatabaseUrl),
		User:           GetSecret(EnvSecretDbUser),
		DbName:         os.Getenv(EnvDatabaseName),
		Password:       GetSecret(EnvSecretDbPassword),
		Port:           os.Getenv(EnvDatabasePort),
		PrivateIp:      os.Getenv(EnvDatabasePrivateIp),
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
	return options.User + ":" + options.Password + "@tcp(" + options.Url + ":" + options.Port + ")/" + options.DbName
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

func DisconnectMongoDb(dbClient *mongo.Client) {
	dbClient.Disconnect(context.TODO())
}

func ConnectDbMongo() (*mongo.Client, *mongo.Database) {

	dbUrl := os.Getenv("MONGO_URL")
	dbPassword := os.Getenv("MONGO_PASSWORD")
	dbName := os.Getenv("MONGO_DBNAME")
	dbUser := os.Getenv("MONGO_USER")

	context, _ := context.WithTimeout(context.Background(), 10*time.Second)

	//dbClient, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	//"mongodb+srv://admin:<password>@tracker-mongo.3dzjg.mongodb.net/<dbname>?retryWrites=true&w=majority"
	uri := fmt.Sprintf("mongodb+srv://%v:%v@%v/%v?retryWrites=true&w=majority", dbUser, dbPassword, dbUrl, dbName)

	logUri := strings.Replace(uri, dbPassword, "password", -1)
	Log("URI " + logUri)

	dbClient, err := mongo.NewClient(options.Client().ApplyURI(uri))
	CheckError(err)

	err = dbClient.Connect(context)
	CheckError(err)

	err = dbClient.Ping(context, readpref.Primary())
	CheckError(err)

	databases, err := dbClient.ListDatabaseNames(context, bson.M{})
	CheckError(err)

	fmt.Println(databases)

	database := dbClient.Database(dbName)
	return dbClient, database
}
