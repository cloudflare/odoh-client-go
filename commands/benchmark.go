package commands

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	odoh "github.com/cloudflare/odoh-go"
	"github.com/miekg/dns"
	"github.com/urfave/cli"
	"log"
	mathrand "math/rand"
	"net/http"
	"os"
	"time"
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
	DnsType         uint16
	TargetPublicKey odoh.ObliviousDoHConfigContents
	// Instrumentation
	Proxy  string
	Target string
	// Timing parameters
	IngestedFrom string
}

type experimentResult struct {
	Hostname        string
	DnsType         uint16
	TargetPublicKey odoh.ObliviousDoHConfigContents
	// Timing parameters
	STime time.Time
	ETime time.Time
	// Instrumentation
	RequestID   string
	DnsQuestion []byte
	DnsAnswer   []byte
	Proxy       string
	Target      string
	Timestamp   runningTime
	// experiment status
	Status       bool
	IngestedFrom string
	ProtocolType string
	ExperimentID string
}

func (e *experimentResult) serialize() string {
	exp := &e
	response, err := json.Marshal(exp)
	if err != nil {
		log.Printf("Unable to log the information correctly.")
	}
	return string(response)
}

type DiscoveryServiceResponse struct {
	Proxies []string `json:"proxies"`
	Targets []string `json:"targets"`
}

func (e *experiment) run(client *http.Client, channel chan experimentResult) {
	hostname := e.Hostname
	dnsType := e.DnsType
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
			DnsType:         dnsType,
			TargetPublicKey: targetPublicKey,
			Target:          target,
			Proxy:           proxy,
			STime:           start,
			ETime:           time.Now(),
			DnsAnswer:       []byte(err.Error()),
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
			DnsType:         dnsType,
			TargetPublicKey: targetPublicKey,
			Target:          target,
			Proxy:           proxy,
			STime:           start,
			ETime:           time.Now(),
			DnsAnswer:       []byte("dnsAnswer incorrectly and unable to Pack"),
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
		DnsType:         dnsType,
		TargetPublicKey: targetPublicKey,
		// Overall timing parameters
		STime: start,
		ETime: time.Now(),
		// Instrumentation
		RequestID:   requestIDString,
		DnsQuestion: odohMessage.Marshal(),
		DnsAnswer:   dnsAnswerBytes,
		Proxy:       proxy,
		Target:      target,
		Timestamp:   rt,
		// Experiment status
		Status:       true,
		IngestedFrom: e.IngestedFrom,
		ProtocolType: "ODOH",
		ExperimentID: expId,
	}
	log.Printf("experiment : %v", exp.serialize())
	channel <- exp
}

func responseHandler(numberOfChannels int, responseChannel chan experimentResult) []string {
	responses := make([]string, 0)
	for index := 0; index < numberOfChannels; index++ {
		answerStructure := <-responseChannel
		answer := answerStructure.DnsAnswer
		sTime := answerStructure.STime
		eTime := answerStructure.ETime
		hostname := answerStructure.Hostname
		target := answerStructure.Target
		proxy := answerStructure.Proxy
		log.Printf("Response %v\n", index)
		log.Printf("Size of the Response for [%v] is [%v] and [%v] to [%v] = [%v] using Proxy [%v] using Target [%v]",
			hostname, len(answer), sTime.UnixNano(), eTime.UnixNano(), eTime.Sub(sTime).Microseconds(), proxy, target)
		responses = append(responses, answerStructure.serialize())
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

	outputFilePath := c.String("out")
	f, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Unable to create a log file to log data into.")
	}
	defer f.Close()
	log.SetOutput(f)

	// The Preparation Phase of the request.
	filepath := c.String("data")
	filterCount := c.Uint64("pick")
	numberOfParallelClients := c.Uint64("numclients")
	requestPerMinute := c.Uint64("rate") // requests/minute
	discoveryServiceHostname := c.String("discovery")
	tickTrigger := getTickTriggerTiming(int(requestPerMinute))

	totalResponsesNeeded := numberOfParallelClients * filterCount

	allDomains, err := readLines(filepath)

	if err != nil {
		log.Printf("Failed to read the file correctly. %v", err)
	}

	hostnames := shuffleAndSlice(allDomains, filterCount)
	log.Printf("Now operating on a total size of : [%v] hostnames", len(hostnames))

	// Create a base state of the experiment
	state := GetInstance(numberOfParallelClients)
	telemetryState := getTelemetryInstance()
	//telemetryResponse := telemetryState.getClusterInformation()
	//log.Printf("Server: %s", telemetryResponse["version"].(map[string]interface{})["number"])

	// Create network requests concurrently.
	const dnsMessageType = dns.TypeA

	availableServices, err := fetchProxiesAndTargets(discoveryServiceHostname, instance.client[0])
	if err != nil {
		log.Fatalf("Unable to discover the services available.")
	}

	// Obtain all the keys for the targets.
	targets := availableServices.Targets
	proxies := availableServices.Proxies
	for _, target := range targets {
		configs, err := fetchTargetConfigs(target)
		if err != nil {
			log.Fatalf("Unable to obtain the ObliviousDoHConfigs from %v. Error %v", target, err)
		}
		config := configs.Configs[0]
		state.InsertKey(target, config.Contents)
	}

	keysAvailable := state.TotalNumberOfTargets()
	log.Printf("%v targets available to choose from.", keysAvailable)
	log.Printf("%v proxies available to choose from.", len(proxies))

	start := time.Now()
	responseChannel := make(chan experimentResult, totalResponsesNeeded)

	totalQueries := len(hostnames)
	log.Printf("Tick Trigger : %v %v", tickTrigger, time.Duration(tickTrigger)*time.Minute)

	requestPerMinuteTick := time.NewTicker(time.Duration(tickTrigger) * time.Second)

	// TODO(@sudheesh): Ideally start all the clients at different durations before they enforce --rate.

	for now := range requestPerMinuteTick.C {
		log.Printf("[%v] Firing %v requests at %v\n", totalQueries, requestPerMinute, now)
		startIndex := totalQueries - 1
		endIndex := startIndex - int(requestPerMinute)
		if endIndex < 0 {
			endIndex = 0
		}
		for index := startIndex; index >= endIndex; index-- {
			for clientIndex := 0; clientIndex < int(numberOfParallelClients); clientIndex++ {
				hostname := hostnames[index]
				clientUsed := state.client[clientIndex]
				log.Printf("Choosing [Client %v] to make a query", index%int(numberOfParallelClients))
				chosenTarget := targets[mathrand.Intn(keysAvailable)]
				chosenProxy := proxies[mathrand.Intn(len(proxies))]
				targetConfigContents, err := state.GetTargetConfigContents(chosenTarget)
				if err != nil {
					log.Fatalf("Unable to retrieve the PK requested")
				}
				e := experiment{
					ExperimentID:    experimentID,
					Hostname:        hostname,
					DnsType:         dnsMessageType,
					TargetPublicKey: targetConfigContents,
					Target:          chosenTarget,
					Proxy:           chosenProxy,
					IngestedFrom:    clientInstanceName,
				}

				log.Printf("Request %v%v\n", index, clientIndex)
				go e.run(clientUsed, responseChannel)
			}
			totalQueries--
		}
		if totalQueries <= 0 {
			log.Printf("Breaking out of the request per minute loop.")
			requestPerMinuteTick.Stop()
			break
		}
	}
	log.Printf("Reached here and triggering the responseHandler.\n")
	responses := responseHandler(int(totalResponsesNeeded), responseChannel)
	close(responseChannel)

	totalResponse := time.Since(start)
	log.Printf("Time to perform [%v] workflow tasks : [%v]", len(hostnames), totalResponse.Milliseconds())

	log.Printf("Collected [%v] Responses.", len(responses))
	telemetryState.streamLogsToGCP(responses)
	//telemetryState.streamLogsToELK(responses)

	telemetryState.tearDown()
}
