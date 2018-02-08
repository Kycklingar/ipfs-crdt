package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	DB "github.com/kycklingar/ipfs-crdt/database"
)

type idManager struct {
	d           *data
	ipfs        ipfs
	currentHash string
}

func newManager() *idManager {
	var m idManager
	var err error
	m.d = createData()
	m.ipfs, err = NewIpfsObject("http://localhost:5001")
	if err != nil {
		log.Fatal(err)
	}
	return &m
}

func (m *idManager) Init() {
	m.currentHash = DB.LatestHash(db)
	data := m.cmp(m.currentHash)
	m.d.merge(&data)
}

func (m *idManager) listen(channel string) {
	ch := make(chan psMessage)
	go m.ipfs.Sub(ch, channel)

	time.Sleep(time.Millisecond * 1000)

	for i := 0; !m.ask(); i++ {
		fmt.Println("Ask failed. Retrying")
		if i >= 9 {
			fmt.Println("Aborting. Did you forget to run the daemon with --enable-pubsub-experiment?")
			os.Exit(0)
		}
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

		//cData := m.cmp(p.Data)
		m.cmp(p.Data)

		//m.d.merge(&cData)
		m.publish(true)
	}
}

func (m *idManager) cmp(hash string) data {
	//var cData data

	s := m.ipfs.Cat(hash)
	if s == "" {
		return *m.d
	}

	// Data will look like this "{POST[Qm...,100]}{TAG[Qm...,tag]}"
	spl := strings.Split(s, "{")
	for _, left := range spl {
		var d crdtData
		if len(left) <= 0 || strings.Index(left, "}") == -1 {
			continue
		}
		// dat := strings.Split(data[strings.Index(data, "[")+1:], ",")
		data := strings.Split(left[:strings.Index(left, "}")-1], "[")
		if len(data) < 2 {
			continue
		}
		switch dat := strings.Split(data[1], ","); data[0] {
		// case "POST":
		// 	var p postData
		// 	i, err := strconv.Atoi(dat[1])
		// 	if err != nil {
		// 		log.Print(err)
		// 		continue
		// 	}
		// 	err = p.set(dat[0], i)
		// 	if err != nil {
		// 		log.Print(err)
		// 		continue
		// 	}
		// 	d = &p
		// case "TAG":
		// 	if len(dat) < 2 || len(dat) > 2 {
		// 		continue
		// 	}
		// 	var t tagData
		// 	err := t.set(dat[0], dat[1])
		// 	if err != nil {
		// 		log.Print(err)
		// 		continue
		// 	}
		// 	d = &t
		case "CPOST":
			//fmt.Println(dat)
			var p compactPost
			i := []interface{}{}
			for _, m := range dat {
				i = append(i, m)
			}
			if err := p.set(i...); err != nil {
				log.Println(err)
				continue
			}
			d = &p
		}
		//cData.data = append(cData.data, d)
		m.d.add(d)
	}
	return *m.d
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
	if hash != m.currentHash {
		go DB.NewHash(db, hash)
	}

	m.currentHash = hash
	m.ipfs.Publish(hash)
}

func (m *idManager) ask() bool {
	return m.ipfs.Publish("ASK")
}
