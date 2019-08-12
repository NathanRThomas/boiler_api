/*! \file api.go
    \brief Main file/entry point
    Written in GO, this will be the a restful api.
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
    "context"

	"github.com/NathanRThomas/boiler_api/db"
	"github.com/NathanRThomas/boiler_api/web"
	"github.com/NathanRThomas/boiler_api/toolz"
)

const API_VER = "0.1.0"

type apiHandler func ([]byte, string, string) ([]byte, int)

type handler_c struct {
	Running 	bool
	DebugLevel	int64 
    tutor       toolz.Tutor
    users       db.User_c
	ip          toolz.IP_c
	login 		db.Login_c
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Public call that returns data no matter who's asking
*/
func (this handler_c) public (body []byte, endpoint, call string) (returnMsg []byte, status int) {
    status = http.StatusOK
	var class web.PublicEntry
	var jErr error
	
    switch endpoint {
    case "login":
        class = &web.Login_c { Post : body }
    default:
		status = http.StatusNotFound
		return
	}
	
	returnMsg, jErr = class.Entry(call)
	toolz.ErrChk(jErr)
    return
}

/*! \brief Private call that returns data only if the request is being made by a valid logged in user
*/
func (this handler_c) private (body []byte, endpoint, call string) (returnMsg []byte, status int) {
    status = http.StatusOK
	var class web.PublicEntry
	var jErr error
	
	user := &db.User_t{}
	toolz.UM(body, user)    //parse our user info
	
	//see if we got a valid user
	if this.login.Token (user) {
		switch endpoint {
		case "user":
			class = &web.User_c { Post : body, User : user }
		default:
			status = http.StatusNotFound
			return
		}
		
		returnMsg, jErr = class.Entry(call)
		toolz.ErrChk(jErr)
	} else {
        returnMsg, _ = json.Marshal(web.APIResponse_t{ Msg: "Please login" })
        status = http.StatusUnauthorized
    }
    return
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \fn apiHandler(w http.ResponseWriter, r *http.Request)
 *  \brief Main handler for the wepage requests.  This figures out where the request is headed
 */
func (this handler_c) Api (w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" { return } //this is a "test" request sent by javascript to test if the call is valid, or something, so just ignore it
    tb := toolz.TimeBox ("top api handler")
	defer tb.Defer()

	body, _ := ioutil.ReadAll(r.Body)	//reads the entire body posted by the user
	w.Header().Set("Content-Type", "application/json")	//we're handling things via json objects in the body of the request
	returnMsg, _ := json.Marshal(web.APIResponse_t{false, "You did it wrong!", web.API_error_bad_call}) //seed a bad call here
	returnStatus := http.StatusMethodNotAllowed
    parts := this.ip.ParseUrlParts(r.URL.Path, "")
    
	if len(parts) == 2 {
		this.tutor.Log (this.DebugLevel, "api", string(body), parts)
		var apiFunc apiHandler
        switch parts[0] { //figure out where this request is headed
		case "user":
			apiFunc = this.private
        default:
            apiFunc = this.public
		}
		
		returnMsg, returnStatus = apiFunc (body, parts[0], parts[1])
		tb.Note("switch done : %s : %s", parts[0], parts[1])
	}

	this.tutor.Log (this.DebugLevel, "api", string(returnMsg[:]))	//record this to our tutor
	w.WriteHeader(returnStatus)	//always write out the new status
	w.Write(returnMsg)	//write our result ot the output http response
}

/*! \brief Special handler for when we're uploading multi-part files.  We need to treat things differently
 */
func (this handler_c) Import (w http.ResponseWriter, r *http.Request) {
	returnMsg, _ := json.Marshal(web.APIResponse_t{false, "You did it wrong!", web.API_error_bad_call}) //seed a bad call here
	returnStatus := http.StatusMethodNotAllowed
	var jErr error
	tb := toolz.TimeBox ("top import handler")
	defer tb.Defer()

	nErr := r.ParseMultipartForm(15728640) ///! 15 mb limit
	if nErr == nil {
		user := &db.User_t{}
		user.ID = r.FormValue("user_id")
		user.Token = r.FormValue("user_token")
        if this.login.Token (user) {
			tb.Note("token login")
			this.tutor.Log (this.DebugLevel, "import", r.URL.Path)
			imp := web.Import_c { }

			if parts := this.ip.ParseUrlParts(r.URL.Path, ""); len(parts) > 2 {
				tb.Note("entry")
				returnMsg, jErr = imp.Entry(parts[2], r)
				tb.Note("completed")
				toolz.ErrChk(jErr)
			}
		} else {
			returnMsg, _ = json.Marshal(web.APIResponse_t{ Msg: "Please login" })
        	returnStatus = http.StatusUnauthorized
		}
	} else {
		returnMsg, _ = json.Marshal(web.APIResponse_t{ Msg : "Image too large! - " + nErr.Error() }) //seed a bad call here
		returnStatus = http.StatusRequestEntityTooLarge
	}

	this.tutor.Log (this.DebugLevel, "import", string(returnMsg[:]))
	w.WriteHeader(returnStatus)	//always write out the new status
	w.Write(returnMsg)	//write our result ot the output http response
}

/*! \brief Our re-direction site ported over.  This handles meta data so we can keep the dynamic websites
 */
func (this handler_c) Links (w http.ResponseWriter, r *http.Request) {
    tb := toolz.TimeBox ("top links handler")
	defer tb.Defer()
	compass := web.Compass_c {}	//init our object

    parts, ok := this.ip.ParseLinksUrl (r.URL.Path, r.UserAgent())
    if ok {
		tb.Note("links entry : %v", parts)
		url, html := compass.Entry(parts[1:])
        tb.Note("links result")
        
        if len(html) > 0 { //we want to return this as a page
            io.WriteString(w, html) //Done!
        } else {
            w.Header().Set("Location", url)    //now we write the redirect
            w.WriteHeader(http.StatusTemporaryRedirect)
        }
    } else {
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

/*! \brief Allows for a http based ping of the service to check for status
*/
func (this handler_c) Health (w http.ResponseWriter, r *http.Request) {
    var err error
	if this.Running {	//make sure we're still running
		/*
		if err = db.TestCockroach(); err == nil {	//now try cockroach
			if err = db.TestCouchbase(); err == nil {	//now try couchbase
				w.Write([]byte("Things look good")) //we're good
				return
			}
		}
		*/
		w.Write([]byte("Things look good")) //we're good
	}

	//if we're here it's bad
	w.WriteHeader(http.StatusServiceUnavailable)
	if err != nil { w.Write([]byte(err.Error())) }
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MAIN --------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds) //configure the logging for this application
	
	versionFlag := flag.Bool("v", false, "Returns the version of the api")
	port := flag.Int("p", 8080, "Port to run the server on")
	oneTime := flag.String("onetime", "", "If we need to run something across the entire database once")

	queenFlag := flag.Bool("queen", false, "Runs as the primary for cooridnation in a single threaded app")
	workerFlag := flag.Bool("worker", false, "Workers take tasks queued up for background processing")
	apiFlag := flag.Bool("api", false, "Runs as an api server")
	debugFlag := flag.Bool("debug", false, "writes logs to stdout")

	flag.Parse()

	if *versionFlag {
        fmt.Printf("\nVersion: %s\n\n", API_VER)
		os.Exit(0)
	}

	handler := handler_c { Running: true }	//using this as our "global" running flag, so i can tell the other threads that they should stop
	if *debugFlag { handler.DebugLevel = toolz.DebugLevel_all }

	toolz.LoadConfig() //figure out our defaults from our local config file
	db.ConnectEm(&handler.Running)	//this will throw a panic if anythings really wrong
	
	if len(*oneTime) > 0 {

	} else if *queenFlag || *workerFlag || *apiFlag {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		
		//for the backend proccessing, doing this so we can run master and a web server
		tickers := make([]*time.Ticker, 0)
		wg := new(sync.WaitGroup)

		if *queenFlag {
			//tickers for the tasks scheduled at intervals
			for _, m := range [2]int{1, 60} { tickers = append(tickers, time.NewTicker(time.Minute * time.Duration(m))) }

			queen := queen_c{ wg: wg, running: &handler.Running }
			
			// 1 minute ticker
			go func() {
				queen.OneMinute ()
				for range tickers[0].C {
					queen.OneMinute ()
				}
			}()

			// 60 minute
			go func() {
				queen.OneHour ()
				for range tickers[1].C {
					queen.OneHour ()
				}
			}()
		}

		if *workerFlag {
			worker := worker_c{ wg: wg, running: &handler.Running }
			go worker.Tasks ()	//always launch the slave processing
		}

		// all instances need a health check
		srv := &http.Server { Addr: fmt.Sprintf(":%d", *port) } //create an instance of the server
		go func() {
			<-c
			handler.Running = false
			time.Sleep(time.Second * 5)	//sleep here, this is needed for this to be taken out of the load balancer before stopping to handle requests
			srv.Shutdown (context.Background())  // graceful shutdown
		}()

		http.HandleFunc("/health/", handler.Health)	//health check for all
		
		if *apiFlag {	//now if this is running the api server, we want to add some more endpoints
			http.HandleFunc("/", handler.Api)
			http.HandleFunc("/import/", handler.Import)
			http.HandleFunc("/l/", handler.Links)
		}

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			toolz.Err("HTTP server ListenAndServe: %v", err)   // this is bad
		}
		//if we're here it's cause the http service is shutting down
	
		for _, t := range tickers { t.Stop() } //stop all the tickers

		wg.Wait() //wait for the slave and possible master to finish
		time.Sleep(time.Second) //just to make sure we're really done
	}
    
	db.CloseConnections() //close the database connections
    os.Exit(0)
}
