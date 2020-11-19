# odoh-client

`odoh-client` is a command line interface as a client for performing oblivious dns-over-https queries, based on [chris-wood/odoh-client](https://github.com/chris-wood/odoh-client).

It currently supports the following functionalities:

- [x] DoH Query : `odoh-client doh --domain www.cloudflare.com. --dnsType AAAA`
- [x] ODoH Query: `odoh-client odoh --domain www.cloudflare.com. --dnsType AAAA --target <target>`
- [x] ODoH Query via Proxy: `odoh-client odoh --domain www.cloudflare.com. --dnsType AAAA --target <target> --use-proxy true --proxy <proxy>`

## Usage

### DoH query to target

```sh
./odoh-client doh --domain www.cloudflare.com. --target<target> --dnstype AAAA
```

### ODoH query to target

```sh
./odoh-client odoh --domain www.cloudflare.com. --dnstype AAAA --target <target>
```

### ODoH query to target via a proxy

```sh
./odoh-client odoh --domain www.cloudflare.com. --dnstype AAAA --target <target> --proxy <proxy>
```

### Fetch ODoH configuration

```sh
./odoh-client odohconfig-fetch --target <target> --pretty
```
