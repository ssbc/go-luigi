// SPDX-FileCopyrightText: 2021 The Luigi Authors
//
// SPDX-License-Identifier: MIT

package lexpvar

import (
	"expvar"

	"go.cryptoscope.co/luigi"
)

// Expvar returns an expvar.Var for the given observable.
func Expvar(o luigi.Observable) expvar.Var {
	return expvar.Func(func() interface{} {
		v, err := o.Value()
		if err != nil {
			return err
		}

		return v
	})
}
