package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	JWT       JWTConfig       `yaml:"jwt"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

type ServerConfig struct {
	Port            string        `yaml:"port"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	ActiveLevel     string        `yaml:"active_level"`
	EnvPath         string        `yaml:"env_path"`
	RecapMode       bool          `yaml:"recap_mode"`
}

type DatabaseConfig struct {
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnTimeout     time.Duration `yaml:"conn_timeout"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	PingPeriod      time.Duration `yaml:"ping_period"`
}

type RedisConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

type JWTConfig struct {
	Expiration time.Duration `yaml:"expiration"`
}

type RateLimitConfig struct {
	RequestsPerMinute int64         `yaml:"rpm"`
	BurstSize         int64         `yaml:"burst_size"`
	WindowDuration    time.Duration `yaml:"window_duration"`
}

func LoadConfig(filename string) (*Config, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
