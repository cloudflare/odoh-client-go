package commands

import (
	"log"
	"net/url"
	"strings"

	odoh "github.com/cloudflare/odoh-go"
	"github.com/miekg/dns"
)

// Function for Converting CLI DNS Query Type to the uint16 Datatype
func dnsQueryStringToType(stringType string) uint16 {
	t, ok := dns.StringToType[strings.ToUpper(stringType)]
	if !ok {
		log.Fatalf("unknown query type: \"%v\"", stringType)
	}
	return t
}

func parseDnsResponse(data []byte) (*dns.Msg, error) {
	msg := &dns.Msg{}
	err := msg.Unpack(data)
	return msg, err
}

func createOdohQuestion(dnsMessage []byte, publicKey odoh.ObliviousDoHConfigContents) (odoh.ObliviousDNSMessage, odoh.QueryContext, error) {
	odohQuery := odoh.CreateObliviousDNSQuery(dnsMessage, 0)
	odnsMessage, queryContext, err := publicKey.EncryptQuery(odohQuery)
	if err != nil {
		return odoh.ObliviousDNSMessage{}, odoh.QueryContext{}, err
	}

	return odnsMessage, queryContext, nil
}

func buildURL(s, defaultPath string) *url.URL {
	if !strings.HasPrefix(s, "https://") && !strings.HasPrefix(s, "http://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		log.Fatalf("failed to parse url: %v", err)
	}
	if u.Path == "" {
		u.Path = defaultPath
	}
	return u
}

func buildDohURL(s string) *url.URL {
	return buildURL(s, DOH_DEFAULT_PATH)
}

func buildOdohTargetURL(s string) *url.URL {
	return buildURL(s, ODOH_DEFAULT_PATH)
}

func buildOdohProxyURL(proxy, target string) *url.URL {
	p := buildURL(proxy, ODOH_PROXY_DEFAULT_PATH)
	t := buildOdohTargetURL(target)
	qry := p.Query()
	if qry.Get("targethost") == "" {
		qry.Set("targethost", t.Host)
	}
	if qry.Get("targetpath") == "" {
		qry.Set("targetpath", t.Path)
	}
	p.RawQuery = qry.Encode()
	return p
}

func buildOdohConfigURL(s string) *url.URL {
	return buildURL(s, ODOH_CONFIG_WELLKNOWN_PATH)
}
