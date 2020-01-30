/*! \file routes.go
	\brief Pulls out the routing of the urls to functions
*/

package cmd

import (
	"github.com/NathanRThomas/boiler_api/pkg/models/redis"
	
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/justinas/alice"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"fmt"
	"net/http"
	"io/ioutil"
	"context"
	"database/sql/driver"
	"database/sql"
)


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type Handler_c struct {
	App_c
	ApiRequests *prometheus.CounterVec 
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MIDDLEWARE --------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Update our headers to allow cors from anywhere
*/
func (this *Handler_c) cors (next http.Handler) http.Handler {
	return http.HandlerFunc (func(w http.ResponseWriter, r *http.Request) {
		//app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		w.Header().Set("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,Accept,Origin,User-Agent,DNT,Cache-Control,X-Mx-ReqToken,Keep-Alive,X-Requested-With,If-Modified-Since,Content-Range, Content-Disposition, Content-Description")
		//w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")  //we're handling things via json objects in the body of the request

        next.ServeHTTP(w, r)
    })
}

/*! \brief Depending on our logging level we can write incoming and out-going json body objects to the info log
*/
func (this *Handler_c) log (next http.Handler) http.Handler {
    return http.HandlerFunc (func(w http.ResponseWriter, r *http.Request) {
		if this.LoggingLevel >= LogLevel_basic {
			this.InfoLog.Printf("%s %s %s", r.Proto, r.Method, r.URL.RequestURI())
		}
        next.ServeHTTP(w, r)
    })
}

/*! \brief Our api works using json encoded bodies in the requests, this reads it out for us and puts it into our context
*/
func (this *Handler_c) readBody (next http.Handler) http.Handler {
    return http.HandlerFunc (func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)	//reads the entire body posted by the user
		if err == nil {
			if this.LoggingLevel >= LogLevel_full { this.InfoLog.Printf("%s", string(body)) }
			ctx := r.Context()
			ctx = context.WithValue(ctx, "body", body)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			this.StackTrace (errors.WithStack (err))
			next.ServeHTTP(w, r)
		}
    })
}

/*! \brief Starts our database transaction for the duration of this request
	This will alos check for a panic and try to recover from it
*/
func (this *Handler_c) sqlTransaction (next http.Handler) http.Handler {
    return http.HandlerFunc (func(w http.ResponseWriter, r *http.Request) {
		
		ctx, tx, err := this.DB.Begin (r.Context())	// create a transaction for this process
		if err != nil {
			this.StackTrace (err) // record this error, no database connection available
			return // bail, we can't continue
		}
		
		defer func() { // using a defer as even in the event of a panic this still will be executed
			err := tx.Rollback() // always do this, if it was successful and committed then this will fail
			switch errors.Cause(err) {
			case nil, driver.ErrBadConn, sql.ErrTxDone, sql.ErrConnDone: // don't record these "expected" errors

			default:
				fmt.Println("rollback failed: ", errors.Cause(err))
			}
		}()

		next.ServeHTTP(w, r.WithContext(ctx)) // continue on
    })
}

/*! \brief Tries to recover from any panics this go routine hit on its journey
*/
func (this *Handler_c) recoverPanic (next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Create a deferred function (which will always be run in the event
        // of a panic as Go unwinds the stack).
        defer func() {
            // Use the builtin recover function to check if there has been a
            // panic or not. If there has...
            if err := recover(); err != nil {
                // Set a "Connection: close" header on the response.
                w.Header().Set("Connection", "close")
                // Call the app.serverError helper method to return a 500
				// Internal Server response.
				//this.ErrorLog.Println(err)
				fmt.Println(err)
				this.ServerError (nil, w)
            }
        }()

        next.ServeHTTP(w, r)
    })
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ROUTES ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *Handler_c) ready (next http.Handler) http.Handler {
	return http.HandlerFunc (func(w http.ResponseWriter, r *http.Request) {
		if *this.PRunning {
			next.ServeHTTP(w, r)
		} else {
			//if we're here it's bad
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Running"))
		}
    })
}

func (this *Handler_c) live (next http.Handler) http.Handler {
    return http.HandlerFunc (func(w http.ResponseWriter, r *http.Request) {
		if err := this.DB.Test(); err == nil {
			err := this.Redis.Ping()
			if err == nil || errors.Cause(err) == redis.ErrServiceDown {
				next.ServeHTTP(w, r)
			} else {
				//if we're here it's bad
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Redis cache service is down"))
			}
		} else {
			//if we're here it's bad
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(err.Error()))
		}
    })
}

func (this *Handler_c) allGood (w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Things look good")) //we're good
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ENTRY POINTS ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Creates our base "routes"
	These are used for all the differen cmd's you can build, these are endpoint needed for kubernetes to monitor the health of a service
*/
func (this *Handler_c) Routes () *mux.Router {
	mux := mux.NewRouter().StrictSlash(true)

	// standard chain that all calls make
	readyCheck := alice.New (this.ready)
	liveCheck := readyCheck.Append (this.live)

	mux.Handle("/ready", readyCheck.ThenFunc (this.allGood)).Methods(http.MethodGet) // default check
	mux.Handle("/live", liveCheck.ThenFunc(this.allGood)).Methods(http.MethodGet) // database connection check

	// metrics handled through prometheus
	this.ApiRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "How many requests processed, partitioned by bot.",
		},
		[]string{"bot", "endpoint", "code"},
	)

	prometheus.MustRegister(this.ApiRequests)
	http.Handle("/metrics", promhttp.Handler())

    return mux
}

/*! \brief Re-used default starting point for any api endpoint
*/
func (this *Handler_c) ApiChain () alice.Chain  {
	return alice.New (this.recoverPanic, this.log, this.cors, this.readBody, this.sqlTransaction)
}