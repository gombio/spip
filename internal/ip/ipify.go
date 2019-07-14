package ip

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Response struct {
	IP string `json:"ip"`
}

type Ipify struct {
}

func (s Ipify) GetIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var r Response
	err = json.Unmarshal(body, &r)
	if err != nil {
		return "", err
	}

	return r.IP, nil
}
