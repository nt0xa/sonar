package models

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type Proto struct {
	Name string
}

func (p Proto) String() string {
	return string(p.Name)
}

func (p Proto) Value() (driver.Value, error) {
	return p.String(), nil
}

func (p *Proto) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.New("type assertion to string failed")
	}
	*p = Proto{s}
	return nil
}

func (p Proto) Category() ProtoCategory {
	return ProtoToCategory(p)
}

var (
	ProtoDNS   = Proto{"dns"}
	ProtoHTTP  = Proto{"http"}
	ProtoHTTPS = Proto{"https"}
	ProtoSMTP  = Proto{"smtp"}
	ProtoFTP   = Proto{"ftp"}
)

type ProtoCategory struct {
	Name string
}

func (p ProtoCategory) String() string {
	return string(p.Name)
}

var (
	ProtoCategoryDNS  = ProtoCategory{"dns"}
	ProtoCategoryHTTP = ProtoCategory{"http"}
	ProtoCategorySMTP = ProtoCategory{"smtp"}
	ProtoCategoryFTP  = ProtoCategory{"ftp"}
)

var ProtoCategoriesAll = ProtoCategoryArray{
	ProtoCategoryDNS,
	ProtoCategoryHTTP,
	ProtoCategorySMTP,
	ProtoCategoryFTP,
}

func ProtoToCategory(p Proto) ProtoCategory {
	switch p {
	case ProtoDNS:
		return ProtoCategoryDNS
	case ProtoHTTP, ProtoHTTPS:
		return ProtoCategoryHTTP
	case ProtoSMTP:
		return ProtoCategorySMTP
	case ProtoFTP:
		return ProtoCategoryFTP
	}

	panic(fmt.Sprintf("invalid protocol: %s", p))
}

type ProtoCategoryArray []ProtoCategory

func ProtoCategories(cc ...string) ProtoCategoryArray {
	res := make([]ProtoCategory, len(cc))
	for i, c := range cc {
		res[i] = ProtoCategory{c}
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

	*p = ProtoCategories(a...)

	return nil
}
