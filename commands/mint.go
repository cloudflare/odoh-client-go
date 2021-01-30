package commands

import (
	"crypto/rand"
	"encoding/pem"
	"log"
	"os"
	"strconv"

	hpke "github.com/cisco/go-hpke"
	odoh "github.com/cloudflare/odoh-go"
	"github.com/urfave/cli"
)

func createConfigurations(c *cli.Context) error {
	kemID, err := strconv.ParseUint(c.String("kemid"), 10, 16)
	if err != nil {
		return err
	}
	kdfID, err := strconv.ParseUint(c.String("kdfid"), 10, 16)
	if err != nil {
		return err
	}
	aeadID, err := strconv.ParseUint(c.String("aeadid"), 10, 16)
	if err != nil {
		return err
	}

	suite, err := hpke.AssembleCipherSuite(hpke.KEMID(kemID), hpke.KDFID(kdfID), hpke.AEADID(aeadID))
	if err != nil {
		return err
	}

	ikm := make([]byte, suite.KEM.PrivateKeySize())
	rand.Reader.Read(ikm)
	privateKey, publicKey, err := suite.KEM.DeriveKeyPair(ikm)
	if err != nil {
		return err
	}

	configContents, err := odoh.CreateObliviousDoHConfigContents(hpke.KEMID(kemID), hpke.KDFID(kdfID), hpke.AEADID(aeadID), suite.KEM.SerializePublicKey(publicKey))
	if err != nil {
		return err
	}

	config := odoh.CreateObliviousDoHConfig(configContents)

	configsBlock := &pem.Block{
		Type:  "ODOH CONFIGS",
		Bytes: config.Marshal(),
	}
	if err := pem.Encode(os.Stdout, configsBlock); err != nil {
		log.Fatal(err)
	}

	privateConfigsBlock := &pem.Block{
		Type:  "ODOH PRIVATE KEY",
		Bytes: suite.KEM.SerializePrivateKey(privateKey),
	}
	if err := pem.Encode(os.Stdout, privateConfigsBlock); err != nil {
		log.Fatal(err)
	}

	return nil
}
