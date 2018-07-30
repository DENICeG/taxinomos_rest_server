package domains

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

type DomainDummy struct {
	Data struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			DomainID   int       `json:"domain-id"`
			DomainName string    `json:"domain-name"`
			ULabel     string    `json:"u-label"`
			CreateDate time.Time `json:"create-date"`
		} `json:"attributes"`
		Relationships struct{} `json:"relationships"`
		Links         struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"data"`
}

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

func LoadDomainsFromTxtFile(filename string) ([]*DomainDummy, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	i := 1
	var domainlist []*DomainDummy
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		id := strconv.Itoa(i)
		domain := new(DomainDummy)
		domain.Data.ID = id
		domain.Data.Type = "domains"
		domain.Data.Attributes.DomainID = i
		domain.Data.Attributes.DomainName = scanner.Text()
		domain.Data.Attributes.ULabel = scanner.Text()
		domain.Data.Links.Self = fmt.Sprintf("https://classify-rest.labs.nic.at/api/v1/domains/%s", id)
		domainlist = append(domainlist, domain)
		i++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return domainlist, nil
}
