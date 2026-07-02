//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package token

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"

	log "github.com/sirupsen/logrus"
)

var symmetricKey paseto.V4SymmetricKey

func SetSymmetricKey(key paseto.V4SymmetricKey) {
	symmetricKey = key
}

func GetSymmetricKey() paseto.V4SymmetricKey {
	return symmetricKey
}

func InitSymmetricKey(configKey string) error {
	if configKey != "" {
		trimmed := configKey
		raw, err := base64.StdEncoding.DecodeString(trimmed)
		if err != nil {
			return fmt.Errorf("failed to decode token_key from config: %w", err)
		}
		key, err := paseto.V4SymmetricKeyFromBytes(raw)
		if err != nil {
			return fmt.Errorf("invalid token_key in config: %w", err)
		}
		SetSymmetricKey(key)
		log.Info("Loaded token key from config")
		return nil
	}

	key := paseto.NewV4SymmetricKey()
	encoded := base64.StdEncoding.EncodeToString(key.ExportBytes())

	SetSymmetricKey(key)
	log.Warn("========================================")
	log.Warn("No token_key found in config!")
	log.Warn("Generated a new ephemeral token key.")
	log.Warn("Add this to your config file for persistence:")
	log.Warnf("  token_key = %s", encoded)
	log.Warn("WARNING: Without a persistent key, all sessions will become invalid on restart.")
	log.Warn("========================================")

	return nil
}

func CreateToken(userID string, userCategory string) string {
	// Create a new PASETO v4.local token
	token := paseto.NewToken()
	now := time.Now()
	token.SetIssuedAt(now)
	token.SetNotBefore(now)
	token.SetExpiration(now.Add(2 * time.Hour))
	token.SetString("userID", userID)
	token.SetString("userCategory", userCategory)
	// TODO add token fingerprint
	// Encrypt the token with our symmetric key
	signedToken := token.V4Encrypt(GetSymmetricKey(), nil)
	return signedToken
}

func Verify(tokenString string) error {
	// Parse and verify the token
	parser := paseto.NewParser()
	token, err := parser.ParseV4Local(GetSymmetricKey(), tokenString, nil)
	if err != nil {
		err = errors.New("auth token invalid")
		return err
	}

	// Check if token is expired
	expiration, err := token.GetExpiration()
	if err != nil || time.Now().After(expiration) {
		err = errors.New("auth token expired")
		return err
	}

	// TODO verify token fingerprint

	userID, err := token.GetString("userID")
	if err != nil {
		log.Warnf("token.Verify: failed to get userID from token: %v", err)
	}
	log.Debugf("token.Verify: Token for user \"%s\" is valid.", userID)

	// Token is valid, proceed with the request, return no error
	return nil
}

func GetUserID(tokenString string) (string, error) {
	parser := paseto.NewParser()
	token, err := parser.ParseV4Local(GetSymmetricKey(), tokenString, nil)
	if err != nil {
		return "", err
	}
	return token.GetString("userID")
}

func GetUserCategory(tokenString string) (string, error) {
	parser := paseto.NewParser()
	token, err := parser.ParseV4Local(GetSymmetricKey(), tokenString, nil)
	if err != nil {
		return "", err
	}
	return token.GetString("userCategory")
}

func GetExpiration(tokenString string) (time.Time, error) {
	parser := paseto.NewParser()
	token, err := parser.ParseV4Local(GetSymmetricKey(), tokenString, nil)
	if err != nil {
		return time.Time{}, err
	}
	return token.GetExpiration()
}

// ExtractFromAuthHeader Extract the token from the Bearer scheme
func ExtractFromAuthHeader(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return parts[1], nil
}
