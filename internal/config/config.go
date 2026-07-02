//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

type Config struct {
	Director     DirectorConfig
	AuthPAM      AuthPAM
	AuthLDAP     AuthLDAP
	AuthRadius   AuthRadius
	AuthExternal AuthExternal
	Housekeeper  HousekeeperConfig
	Log          LogConfig
	Gateway      GatewayConfig
}

type DirectorConfig struct {
	DirectorHost   string
	DirectorPort   int
	AgentPort      int
	AuthType       string
	TokenKey       string
	PolicyDBFolder string
	DataDir        string
}

type AuthPAM struct {
	PAMServiceName string
}

type AuthLDAP struct {
	LDAPAddress             string
	LDAPBaseDN              string
	LDAPBindUser            string
	LDAPBindPass            string
	LDAPFilter              string
	OTPType                 string
	OTPPrivacyIdeaURL       string
	OTPPrivacyIdeaTLSStrict bool
	OTPCommand              string
	OTPArgs                 []string
}

type AuthRadius struct {
	RadiusServer string
	RadiusPort   int
	RadiusSecret string
}

type AuthExternal struct {
	Command string
	Args    []string
}

type HousekeeperConfig struct {
	HousekeeperFrequency string
	MaxAge               string
}

type GatewayConfig struct {
	GatewaysList []string
}

type LogConfig struct {
	Level     string
	Directory string
	Rotation  int
}

func getExecutablePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not get executable path: %w", err)
	}
	return filepath.Dir(ex), nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func findConfig(configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}

	path := "/etc/dcvix-director/dcvix-director.conf"
	if fileExists(path) {
		return path, nil
	}

	execPath, err := getExecutablePath()
	if err == nil {
		fallbackPath := filepath.Join(execPath, "dcvix-director.conf")
		if fileExists(fallbackPath) {
			return fallbackPath, nil
		}
	}

	if fileExists("./dcvix-director.conf") {
		return "./dcvix-director.conf", nil
	}

	return "", fmt.Errorf("configuration file not found")
}

func loadEmbeddedConfig() (*Config, error) {
	defaultFile, err := ini.Load(defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded default config: %w", err)
	}

	dDirector := defaultFile.Section("director")
	dPam := defaultFile.Section("pam_auth")
	dLdap := defaultFile.Section("ldap_auth")
	dRadius := defaultFile.Section("radius_auth")
	dExternal := defaultFile.Section("external_auth")
	dHousekeeper := defaultFile.Section("housekeeper")
	dLog := defaultFile.Section("log")
	dGateway := defaultFile.Section("gateway")

	return &Config{
		Director: DirectorConfig{
			DirectorHost:   dDirector.Key("director_host").String(),
			DirectorPort:   dDirector.Key("director_port").MustInt(8445),
			AgentPort:      dDirector.Key("agent_port").MustInt(8446),
			AuthType:       dDirector.Key("auth_type").String(),
			TokenKey:       dDirector.Key("token_key").String(),
			PolicyDBFolder: dDirector.Key("policydb_folder").String(),
			DataDir:        dDirector.Key("data_dir").String(),
		},
		AuthPAM: AuthPAM{
			PAMServiceName: dPam.Key("pam_service_name").String(),
		},
		AuthLDAP: AuthLDAP{
			LDAPAddress:             dLdap.Key("ldap_address").String(),
			LDAPBaseDN:              dLdap.Key("ldap_base_dn").String(),
			LDAPBindUser:            dLdap.Key("ldap_bind_user").String(),
			LDAPBindPass:            dLdap.Key("ldap_bind_password").String(),
			LDAPFilter:              dLdap.Key("ldap_filter").String(),
			OTPType:                 dLdap.Key("otp_type").String(),
			OTPPrivacyIdeaURL:       dLdap.Key("privacyidea_url").String(),
			OTPPrivacyIdeaTLSStrict: dLdap.Key("privacyidea_tls_strict").MustBool(true),
			OTPCommand:              dLdap.Key("otp_external_command").String(),
			OTPArgs:                 strings.Fields(dLdap.Key("otp_external_args").String()),
		},
		AuthRadius: AuthRadius{
			RadiusServer: dRadius.Key("radius_server").String(),
			RadiusPort:   dRadius.Key("radius_port").MustInt(1812),
			RadiusSecret: dRadius.Key("radius_secret").String(),
		},
		AuthExternal: AuthExternal{
			Command: dExternal.Key("external_command").String(),
			Args:    strings.Fields(dExternal.Key("args").String()),
		},
		Housekeeper: HousekeeperConfig{
			HousekeeperFrequency: dHousekeeper.Key("housekeeper_frequency").String(),
			MaxAge:               dHousekeeper.Key("max_age").String(),
		},
		Log: LogConfig{
			Level:     dLog.Key("level").String(),
			Directory: dLog.Key("directory").String(),
			Rotation:  dLog.Key("rotation").MustInt(2),
		},
		Gateway: GatewayConfig{
			GatewaysList: parseGatewayList(dGateway.Key("gateways_list").String()),
		},
	}, nil
}

func parseGatewayList(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func LoadConfig(configPath string) (*Config, error) {
	cfg, err := loadEmbeddedConfig()
	if err != nil {
		return nil, err
	}

	configPath, err = findConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to locate config file: %w", err)
	}

	file, err := ini.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Load Director section
	directorSection := file.Section("director")
	if directorSection.HasKey("director_host") {
		cfg.Director.DirectorHost = directorSection.Key("director_host").String()
	}
	cfg.Director.DirectorPort = directorSection.Key("director_port").MustInt(cfg.Director.DirectorPort)
	cfg.Director.AgentPort = directorSection.Key("agent_port").MustInt(cfg.Director.AgentPort)
	if directorSection.HasKey("auth_type") {
		cfg.Director.AuthType = directorSection.Key("auth_type").String()
	}
	if directorSection.HasKey("token_key") {
		cfg.Director.TokenKey = directorSection.Key("token_key").String()
	}
	if directorSection.HasKey("policydb_folder") {
		cfg.Director.PolicyDBFolder = directorSection.Key("policydb_folder").String()
	}
	if directorSection.HasKey("data_dir") {
		cfg.Director.DataDir = directorSection.Key("data_dir").String()
	}

	// Load PAM Auth section
	pamAuthSection := file.Section("pam_auth")
	if pamAuthSection.HasKey("pam_service_name") {
		cfg.AuthPAM.PAMServiceName = pamAuthSection.Key("pam_service_name").String()
	}

	// Load LDAP Auth section
	ldapAuthSection := file.Section("ldap_auth")
	if ldapAuthSection.HasKey("ldap_address") {
		cfg.AuthLDAP.LDAPAddress = ldapAuthSection.Key("ldap_address").String()
	}
	if ldapAuthSection.HasKey("ldap_base_dn") {
		cfg.AuthLDAP.LDAPBaseDN = ldapAuthSection.Key("ldap_base_dn").String()
	}
	if ldapAuthSection.HasKey("ldap_bind_user") {
		cfg.AuthLDAP.LDAPBindUser = ldapAuthSection.Key("ldap_bind_user").String()
	}
	if ldapAuthSection.HasKey("ldap_bind_password") {
		cfg.AuthLDAP.LDAPBindPass = ldapAuthSection.Key("ldap_bind_password").String()
	}
	if ldapAuthSection.HasKey("ldap_filter") {
		cfg.AuthLDAP.LDAPFilter = ldapAuthSection.Key("ldap_filter").String()
	}
	if ldapAuthSection.HasKey("otp_type") {
		cfg.AuthLDAP.OTPType = ldapAuthSection.Key("otp_type").String()
	}
	if ldapAuthSection.HasKey("privacyidea_url") {
		cfg.AuthLDAP.OTPPrivacyIdeaURL = ldapAuthSection.Key("privacyidea_url").String()
	}
	cfg.AuthLDAP.OTPPrivacyIdeaTLSStrict = ldapAuthSection.Key("privacyidea_tls_strict").MustBool(cfg.AuthLDAP.OTPPrivacyIdeaTLSStrict)
	if ldapAuthSection.HasKey("otp_external_command") {
		cfg.AuthLDAP.OTPCommand = ldapAuthSection.Key("otp_external_command").String()
	}
	if ldapAuthSection.HasKey("otp_external_args") {
		cfg.AuthLDAP.OTPArgs = strings.Fields(ldapAuthSection.Key("otp_external_args").String())
	}

	// Load Radius Auth section
	radiusAuthSection := file.Section("radius_auth")
	if radiusAuthSection.HasKey("radius_server") {
		cfg.AuthRadius.RadiusServer = radiusAuthSection.Key("radius_server").String()
	}
	cfg.AuthRadius.RadiusPort = radiusAuthSection.Key("radius_port").MustInt(cfg.AuthRadius.RadiusPort)
	if radiusAuthSection.HasKey("radius_secret") {
		cfg.AuthRadius.RadiusSecret = radiusAuthSection.Key("radius_secret").String()
	}

	// Load External Auth section
	externalAuthSection := file.Section("external_auth")
	if externalAuthSection.HasKey("external_command") {
		cfg.AuthExternal.Command = externalAuthSection.Key("external_command").String()
	}
	if externalAuthSection.HasKey("args") {
		cfg.AuthExternal.Args = strings.Fields(externalAuthSection.Key("args").String())
	}

	// Load Housekeeper section
	housekeeperSection := file.Section("housekeeper")
	if housekeeperSection.HasKey("housekeeper_frequency") {
		cfg.Housekeeper.HousekeeperFrequency = housekeeperSection.Key("housekeeper_frequency").String()
	}
	if housekeeperSection.HasKey("max_age") {
		cfg.Housekeeper.MaxAge = housekeeperSection.Key("max_age").String()
	}

	// Load Log section
	logSection := file.Section("log")
	if logSection.HasKey("level") {
		cfg.Log.Level = logSection.Key("level").String()
	}
	if logSection.HasKey("directory") {
		cfg.Log.Directory = logSection.Key("directory").String()
	}
	cfg.Log.Rotation = logSection.Key("rotation").MustInt(cfg.Log.Rotation)

	// Load Gateway section
	gatewaySection := file.Section("gateway")
	if gatewaySection.HasKey("gateways_list") {
		cfg.Gateway.GatewaysList = parseGatewayList(gatewaySection.Key("gateways_list").String())
	}

	log.Infof("Loaded configuration file: %s", configPath)

	if cfg.Director.DataDir == "" {
		cfg.Director.DataDir = "/var/lib/dcvix-director"
	}
	if cfg.Log.Directory == "" {
		cfg.Log.Directory = "/var/log/dcvix-director"
	}

	return cfg, nil
}
