// SPDX-License-Identifier: MIT

package lexpvar

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.cryptoscope.co/luigi"
)

var i int

func TestExpvar(t *testing.T) {
	a := assert.New(t)
	o := luigi.NewObservable("value")
	k := fmt.Sprintf("key_%d", i)
	expvar.Publish(k, Expvar(o))

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/debug/vars", nil)
	a.NoError(err, "error making request")

	expvar.Handler().ServeHTTP(rr, req)

	var d map[string]interface{}
	err = json.NewDecoder(rr.Body).Decode(&d)
	a.NoError(err, "error decoding body")

	v, has := d[k]
	a.True(has)
	a.Equal("value", v)

	newV := "a thing used to open locks."
	o.Set(newV)

	expvar.Handler().ServeHTTP(rr, req)

	var d2 map[string]interface{}
	err = json.NewDecoder(rr.Body).Decode(&d2)

	v2, has := d2[k]
	a.True(has)
	a.Equal(newV, v2)

	i++
}
