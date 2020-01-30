/*! \file main.go
	\brief API version of our V4 api

	Created 2019-11-19 by NateDogg
*/

package main 

 import (
	"github.com/NathanRThomas/boiler_api/cmd"
	"github.com/NathanRThomas/boiler_api/pkg/models/cockroach"
	"github.com/NathanRThomas/boiler_api/pkg/models/redis"
		
	"github.com/patrickmn/go-cache"

	"fmt"
	"os"
	"net/http"
	"flag"
	"time"
	"sync"
 )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type app_c struct {
	cmd.App_c
	
	users 		cockroach.User_c
}


//global config object
var cfg struct {
	cmd.CFG
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MAIN --------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief You got to start somewhere
*/
func main() {
	// create our logs
	errorLog, infoLog := cmd.CreateLoggers ()
	
	// parse from file first
	err := cmd.ParseConfig (&cfg)
	if err != nil { errorLog.Fatalf ("Invalid Config: %s", err.Error()) }

	// now handle command line flags, these override the config file
	flag.BoolVar (&cfg.Version, "v", false, "Returns the version of the api")
	flag.StringVar (&cfg.Port, "p", "8082", "Port to run the cli service on")
	flag.IntVar (&cfg.LoggingLevel, "log", 0, "Debug log level")
	
	flag.Parse()

	if cfg.Version {
		fmt.Printf("\nLabs API Version: %s\n\n", cmd.API_ver)
		os.Exit(0)
	}

	// connect to our database(s)
	// redis
	redisDB, err := cmd.ConnectRedis (cfg.Redis.Cache.IP, cfg.Redis.Cache.Port)
	if err != nil { errorLog.Fatal (err) }	// we can't start without the cache service running
	
	// cockroach
	cockDB, err := cmd.ConnectCockroach (cfg.Cockroach.IP, cfg.Cockroach.Port, cfg.Cockroach.Database, cfg.Cockroach.User, cfg.ProductionLevel)
	if err != nil { errorLog.Fatal (err) }

	app := &app_c { App_c: cmd.App_c { Running: true,
						WG: new(sync.WaitGroup),
						ProductionLevel: cfg.ProductionLevel,
						LoggingLevel: cmd.LogLevel(cfg.LoggingLevel),
						ErrorLog: errorLog, 
						InfoLog: infoLog,
						Redis: &redis.DB_c { DB: redisDB }, 
						DB: &cockroach.DB_c { DB: cockDB },
						Cache: cache.New(60*time.Second, 10*time.Minute),	// local cache
					},
	}

	// task handlers
	app.TaskQue = cmd.NewQue (app, &cfg.Slack, &cfg.Mailgun)
	
	// server
	srv := &http.Server {
        Addr:     ":" + cfg.Port,
        ErrorLog: errorLog,
        Handler:  app.routes(),
	}

	// signal handling
	cmd.MonitorSignals(&app.Running, srv)
	
	infoLog.Printf("Starting API server on port %s\n", cfg.Port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {	// Error starting or closing listener:
		errorLog.Printf("API server ListenAndServe: %v", err)
	}

	// we're out of the LB at this point, but we still need to make sure we aren't still adding to the queues before we close them
	time.Sleep(time.Second * 3)
	close(app.TaskQue) // close this que, this will wait for all the queued items to get processed
	
	app.WG.Wait() //wait for the workers and queen to finish
	time.Sleep(time.Second * 3) //just to make sure we're really done

	// close down the database connections now that we're done handling requests
	cockDB.Close ()
	redisDB.Close ()
	
	os.Exit(0)	//final exit
}