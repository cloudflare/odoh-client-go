package commands

const (
	DEFAULT_DOH_SERVER         = "cloudflare-dns.com"
	DOH_CONTENT_TYPE           = "application/dns-message"
	OBLIVIOUS_DOH_CONTENT_TYPE = "application/oblivious-dns-message"
	TARGET_HTTP_MODE           = "https"
	PROXY_HTTP_MODE            = "https"
	ODOH_CONFIG_WELLKNOWN_URL  = "/.well-known/odohconfigs"
)
