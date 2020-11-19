package commands

import (
	"errors"
	odoh "github.com/cloudflare/odoh-go"
	"net/http"
	"sync"
	"time"
)

type state struct {
	sync.RWMutex
	configContents map[string]odoh.ObliviousDoHConfigContents
	client         []*http.Client
}

var instance state

func GetInstance(N uint64) *state {
	instance.client = make([]*http.Client, N)
	for index := 0; index < int(N); index++ {
		tr := &http.Transport{
			MaxIdleConnsPerHost: 1024,
			TLSHandshakeTimeout: 0 * time.Second,
			// Uncomment the line below to explicitly disable http/2 in the clients.
			//TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		}
		instance.client[index] = &http.Client{Transport: tr}
	}
	instance.configContents = make(map[string]odoh.ObliviousDoHConfigContents)
	return &instance
}

func (s *state) InsertKey(targethost string, key odoh.ObliviousDoHConfigContents) {
	s.Lock()
	defer s.Unlock()
	s.configContents[targethost] = key
}

func (s *state) GetTargetConfigContents(targethost string) (odoh.ObliviousDoHConfigContents, error) {
	s.RLock()
	defer s.RUnlock()
	if key, ok := s.configContents[targethost]; ok {
		return key, nil
	}
	return odoh.ObliviousDoHConfigContents{}, errors.New("public key for target not available")
}

func (s *state) TotalNumberOfTargets() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.configContents)
}
