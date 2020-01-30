/*! \file routes_test.go
	\brief Handles testing of our shared routes
*/

package cmd 

import (
	"github.com/NathanRThomas/boiler_api/pkg/models/mock"

	"io/ioutil"
    "net/http"
    "net/http/httptest"
    "testing"
)

func newTestHandler (t *testing.T) *Handler_c {
	running := true
	return &Handler_c { running: &running, db: &mock.DB_c{} }
}

// Define a custom testServer type which anonymously embeds a httptest.Server
// instance.
type testServer struct {
    *httptest.Server
}

// Create a newTestServer helper which initalizes and returns a new instance
// of our custom testServer type.
func newTestServer(t *testing.T, h http.Handler) *testServer {
    ts := httptest.NewServer(h)
    return &testServer{ts}
}

// Implement a get method on our custom testServer type. This makes a GET
// request to a given url path on the test server, and returns the response
// status code, headers and body.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
    rs, err := ts.Client().Get(ts.URL + urlPath)
    if err != nil {
        t.Fatal(err)
    }

    defer rs.Body.Close()
    body, err := ioutil.ReadAll(rs.Body)
    if err != nil {
        t.Fatal(err)
    }

    return rs.StatusCode, rs.Header, body
}

func TestRoutes (t *testing.T) {
	handler := newTestHandler(t)
    ts := newTestServer(t, Routes(handler.running, handler.db))
	defer ts.Close()
	
	tests := []struct {
		name, urlPath, body string
		code int 
	}{
		{"Ready", "/ready", "Things look good", http.StatusOK},
		{"Live", "/live", "Things look good", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			if code != tt.code {
				t.Errorf("want %d; got %d", tt.code, code)
			}

			if string(body) != tt.body {
				t.Errorf("want body to equal %s and got %s", tt.body, string(body))
			}
		})
	}    
}
