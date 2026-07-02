//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package auth

import (
	"context"
	"fmt"

	"github.com/dcvix/dcvix-director/internal/config"
	log "github.com/sirupsen/logrus"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type RadiusAuthenticator struct {
	RadiusServer string
	RadiusPort   int
	RadiusSecret string
}

func NewRadiusAuthenticator(cfg config.AuthRadius) Authenticator {
	return &RadiusAuthenticator{
		RadiusServer: cfg.RadiusServer,
		RadiusPort:   cfg.RadiusPort,
		RadiusSecret: cfg.RadiusSecret,
	}
}

func (a *RadiusAuthenticator) Auth(user UserLogin) (bool, error) {

	packet := radius.New(radius.CodeAccessRequest, []byte(a.RadiusSecret))
	rfc2865.UserName_SetString(packet, user.UserID)
	var password string
	// Add OTP to password if provided
	if user.OTP != "" {
		password = fmt.Sprintf("%s%s", user.Password, user.OTP)
	} else {
		password = user.Password
	}
	rfc2865.UserPassword_SetString(packet, password)
	response, err := radius.Exchange(
		context.Background(),
		packet,
		fmt.Sprintf("%s:%v", a.RadiusServer, a.RadiusPort),
	)
	if err != nil {
		log.Errorf("RADIUS exchange failed: %v", err)
		return false, fmt.Errorf("RADIUS server unavailable: %w", err)
	}

	log.Debugf("RADIUS response code: %v", response.Code)

	if response.Code == radius.CodeAccessAccept {
		return true, nil
	}

	return false, nil
}
