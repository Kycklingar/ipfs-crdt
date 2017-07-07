package main

import (
	"fmt"
	"strings"
	"time"
)

type idManager struct {
	d           *data
	ipfs        ipfs
	currentHash string
}

func newManager() *idManager {
	var m idManager
	m.d = createData()
	m.ipfs = NewIpfsObject("http://localhost:5001")

	return &m
}

func (m *idManager) listen() {
	ch := make(chan psMessage)
	go m.ipfs.Sub(ch, "test")

	time.Sleep(time.Millisecond * 1000)

	for !m.ask() {
		time.Sleep(time.Millisecond * 1000)
	}

	for p := range ch {
		fmt.Println(p.Data)
		if p.Data == m.currentHash {
			continue
		}
		if p.Data == "ASK" {
			m.publish()
			continue
		}
		s := m.ipfs.Cat(p.Data)
		if s == nil {
			continue
		}

		m.d.merge(s)
		m.publish()
	}
}

func (m *idManager) add(a interface{}) {
	m.d.add(a)
	m.publish()
}

func (m *idManager) publish() {
	var data string
	for _, str := range m.d.multihash {
		if strings.TrimSpace(str) == "" {
			continue
		}
		data += " " + strings.TrimSpace(str)
	}
	hash := m.ipfs.CreateObject(data)

	m.currentHash = hash
	m.ipfs.Publish(hash)
}

func (m *idManager) ask() bool {
	return m.ipfs.Publish("ASK")
}
