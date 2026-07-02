//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package auth

import (
	"fmt"
	"strings"

	"github.com/dcvix/dcvix-director/internal/config"
	"github.com/dcvix/dcvix-director/internal/otp"
	"github.com/go-ldap/ldap/v3"
	log "github.com/sirupsen/logrus"
)

type LDAPAuthenticator struct {
	Address                 string
	BaseDN                  string
	BindUser                string
	BindPass                string
	Filter                  string
	OTPType                 string
	OTPPrivacyIdeaURL       string
	OTPPrivacyIdeaTLSStrict bool
	OTPCommand              string
	OTPArgs                 []string
}

func NewLDAPAuthenticator(cfg config.AuthLDAP) Authenticator {
	return &LDAPAuthenticator{
		Address:                 cfg.LDAPAddress,
		BaseDN:                  cfg.LDAPBaseDN,
		BindUser:                cfg.LDAPBindUser,
		BindPass:                cfg.LDAPBindPass,
		Filter:                  cfg.LDAPFilter,
		OTPType:                 cfg.OTPType,
		OTPPrivacyIdeaURL:       cfg.OTPPrivacyIdeaURL,
		OTPPrivacyIdeaTLSStrict: cfg.OTPPrivacyIdeaTLSStrict,
		OTPCommand:              cfg.OTPCommand,
		OTPArgs:                 cfg.OTPArgs,
	}
}

func (a *LDAPAuthenticator) Auth(user UserLogin) (bool, error) {

	// if configured, verify OTP
	switch a.OTPType {
	case "privacyidea":
		var verifier otp.OtpVerifier = otp.NewPrivacyIdeaVerifier(a.OTPPrivacyIdeaURL, a.OTPPrivacyIdeaTLSStrict)
		otpRes, err := verifier.Verify(user.UserID, user.Password, user.OTP)
		if err != nil {
			return false, fmt.Errorf("privacy idea api check failed: %v", err)
		}
		if !otpRes {
			return false, fmt.Errorf("invalid OTP")
		}
	case "external":
		var verifier otp.OtpVerifier = otp.NewExternalVerifier(a.OTPCommand, a.OTPArgs)
		otpRes, err := verifier.Verify(user.UserID, user.Password, user.OTP)
		if err != nil {
			return false, fmt.Errorf("external otp check failed: %v", err)
		}
		if !otpRes {
			return false, fmt.Errorf("invalid OTP")
		}
	}

	conn, err := ldap.DialURL(a.Address)
	if err != nil {
		log.Errorf("LDAP connection failed: %v", err)
		return false, err
	}
	defer conn.Close()

	if err := conn.Bind(a.BindUser, a.BindPass); err != nil {
		log.Errorf("LDAP bind failed: %v", err)
		return false, err
	}

	escapedUserID := ldap.EscapeFilter(user.UserID)
	var filter string
	if a.Filter == "" {
		filter = fmt.Sprintf("(sAMAccountName=%s)", escapedUserID)
	} else {
		filter = strings.Replace(a.Filter, "%s", escapedUserID, 1)
	}

	searchRequest := ldap.NewSearchRequest(
		a.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn"},
		nil,
	)

	searchResp, err := conn.Search(searchRequest)
	if err != nil || len(searchResp.Entries) != 1 {
		return false, fmt.Errorf("user not found or multiple entries")
	}

	userDN := searchResp.Entries[0].DN
	if err := conn.Bind(userDN, user.Password); err != nil {
		return false, fmt.Errorf("LDAP authentication failed for user %s", user.UserID)
	}
	return true, nil
}
