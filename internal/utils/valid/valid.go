package valid

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/bi-zone/sonar/internal/models"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

var (
	subdomainRegexp = regexp.MustCompile(`^(\*|[a-z0-9]{1}[a-z0-9-]+[a-z0-9]{1})(\.[a-z0-9]{1}[a-z0-9-]+[a-z0-9]{1})*$`)
	fqdnRegexp      = regexp.MustCompile(`^([a-z0-9]{1}[a-z0-9-]+[a-z0-9]{1}\.)+$`)
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

func DNSRecord(typ string) validation.Rule {

	switch typ {

	case models.DNSTypeA:
		return is.IPv4

	case models.DNSTypeAAAA:
		return is.IPv6

	case models.DNSTypeMX:
		return validation.By(MX)

	case models.DNSTypeTXT:
		return validation.Required

	case models.DNSTypeCNAME:
		return validation.By(FQDN)
	}

	return nil
}

type OneOfRule struct {
	values []string
}

func (r *OneOfRule) Validate(value interface{}) error {
	val, _ := value.(string)

	for _, v := range r.values {
		if v == val {
			return nil
		}
	}

	return fmt.Errorf("invalid value, expected one of %s", strings.Join(r.values, ","))
}

func OneOf(values []string) validation.Rule {
	return &OneOfRule{values}
}
