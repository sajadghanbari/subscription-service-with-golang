package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "8080"

func main() {
	//connect db
	db := initDB()

	// create sessions
	session := initSession()

	//create loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	//create channels

	//create wait groups
	wg := sync.WaitGroup{}
	//set up application config
	app := Config{
		Session: session,
		DB:      db,
		InfoLog: infoLog,
		ErrorLog: errorLog,
		Wait:    &wg,
	}
	// set up email

	//listen for web connection
	app.serve()
}

func (app *Config) serve(){
	stv := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	app.InfoLog.Println("Starting server on port", webPort)
	err := stv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}


func initDB() *sql.DB {
	conn := connectToDB()
	if conn == nil{
		log.Panic("can't connect to database")
	}

	return conn
}

func connectToDB() *sql.DB{
	counts := 0

	dsn := os.Getenv("DSN")

	for {
		connection , err := openDB(dsn)
		if err != nil{
			log.Println("postgres not yet ready...")
		}else{
			log.Println("connected to database....")
			return connection
		}

		if counts > 10 {
			return nil
		}

		log.Print("Backing off fot 1 second")
		time.Sleep(1*time.Second)
		counts++
		continue
	}
}

func openDB(dsn string) (*sql.DB,error){
	db , err := sql.Open("pgx" , dsn)
	if err  != nil {
		return nil , err
	}

	err = db.Ping()
		if err  != nil {
		return nil , err
	}
	return db, nil
}

func initSession() *scs.SessionManager{
	session := scs.New()
	// Initialize Redis for session storage
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode

	return session
}

func initRedis() *redis.Pool{
	redisPool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}
	return redisPool
}