package fastcontroller

import (
	"fmt"

	"github.com/golang-jwt/jwt"
)

var (
	JwtAlgorithms = map[string]*jwt.SigningMethodHMAC{
		"HS256": jwt.SigningMethodHS256,
		"HS384": jwt.SigningMethodHS384,
		"HS512": jwt.SigningMethodHS512,
	}
)

type Config struct {
	DevMode   bool
	SecretKey string
	JWT       JWT
	HTTPPort  int
	DbSession SessionConfig
	Agent     Agent
}

type Agent struct {
	Port uint16
	API  AgentApis
}

type AgentApis struct {
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
