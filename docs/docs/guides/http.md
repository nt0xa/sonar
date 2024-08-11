---
sidebar_position: 5
---

# HTTP routes

```
Manage HTTP routes

Usage:
 /http [command]

Available Commands:
  /new         Create new HTTP route
  /del         Delete HTTP route
  /list        List HTTP routes
  /clr         Delete multiple HTTP routes

Flags:
  -h, --help   help for http

Use "/http [command] --help" for more information about a command.
```

- You can create HTTP routes on your payloads, which will respond with any content you want.
- It is possible to set response status code, headers and body.
- It is possible to create dynamic response based on request headers, query parameters, path
parameters, etc. using Golang template language.
- Data available to use in templates. Example HTTP request for reference:

	```http
	POST /test/PATH?param=QUERY HTTP/1.1
	Host: d14a68e4.sonar.test 
	User-Agent: curl/7.86.0
	Accept: */*
	Param: HEADER
	Cookie: param=COOKIE;
	Content-Length: 10
	Content-Type: application/x-www-form-urlencoded
	Connection: close

	param=FORM
	```

	- `{{ .Host }}` - HTTP `Host` header: `d14a68e4.sonar.test`

	- `{{ .Method }}` - HTTP method: `POST`

	- `{{ .Path }}` - HTTP path: `/test/PATH`
	- `{{ .RawQuery }}` - query part after `?`: `param=QUERY`
	- `{{ .RequestURI }}` - fullURI (path + query): `/test/PATH?param=QUERY`
	- `{{ .Header "Param" }}` - value of `Param` header: `HEADER`
	- `{{ .URLParam "param" }}` - path parameter defined during route creation like
	this `/test/{param}`: `PATH`
	- `{{ .Query "param" }}` - value of query paramter with name `param`:  `QUERY`
	- `{{ .Form "param" }}` - value of body parameter with name `param`: `FORM`
	- `{{ .Cookie "param" }}` — value of cookie with name `param`: `COOKIE`

## Create new HTTP route

```
Create new HTTP route

Usage:
  /http new BODY [flags]

Flags:
  -c, --code int             Response status code (default 200)
  -d, --dynamic              Interpret body and headers as templates
  -H, --header stringArray   Response header
  -h, --help                 help for new
  -m, --method string        Request method (one of "ANY", "CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT", "TRACE") (default "GET")
  -P, --path string          Request path (default "/")
  -p, --payload string       Payload name
```

### Create HTTP GET route with 302 redirect to https://example.com on path "/redirect"

```
/http new --payload <NAME> --path /redirect --code 302 --header 'Location: https://example.com' "Redirecting..."
```

![Create new HTTP route with redirect](../assets/http_new_redirect_telegram_dark.png#gh-dark-mode-only)![Create new HTTP route with redirect](../assets/http_new_redirect_telegram_light.png#gh-light-mode-only)

Now `/redirect` on the payload will return code 302, header `Location: https://example.com` and body `Redirecting...`:

![Test HTTP redirect route](../assets/http_test_redirect_dark.png#gh-dark-mode-only)![Test HTTP redirect route](../assets/http_test_redirect_light.png#gh-light-mode-only)

And you will also receive an alert:

![Test HTTP redirect alert](../assets/http_test_alert_redirect_telegram_dark.png#gh-dark-mode-only)![Test HTTP redirect alert](../assets/http_test_alert_redirect_telegram_light.png#gh-light-mode-only)

### Create HTTP POST route which uses request form parameter in response

```
/http new -p <NAME> -m POST -P /hello -c 200 -d 'Hello {{ .Query "name" }}!'
```

![Create new HTTP dynamic POST](../assets/http_new_dynamic_telegram_dark.png#gh-dark-mode-only)![Create new HTTP dynamic POST](../assets/http_new_dynamic_telegram_light.png#gh-light-mode-only)

Now you can test it by seding `POST` request with form data `name=peter` to `d14a68e4.sonar.test/hello`:

![Test dynamic HTTP POST route](../assets/http_test_dynamic_dark.png#gh-dark-mode-only)![Test dynamic HTTP POST route](../assets/http_test_dynamic_light.png#gh-light-mode-only)

And you will also receive an alert:

![Test HTTP dynamic route alert](../assets/http_test_alert_dynamic_telegram_dark.png#gh-dark-mode-only)![Test HTTP dynamic route alert](../assets/http_test_alert_dynamic_telegram_light.png#gh-light-mode-only)

### Create HTTP route for any method with all possible dynamic variables in response

```
/http new -p test -m ANY -P "/test/{param}" -d '.Host => {{ .Host }}, .Method => {{ .Method }}, .Path => {{ .Path }}, .RawQuery => {{ .RawQuery }}, .RequestURI => {{ .RequestURI }}, .Header "param" => {{ .Header "param" }}, .URLParam "param" => {{ .URLParam "param" }}, .Query "param" => {{ .Query "param" }}, .Form "param" => {{ .Form "param" }}'
```

![Create new HTTP dynamic route with all parameters](../assets/http_new_dynamic_all_telegram_dark.png#gh-dark-mode-only)![Create new HTTP dynamic route with all parameters](../assets/http_new_dynamic_all_telegram_light.png#gh-light-mode-only)

Now you can test it by seding request with `curl`:

![Test dynamic HTTP route with all parameters](../assets/http_test_dynamic_all_dark.png#gh-dark-mode-only)![Test dynamic HTTP route with all parameters](../assets/http_test_dynamic_all_light.png#gh-light-mode-only)

And you will also receive an alert:

![Test HTTP dynamic route alert](../assets/http_test_alert_dynamic_all_telegram_dark.png#gh-dark-mode-only)![Test HTTP dynamic route alert](../assets/http_test_alert_dynamic_all_telegram_light.png#gh-light-mode-only)

## List HTTP routes

```
List HTTP routes

Usage:
  /http list [flags]

Flags:
  -h, --help             help for list
  -p, --payload string   Payload name
```

### List HTTP routes for payload

```
/http list -p <NAME>
```

- Every HTTP route has an index, which can be used in `/del` command to remove the record.

![List HTTP routes for payload](../assets/http_list_telegram_dark.png#gh-dark-mode-only)![List HTTP routes for payload](../assets/http_list_telegram_light.png#gh-light-mode-only)

## Delete HTTP route

```
Delete HTTP route identified by INDEX

Usage:
  /http del INDEX [flags]

Flags:
  -h, --help             help for del
  -p, --payload string   Payload name
```

### Delete HTTP route for payload by index

```
/http del -p <NAME> <INDEX>
```

![Delete HTTP route for payload by index](../assets/http_del_telegram_dark.png#gh-dark-mode-only)![Delete HTTP route for payload by index](../assets/http_del_telegram_light.png#gh-light-mode-only)

### Clear HTTP routes for payload

```
/http clr -p <NAME>
```

![Delete all HTTP routes for payload](../assets/http_clear_telegram_dark.png#gh-dark-mode-only)![Delete all HTTP routes for payload](../assets/http_clear_telegram_light.png#gh-light-mode-only)
