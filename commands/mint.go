package commands

import (
	"crypto/rand"
	"encoding/pem"
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/cloudflare/circl/hpke"
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

	kem := hpke.KEM(kemID)
	if !(kem.IsValid()) {
		return errors.New("invalid kemid")
	}
	kdf := hpke.KDF(kdfID)
	if !(kdf.IsValid()) {
		return errors.New("invalid kdfid")
	}
	aead := hpke.AEAD(aeadID)
	if !(aead.IsValid()) {
		return errors.New("invalid aeadid")
	}

	suite := hpke.NewSuite(kem, kdf, aead)
	scheme := kem.Scheme()
	ikm := make([]byte, scheme.SeedSize())
	rand.Reader.Read(ikm)
	publicKey, privateKey := scheme.DeriveKeyPair(ikm)
	publicKeyBytes, err := publicKey.MarshalBinary()
	if err != nil {
		return err
	}

	configContents, err := odoh.CreateObliviousDoHConfigContents(suite, publicKeyBytes)
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

	privateKeyBytes, err := privateKey.MarshalBinary()
	if err != nil {
		return err
	}
	privateConfigsBlock := &pem.Block{
		Type:  "ODOH PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	if err := pem.Encode(os.Stdout, privateConfigsBlock); err != nil {
		log.Fatal(err)
	}

	return nil
}
