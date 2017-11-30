package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type crdt interface {
	add(crdtData)
	query(crdtData) bool
	compare(crdt) bool
	merge(crdt) crdt
}

type data struct {
	data  []crdtData
	mutex sync.Mutex
}

type postData struct {
	Hash string
	Size int
}

type tagData struct {
	PostHash string
	Tag      string
}

type crdtData interface {
	string() string
	set(...interface{}) error
	same(crdtData) bool
}

func (c *postData) string() string {
	return fmt.Sprintf("{POST[%s,%d]}", c.Hash, c.Size)
}

func (c *postData) set(vars ...interface{}) error {
	if len(vars) < 2 || len(vars) > 2 {
		return errors.New(fmt.Sprintf("invalid argument: ", vars...))
	}

	//TODO: Verify the content

	// if _, ok := vars[1].(string); ok {
	// 	var err error
	// 	c.Size, err = strconv.Atoi(vars[1].(string))
	// 	if err != nil {
	// 		return err
	// 	}
	// } else {
	// 	c.Size = vars[1].(int)
	// }

	hash, ok := vars[0].(string)
	if !ok || len(hash) < 46 || len(hash) > 49 {
		return errors.New("incorrect hash")
	}
	c.Hash = hash

	size, ok := vars[1].(int)
	if !ok || size <= 0 {
		return errors.New("size not defined")
	}
	c.Size = size
	return nil
}

func (c *postData) same(a crdtData) bool {
	if _, ok := a.(*postData); !ok {
		return false
	}
	return c.Hash == a.(*postData).Hash && c.Size == a.(*postData).Size
}

func (c *tagData) string() string {
	return fmt.Sprintf("{TAG[%s,%s]}", c.PostHash, c.Tag)
}

func (c *tagData) set(vars ...interface{}) error {
	if len(vars) < 2 || len(vars) > 2 {
		return errors.New("invalid argument")
	}

	//TODO: Verify the content
	c.PostHash = strings.TrimSpace(vars[0].(string))
	c.Tag = strings.TrimSpace(vars[1].(string))
	return nil
}

func (c *tagData) same(a crdtData) bool {
	if _, ok := a.(*tagData); !ok {
		return false
	}
	return c.PostHash == a.(*tagData).PostHash && c.Tag == a.(*tagData).Tag
}

func createData() *data {
	var d = &data{make([]crdtData, 0), sync.Mutex{}}
	return d
}

func (d *data) add(p crdtData) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.query(p) {
		return
	}

	d.data = append(d.data, p)
}

func (d *data) query(a crdtData) bool {
	if d.data != nil {
		for _, data := range d.data {
			if data.same(a) {
				return true
			}
		}
	}
	return false
}

func (d *data) compare(ai crdt) bool {
	a := ai.(*data)
	if len(d.data) != len(a.data) {
		return false
	}
	for _, dx := range d.data {
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
	for _, ax := range a.data {
		if !d.query(ax) {
			d.data = append(d.data, ax)
		}
	}

	return d
}
