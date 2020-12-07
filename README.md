# odoh-client-go

`odoh-client` is a command line interface as a client for making [Oblivious DNS-over-HTTPS](https://tools.ietf.org/html/draft-pauly-dprive-oblivious-doh-03) queries. The implementation is based on [chris-wood/odoh-client](https://github.com/chris-wood/odoh-client).

It currently supports the following functionalities:

- [x] DoH Query: `odoh-client doh --domain www.cloudflare.com. --dnstype AAAA --target <target>` where `<target>` is the name of the target resolver, e.g., `odoh.cloudflare-dns.com`.
- [x] ODoH Query: `odoh-client odoh --domain www.cloudflare.com. --dnstype AAAA --target <target>`
- [x] ODoH Query via Proxy: `odoh-client odoh --domain www.cloudflare.com. --dnstype AAAA --target <target> --proxy <proxy>` where `<proxy>` is the name of a proxy server.

## Usage

To build the executable, do:

```sh
go build -o odoh-client ./cmd/...
```

### DoH query to target

```sh
./odoh-client doh --domain www.cloudflare.com. --target odoh.cloudflare-dns.com --dnstype AAAA
```

### ODoH query to target

```sh
./odoh-client odoh --domain www.cloudflare.com. --dnstype AAAA --target odoh.cloudflare-dns.com
```

### ODoH query to target via a proxy

```sh
./odoh-client odoh --domain www.cloudflare.com. --dnstype AAAA --target odoh.cloudflare-dns.com --proxy odoh1.surfdomeinen.nl
```

### Fetch ODoH configuration

```sh
./odoh-client odohconfig-fetch --target odoh.cloudflare-dns.com --pretty
```
