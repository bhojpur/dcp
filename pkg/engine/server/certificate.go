package server

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bhojpur/dcp/cmd/grid/dcpsvr/config"
	"github.com/bhojpur/dcp/pkg/engine/certificate/interfaces"
)

const (
	tokenKey = "jointoken"
)

// updateTokenHandler returns a http handler that update bootstrap token in the bootstrap-hub.conf file
// in order to update node certificate when both node certificate and old join token expires
func updateTokenHandler(certificateMgr interfaces.EngineCertificateManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokens := make(map[string]string)
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&tokens)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, "could not decode tokens, %v", err)
			return
		}

		joinToken := tokens[tokenKey]
		if len(joinToken) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "no join token is set")
			return
		}

		err = certificateMgr.Update(&config.EngineConfiguration{JoinToken: joinToken})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "could not update bootstrap token, %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "update bootstrap token successfully")
		return
	})
}
