// Copyright 2018 The Prizem Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package logging

type logger interface {
	Print(args ...interface{})
	Printf(fmt string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}
