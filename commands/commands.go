package commands

import (
	"github.com/urfave/cli"
)

var Commands = []cli.Command{
	{
		Name:   "doh",
		Usage:  "An application/dns-message request",
		Action: plainDnsRequest,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "domain, d",
				Value: "www.cloudflare.com.",
			},
			cli.StringFlag{
				Name:  "dnstype, t",
				Value: "AAAA",
			},
			cli.StringFlag{
				Name:  "target",
				Value: "localhost:8080",
			},
		},
	},
	{
		Name:   "odoh",
		Usage:  "An application/oblivious-dns-message request",
		Action: obliviousDnsRequest,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "domain, d",
				Value: "www.cloudflare.com.",
				Usage: "Domain name which needs to be resolved. Use trailing period (.).",
			},
			cli.StringFlag{
				Name:  "dnstype, t",
				Value: "AAAA",
				Usage: "Type of DNS Question. Currently supports A, AAAA, CAA, CNAME",
			},
			cli.StringFlag{
				Name:  "target",
				Value: "localhost:8080",
				Usage: "Hostname:Port format declaration of the target resolver hostname",
			},
			cli.StringFlag{
				Name:  "proxy, p",
				Usage: "Hostname:Port format declaration of the proxy hostname",
			},
			cli.StringFlag{
				Name:  "config, c",
				Usage: "ODoHConfigs to use for the query, encoded as a hexadecimal string",
			},
		},
	},
	{
		Name:   "odohconfig-fetch",
		Usage:  "Retrieves the ObliviousDoHConfigs of the target resolver",
		Action: getTargetConfigs,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "target",
				Value: "localhost:8080",
			},
			cli.BoolFlag{
				Name: "pretty",
			},
		},
	},
	{
		Name:   "odohconfig-mint",
		Usage:  "Mints a singleton ObliviousDoHConfig with the specified (KEM, KDF, AEAD) HPKE ciphersuite",
		Action: createConfigurations,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "kemid",
				Value: "32",
			},
			cli.StringFlag{
				Name:  "kdfid",
				Value: "1",
			},
			cli.StringFlag{
				Name:  "aeadid",
				Value: "1",
			},
		},
	},
	{
		Name:   "bench",
		Usage:  "Performs a benchmark for ODOH Target Resolver",
		Action: benchmarkClient,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "data",
				Value: "dataset.csv",
			},
			cli.Uint64Flag{
				Name:  "pick",
				Value: 10,
			},
			cli.Uint64Flag{
				Name:  "numclients",
				Value: 10,
			},
			cli.Uint64Flag{
				Name:  "rate", // We default to the rate per minute. Please provide this rate in req/min to make.
				Value: 15,
			},
			cli.StringFlag{
				Name:  "logout",
				Value: "log.txt",
			},
			cli.StringFlag{
				Name:  "out",
				Value: "",
				Usage: "Filename to save serialized JSON response from benchmark execution (eg. output.json). " +
					"If no filename is provided, or failure to write to file, the default will print to console.",
			},
			cli.StringFlag{
				Name:  "target",
				Value: "localhost:8080",
				Usage: "Hostname:Port format declaration of the target resolver hostname",
			},
			cli.StringFlag{
				Name:  "proxy, p",
				Usage: "Hostname:Port format declaration of the proxy hostname",
			},
			cli.StringFlag{
				Name:  "dnstype, t",
				Value: "A",
			},
		},
	},
}
