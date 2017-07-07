package main

import "sync"

type crdt interface {
	add(interface{})
	query(interface{}) bool
	compare(crdt) bool
	merge(crdt) crdt
}

type data struct {
	multihash []string
	mutex     sync.Mutex
}

func createData() *data {
	var d = &data{make([]string, 0), sync.Mutex{}}
	return d
}

func (d *data) add(p interface{}) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.query(p) {
		return
	}

	d.multihash = append(d.multihash, p.(string))
}

func (d *data) query(a interface{}) bool {
	if d.multihash != nil {
		for _, str := range d.multihash {
			if str == a {
				return true
			}
		}
	}
	return false
}

func (d *data) compare(ai crdt) bool {
	a := ai.(*data)
	if len(d.multihash) != len(a.multihash) {
		return false
	}
	for _, dx := range d.multihash {
		if !a.query(dx) {
			return false
		}
	}
	return true
}

func (d *data) merge(ai crdt) crdt {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	a := ai.(*data)
	for _, ax := range a.multihash {
		if !d.query(ax) {
			d.multihash = append(d.multihash, ax)
		}
	}

	return d
}
