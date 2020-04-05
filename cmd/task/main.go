/*! \file main.go
	\brief CLI version, or could be used a
*/

package main 

 import (
	"github.com/NathanRThomas/boiler_api/cmd"
	"github.com/NathanRThomas/boiler_api/pkg/models/redis"
	"github.com/NathanRThomas/boiler_api/pkg/models/cockroach"
		
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

const api_ver = cmd.API_major + "0.1"


type app_c struct {
	cmd.App_c

	tasks	cockroach.Task_c
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
	flag.BoolVar (&cmd.CFG.Version, "v", false, "Returns the version of the backend task service")
	flag.StringVar (&cmd.CFG.Port, "p", "8051", "Port to run the task service on")
	flag.BoolVar (&cmd.CFG.LocalRun, "local", false, "Used for regression testing, runs as a 'fails fast' setup")
	
	flag.Parse()

	if cmd.CFG.Version {
		fmt.Printf("\nCLI Version: %s\n\n", api_ver)
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

	// local cache
	cacheDB := cache.New(120*time.Second, 10*time.Minute)

	app := &app_c { App_c: cmd.App_c { 
			Running: true,
			WG: new(sync.WaitGroup),
			ErrorLog: errorLog, 
			InfoLog: infoLog,
			Redis: &redis.DB_c { DB: redisDB }, 
			Cache: cacheDB,
		},
	}

	// task handlers
	app.StartTaskQue()

	// server
	srv := &http.Server {
        Addr:     ":" + cmd.CFG.Port,
        ErrorLog: errorLog,
		Handler:  app.routes(),
		//WriteTimeout: 15 * time.Second, // don't set this, it prevents us from writing a response to the request after the timeout
		ReadTimeout:  15 * time.Second,
	}

	// signal handling
	cmd.MonitorSignals(&app.Running, srv)
	
	// start the background processes
	app.WG.Add(1)
	go app.queen()	// this gets its own thread
	
	infoLog.Printf("Starting task server on port %s\n", cmd.CFG.Port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {	// Error starting or closing listener:
		errorLog.Printf("Task server ListenAndServe: %v", err)
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