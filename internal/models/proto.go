package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/lib/pq"
)

type Proto string

func (p Proto) String() string {
	return string(p)
}

const (
	ProtoDNS   Proto = "dns"
	ProtoHTTP  Proto = "http"
	ProtoHTTPS Proto = "https"
	ProtoSMTP  Proto = "smtp"
)

type ProtoCategory string

func (p ProtoCategory) String() string {
	return string(p)
}

const (
	ProtoCategoryDNS  ProtoCategory = "dns"
	ProtoCategoryHTTP ProtoCategory = "http"
	ProtoCategorySMTP ProtoCategory = "smtp"
)

var ProtoCategoriesAll = ProtoCategoryArray{
	ProtoCategoryDNS,
	ProtoCategoryHTTP,
	ProtoCategorySMTP,
}

func ProtoToCategory(p Proto) ProtoCategory {
	switch p {
	case ProtoDNS:
		return ProtoCategoryDNS
	case ProtoHTTP, ProtoHTTPS:
		return ProtoCategoryHTTP
	case ProtoSMTP:
		return ProtoCategorySMTP
	}

	panic(fmt.Sprintf("invalid protocol: %s", p))
}

type ProtoCategoryArray []ProtoCategory

func ProtoCagories(cc ...string) ProtoCategoryArray {
	res := make([]ProtoCategory, len(cc))
	for i, c := range cc {
		res[i] = ProtoCategory(c)
	}
	return res
}

func (pc ProtoCategoryArray) Strings() []string {
	res := make([]string, len(pc))
	for i, c := range pc {
		res[i] = c.String()
	}
	return res
}

func (p ProtoCategoryArray) Contains(c ProtoCategory) bool {
	for _, cc := range p {
		if cc == c {
			return true
		}
	}
	return false
}

func (p ProtoCategoryArray) Value() (driver.Value, error) {
	return pq.StringArray(p.Strings()).Value()
}

func (p *ProtoCategoryArray) Scan(value interface{}) error {
	a := make(pq.StringArray, 0)

	if err := a.Scan(value); err != nil {
		return err
	}

	*p = ProtoCagories(a...)

	return nil
}
