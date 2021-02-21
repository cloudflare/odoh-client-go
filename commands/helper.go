package commands

import (
	"log"
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
