//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package auth

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/dcvix/dcvix-director/internal/config"
)

type ExternalAuthenticator struct {
	Command string   // Path to the external executable
	Args    []string // Optional arguments
}

func NewExternalAuthenticator(cfg config.AuthExternal) Authenticator {
	return &ExternalAuthenticator{
		Command: cfg.Command,
		Args:    cfg.Args,
	}
}

func (a *ExternalAuthenticator) Auth(user UserLogin) (bool, error) {
	cmd := exec.Command(a.Command, a.Args...)

	// Prepare input: UserID\nPassword\notp\n
	input := []byte(fmt.Sprintf("%s\n%s\n%s\n", user.UserID, user.Password, user.OTP))
	cmd.Stdin = bytes.NewReader(input)

	// Capture output for debugging
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("external auth failed: %v, stderr: %s", err, stderr.String())
	}

	return true, nil
}
