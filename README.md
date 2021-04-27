# odoh-client-go

`odoh-client` is a command line interface as a client for making [Oblivious DNS-over-HTTPS](https://tools.ietf.org/html/draft-pauly-dprive-oblivious-doh-06) queries. The implementation is based on [chris-wood/odoh-client](https://github.com/chris-wood/odoh-client).

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

### Running benchmarks

#### Obtaining a dataset

We use the Tranco Top Million dataset as an example dataset for this repository. The file `fetch-datasets.sh` retrieves
a sample Tranco top million dataset which is then parsed to remove the ranking numbers modifying the dataset file to
contain only the hostnames. Any other dataset which follows the same pattern can be directly used as a test file for 
benchmarking `ODoH`. By running the `make fetch` command, a new directory called `dataset/` is created and the tranco
top million file is obtained and stored.

```
make fetch
```

#### Running the Benchmark

`odoh-client` takes in the path to the dataset `--data` and reads the list, shuffles it to obtain a random order and 
performs the `ODoH` queries to a chosen `--target` and `--proxy`. By default if the `--out` is not specified, the output
is printed to console or is written to the file provided in the `--out` argument.

For reading the current defaults and additional configuration options, please run `odoh-client bench --help`

An example command for running the benchmark is as follows:

```sh
odoh-client bench --data dataset/tranco-1m.csv --target odoh.cloudflare-dns.com --proxy <instance> --out odoh-cf.json
``` 

### Note

> This tool includes a sub command for benchmarking various protocols and has been
> used for performing measurements presented in this [arxiv paper](https://arxiv.org/abs/2011.10121). There are also
> traces of telemetry which are used for the same purpose in an effort to reproduce the results of the paper.
