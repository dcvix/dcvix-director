//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package auth

type UserLogin struct {
	UserID   string `json:"userID"`
	Password string `json:"password"`
	OTP      string `json:"otp"`
}

type Authenticator interface {
	Auth(user UserLogin) (bool, error)
}
