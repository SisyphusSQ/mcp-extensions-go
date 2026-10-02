// Package schemautil validates the explicit schema subset shared by helpers.
package schemautil

import (
	"fmt"
	"math"
	"net/mail"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// String checks string declarations. Patterns use Go's RE2 syntax.
func String(format, pattern string, min, max *int) error {
	if min != nil && *min < 0 || max != nil && *max < 0 || min != nil && max != nil && *min > *max {
		return fmt.Errorf("inconsistent string bounds")
	}
	if pattern != "" {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid RE2 pattern: %w", err)
		}
	}
	switch format {
	case "", "email", "uri", "date", "date-time":
		return nil
	default:
		return fmt.Errorf("unsupported string format %q", format)
	}
}

// Number checks numeric bounds, including intersections of inclusive/exclusive bounds.
func Number(min, max, exclusiveMin, exclusiveMax, multiple *float64) error {
	for _, n := range []*float64{min, max, exclusiveMin, exclusiveMax, multiple} {
		if n != nil && (math.IsNaN(*n) || math.IsInf(*n, 0)) {
			return fmt.Errorf("numeric constraints must be finite")
		}
	}
	if multiple != nil && *multiple <= 0 {
		return fmt.Errorf("multipleOf must be positive")
	}
	for _, lower := range []*float64{min, exclusiveMin} {
		for _, upper := range []*float64{max, exclusiveMax} {
			if lower != nil && upper != nil && (*lower > *upper || *lower == *upper && (lower == exclusiveMin || upper == exclusiveMax)) {
				return fmt.Errorf("inconsistent numeric bounds")
			}
		}
	}
	return nil
}

var dateTime = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[Tt]\d{2}:\d{2}:\d{2}(\.\d+)?([Zz]|[+-](0\d|1\d|2[0-3]):[0-5]\d)$`)

const uriChar = `(?:[A-Za-z0-9._~!$&'()*+,;=:@-]|%[0-9A-Fa-f]{2})`
const uriName = `(?:[A-Za-z0-9._~!$&'()*+,;=-]|%[0-9A-Fa-f]{2})`

var uriPath = regexp.MustCompile(`^(?:` + uriChar + `|/)*$`)
var uriQuery = regexp.MustCompile(`^(?:` + uriChar + `|[/?])*$`)
var authority = regexp.MustCompile(`^(?:(?:` + uriName + `|:)*@)?(?:` + uriName + `*|\[[A-Za-z0-9._~!$&'()*+,;=:-]+\])(?::[0-9]*)?$`)
var ipvFuture = regexp.MustCompile(`^[vV][0-9A-Fa-f]+\.[A-Za-z0-9._~!$&'()*+,;=:-]+$`)

// Format validates the supported formats without network access. Date-time uses
// RFC 3339 with no leap seconds; URI requires an absolute ASCII URI.
func Format(format, value string) error {
	valid := true
	switch format {
	case "":
		return nil
	case "email":
		address, err := mail.ParseAddress(value)
		valid = err == nil && address.Address == value && !strings.ContainsAny(value, "\r\n")
		if valid {
			local, domain, _ := strings.Cut(value, "@")
			valid = !strings.HasPrefix(local, `"`) && len(local) <= 64 && len(value) <= 254 && strings.Contains(domain, ".")
			if _, err := netip.ParseAddr(domain); err == nil {
				valid = false
			}
			for _, label := range strings.Split(domain, ".") {
				if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
					valid = false
				}
				for _, c := range label {
					if c != '-' && !unicode.IsLetter(c) && !unicode.IsNumber(c) {
						valid = false
					}
				}
			}
		}
	case "uri":
		valid = URI(value)
	case "date":
		_, err := time.Parse("2006-01-02", value)
		valid = err == nil && len(value) == 10 && !strings.HasPrefix(value, "0000-")
	case "date-time":
		_, err := time.Parse(time.RFC3339Nano, strings.ToUpper(value))
		valid = err == nil && dateTime.MatchString(value) && !strings.HasPrefix(value, "0000-")
	default:
		valid = false
	}
	if !valid {
		return fmt.Errorf("value does not match format %q", format)
	}
	return nil
}

// URI accepts opaque resource URIs as well as hierarchical URIs. It performs
// syntax validation only, never authorization, DNS lookup, or a resource read.
func URI(value string) bool {
	for _, c := range value {
		if c <= 0x20 || c >= 0x7f || strings.ContainsRune(`<>"{}|\^`+"`", c) {
			return false
		}
	}
	u, err := url.Parse(value)
	if err != nil || u.Scheme == "" {
		return false
	}
	// Validate escaped components: net/url accepts characters in opaque paths
	// and bracketed authorities that are outside RFC 3986's grammar.
	_, rest, _ := strings.Cut(value, ":")
	pathQuery, fragment, hasFragment := strings.Cut(rest, "#")
	if hasFragment && !uriQuery.MatchString(fragment) {
		return false
	}
	path, query, hasQuery := strings.Cut(pathQuery, "?")
	if hasQuery && !uriQuery.MatchString(query) {
		return false
	}
	if strings.HasPrefix(path, "//") {
		authPath := path[2:]
		auth, tail, hasPath := strings.Cut(authPath, "/")
		if !authority.MatchString(auth) {
			return false
		}
		if strings.HasPrefix(u.Hostname(), "v") || strings.HasPrefix(u.Hostname(), "V") {
			if strings.Contains(auth, "[") && !ipvFuture.MatchString(u.Hostname()) {
				return false
			}
		} else if strings.Contains(auth, "[") {
			address, err := netip.ParseAddr(u.Hostname())
			if err != nil || !address.Is6() {
				return false
			}
		}
		path = ""
		if hasPath {
			path = "/" + tail
		}
	}
	return uriPath.MatchString(path)
}
