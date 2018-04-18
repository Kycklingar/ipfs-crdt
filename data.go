package main

import (
	"fmt"
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

type identifier struct {
	m map[string]func() crdtData
	l sync.RWMutex
}

var id identifier

// Register a CRDTData creation function
func RegisterCRDTData(name string, fn func() crdtData) error {
	if id.m == nil {
		id.m = make(map[string]func() crdtData)
	}
	id.l.Lock()
	defer id.l.Unlock()
	if _, ok := id.m[name]; ok {
		return fmt.Errorf("Function already registered: %s", name)
	}
	id.m[name] = fn
	return nil
}

// Identify the crdtData type and return a new
func Identify(a string) crdtData {
	id.l.RLock()
	defer id.l.RUnlock()
	if fn, ok := id.m[a]; ok {
		return fn()
	}
	return nil
}

// type postData struct {
// 	Hash string
// 	Size int
// }

// type tagData struct {
// 	PostHash string
// 	Tag      string
// }

type crdtData interface {
	string() string
	set(...interface{}) error
	same(crdtData) bool
	id() string
	smash(crdtData) crdtData
}

// func (c *postData) string() string {
// 	return fmt.Sprintf("{POST[%s,%d]}", c.Hash, c.Size)
// }

// func (c *postData) set(vars ...interface{}) error {
// 	if len(vars) < 2 || len(vars) > 2 {
// 		return errors.New(fmt.Sprintf("invalid argument: ", vars...))
// 	}

// 	//TODO: Verify the content

// 	// if _, ok := vars[1].(string); ok {
// 	// 	var err error
// 	// 	c.Size, err = strconv.Atoi(vars[1].(string))
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// } else {
// 	// 	c.Size = vars[1].(int)
// 	// }

// 	hash, ok := vars[0].(string)
// 	if !ok || len(hash) < 46 || len(hash) > 49 {
// 		return errors.New("incorrect hash")
// 	}
// 	c.Hash = hash

// 	size, ok := vars[1].(int)
// 	if !ok || size <= 0 {
// 		return errors.New("size not defined")
// 	}
// 	c.Size = size
// 	return nil
// }

// func (c *postData) same(a crdtData) bool {
// 	if _, ok := a.(*postData); !ok {
// 		return false
// 	}
// 	return c.Hash == a.(*postData).Hash && c.Size == a.(*postData).Size
// }

// func (c *tagData) string() string {
// 	return fmt.Sprintf("{TAG[%s,%s]}", c.PostHash, c.Tag)
// }

// func (c *tagData) set(vars ...interface{}) error {
// 	if len(vars) < 2 || len(vars) > 2 {
// 		return errors.New("invalid argument")
// 	}

// 	//TODO: Verify the content
// 	c.PostHash = strings.TrimSpace(vars[0].(string))
// 	c.Tag = strings.TrimSpace(vars[1].(string))
// 	return nil
// }

// func (c *tagData) same(a crdtData) bool {
// 	if _, ok := a.(*tagData); !ok {
// 		return false
// 	}
// 	return c.PostHash == a.(*tagData).PostHash && c.Tag == a.(*tagData).Tag
// }

func createData() *data {
	var d = &data{make([]crdtData, 0), sync.Mutex{}}
	return d
}

func (d *data) add(p crdtData) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// fmt.Println("Add", p.string())
	//fmt.Println(d.query(p), p.id())
	if d.query(p) {
		//fmt.Println("Smash")
		d.smash(p)
		return
	}

	d.data = append(d.data, p)
}

func (d *data) smash(a crdtData) {
	for i, _ := range d.data {
		//fmt.Println("!smashing", d.data[i].string(), a.string())
		if d.data[i].same(a) {
			d.data[i] = d.data[i].smash(a)
		}
	}
}

func (d *data) query(a crdtData) bool {
	if d.data != nil {
		for _, data := range d.data {
			if data.same(a) {
				//fmt.Println(data.id(), a.id())
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
