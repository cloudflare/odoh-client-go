package commands

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	odoh "github.com/cloudflare/odoh-go"
	"github.com/miekg/dns"
	"github.com/urfave/cli"
)

func createPlainQueryResponse(hostname string, serializedDnsQueryString []byte) (response *dns.Msg, err error) {
	client := http.Client{}
	queryUrl := fmt.Sprintf(TARGET_HTTP_MODE+"://%s/dns-query", hostname)
	req, err := http.NewRequest(http.MethodGet, queryUrl, nil)
	if err != nil {
		log.Fatalln(err)
	}

	queries := req.URL.Query()
	encodedString := base64.RawURLEncoding.EncodeToString(serializedDnsQueryString)
	queries.Add("dns", encodedString)
	req.Header.Set("Content-Type", DOH_CONTENT_TYPE)
	req.URL.RawQuery = queries.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	dnsBytes, err := parseDnsResponse(bodyBytes)

	return dnsBytes, nil
}

func prepareHttpRequest(serializedBody []byte, useProxy bool, targetIP string, proxy string) (req *http.Request, err error) {
	var baseurl string
	var queries url.Values

	if useProxy != true {
		baseurl = fmt.Sprintf("%s://%s/%s", TARGET_HTTP_MODE, targetIP, "dns-query")
		req, err = http.NewRequest(http.MethodPost, baseurl, bytes.NewBuffer(serializedBody))
		queries = req.URL.Query()
	} else {
		baseurl = fmt.Sprintf("%s://%s/%s", PROXY_HTTP_MODE, proxy, "proxy")
		req, err = http.NewRequest(http.MethodPost, baseurl, bytes.NewBuffer(serializedBody))
		queries = req.URL.Query()
		queries.Add("targethost", targetIP)
		queries.Add("targetpath", "/dns-query")
	}

	req.Header.Set("Content-Type", OBLIVIOUS_DOH_CONTENT_TYPE)
	req.Header.Set("Accept", OBLIVIOUS_DOH_CONTENT_TYPE)
	req.URL.RawQuery = queries.Encode()

	return req, err
}

func resolveObliviousQuery(query odoh.ObliviousDNSMessage, useProxy bool, targetIP string, proxy string, client *http.Client) (response odoh.ObliviousDNSMessage, err error) {
	serializedQuery := query.Marshal()
	req, err := prepareHttpRequest(serializedQuery, useProxy, targetIP, proxy)
	if err != nil {
		return odoh.ObliviousDNSMessage{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return odoh.ObliviousDNSMessage{}, err
	}

	responseHeader := resp.Header.Get("Content-Type")
	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return odoh.ObliviousDNSMessage{}, err
	}
	if responseHeader != OBLIVIOUS_DOH_CONTENT_TYPE {
		return odoh.ObliviousDNSMessage{}, fmt.Errorf("Did not obtain the correct headers from %v with response %v", targetIP, string(bodyBytes))
	}

	odohQueryResponse, err := odoh.UnmarshalDNSMessage(bodyBytes)
	if err != nil {
		return odoh.ObliviousDNSMessage{}, err
	}

	return odohQueryResponse, nil
}

func plainDnsRequest(c *cli.Context) error {
	domainName := dns.Fqdn(c.String("domain"))
	dnsTypeString := c.String("dnstype")
	dnsTargetServer := c.String("target")
	dnsType := dnsQueryStringToType(dnsTypeString)

	dnsQuery := new(dns.Msg)
	dnsQuery.SetQuestion(domainName, dnsType)
	packedDnsQuery, err := dnsQuery.Pack()
	if err != nil {
		return err
	}

	response, err := createPlainQueryResponse(dnsTargetServer, packedDnsQuery)
	if err != nil {
		return err
	}

	fmt.Println(response)
	return nil
}

func obliviousDnsRequest(c *cli.Context) error {
	domainName := dns.Fqdn(c.String("domain"))
	dnsTypeString := c.String("dnstype")
	targetName := c.String("target")
	proxy := c.String("proxy")

	var useproxy bool
	if len(proxy) > 0 {
		useproxy = true
	}

	odohConfigs, err := fetchTargetConfigs(targetName)
	if err != nil {
		return err
	}
	if len(odohConfigs.Configs) == 0 {
		err := errors.New("target provided no valid odoh configs")
		fmt.Println(err)
		return err
	}
	odohConfig := odohConfigs.Configs[0]

	dnsType := dnsQueryStringToType(dnsTypeString)

	dnsQuery := new(dns.Msg)
	dnsQuery.SetQuestion(domainName, dnsType)
	packedDnsQuery, err := dnsQuery.Pack()
	if err != nil {
		fmt.Println(err)
		return err
	}

	odohQuery, queryContext, err := createOdohQuestion(packedDnsQuery, odohConfig.Contents)
	if err != nil {
		fmt.Println(err)
		return err
	}

	client := http.Client{}
	odohMessage, err := resolveObliviousQuery(odohQuery, useproxy, targetName, proxy, &client)
	if err != nil {
		fmt.Println(err)
		return err
	}

	dnsResponse, err := validateEncryptedResponse(odohMessage, queryContext)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(dnsResponse)
	return nil
}

func validateEncryptedResponse(message odoh.ObliviousDNSMessage, queryContext odoh.QueryContext) (response *dns.Msg, err error) {
	decryptedResponse, err := queryContext.OpenAnswer(message)
	if err != nil {
		return nil, err
	}

	dnsBytes, err := parseDnsResponse(decryptedResponse)
	if err != nil {
		return nil, err
	}

	return dnsBytes, nil
}
