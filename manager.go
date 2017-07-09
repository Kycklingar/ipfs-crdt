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
			m.publish(false)
			continue
		}
		s := m.ipfs.Cat(p.Data)
		if s == "" {
			continue
		}

		var cData data

		// Data will look like this "{POST[Qm...,100]}{TAG[Qm...,tag]}"
		spl := strings.Split(s, "{")
		for _, left := range spl {
			var d crdtData
			if len(left) <= 0 || strings.Index(left, "}") == -1 {
				continue
			}
			// dat := strings.Split(data[strings.Index(data, "[")+1:], ",")
			data := strings.Split(left[:strings.Index(left, "}")-1], "[")
			switch dat := strings.Split(data[1], ","); data[0] {
			case "POST":
				var p postData
				err := p.set(dat[0], dat[1])
				if err != nil {
					fmt.Println(err)
					continue
				}
				d = &p
			case "TAG":
				if len(dat) < 2 || len(dat) > 2 {
					continue
				}
				var t tagData
				err := t.set(dat[0], dat[1])
				if err != nil {
					fmt.Println(err)
					continue
				}
				d = &t
			}
			cData.data = append(cData.data, d)
		}

		m.d.merge(&cData)
		m.publish(true)
	}
}

func (m *idManager) add(a ...crdtData) {
	for _, data := range a {
		m.d.add(data)
	}
	m.publish(false)
}

func (m *idManager) publish(checkNew bool) {
	var data string
	for _, d := range m.d.data {
		if d == nil {
			continue
		}
		data += d.string()
	}
	hash := m.ipfs.CreateObject(data)

	if checkNew && hash == m.currentHash {
		return
	}

	m.currentHash = hash
	m.ipfs.Publish(hash)
}

func (m *idManager) ask() bool {
	return m.ipfs.Publish("ASK")
}
