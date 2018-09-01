package lexpvar

import (
	"expvar"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.cryptoscope.co/luigi"
)

func TestExpvar(t *testing.T) {
	a := assert.New(t)
	o := luigi.NewObservable("value")
	expvar.Publish("key", Expvar(o))

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/debug/vars", nil)
	a.NoError(err, "error making request")

	expvar.Handler().ServeHTTP(rr, req)
	a.Contains(rr.Body.String(), `"key": "value"`, "could not find key value pair in output")
	t.Log(rr.Body.String())

	o.Set("a thing used to open locks.")

	expvar.Handler().ServeHTTP(rr, req)
	a.Contains(rr.Body.String(), `"key": "a thing used to open locks."`, "could not find dictionary entry in output")
	t.Log(rr.Body.String())
}
