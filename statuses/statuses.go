package statuses

import (
	"encoding/json"
	"io/ioutil"
)

/*
{
	"meta": {
		"page": ["total":5]
	}
    "data": {
        "type": "statuses",
        "id": "1",
        "attributes": {
            "measurement-status-id": 1,
            "status": "success"
        },
        "relationships": {
            "measurements": {
                "meta": {
                    "total": 1
                },
                "links": {
                    "self": "https://classify-rest.labs.nic.at/api/v1/statuses/1/relationships/measurements",
                    "related": "https://classify-rest.labs.nic.at/api/v1/statuses/1/measurements"
                }
            }
        },
        "links": {
            "self": "https://classify-rest.labs.nic.at/api/v1/statuses/1"
        }
    }
}
*/

type Status struct {
	Data StatusData `json:"data"`
}

type StatusData struct {
	Type          string              `json:"type"`
	Id            string              `json:"id"`
	Attributes    StatusAttributes    `json:"attributes"`
	Relationships StatusRelationships `json:"relationships"`
	Links         StatusLinks         `json:"links"`
}

type StatusAttributes struct {
	MeasurementStatusId int    `json:"measurement-status-id"`
	Status              string `json:"status"`
}

type StatusRelationships struct {
	Measurements struct {
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
		Links struct {
			Self    string `json:"self"`
			Related string `json:"related"`
		} `json:"links"`
	} `json:"measurements"`
}

type StatusLinks struct {
	Self string `json:"self"`
}

func LoadStatusesFromFile(filename string, statuses interface{}) error {
	statusesbytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(statusesbytes, statuses)
	if err != nil {
		return err
	}
	return nil
}
