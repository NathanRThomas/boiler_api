/*! \file main.go
	\brief Shared defines required across multiple of our cmd apps

	Created 2020-01-29 by NateDogg
*/

package cmd 

 import (
	"github.com/NathanRThomas/boiler_api/pkg/models"
	"github.com/NathanRThomas/boiler_api/pkg/models/redis"
	"github.com/NathanRThomas/boiler_api/pkg/models/cockroach"
	
	"github.com/NathanRThomas/boiler_api/pkg/toolz"
	
	"github.com/pkg/errors"
	"github.com/mediocregopher/radix/v3"
	"github.com/patrickmn/go-cache"
	
	"fmt"
	"os"
	"encoding/json"
	"database/sql"
	_ "github.com/lib/pq"
	"time"
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"log"
	"sync"
 )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- DEFINES -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

const API_ver = "0.1.0"

const ContextTimeout = 55 // seconds to timeout any of our default context's
const JWT_secret_key	= "doesn't have to be that secret"

// global config object
type CFG struct {
	Port, UrlAPI string
	ProductionLevel, LoggingLevel int
	Version bool
	Cockroach struct {
		IP, Database, User string
		Port int
	}
	Redis struct {
		Cache struct {
			IP string
			Port int
		}
	}
	Slack toolz.SlackConfig_t
	Mailgun toolz.MailgunConfig_t
	DigitalOcean struct {
		Key string
	}
}

type LogLevel int
const (
	LogLevel_none 		LogLevel = iota
	LogLevel_basic
	LogLevel_full
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- TYPES -------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type stackTracer interface {
	StackTrace() errors.StackTrace
}

//----- HANDLER
type App_c struct {
	InfoLog, ErrorLog  *log.Logger
	Running bool
	PRunning *bool	// sorry, this is confusing, this is not a pointer to the Running bool within this struct.  It's used to "pass" informatino to another instance of this class that the service is "stopped" or shutting down
	ProductionLevel int
	LoggingLevel	LogLevel
	WG *sync.WaitGroup
	Redis 		*redis.DB_c
	DB			*cockroach.DB_c
	Cache 		*cache.Cache
	TaskQue chan<- *models.Que_t
}

//----- Public Handler
func (this *App_c) GetLogs () (*log.Logger, *log.Logger) {
	return this.InfoLog, this.ErrorLog
}

func (this *App_c) GetDatabases () (*redis.DB_c, *cockroach.DB_c, *cache.Cache) {
	return this.Redis, this.DB, this.Cache
}

func (this *App_c) GetWaitGroup () (*sync.WaitGroup) {
	return this.WG
}

func (this *App_c) GetProductionLevel () int {
	return this.ProductionLevel
}

func (this *App_c) NewHandler () Handler_c {
	return Handler_c { 
		App_c: App_c { PRunning: &this.Running,
		LoggingLevel: this.LoggingLevel,
		Redis: this.Redis, 
		DB: this.DB,
		},
	}
}

/*! \brief Pulls out the stack trace error info
*/
func (this *App_c) StackTrace (err error) {
	if err == nil { return }
	this.ErrorLog.Println(err) 

	if err, ok := err.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			fmt.Printf("%+s:%d\n", f, f)
		}
	}
}

type Handler_i interface {
	GetLogs () (*log.Logger, *log.Logger)
	GetDatabases () (*redis.DB_c, *cockroach.DB_c, *cache.Cache)
	GetWaitGroup () (*sync.WaitGroup)
	GetProductionLevel () int
}

//----- SHARED
type shared_c struct {
	App_c

	slack 		toolz.Slack_c
	self 		chan<- *models.Que_t
	
	users 		cockroach.User_c
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func ParseConfig (data interface{}) error {
	configFile, err := os.Open(os.Getenv("API_CONFIG")) //try the file
	
	if err == nil {
        jsonParser := json.NewDecoder(configFile)
        err = jsonParser.Decode(data)
	}
	return err
}

func CreateLoggers () (errorLog, infoLog *log.Logger) {
	errorLog = log.New(os.Stderr, "ERROR\t", log.LstdFlags | log.Lmicroseconds | log.Llongfile | log.LUTC)	// error log handler
	infoLog = log.New(os.Stdout, "INFO\t", log.LstdFlags | log.Lmicroseconds)	// for info level messages
	if errorLog == nil || infoLog == nil { log.Panic ("unable to create loggers") }
	return
}

func ConnectCockroach (ip string, port int, database, user string, productionLevel int) (*sql.DB, error) {
    sslmode := "sslmode=disable"
    if productionLevel == models.ProductionType_Production {
		sslmode = fmt.Sprintf("sslmode=verify-full&sslcert=/cockroach-certs/client.%s.crt&sslkey=/cockroach-certs/client.%s.key&sslrootcert=/cockroach-certs/ca.crt",
			user, user)
	}
	var err error
    sqlDB, err := sql.Open("postgres", fmt.Sprintf("postgres://%s@%s:%d/%s?%s", user, ip, port, database, sslmode))

    if err == nil {
		_, err = sqlDB.Exec(`SET TIME ZONE 'UTC'`)   //set our default timezone
	}
	return sqlDB, err
}

func ConnectRedis (ip string, port int) (*radix.Pool, error) {
	return radix.NewPool("tcp", fmt.Sprintf("%s:%d", ip, port), redis.MaxPoolSize)
}

/*! \brief Handles marking the server as shutting down based on a cancel call
*/
func MonitorSignals (running *bool, srv *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {	//listen for kill messages
		<-c
		*running = false 	//stops the queen/workers background tasks.  This also stops health checks so this looks unhealthy now
		time.Sleep(time.Second * 5)	//sleep here, this is needed for this to be taken out of the load balancer before stopping to handle requests
		srv.Shutdown (context.Background())  //shutsdown the http server
	}()
}

