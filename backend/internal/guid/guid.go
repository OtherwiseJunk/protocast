// package guid implements logic deriving a GUID from a feed URI, using the Podcast Namespace UUIDv5 standard.
//
// Additional information on the methodology and rationale for this approach can be found in the Podcasting 2.0 specification: https://podcasting2.org/docs/podcast-namespace/tags/guid
// Before deriving the GUID, the feed URI is canonicalized to ensure consistency in the hashing process. The canonicalization process includes:
// - Lowercasing the host component of the URI.
// - Removing the scheme (e.g., "http://", "https://") from the URI.
// - Excluding default ports (80 for HTTP and 443 for HTTPS) from the host.
// - Stripping any single trailing slash from the URI.
//
// The resulting canonicalized URI is then hashed using SHA-1, combined with the Podcast Namespace UUIDv5, to produce a unique GUID for the feed.
package guid

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/net/idna"
)

// namespace is the industry-standard UUIDv5 Podcast Namespace.
var namespace = uuid.MustParse("ead4c236-bf58-58c6-a2c6-a6b28d128cb6")

// schemeName is the scheme production from RFC 3986 section 3.1:
// ALPHA *( ALPHA / DIGIT / "+" / "-" / "." ).
var schemeName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*$`)

// percentEscape matches one percent escape. RFC 3986 section 6.2.2.1 makes its
// hex digits case-insensitive, so %2F and %2f are the same octet and must not
// reach the hash as different strings.
var percentEscape = regexp.MustCompile(`%[0-9a-fA-F]{2}`)

func Namespace() uuid.UUID {
	return namespace
}

// Canonicalise normalises a feed URI for hashing: no scheme, no userinfo,
// lowercase A-label host, no trailing dot, no default port, no trailing
// slashes, uppercase percent escapes.
//
// It returns an empty string when the input has no usable host.
func Canonicalise(rawURI string) string {
	trimmed := strings.TrimSpace(rawURI)
	if trimmed == "" {
		return ""
	}

	if !isFetchableScheme(getScheme(trimmed)) {
		return ""
	}

	// url.Parse reads a host only after "//"
	toParse := trimmed
	if !hasScheme(toParse) {
		toParse = "//" + toParse
	}

	parsed, err := url.Parse(toParse)
	if err != nil || parsed.Hostname() == "" {
		return canonicaliseFallback(trimmed)
	}

	// if Userinfo is present, the Uri is not a valid feed URI like in the case of
	// "mailto:hi@example.com", or it is a private feed URI with credentials, which
	// we MUST NOT publish on AtProto, as it would be public. Returning an empty string.
	if parsed.User != nil {
		return ""
	}

	host := canonicalHost(parsed.Hostname(), parsed.Port())
	if host == "" {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(host)
	builder.WriteString(canonicalPath(parsed.EscapedPath()))

	if parsed.RawQuery != "" {
		builder.WriteString("?")
		builder.WriteString(upperPercentEscapes(parsed.RawQuery))
	}

	return builder.String()
}

func Derive(uri string) (canonical string, guidStr string) {
	canonical = Canonicalise(uri)
	return canonical, uuid.NewSHA1(Namespace(), []byte(canonical)).String()
}

func canonicaliseFallback(trimmed string) string {
	authority, remainder := splitAuthority(stripScheme(trimmed))

	parsed, err := url.Parse("//" + authority)
	if err != nil || parsed.Hostname() == "" || parsed.User != nil {
		return ""
	}

	host := canonicalHost(parsed.Hostname(), parsed.Port())
	if host == "" {
		return ""
	}
	return host + canonicalPath(remainder)
}

// canonicalHost reduces a host to the one spelling every alias of it shares.
//
// A trailing dot is the same DNS name, and an IDN U-label is the same host as
// its A-label.Because the scheme is not part of the canonical form, :80 and :443 carry no
// information; any other port is part of the address and is preserved.
func canonicalHost(hostname string, port string) string {
	host := strings.TrimSuffix(hostname, ".")

	for _, label := range strings.Split(host, ".") {
		if label == "" {
			return ""
		}
	}

	if !strings.Contains(host, ".") && !strings.Contains(host, ":") {
		return ""
	}

	// Hosts idna rejects - IPv6 literals, underscores - keep their own
	// spelling; they are already unambiguous.
	if ascii, err := idna.Lookup.ToASCII(host); err == nil {
		host = ascii
	}
	host = strings.ToLower(host)
	if host == "" {
		return ""
	}

	if strings.Contains(host, ":") {
		host = "[" + host + "]" // IPv6 literal, unbracketed by Hostname
	}
	if port != "" && port != "80" && port != "443" {
		host += ":" + port
	}
	return host
}

// canonicalPath strips every trailing slash and normalises escape case.
func canonicalPath(escapedPath string) string {
	return upperPercentEscapes(strings.TrimRight(escapedPath, "/"))
}

func hasScheme(rawURI string) bool {
	return strings.HasPrefix(rawURI, "//") || getScheme(rawURI) != ""
}

func isFetchableScheme(scheme string) bool {
	return scheme == "" || scheme == "http" || scheme == "https"
}

func getScheme(rawURI string) string {
	sep := strings.Index(rawURI, "://")
	if sep <= 0 {
		return ""
	}
	scheme := rawURI[:sep]
	if !schemeName.MatchString(scheme) {
		return ""
	}
	return strings.ToLower(scheme)
}

func splitAuthority(withoutScheme string) (authority string, remainder string) {
	if cut := strings.IndexAny(withoutScheme, "/?#"); cut != -1 {
		return withoutScheme[:cut], withoutScheme[cut:]
	}
	return withoutScheme, ""
}

func stripScheme(rawURI string) string {
	if sep := strings.Index(rawURI, "://"); sep > 0 && schemeName.MatchString(rawURI[:sep]) {
		rawURI = rawURI[sep+3:]
	}
	return strings.TrimPrefix(rawURI, "//")
}

func upperPercentEscapes(escaped string) string {
	return percentEscape.ReplaceAllStringFunc(escaped, strings.ToUpper)
}
