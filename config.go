package main

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

const DefaultConfigPath = "config.toml"
const DefaultBlockingIP = "0.0.0.0"
const DefaultTTL = 60
const DefaultDNSPort = 8053

func allDays() []int {
	days := make([]int, 7)
	for i := 0; i < 7; i++ {
		days[i] = i
	}
	return days
}

type Config struct {
	DNS struct {
		// Port to listen on for DNS requests
		Port int `toml:"port"`
	} `toml:"dns"`

	// Array of rules
	Rules map[string]Rule `toml:"rules"`

	// IP to return when something should be blocked
	BlockingIP string `toml:"blocking_ip"`

	// Default resolver to use for forwarding requests
	Resolver string `toml:"resolver"`

	// How long to cache blocked domains (in seconds)
	TTL uint32 `toml:"ttl"`
}

type Rule struct {
	// Regex for matching the domain name
	Domain string `toml:"domain"`

	// Blocking hours (0-23)
	Hours []int `toml:"hours"`

	// Blocking days of the week (0 = Sunday, 6 = Saturday)
	Days []int `toml:"days"`
}

func loadConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	config := &Config{}

	decoder := toml.NewDecoder(file)
	decoder.Decode(&config)

	// Set defaults
	if config.BlockingIP == "" {
		config.BlockingIP = DefaultBlockingIP
	}

	if config.Resolver == "" {
		config.Resolver = "1.1.1.1:53"
	}

	if config.TTL == 0 {
		config.TTL = DefaultTTL
	}

	if config.DNS.Port == 0 {
		config.DNS.Port = DefaultDNSPort
	}

	for name, rule := range config.Rules {
		// No hours = block all hours
		if rule.Hours == nil {
			rule.Hours = []int{0, 23}
		}

		// No days = block all days
		if rule.Days == nil {
			rule.Days = allDays()
		}

		// Hours must be one or two.
		if len(rule.Hours) != 1 && len(rule.Hours) != 2 {
			return nil, fmt.Errorf("invalid number of hours for rule %s: %d", name, len(rule.Hours))
		}

		// Sanity checks
		for _, hour := range rule.Hours {
			if hour < 0 || hour > 23 {
				return nil, fmt.Errorf("invalid hour for rule %s: %d", name, hour)
			}
		}

		// If only one hour is specified, block only for that hour.
		if len(rule.Hours) == 1 {
			rule.Hours = []int{rule.Hours[0], rule.Hours[0]}
		}

		// Check that the start hour is less than the end hour
		if len(rule.Hours) == 2 && rule.Hours[0] > rule.Hours[1] {
			return nil, fmt.Errorf("invalid hours for rule %s: start hour must be less than end hour", name)
		}

		config.Rules[name] = rule
	}

	return config, nil
}
