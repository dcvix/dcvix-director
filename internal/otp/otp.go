//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package otp

type OtpVerifier interface {
	Verify(userID, password, otp string) (bool, error)
}
