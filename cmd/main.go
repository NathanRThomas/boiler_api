/*! \file main.go
	\brief Shared defines required across multiple of our cmd apps
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
	"github.com/prometheus/client_golang/prometheus"
	
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

const API_ver = "0.1.0" // major version for all related builds

// global config object
var CFG struct {
	Port string
	ApiUrl, WebsiteUrl models.ApiString
	ProductionLevel models.ProductionLevel
	Version, LocalRun bool
	Cockroach struct {
		IP, Database, User string
		Port int
	}
	Redis struct {
		IPs []string 
		Port int
	}
	Slack toolz.SlackConfig_t
	Mailgun toolz.MailgunConfig_t
}

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
	WG *sync.WaitGroup

	ApiRequests *prometheus.CounterVec 
	Redis 		*redis.DB_c
	Cache 		*cache.Cache
	TaskQue chan *models.Que_t

	Users		cockroach.User_c
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

/*! \brief Wrapper around stacktrace so we don't have to create the error each time
*/
func (this *App_c) StackRecord (msg string, params ...interface{}) {
	this.StackTrace (errors.Errorf (msg, params...))
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func ParseConfig () error {
	configFile, err := os.Open(os.Getenv("API_CONFIG")) //try the file
	if err != nil { return errors.WithStack (err) }
	
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&CFG)
	if err != nil { return errors.WithStack (err) }

	// validate some expected values
	if !CFG.ApiUrl.Url() {
		return errors.Errorf ("ApiUrl from config file appears invalid, this shoudl be a url that this service is listening on")
	}

	// validate anything else
	
	return nil
}

func CreateLoggers () (errorLog, infoLog *log.Logger) {
	errorLog = log.New(os.Stderr, "ERROR\t", log.LstdFlags | log.Lmicroseconds | log.Llongfile | log.LUTC)	// error log handler
	infoLog = log.New(os.Stdout, "INFO\t", log.LstdFlags | log.Lmicroseconds)	// for info level messages
	if errorLog == nil || infoLog == nil { log.Panic ("unable to create loggers") }
	return
}

func ConnectCockroach (ip string, port int, database, user string) (*sql.DB, error) {
    sslmode := "sslmode=disable"
    if CFG.ProductionLevel == models.ProductionLevel_Production {
		sslmode = fmt.Sprintf("sslmode=verify-full&sslcert=/cockroach-certs/client.%s.crt&sslkey=/cockroach-certs/client.%s.key&sslrootcert=/cockroach-certs/ca.crt",
			user, user)
	}
	
	sqlDB, err := sql.Open("postgres", fmt.Sprintf("postgres://%s@%s:%d/%s?%s", user, ip, port, database, sslmode))
	if err != nil { return nil, errors.WithStack (err) }

	_, err = sqlDB.Exec(`SET TIME ZONE 'UTC'`)   //set our default timezone
	if err != nil { return nil, errors.WithStack (err) }

	return sqlDB, cockroach.SetDB (sqlDB) // save this to our global in the database layer
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

