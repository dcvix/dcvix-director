//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package user

import (
	"bytes"
	"encoding/xml"
	"net/http"

	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/token"
	log "github.com/sirupsen/logrus"
)

// AuthTokenVerify handles POST /v1/user/authtokenverify.
func AuthTokenVerify(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("POST /v1/user/authtokenverify: request from: %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "text/plain")

	authenticationToken := r.FormValue("authenticationToken")

	err := token.Verify(authenticationToken)
	if err != nil {
		w.Write([]byte(`<auth result="no"><message>dcvix-director: authentication token verification failed</message></auth>`))
		return
	}

	userID, err := token.GetUserID(authenticationToken)
	if err != nil {
		w.Write([]byte(`<auth result="no"><message>dcvix-director: could not get UserID from authentication token</message></auth>`))
		return
	}

	var buf bytes.Buffer
	buf.WriteString(`<auth result="yes"><username>`)
	xml.EscapeText(&buf, []byte(userID))
	buf.WriteString(`</username></auth>`)
	w.Write(buf.Bytes())
}
