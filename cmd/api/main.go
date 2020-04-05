/*! \file main.go
	\brief API code
*/

package main 

 import (
	"github.com/NathanRThomas/boiler_api/cmd"
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

const api_ver = cmd.API_major + "0.1" // "final" version of where we're at with just the api part of this

type app_c struct { // re-init this locally so we can add scope to functions we only need here
	cmd.App_c
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
	err := cmd.ParseConfig ()
	if err != nil { errorLog.Fatalf ("Invalid Config: %s", err.Error()) }

	// now handle command line flags, these override the config file
	flag.BoolVar (&cmd.CFG.Version, "v", false, "Returns the version of the api")
	flag.StringVar (&cmd.CFG.Port, "p", "8050", "Port to run the cli service on")
	
	flag.Parse()

	if cmd.CFG.Version {
		fmt.Printf("\n API Version: %s\n\n", api_ver)
		os.Exit(0)
	}

	// connect to our database(s)
	// redis
	if len(cmd.CFG.Redis.IPs) == 0 { errorLog.Fatal ("no redis ip address") }
	ip := cmd.CFG.Redis.IPs[len(cmd.CFG.Redis.IPs) -1] // always get the last one

	redisDB, err := cmd.ConnectRedis (ip, cmd.CFG.Redis.Port)
	if err != nil { errorLog.Fatal (err) }	// we can't start without the cache service running
	
	// cockroach
	cockDB, err := cmd.ConnectCockroach (cmd.CFG.Cockroach.IP, cmd.CFG.Cockroach.Port, cmd.CFG.Cockroach.Database, cmd.CFG.Cockroach.User)
	if err != nil { errorLog.Fatal (err) }

	app := &app_c { App_c: cmd.App_c {
			Running: true,
			WG: new(sync.WaitGroup),
			ErrorLog: errorLog, 
			InfoLog: infoLog,
			Redis: &redis.DB_c { DB: redisDB }, 
			Cache: cache.New(60*time.Second, 10*time.Minute),	// local cache
		},
	}
	

	// task handlers
	app.StartTaskQue()
	
	// server
	srv := &http.Server {
        Addr:     ":" + cmd.CFG.Port,
        ErrorLog: errorLog,
        Handler:  app.routes(),
	}

	// signal handling
	cmd.MonitorSignals(&app.Running, srv)
	
	infoLog.Printf("Starting API server on port %s\n", cmd.CFG.Port)
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