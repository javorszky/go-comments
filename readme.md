# Go Comments

A work in progress.

Like Disqus, but doesn't suck.

## Installation / Usage

### How to host it locally with https and a domain

The domain is assumed to be `goapp.test`. You may need to configure your local tlds with dnsmasq or using your hosts file, depending on your system.

It's also assumed you have nginx installed and that it responds to requests.

#### Nginx forwarding

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

### Create the self signed certificate

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

Once you have the files (`cert.pem` and `key.pem`), rename them to `cert.crt` and `key.key`. Find the path with `pwd`, and make sure the nginx config file points to these two files. These files are ignored by the `.gitignore` file.

If you're on mac, open the Keychain Access app, and drag the `cert.crt` file into there. Search for **ACME**, and set it to always trust the certificate.

### Set a .env file

The `.env` file needs to contain the following four things:

```dotenv
DB_USER=<your mysql db username>
DB_PASS=<your mysql db user's password>
DB_TABLE=<your mysql database name>
DB_ADDRESS=""
PORT=<port for http. HTTPS is currently always on 1323>
```

### How to use this with Docker?

The repo has three docker related files:

- `Dockerfile`
- `docker-compose.yml`
- `.env.docker`

All three files are needed for a successful docker initialisation. Then the only thing you should need to do is issue this command:

```dotenv
$ docker-compose up --build
```
That way the app should be accessible on `https://localhost:5000`, which should be SSL, but the certificate should be untrusted.

## Tooling decision

### Password

It was between `bcrypt` and `argon2`. In the end I went with Argon2 as that's the stronger of the two. I've essentially followed [How to Hash and Verify Passwords With Argon2 in Go](https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go) by Alex Edwards (dated 10th December 2018) with some minor modifications around wrapping the functionality into a package I can pass into the app.

### 2FA

**Definitely not SMS based.**

Initially I wanted to use the native Authy app and dashboard, but following some research, I decided against it.

One reason is because in order for you to add an Authy token, you need the Authy app, and you need to tell them your phone number. As we've recently seen with Facebook, phone numbers are used to identify a user across services even if they didn't specifically consent. See the following reading material:

* This entire Twitter thread: https://twitter.com/jeremyburge/status/1101402001907372032?s=19
* This article about 2FA QR codes containing more information than they should: https://medium.com/crypto-punks/why-you-shouldnt-scan-two-factor-authentication-qr-codes-e2a44876a524

Because of this I decided to roll my own using the standard itself, and only including the information necessary to create the TOTP (ie no domain, no email address) based on this article: [A DIY Two-Factor Authenticator in Golang](https://blog.gojekengineering.com/a-diy-two-factor-authenticator-in-golang-32e5641f6ec5) by Tilak Lodha (31st May 2018).

### Magic link login

Yes, to be developed...

### Yubikey

Yes, to be developed...
