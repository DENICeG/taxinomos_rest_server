package domains

import (
	"encoding/json"
	"io/ioutil"
)

func LoadDomainsFromFile(filename string, domains interface{}) error {
	domainbytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(domainbytes, domains)
	if err != nil {
		return err
	}
	return nil
}
