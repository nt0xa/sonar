package valid

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

var (
	namePattern     = `[a-z0-9]{1}([a-z0-9-]*[a-z0-9]{1})?`
	subdomainRegexp = regexp.MustCompile(fmt.Sprintf(`^(\*|%[1]s)(\.%[1]s)*$`, namePattern))
	fqdnRegexp      = regexp.MustCompile(fmt.Sprintf(`^(%s\.)+$`, namePattern))
)

func File(value interface{}) error {
	path, _ := value.(string)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	return nil
}

func Directory(value interface{}) error {
	path, _ := value.(string)

	if fi, err := os.Stat(path); os.IsNotExist(err) {
		return err
	} else if fi.Mode().IsRegular() {
		return errors.New("must be directory")
	}

	return nil
}

func Subdomain(value interface{}) error {
	val, _ := value.(string)

	if !subdomainRegexp.MatchString(val) {
		return errors.New("invalid subdomain")
	}

	return nil
}

func FQDN(value interface{}) error {
	val, _ := value.(string)

	if !fqdnRegexp.MatchString(val) {
		return errors.New("invalid fqdn")
	}

	return nil
}

func MX(value interface{}) error {
	val, _ := value.(string)

	parts := strings.Split(val, " ")

	_, err := strconv.Atoi(parts[0])

	if len(parts) == 2 &&
		err == nil &&
		fqdnRegexp.MatchString(parts[1]) {
		return nil
	}

	return errors.New("invalid mx record")
}

func CAA(value interface{}) error {
	v, _ := value.(string)

	var (
		flag uint8
		tag  string
		val  string
	)
	_, err := fmt.Sscanf(v, "%d %s %q", &flag, &tag, &val)
	if err != nil {
		return fmt.Errorf("invalid caa record: %w", err)
	}

	return nil
}

func DNSRecord(typ string) validation.Rule {
	switch typ {
	case "A":
		return is.IPv4

	case "AAAA":
		return is.IPv6

	case "MX":
		return validation.By(MX)

	case "TXT":
		return validation.Required

	case "CNAME":
		return validation.By(FQDN)

	case "CAA":
		return validation.By(CAA)
	}

	return validation.Required
}

func Base64(value interface{}) error {
	val, _ := value.(string)

	_, err := base64.StdEncoding.DecodeString(val)

	if err != nil {
		return fmt.Errorf("invalid base64 data")
	}

	return nil
}

type OneOfRule struct {
	values        []string
	caseSensetive bool
}

func (r *OneOfRule) Validate(value interface{}) error {
	val, _ := value.(string)

	if !r.caseSensetive {
		val = strings.ToLower(val)
	}

	for _, v := range r.values {
		if val == strings.ToLower(v) {
			return nil
		}
	}

	return fmt.Errorf("invalid value, expected one of %s", strings.Join(r.values, ","))
}

func OneOf(values []string, caseSensetive bool) validation.Rule {
	return &OneOfRule{values, caseSensetive}
}
