package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type compactPost struct {
	Hash string
	Tags []string
}

func (c *compactPost) string() string {
	var tgstr string
	for _, tag := range c.Tags {
		tgstr += tag + ","
	}
	if len(tgstr) > 0 {
		tgstr = tgstr[:len(tgstr)-1]
	}
	return fmt.Sprintf("{CPOST[%s, %s]}", c.Hash, tgstr)
}

func (c *compactPost) set(vars ...interface{}) error {
	if len(vars) < 1 {
		return errors.New(fmt.Sprintf("Invalid argument: ", vars...))
	}

	hash, ok := vars[0].(string)
	if !ok || len(hash) < 46 || len(hash) > 49 {
		return fmt.Errorf("incorrect hash", hash)
	}
	c.Hash = hash

	if len(vars) >= 2 {
		for _, tag := range vars[1:] {
			t, ok := tag.(string)
			if !ok {
				log.Println("Invalid string: ", tag)
				continue
			}
			c.Tags = append(c.Tags, strings.TrimSpace(t))
		}
		// Sort?
	}

	return nil
}

func (c *compactPost) same(a crdtData) bool {
	if _, ok := a.(*compactPost); !ok {
		return false
	}

	// if len(a.(*compactPost).Tags) != len(c.Tags) {
	// 	return false
	// }

	// for _, ta := range a.(*compactPost).Tags {
	// 	var b bool
	// 	for _, tc := range c.Tags {
	// 		if ta == tc {
	// 			b = true
	// 			break
	// 		}
	// 	}
	// 	if !b {
	// 		return false
	// 	}
	// }

	// return c.Hash == a.(*compactPost).Hash
	return c.id() == a.(*compactPost).id()
}

func (c *compactPost) id() string {
	return c.Hash
}

func (c *compactPost) smash(a crdtData) crdtData {
	if !c.same(a) {
		return c
	}

	var appTags []string
	for _, at := range a.(*compactPost).Tags {
		f := false
		for _, ct := range c.Tags {
			if at == ct {
				f = true
				break
			}
		}
		if !f {
			appTags = append(appTags, at)
		}
	}
	c.Tags = append(c.Tags, appTags...)
	return c
}
