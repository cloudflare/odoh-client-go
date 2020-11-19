## Usage

#### DOH Query to target

```sh
./odoh-client doh --domain www.apple.com. --target odoh-target-dot-odoh-target.wm.r.appspot.com --dnstype AAAA
```


#### ODOH Query to target

```sh
./odoh-client odoh --domain www.cloudflare.com. --dnstype AAAA --target odoh-target-dot-odoh-target.wm.r.appspot.com --key 01234567890123456789012345678912
```

#### ODOH Query to target via a proxy

```sh
./odoh-client odoh --domain www.cloudflare.com. --dnstype AAAA --target odoh-target-dot-odoh-target.wm.r.appspot.com --key 01234567890123456789012345678912 --proxy odoh-proxy-dot-odoh-target.wm.r.appspot.com
```

#### Get Public Key of a target

```sh
./odoh-client get-publickey --ip odoh-target-dot-odoh-target.wm.r.appspot.com
```
