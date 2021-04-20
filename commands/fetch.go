package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"

	odoh "github.com/cloudflare/odoh-go"
	"github.com/urfave/cli"
)

func fetchTargetConfigsFromWellKnown(url string) (odoh.ObliviousDoHConfigs, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return odoh.ObliviousDoHConfigs{}, err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return odoh.ObliviousDoHConfigs{}, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return odoh.ObliviousDoHConfigs{}, err
	}

	return odoh.UnmarshalObliviousDoHConfigs(bodyBytes)
}

func fetchTargetConfigs(targetName string) (odoh.ObliviousDoHConfigs, error) {
	return fetchTargetConfigsFromWellKnown(buildOdohConfigURL(targetName).String())
}

func getTargetConfigs(c *cli.Context) error {
	targetName := c.String("target")
	pretty := c.Bool("pretty")

	odohConfigs, err := fetchTargetConfigs(targetName)
	if err != nil {
		return err
	}

	if pretty {
		fmt.Println("ObliviousDoHConfigs:")
		for i, config := range odohConfigs.Configs {
			configContents := config.Contents
			fmt.Printf("  Config %d: Version(0x%04x), KEM(0x%04x), KDF(0x%04x), AEAD(0x%04x) KeyID(%x)\n", (i + 1), config.Version, configContents.KemID, configContents.KdfID, configContents.AeadID, configContents.KeyID())
		}
	} else {
		fmt.Printf("%x", odohConfigs.Marshal())
	}
	return nil
}
