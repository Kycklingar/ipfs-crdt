package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
)

type ipfs struct {
	api     string
	subject string
	client  *http.Client // TODO: Check if there is a client timeout limt
}

type psMessage struct {
	From     string   `json:"from"`
	Data     string   `json:"data"`
	Seqno    string   `json:"seqno"`
	TopicIDs []string `json:"topicIDs"`
}

func NewIpfsObject(api string) ipfs {
	return ipfs{api: api + "/api/v0/", client: &http.Client{}}
}

func (i *ipfs) Sub(ch chan psMessage, name string) {
	i.subject = name

	res, err := i.client.Get(fmt.Sprintf(i.api+"pubsub/sub?arg=%s", i.subject))
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	for {
		var ps psMessage
		err = json.NewDecoder(res.Body).Decode(&ps)
		if err != nil {
			fmt.Println(err)
		}
		encd, err := base64.StdEncoding.DecodeString(ps.Data)
		if err != nil {
			fmt.Println(err)
			continue
		}
		ps.Data = string(encd)
		ch <- ps
	}
}

func (i *ipfs) Cat(hash string) string {
	if len(hash) < 48 {
		return ""
	}
	res, err := i.client.Get(fmt.Sprintf(i.api+"cat?arg=%s", hash))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(body)
}

func (i *ipfs) Publish(hash string) bool {
	res, err := i.client.Get(fmt.Sprintf(i.api+"pubsub/pub?arg=%s&arg=%s", i.subject, hash))
	if err != nil {
		fmt.Println(err)
		return false
	}
	res.Body.Close()

	return res.StatusCode == 200
}

func (i *ipfs) CreateObject(data string) string {
	bodyBuff := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuff)

	fileWriter, err := bodyWriter.CreateFormFile("arg", "text.txt")
	if err != nil {
		return ""
	}

	r := strings.NewReader(data)

	_, err = io.Copy(fileWriter, r)
	if err != nil {
		return ""
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	// resp, err := http.Post("http://192.168.1.100:5001/api/v0/add?cid-version=1&fscache", contentType, bodyBuff)
	resp, err := i.client.Post(i.api+"add?pin=false&cid-version=1", contentType, bodyBuff)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return ""
	}

	f := make(map[string]interface{})

	json.Unmarshal(body.Bytes(), &f)

	m, ok := f["Hash"].(string)
	if !ok {
		return ""
	}

	return m
}
