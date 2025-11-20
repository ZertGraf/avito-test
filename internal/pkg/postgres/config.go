package postgres

import (
	"fmt"
	. "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"time"
)

type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Schema   string `json:"schema"`
	SSLMode  string `json:"ssl_mode"`

	MaxConns          int32         `json:"max_conns"`
	MinConns          int32         `json:"min_conns"`
	MaxConnLifetime   time.Duration `json:"max_conn_lifetime"`
	MaxConnIdleTime   time.Duration `json:"max_conn_idle_time"`
	HealthCheckPeriod time.Duration `json:"health_check_period"`
	ConnectTimeout    time.Duration `json:"connect_timeout"`
	AcquireTimeout    time.Duration `json:"acquire_timeout"`
}

func (c *Config) DSN() string {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
	)

	// add schema to search_path if specified
	if c.Schema != "" && c.Schema != "public" {
		dsn += fmt.Sprintf(" search_path=%s", c.Schema)
	}

	if c.ConnectTimeout > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", int(c.ConnectTimeout.Seconds()))
	}

	return dsn
}

func (c *Config) Validate() error {
	return ValidateStruct(c,
		Field(&c.Host, Required, is.Host),
		Field(&c.Port, Required, Min(1), Max(65535)),
		Field(&c.Username, Required, Length(1, 63)),
		Field(&c.Password, Required, Length(0, 1000)),
		Field(&c.Database, Required, Length(1, 63)),
		Field(&c.SSLMode, Required, In("disable", "allow", "prefer", "require", "verify-ca", "verify-full")),

		Field(&c.MaxConns, Required, Min(int32(1)), Max(int32(1000))),
		Field(&c.MinConns, Required, Min(int32(1)), By(c.validateMinConns)),
		Field(&c.MaxConnLifetime, Required, Min(time.Minute), Max(24*time.Hour)),
		Field(&c.MaxConnIdleTime, Required, Min(time.Second), Max(time.Hour)),
		Field(&c.HealthCheckPeriod, Required, Min(10*time.Second), Max(10*time.Minute)),
		Field(&c.ConnectTimeout, Min(time.Duration(0)), Max(time.Minute)),
		Field(&c.AcquireTimeout, Min(time.Duration(0)), Max(time.Minute)),
	)
}

func (c *Config) validateMinConns(value interface{}) error {
	minConns, ok := value.(int32)
	if !ok {
		return fmt.Errorf("min_conns must be an int32")
	}

	if minConns > c.MaxConns {
		return fmt.Errorf("min_conns (%d cannot be greater than max_conns (%d)", minConns, c.MaxConns)
	}
	return nil
}
