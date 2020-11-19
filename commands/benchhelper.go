package commands

import (
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"
)

func readLines(path string) (lines []string, err error) {
	bytesRead, _ := ioutil.ReadFile(path)
	fileContent := string(bytesRead)
	records := strings.Split(fileContent, "\n")
	log.Printf("Read into memory %v total number of hostnames.\n", len(records))
	return records, nil
}

func shuffleAndSlice(records []string, slice uint64) (lines []string) {
	rand.Seed(time.Now().UnixNano())
	start := time.Now()
	rand.Shuffle(len(records), func(i, j int) { records[i], records[j] = records[j], records[i] })
	log.Printf("Time (ms) to shuffle %v records : [%v]", len(records), time.Since(start).Milliseconds())
	chosen_records := records[0:slice]
	// Append a '.' to the end of the message for it to be a valid DNS Question about the Hostname
	for index, record := range chosen_records {
		chosen_records[index] = record + "."
	}
	return chosen_records
}
