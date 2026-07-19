package main

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"
)

type Context struct {
	// Configuration loaded from the config file
	Config *Config

	// DNS client for forwarding requests
	Client *dns.Client

	// Logger
	Logger zerolog.Logger
}

func makeDNSHandler(ctx *Context) func(dns.ResponseWriter, *dns.Msg) {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true

		// Check if the request's domain matches any of the rules in the config
		domain := r.Question[0].Name

		// If it's not in the block cache and it doesn't match any rules, we will forward the request to the upstream resolver
		if !shouldBlock(domain, ctx) {
			ctx.Logger.Info().Str("domain", domain).Msg("Forwarding request to upstream resolver")

			response, _, err := ctx.Client.Exchange(r, ctx.Config.Resolver)
			if err == nil {
				m = response
			}

			// Override the TTL to the lowest value of 60 seconds to prevent caching of blocked domains
			for _, answer := range m.Answer {
				answer.Header().Ttl = ctx.Config.TTL
			}
		} else {
			ctx.Logger.Info().Str("domain", domain).Msg("Blocking request and returning unreachable IP address")

			// Block the request by returning a response with the unreachable IP address
			rr, err := dns.NewRR(fmt.Sprintf("%s A %s", domain, ctx.Config.BlockingIP))
			rr.Header().Ttl = ctx.Config.TTL

			if err == nil {
				m.Answer = append(m.Answer, rr)
			}
		}

		w.WriteMsg(m)
	}
}

func shouldBlock(domain string, ctx *Context) bool {
	// Strip the trailing dot from the domain name
	if domain[len(domain)-1] == '.' {
		domain = domain[:len(domain)-1]
	}

	// Check if the domain matches any of the rules in the config
	for _, rule := range ctx.Config.Rules {
		if rule.Domain != "" {
			regexp, err := regexp.Compile(rule.Domain)
			if err != nil {
				fmt.Printf("Invalid regex in config: %s\n", err.Error())
				continue
			}

			if regexp.MatchString(domain) {

				// Check if the current time is within the blocking hours and days
				currentHour := time.Now().Hour()
				currentDay := int(time.Now().Weekday())

				// Hours are inclusive
				if currentHour >= rule.Hours[0] && currentHour <= rule.Hours[1] {
					for _, day := range rule.Days {
						if day == currentDay {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	// get `CONFIG_PATH` environment variable, if not set, use default config path
	cfgPath := os.Getenv("CONFIG_PATH")
	cfg, err := loadConfig(cfgPath)

	if err != nil {
		logger.Error().Err(err).Msg("Error loading config")
		os.Exit(1)
	}

	ctx := &Context{
		Config: cfg,
		Client: &dns.Client{},
		Logger: logger,
	}

	dns.HandleFunc(".", makeDNSHandler(ctx))

	servers := []struct {
		Addr string
		Net  string
	}{
		{Addr: ":53", Net: "udp"},
		{Addr: ":53", Net: "tcp"},
	}

	for _, server := range servers {
		go func(s struct{ Addr, Net string }) {
			srv := &dns.Server{Addr: s.Addr, Net: s.Net}
			ctx.Logger.Info().Str("network", s.Net).Msgf("Starting DNS server on %s...", s.Addr)
			if err := srv.ListenAndServe(); err != nil {
				ctx.Logger.Error().Err(err).Msgf("Failed to start %s server", s.Net)
				os.Exit(1)
			}
		}(server)
	}

	// Wait until a termination signal is received
	select {}
}
