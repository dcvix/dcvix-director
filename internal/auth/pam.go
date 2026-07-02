//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package auth

import (
	"fmt"

	"github.com/dcvix/dcvix-director/internal/config"
	"github.com/msteinert/pam"
	log "github.com/sirupsen/logrus"
)

type PamAuthenticator struct {
	PAMServiceName string
}

func NewPamAuthenticator(cfg config.AuthPAM) Authenticator {
	return &PamAuthenticator{
		PAMServiceName: cfg.PAMServiceName,
	}
}

func (a *PamAuthenticator) Auth(user UserLogin) (bool, error) {
	t, err := pam.StartFunc(a.PAMServiceName, user.UserID, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return user.Password, nil
		}
		return "", fmt.Errorf("unsupported PAM style: %v", s)
	})

	if err != nil {
		return false, fmt.Errorf("pam start failed: %w", err)
	}

	err = t.Authenticate(0)
	if err != nil {
		log.Errorf("PAM authentication failed: %v", err)
		return false, nil
	}

	log.Info("PAM authentication successful")
	return true, nil
}
