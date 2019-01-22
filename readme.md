# Go Comments

A work in progress.

Like Disqus, but doesn't suck.

## How to host it locally with https and a domain

The domain is assumed to be `goapp.test`. You may need to configure your local tlds with dnsmasq or using your hosts file, depending on your system.

It's also assumed you have nginx installed and that it responds to requests.

### Nginx forwarding

```
server {
	listen 80;
	server_name goapp.test;
	return 301 https://$host$request_uri;
}

server {
	listen 443 ssl http2;
	server_name goapp.test;
	charset utf-8;
	root /;
	ssl_certificate         /path/to/your/self/signed/cert/file/cert.crt;
	ssl_certificate_key     /path/to/your/self/signed/key/file/key.key;

	location ~ {
		proxy_ssl_session_reuse on;
		proxy_pass              https://localhost:1323;
		proxy_ssl_protocols     TLSv1 TLSv1.1 TLSv1.2;
        proxy_ssl_ciphers       HIGH:!aNULL:!MD5;
	}
}

```
Save this in a `.conf` file, and get nginx to load it for you, then restart nginx.

This config assumes you've already created your self signed certificates (see below).

## Create the self signed certificate

As per https://echo.labstack.com/cookbook/http2, you can use the following command in your app's directory to generate your certificate:

```
$ go run $GOROOT/src/crypto/tls/generate_cert.go --host=goapp.test --ca=true --ecdsa-curve=P384
```

Make sure your `$GOROOT` is set up properly. This should be in your profile file for your terminal:

```
# don't forget to change your path correctly!

export GOPATH=$HOME/golang
export GOROOT=/usr/local/opt/go/libexec
export PATH=$PATH:$GOPATH/bin
export PATH=$PATH:$GOROOT/bin
```

See https://gist.github.com/vsouza/77e6b20520d07652ed7d for more variations.

Once you have the files (`cert.pem` and `key.pem`), rename them to `cert.crt` and `key.key`. Find the path with `pwd`, and make sure the nginx config file points to these two files.

If you're on mac, open the Keychain Access app, and drag the `cert.crt` file into there. Search for **ACME**, and set it to always trust the certificate.

## Set a .env file

The `.env` file needs to contain the following four things:

```dotenv
DB_USER=<your mysql db username>
DB_PASS=<your mysql db user's password>
DB_TABLE=<your mysql database name>
PORT=<port for http. HTTPS is currently always on 1323>
```
