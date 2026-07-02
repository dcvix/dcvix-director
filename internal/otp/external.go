//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package otp

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ExternalVerifier implements the OtpVerifier interface via external command
type ExternalVerifier struct {
	Command string
	Args    []string
}

// NewExternalVerifier creates a new ExternalVerifier instance
func NewExternalVerifier(command string, args []string) *ExternalVerifier {
	return &ExternalVerifier{
		Command: command,
		Args:    args,
	}
}

// Verify implements the OtpVerifier interface
func (v *ExternalVerifier) Verify(userID, password, otp string) (bool, error) {
	cmd := exec.Command(v.Command, v.Args...)

	// Prepare input: UserID\nPassword\n
	input := []byte(fmt.Sprintf("%s\n%s\n%s\n", userID, password, otp))
	cmd.Stdin = bytes.NewReader(input)

	// Capture output for debugging
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("external OTP verification failed: %v, stderr: %s", err, stderr.String())
	}

	return true, nil
}
