/*
 *
 * Copyright 2015, Google Inc.
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     * Redistributions of source code must retain the above copyright
 * notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above
 * copyright notice, this list of conditions and the following disclaimer
 * in the documentation and/or other materials provided with the
 * distribution.
 *     * Neither the name of Google Inc. nor the names of its
 * contributors may be used to endorse or promote products derived from
 * this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

/*
Package glogger defines glog-based logging for grpc.
*/
package common

import (
	"fmt"

	"github.com/golang/glog"
)

// GLogger defines glog-based logging for grpc. This is a modification of
// glogger (https://github.com/grpc/grpc-go/blob/master/grpclog/glogger/glogger.go)
// that includes a new Disable function to disable logging.
type GLogger struct {
	disable      bool
	disableFatal bool
}

func (g *GLogger) Fatal(args ...interface{}) {
	if g.disable && g.disableFatal {
		return
	}
	glog.FatalDepth(2, args...)
}

func (g *GLogger) Fatalf(format string, args ...interface{}) {
	if g.disable && g.disableFatal {
		return
	}
	glog.FatalDepth(2, fmt.Sprintf(format, args...))
}

func (g *GLogger) Fatalln(args ...interface{}) {
	if g.disable && g.disableFatal {
		return
	}
	glog.FatalDepth(2, fmt.Sprintln(args...))
}

func (g *GLogger) Print(args ...interface{}) {
	if g.disable {
		return
	}
	glog.InfoDepth(2, args...)
}

func (g *GLogger) Printf(format string, args ...interface{}) {
	if g.disable {
		return
	}
	glog.InfoDepth(2, fmt.Sprintf(format, args...))
}

func (g *GLogger) Println(args ...interface{}) {
	if g.disable {
		return
	}
	glog.InfoDepth(2, fmt.Sprintln(args...))
}

// Disable prevents the logger from calling glog in
// non-fatal log methods. If includeFatal is set to true,
// fatal logs will be disabled.
func (g *GLogger) Disable(v bool, includeFatal bool) {
	g.disable = v
	g.disableFatal = includeFatal
}
