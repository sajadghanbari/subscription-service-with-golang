package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"subscription-service/data"
	"sync"
	"syscall"
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
		Session:  session,
		DB:       db,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
		Wait:     &wg,
		Models:   data.New(db),
		ErrorChan: make(chan error),
		ErrorChanDone: make(chan bool),
	}
	// set up email
	app.Mailer = app.createMail()
	go app.listenForMail()
	//listen for signals
	go app.listenFotShutdown()
	//listen for errors
	go app.listenForErrors()
	//listen for web connection
	app.serve()
}

func (app *Config) listenForErrors(){
	for {
		select{
		case err := <-app.ErrorChan:
			app.ErrorLog.Panicln(err)
		case <-app.ErrorChanDone:
			return	
		}
	}
}

func (app *Config) serve() {
	stv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
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
	if conn == nil {
		log.Panic("can't connect to database")
	}

	return conn
}

func connectToDB() *sql.DB {
	counts := 0

	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("postgres not yet ready...")
		} else {
			log.Println("connected to database....")
			return connection
		}

		if counts > 10 {
			return nil
		}

		log.Print("Backing off fot 1 second")
		time.Sleep(1 * time.Second)
		counts++
		continue
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func initSession() *scs.SessionManager {
	gob.Register(data.User{})
	

	session := scs.New()
	// Initialize Redis for session storage
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode

	return session
}

func initRedis() *redis.Pool {
	redisPool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}
	return redisPool
}

func (app *Config) listenFotShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.shutdown()
	os.Exit(0)
}

func (app *Config) shutdown() {
	app.InfoLog.Println("would run cleanup tasks...")
	// block until wait group
	app.Wait.Wait()
	app.Mailer.DoneChan <- true
	app.ErrorChanDone <- true
	
	app.InfoLog.Println("closing channels and shuting down.....")
	close(app.Mailer.MailerChan)
	close(app.Mailer.DoneChan)
	close(app.Mailer.ErrorChan)
	close(app.ErrorChan)
	close(app.ErrorChanDone)
}

func (app *Config) createMail() Mail{
	errorChan := make(chan error)
	mailerChan := make(chan Message,100)
	mailerDonChan := make(chan bool)

	m :=Mail{
		Domain: "localhost",
		Host: "localhost",
		Port: 1025,
		Encryption: "none",
		FromName: "info",
		FromAddress: "info@company.com",
		Wait: app.Wait,
		ErrorChan: errorChan,
		MailerChan: mailerChan,
		DoneChan: mailerDonChan,
	}
	 return m
}