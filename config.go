package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

var (
	jwtAlgorithms = map[string]*jwt.SigningMethodHMAC{
		"HS256": jwt.SigningMethodHS256,
		"HS384": jwt.SigningMethodHS384,
		"HS512": jwt.SigningMethodHS512,
	}
)

func NewConfig() Config {
	sc := SessionConfig{
		Driver:          "postgres",
		Host:            os.Getenv("POSTGRES_HOST"),
		Port:            os.Getenv("POSTGRES_PORT"),
		User:            os.Getenv("POSTGRES_USER"),
		Password:        os.Getenv("POSTGRES_PASSWORD"),
		DBName:          os.Getenv("POSTGRES_DB"),
		Schema:          os.Getenv("POSTGRES_SCHEMA"),
		TestDBName:      "kariz_test",
		AdminDBName:     "postgres",
		SslMode:         "disable",
		TimeZone:        "Asia/Tehran",
		MigrationsPath:  "file://migrations",
		MigrationsTable: "migration_andsm",
	}

	maxAge, err := strconv.ParseInt(os.Getenv("JWT_MAXAGE"), 10, 64)
	if err != nil {
		logrus.Fatal(err)
	}
	jwtConfig := JWT{
		Secret:    []byte(os.Getenv("JWT_SECRET_KEY")),
		Algorithm: jwtAlgorithms[os.Getenv("JWT_ALGORITHM")],
		MaxAge:    maxAge,
		HTTPOnly:  os.Getenv("JWT_HTTPONLY") == "true",
	}

	cfg := Config{
		HTTPPort:       9000,
		DbSession:      sc,
		DockerRegistry: "reg.bernetco.ir",
		Agent: agent{
			Port: 6000,
			API: agentApis{
				Start: "api/sessions/start",
				Stop:  "api/session/stop",
			},
		},
		JWT:       jwtConfig,
		Templates: os.Getenv("TEMPLATES"),
	}
	mod := os.Getenv("ANDSM_ENV")
	if mod != "release" {
		logrus.Warning("Application run in DEVELOPER mode, set ANDSM_ENV to release for production mod")
		cfg.DevMode = true
	}

	return cfg
}

type Config struct {
	DevMode        bool
	SecretKey      string
	JWT            JWT
	HTTPPort       int
	DbSession      SessionConfig
	DockerRegistry string
	Agent          agent
	Templates      string
}

type agent struct {
	Port uint16
	API  agentApis
}

type agentApis struct {
	Start string
	Stop  string
}

type JWT struct {
	Secret       []byte
	Algorithm    jwt.SigningMethod
	MaxAge       int64
	HTTPOnly     bool
	RefreshToken RefreshToken
	Path         string
	Secure       bool
}

type RefreshToken struct {
	Secret    []byte
	Algorithm jwt.SigningMethodHMAC
	MaxAge    int64
	Secure    bool
	HTTPOnly  bool
	Path      string
}

type SessionConfig struct {
	Driver          string
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	Schema          string
	TestDBName      string
	AdminDBName     string
	SslMode         string
	TimeZone        string
	MigrationsPath  string
	MigrationsTable string
}

func (s SessionConfig) Dsn() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		s.Host, s.Port, s.User, s.Password, s.DBName, s.SslMode,
	)
}

func (s SessionConfig) DsnWithSchema() string {
	dsn := fmt.Sprintf("%s search_path=%s", s.Dsn(), s.Schema)

	return dsn
}

func (s SessionConfig) AmdinDsn() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		s.Host, s.Port, s.User, s.Password, s.AdminDBName, s.SslMode,
	)
}
