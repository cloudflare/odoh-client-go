module github.com/cloudflare/odoh-client-go

go 1.14

require (
	cloud.google.com/go v0.74.0 // indirect
	cloud.google.com/go/logging v1.1.2
	github.com/cloudflare/circl v1.0.1-0.20201214203952-f327aa409851
	github.com/cloudflare/odoh-go v0.1.3
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/elastic/go-elasticsearch/v8 v8.0.0-20201202142044-1e78b5bf06b1
	github.com/miekg/dns v1.1.35
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/urfave/cli v1.22.5
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
	golang.org/x/tools v0.0.0-20201211185031-d93e913c1a58 // indirect
	google.golang.org/genproto v0.0.0-20201214200347-8c77b98c765d // indirect
)

replace github.com/cloudflare/odoh-go v0.1.3 => github.com/cloudflare/odoh-go v0.1.4-0.20201214225149-38529c8b3758
