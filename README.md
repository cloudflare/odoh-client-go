# odoh-client
Oblivious DoH client

This is a command line interface as a client for performing oblivious dns-over-https queries.

### Current Support:

- [x] DoH Query : `odoh-client doh --domain www.cloudflare.com. --dnsType AAAA`
- [x] oDoH Query: `odoh-client odoh --domain www.cloudflare.com. --dnsType AAAA --key 01234567890123456789012345678912 --target 1.1.1.1`
- [x] oDoH Query via Proxy: `odoh-client odoh --domain www.cloudflare.com --dnsType AAAA --key 01234567890123456789012345678912 --target 1.1.1.1 --use-proxy true --proxy sampleproxy.service.hosted.net[:port]`

The current implementation for oDoH uses a dummy Public Key stub on the target server which provides the public key to 
the client. In the ideal implementation, this will be obtained after performing DNSSEC validation + HTTPSSVC.

The explicit query for the public key of a target server without validation can be obtained by performing 
`odoh-client get-publickey --ip 1.1.1.1[:port]`

For the `proxy` usage, the client treats the `target` as the hostname and port of the intended target to which the proxy
needs to forward the ODOH message and obtain a response from. The client then uses the `key` to decrypt the obtained 
response from the Oblivious Target.

### Tests

|  Instances    | Link                                           | Active  | Code           |
|---------------|------------------------------------------------|---------|----------------|
| Target Server | odoh-target-dot-odoh-target.wm.r.appspot.com   | &check; | GCP Go Target  |
| Proxy Server  | odoh-proxy-dot-odoh-target.wm.r.appspot.com    | &check; | GCP Go Proxy   |
| Target Server | odoh-target-rs.crypto-team.workers.dev         | &check; | CF Rust Target |
| Proxy Server  | alpha-odoh-rs-proxy.research.cloudflare.com    | &check; | CF Rust Proxy  |
| Discovery     | odoh-discovery.crypto-team.workers.dev         | &check; | CF Discovery   |

### Experiments

| Proxied Via   | To Target      | Can Resolve? | CLI Call                                                                                                                                                               |
|---------------|----------------|--------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| GCP Go Proxy  | GCP Go Target  | &check;      | `odoh-client odoh --domain www.github.com. --dnstype AAAA --target odoh-target-dot-odoh-target.wm.r.appspot.com --proxy odoh-proxy-dot-odoh-target.wm.r.appspot.com` |
| GCP Go Proxy  | CF Rust Target | &check;      | `odoh-client odoh --domain www.github.com. --dnstype AAAA --target odoh-target-rs.crypto-team.workers.dev --proxy odoh-proxy-dot-odoh-target.wm.r.appspot.com`       |
| CF Rust Proxy | CF Rust Target | &cross;      | `odoh-client odoh --domain www.github.com. --dnstype AAAA --target odoh-target-rs.crypto-team.workers.dev --proxy alpha-odoh-rs-proxy.research.cloudflare.com`       |
| CF Rust Proxy | GCP Go Target  | &check;      | `odoh-client odoh --domain www.github.com. --dnstype AAAA --target odoh-target-dot-odoh-target.wm.r.appspot.com --proxy odoh-rs-proxy.crypto-team.workers.dev`       |

Note: The CF Worker &rightarrow; CF Worker communication will NOT work and is by design from Cloudflare Workers. The 
usage of ODOH Rust Proxy and Target together however does work correctly if the workers are hosted on different zones. 