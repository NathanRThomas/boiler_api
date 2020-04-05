/*! \file routes.go
	\brief Pulls out the routing of the urls to functions
*/

package cmd

import (
	"github.com/NathanRThomas/boiler_api/pkg/models/redis"
	"github.com/NathanRThomas/boiler_api/pkg/models/cockroach"
	
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/justinas/alice"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	//"fmt"
	"net/http"
	"io/ioutil"
	"context"
	"runtime/debug"
	"time"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MIDDLEWARE --------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Update our headers to allow cors from anywhere
*/
func (this *App_c) cors (next http.Handler) http.Handler {
	return http.HandlerFunc (func(w http.ResponseWriter, r *http.Request) {
		//app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		w.Header().Set("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,Accept,Origin,User-Agent,DNT,Cache-Control,X-Mx-ReqToken,Keep-Alive,X-Requested-With,If-Modified-Since,Content-Range, Content-Disposition, Content-Description")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")  //we're handling things via json objects in the body of the request

		if r.Method == http.MethodOptions {
			w.Write(nil)
		} else {
			next.ServeHTTP(w, r)
		}
    })
}

/*! \brief Our api works using json encoded bodies in the requests, this reads it out for us and puts it into our context
*/
func (this *App_c) readBody (next http.Handler) http.Handler {
    return http.HandlerFunc (func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)	//reads the entire body posted by the user
		if err == nil {
			ctx := r.Context()
			ctx = context.WithValue(ctx, "body", body)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			this.StackTrace (errors.WithStack (err))
			next.ServeHTTP(w, r)
		}
    })
}

/*! \brief Tries to recover from any panics this go routine hit on its journey
*/
func (this *App_c) recoverPanic (next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Create a deferred function (which will always be run in the event
        // of a panic as Go unwinds the stack).
        defer func() {
            // Use the builtin recover function to check if there has been a
			// panic or not. If there has...
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
                w.Header().Set("Connection", "close")
                this.ErrorLog.Println(err)
				debug.PrintStack()
				this.ServerError (nil, ApiErrorCode_panicRecovery, w)
			}
        }()

        next.ServeHTTP(w, r)
    })
}

/*! \brief Adds parts of our global config to the context for the rest of the requests
*/
func (this *App_c) contextConfig (next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue (r.Context(), "slackConfig", &CFG.Slack)	// add this to our context, some tasks need it
		ctx = context.WithValue (ctx, "mailgunConfig", &CFG.Mailgun)	// add this to our context, some tasks need it
		
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

/*! \brief Creates a timeout for the request context so we bail on long-running requests
*/
func (this *App_c) requestTimeout (next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel (r.Context())
		defer cancel()

		go func() {
			next.ServeHTTP(w, r.WithContext(ctx))
			cancel()
		}()

		select {
		case <- ctx.Done():
			// we're good
		case <-time.After(time.Second * ContextTimeout):
			this.ErrorLog.Printf("Request timed out: %s %s %s\n", r.Proto, r.Method, r.URL.RequestURI())
			w.WriteHeader(http.StatusRequestTimeout)
		}
    })
}

func (this *App_c) longRequestCheck (next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		next.ServeHTTP(w, r)

		if time.Now().After (startTime.Add(time.Millisecond * 5000)) {
			this.InfoLog.Printf("Request took %s to complete: %s %s %s\n", time.Now().Sub(startTime), r.Proto, r.Method, r.URL.RequestURI())
		}
    })
}

/*! \brief Monitors the ip address of the requester and returns an error if they've had too many requests
*/
func (this *App_c) ddos (next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.RemoteAddr) > 0 {
			//this.InfoLog.Printf("remote address: %s\n", r.RemoteAddr)

			var cnt int64 // start at zero
			key := "ddos:" + r.RemoteAddr
			if data, found := this.Cache.Get (key); found { cnt = data.(int64) }	// we're cached
			
			cnt++	// keep adding to our count
			if cnt > 10 {
				this.InfoLog.Println ("ddos blocked", r.RemoteAddr)
				this.ErrorWithMsg (nil, w, http.StatusTooManyRequests, ApiErrorCode_passwordGuessing, "You're just guessing")
				return // don't serve next
			}

			this.Cache.Set (key, cnt, time.Minute) // cache this again for another minute
		} else {
			this.InfoLog.Println("request had no remote address")
		}
        next.ServeHTTP(w, r)
    })
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ROUTES ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *App_c) ready (next http.Handler) http.Handler {
	return http.HandlerFunc (func(w http.ResponseWriter, r *http.Request) {
		if this.Running {
			next.ServeHTTP(w, r)
		} else {
			//if we're here it's bad
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Running"))
		}
    })
}

func (this *App_c) live (next http.Handler) http.Handler {
    return http.HandlerFunc (func(w http.ResponseWriter, r *http.Request) {
		if err := cockroach.TestDB(); err == nil {
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

func (this *App_c) thingsLookGood (w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Things look good")) //we're good
}

func (this *App_c) notFound (w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ENTRY POINTS ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *App_c) Routes () *mux.Router {
	mux := mux.NewRouter().StrictSlash(true)
	
	// standard chain that all calls make
	readyCheck := alice.New (this.ready)
	liveCheck := readyCheck.Append (this.live)
	cors := alice.New (this.cors)

	mux.Handle ("/", cors.ThenFunc(this.notFound)) // default not found handler

	mux.Handle("/status/ready", readyCheck.ThenFunc(this.thingsLookGood)).Methods(http.MethodGet)	// default check
	mux.Handle("/status/live", liveCheck.ThenFunc(this.thingsLookGood)).Methods(http.MethodGet)	// database connection check

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
func (this *App_c) ApiChain () (alice.Chain, alice.Chain)  {
	std := alice.New (this.recoverPanic, this.requestTimeout, this.longRequestCheck, this.cors, this.contextConfig, this.readBody)
	ddos := std.Append (this.ddos)
	return std, ddos
}
