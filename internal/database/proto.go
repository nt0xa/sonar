package database

import "slices"

const (
	ProtoDNS   = "dns"
	ProtoHTTP  = "http"
	ProtoHTTPS = "https"
	ProtoSMTP  = "smtp"
	ProtoFTP   = "ftp"

	ProtoCategoryDNS  = "dns"
	ProtoCategoryHTTP = "http"
	ProtoCategorySMTP = "smtp"
	ProtoCategoryFTP  = "ftp"
)

var ProtoCategoriesAll = []string{
	ProtoCategoryDNS,
	ProtoCategoryHTTP,
	ProtoCategorySMTP,
	ProtoCategoryFTP,
}

func ProtoToCategory(proto string) string {
	switch proto {
	case ProtoDNS:
		return ProtoCategoryDNS
	case ProtoHTTP, ProtoHTTPS:
		return ProtoCategoryHTTP
	case ProtoSMTP:
		return ProtoCategorySMTP
	case ProtoFTP:
		return ProtoCategoryFTP
	default:
		return ""
	}
}

func ProtoCategoryContains(protocols []string, category string) bool {
	return slices.Contains(protocols, category)
}
