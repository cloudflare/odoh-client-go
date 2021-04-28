package commands

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	odoh "github.com/cloudflare/odoh-go"
	"github.com/miekg/dns"
	"github.com/urfave/cli"
)

// This runningTime structure contains the epoch timestamps for each of the operations
// taking place. The explanations are as follows:
// 1. Start => Epoch time at which the client starts to prepare the question
// 2. ClientQueryEncryptionTime => Epoch time at which the client completes the encryption and serialization of the question.
// 3. ClientUpstreamRequestTime => Epoch time indicating the start of the network request.
// 4. ClientDownstreamResponseTime => Epoch time indicating the receipt of the response and deserialization into ObliviousDNSMessage
// 5. EndTime => Epoch time indicating the end of all tasks for the request.
// NOTE: All timestamps are stored in NanoSecond granularity and need to be converted into microseconds (/1000.0) or milliseconds (/1000.0^2)
type runningTime struct {
	Start                        int64
	ClientQueryEncryptionTime    int64
	ClientUpstreamRequestTime    int64
	ClientDownstreamResponseTime int64
	ClientAnswerDecryptionTime   int64
	EndTime                      int64
}

type experiment struct {
	ExperimentID    string
	Hostname        string
	DNSType         uint16
	TargetPublicKey odoh.ObliviousDoHConfigContents
	// Instrumentation
	Proxy  string
	Target string
	// Timing parameters
	IngestedFrom string
}

type experimentResult struct {
	Hostname        string                          `json:"Hostname"`
	DNSType         uint16                          `json:"DNSType"`
	TargetPublicKey odoh.ObliviousDoHConfigContents `json:"ODoHConfigContents"`
	// Timing parameters
	STime time.Time `json:"StartTime"`
	ETime time.Time `json:"EndTime"`
	// Instrumentation
	RequestID   string      `json:"RequestID"`
	DNSQuestion []byte      `json:"Question"`
	DNSAnswer   []byte      `json:"Answer"`
	Proxy       string      `json:"Proxy"`
	Target      string      `json:"Target"`
	Timestamp   runningTime `json:"Timestamp"`
	// experiment status
	Status       bool   `json:"Status"`
	IngestedFrom string `json:"IngestedFrom"`
	ProtocolType string `json:"ProtocolType"`
	ExperimentID string `json:"ExperimentID"`
}

type DiscoveryServiceResponse struct {
	Proxies []string `json:"proxies"`
	Targets []string `json:"targets"`
}

func (e *experiment) run(client *http.Client, channel chan experimentResult) {
	hostname := e.Hostname
	dnsType := e.DNSType
	targetPublicKey := e.TargetPublicKey
	proxy := e.Proxy
	target := e.Target
	expId := e.ExperimentID

	shouldUseProxy := false

	if proxy != "" {
		shouldUseProxy = true
	}

	rt := runningTime{}

	start := time.Now()
	rt.Start = start.UnixNano()

	dnsQuery := new(dns.Msg)
	dnsQuery.SetQuestion(hostname, dnsType)
	packedDnsQuery, err := dnsQuery.Pack()
	if err != nil {
		log.Fatalf("dns.Pack() failed: %v", err)
	}

	odohQuery, queryContext, err := createOdohQuestion(packedDnsQuery, targetPublicKey)
	if err != nil {
		log.Fatalf("createOdohQuestion failed: %v", err)
	}

	timeToPrepareQuestionAndSerialize := time.Now().UnixNano()
	rt.ClientQueryEncryptionTime = timeToPrepareQuestionAndSerialize
	if err != nil {
		log.Fatalf("Error while preparing OdohQuestion: %v", err)
	}
	requestTime := time.Now().UnixNano()
	rt.ClientUpstreamRequestTime = requestTime
	odohMessage, err := resolveObliviousQuery(odohQuery, shouldUseProxy, target, proxy, client)

	responseTime := time.Now().UnixNano()
	rt.ClientDownstreamResponseTime = responseTime

	if err != nil {
		exp := experimentResult{
			Hostname:        hostname,
			DNSType:         dnsType,
			TargetPublicKey: targetPublicKey,
			Target:          target,
			Proxy:           proxy,
			STime:           start,
			ETime:           time.Now(),
			DNSAnswer:       []byte(err.Error()),
			Status:          false,
			Timestamp:       rt,
			IngestedFrom:    e.IngestedFrom,
			ProtocolType:    "ODOH",
			ExperimentID:    expId,
		}
		channel <- exp
		return
	}

	log.Printf("[DNSANSWER] %v \n", odohMessage)
	dnsAnswer, err := validateEncryptedResponse(odohMessage, queryContext)
	validationTime := time.Now().UnixNano()
	rt.ClientAnswerDecryptionTime = validationTime
	if err != nil || dnsAnswer == nil {
		exp := experimentResult{
			Hostname:        hostname,
			DNSType:         dnsType,
			TargetPublicKey: targetPublicKey,
			Target:          target,
			Proxy:           proxy,
			STime:           start,
			ETime:           time.Now(),
			DNSAnswer:       []byte("dnsAnswer incorrectly and unable to Pack"),
			Status:          false,
			Timestamp:       rt,
			IngestedFrom:    e.IngestedFrom,
			ProtocolType:    "ODOH",
			ExperimentID:    expId,
		}
		channel <- exp
		return
	}
	dnsAnswerBytes, err := dnsAnswer.Pack()
	endTime := time.Now().UnixNano()
	rt.EndTime = endTime

	requestId := make([]byte, 2)
	binary.BigEndian.PutUint16(requestId, uint16(dnsQuery.Id))

	log.Printf("=======ODOH Request for [%v]========\n", hostname)
	log.Printf("Request ID : [%x]\n", requestId)
	log.Printf("Start Time : [%v]\n", start.UnixNano())
	log.Printf("Time @ Prepare Question and Serialize : [%v]\n", timeToPrepareQuestionAndSerialize)
	log.Printf("Time @ Starting ODOH Request  : [%v]\n", requestTime)
	log.Printf("Time @ Received ODOH Response : [%v]\n", responseTime)
	log.Printf("Time @ Finished Validation Response : [%v]\n", validationTime)
	log.Printf("DNS Answer : [%v]\n", dnsAnswerBytes)
	log.Printf("====================================")
	requestIDString := hex.EncodeToString(requestId)
	log.Printf("Requested ID : [%s]", requestIDString)
	exp := experimentResult{
		Hostname:        hostname,
		DNSType:         dnsType,
		TargetPublicKey: targetPublicKey,
		// Overall timing parameters
		STime: start,
		ETime: time.Now(),
		// Instrumentation
		RequestID:   requestIDString,
		DNSQuestion: odohMessage.Marshal(),
		DNSAnswer:   dnsAnswerBytes,
		Proxy:       proxy,
		Target:      target,
		Timestamp:   rt,
		// Experiment status
		Status:       true,
		IngestedFrom: e.IngestedFrom,
		ProtocolType: "ODoH",
		ExperimentID: expId,
	}
	channel <- exp
}

func responseHandler(numberOfChannels int, responseChannel chan experimentResult) []experimentResult {
	responses := make([]experimentResult, 0)
	for index := 0; index < numberOfChannels; index++ {
		answerStructure := <-responseChannel
		answer := answerStructure.DNSAnswer
		sTime := answerStructure.STime
		eTime := answerStructure.ETime
		hostname := answerStructure.Hostname
		target := answerStructure.Target
		proxy := answerStructure.Proxy
		log.Printf("Response %v\n", index)
		log.Printf("Size of the Response for [%v] is [%v] and [%v] to [%v] = [%v] using Proxy [%v] using Target [%v]",
			hostname, len(answer), sTime.UnixNano(), eTime.UnixNano(), eTime.Sub(sTime).Microseconds(), proxy, target)
		responses = append(responses, answerStructure)
	}
	return responses
}

func getTickTriggerTiming(requestsPerMinute int) float64 {
	intervalDuration := time.Minute.Seconds() / float64(requestsPerMinute)
	return intervalDuration
}

// The benchmarkClient creates `--numclients` client instances performing `--pick`
// queries over `--rate` requests/minute uniformly distributed.
func benchmarkClient(c *cli.Context) {
	var clientInstanceName string
	if clientInstanceEnvironmentName := os.Getenv("CLIENT_INSTANCE_NAME"); clientInstanceEnvironmentName != "" {
		clientInstanceName = clientInstanceEnvironmentName
	} else {
		clientInstanceName = "client_localhost_instance"
	}

	var experimentID string
	if experimentID := os.Getenv("EXPERIMENT_ID"); experimentID == "" {
		experimentID = "EXP_LOCAL"
	}

	logFilePath := c.String("logout")
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Unable to create a log file to log data into.")
	}
	defer f.Close()
	log.SetOutput(f)

	outputFilePath := c.String("out")
	filepath := c.String("data")
	filterCount := c.Uint64("pick")
	numberOfParallelClients := c.Uint64("numclients")
	requestPerMinute := c.Uint64("rate") // requests/minute
	target := c.String("target")
	if len(target) < 0 {
		log.Fatal("Missing target parameter")
	}
	proxy := c.String("proxy")
	if len(proxy) < 0 {
		log.Fatal("Missing target parameter")
	}
	dnsTypeString := c.String("dnstype")
	dnsMessageType := dnsQueryStringToType(dnsTypeString)

	tickTrigger := getTickTriggerTiming(int(requestPerMinute))
	hostnames, err := readDomainsFromFile(filepath, filterCount)
	if err != nil {
		log.Printf("Failed to read the file correctly. %v", err)
	}
	if len(hostnames) < int(filterCount) {
		filterCount = uint64(len(hostnames))
	}
	totalResponsesNeeded := numberOfParallelClients * filterCount

	state := GetInstance(numberOfParallelClients)

	configs, err := fetchTargetConfigs(target)
	if err != nil {
		fmt.Println("failed configs")
		log.Fatalf("Unable to obtain the ObliviousDoHConfigs from %v. Error %v", target, err)
	}
	if len(configs.Configs) == 0 {
		log.Fatalf("Empty configs returned for the target %v", target)
	}

	config := configs.Configs[0]
	state.InsertKey(target, config.Contents)

	responseChannel := make(chan experimentResult, totalResponsesNeeded)
	totalQueries := len(hostnames)
	requestPerMinuteTick := time.NewTicker(time.Duration(tickTrigger) * time.Second)

	for range requestPerMinuteTick.C {
		startIndex := totalQueries - 1
		endIndex := startIndex - int(requestPerMinute)
		if endIndex < 0 {
			endIndex = 0
		}
		for index := startIndex; index >= endIndex; index-- {
			for clientIndex := 0; clientIndex < int(numberOfParallelClients); clientIndex++ {
				hostname := hostnames[index]
				clientUsed := state.client[clientIndex]
				targetConfigContents, err := state.GetTargetConfigContents(target)
				if err != nil {
					log.Fatalf("Unable to retrieve the PK requested")
				}
				e := experiment{
					ExperimentID:    experimentID,
					Hostname:        hostname,
					DNSType:         dnsMessageType,
					TargetPublicKey: targetConfigContents,
					Target:          target,
					Proxy:           proxy,
					IngestedFrom:    clientInstanceName,
				}

				go e.run(clientUsed, responseChannel)
			}
			totalQueries--
		}
		if totalQueries <= 0 {
			requestPerMinuteTick.Stop()
			break
		}
	}
	responses := responseHandler(int(totalResponsesNeeded), responseChannel)
	close(responseChannel)

	encResponses, err := json.Marshal(responses)
	if err != nil {
		log.Fatal("Failed to encode results:", err)
	}

	if outputFilePath != "" {
		if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
			err = ioutil.WriteFile(outputFilePath, encResponses, 0644)
			if err != nil {
				log.Printf("Failed writing results to file: %v\nPrinting the results instead\n", err)
				fmt.Println(string(encResponses))
			} else {
				log.Printf("File saved at: %v\n", outputFilePath)
			}
		}
	} else {
		fmt.Println(string(encResponses))
	}

}
