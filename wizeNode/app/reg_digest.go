package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func regDigest() {
	url := "http://" + os.Getenv("DIGEST_NODE") + ":8888/hello/blockchain"
	values := map[string]string{
		"Address":   os.Getenv("USER_ADDRESS"),
		"PrivKey":   os.Getenv("USER_PRIVKEY"),
		"Pubkey":    os.Getenv("USER_PUBKEY"),
		"AES":       os.Getenv("PASSWORD"),
		"Url":       "http://" + os.Getenv("PUBLIC_IP") + ":4000",
		"ServerKey": os.Getenv("SERVER_KEY"),
	}

	jsonValue, _ := json.Marshal(values)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
