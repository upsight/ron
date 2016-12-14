// Package httpstat provides http trace timings for requests.
// This implementation is based on https://github.com/davecheney/httpstat
package httpstat

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

// Define templates for rendering timings.
const (
	HTTPSTemplate = `` +
		`  DNS Lookup   TCP Connection   TLS Handshake   Server Processing   Content Transfer` + "\n" +
		`[%s  |     %s  |    %s  |        %s  |       %s  ]` + "\n" +
		`            |                |               |                   |                  |` + "\n" +
		`   namelookup:%s      |               |                   |                  |` + "\n" +
		`                       connect:%s     |                   |                  |` + "\n" +
		`                                   pretransfer:%s         |                  |` + "\n" +
		`                                                     starttransfer:%s        |` + "\n" +
		`                                                                                total:%s` + "\n"

	HTTPTemplate = `` +
		`   DNS Lookup   TCP Connection   Server Processing   Content Transfer` + "\n" +
		`[ %s  |     %s  |        %s  |       %s  ]` + "\n" +
		`             |                |                   |                  |` + "\n" +
		`    namelookup:%s      |                   |                  |` + "\n" +
		`                        connect:%s         |                  |` + "\n" +
		`                                      starttransfer:%s        |` + "\n" +
		`                                                                 total:%s` + "\n"
	maxRedirects = 10
)

// Command ...
type Command struct {
	Name              string
	W                 io.Writer
	WErr              io.Writer
	AppName           string
	method            string
	body              string
	headers           headers
	redirectsFollowed int
	showBody          bool
}

// Key returns the commands name for sorting.
func (c *Command) Key() string {
	return c.Name
}

// Run ...
func (c *Command) Run(args []string) (int, error) {
	f := flag.NewFlagSet(c.Name, flag.ExitOnError)
	f.Usage = func() {
		fmt.Fprintf(c.W, "Usage: %s %s [OPTIONS] URL\n", c.AppName, c.Name)
		f.PrintDefaults()
	}
	f.StringVar(&c.method, "X", "GET", "HTTP method to use")
	f.StringVar(&c.body, "d", "", "the body for POST or PUT requests")
	f.BoolVar(&c.showBody, "v", false, "show the body for the response")
	f.Var(&c.headers, "H", "set HTTP headers -H 'Accept: ...' -H 'Range: ...'")
	f.Parse(args)
	if len(f.Args()) != 1 {
		f.Usage()
		return 1, nil
	}
	if (c.method == "POST" || c.method == "PUT") && c.body == "" {
		log.Fatal("must supply post body using -d when POST or PUT is used")
	}

	url := c.parseURL(f.Arg(0))
	c.visit(url)

	return 0, nil
}

// visit visits a url and times the interaction.
// If the response is a 30x, visit follows the redirect.
func (c *Command) visit(url *url.URL) {
	req := c.newRequest(c.method, url, c.body)

	var t0, t1, t2, t3, t4 time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { t0 = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t1 = time.Now() },
		ConnectStart: func(_, _ string) {
			if t1.IsZero() {
				// connecting to IP
				t1 = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				log.Fatalf("unable to connect to host %v: %v", addr, err)
			}
			t2 = time.Now()

			fmt.Printf("\nConnected to %s\n", addr)
		},
		GotConn:              func(_ httptrace.GotConnInfo) { t3 = time.Now() },
		GotFirstResponseByte: func() { t4 = time.Now() },
	}
	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	switch url.Scheme {
	case "https":
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}

		tr.TLSClientConfig = &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: true,
		}

		// Because we create a custom TLSClientConfig, we have to opt-in to HTTP/2.
		// See https://github.com/golang/go/issues/14275
		err = http2.ConfigureTransport(tr)
		if err != nil {
			log.Fatalf("failed to prepare transport for HTTP/2: %v", err)
		}
	}

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// always refuse to follow redirects, visit does that
			// manually if required.
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to read response: %v", err)
	}

	bodyMsg := c.readResponseBody(req, resp)
	resp.Body.Close()

	t5 := time.Now() // after read body
	if t0.IsZero() {
		// we skipped DNS
		t0 = t1
	}

	// print status line and headers
	fmt.Printf("\nHTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)

	names := make([]string, 0, len(resp.Header))
	for k := range resp.Header {
		names = append(names, k)
	}
	sort.Sort(headers(names))
	for _, k := range names {
		fmt.Printf("%s: %s\n", k, strings.Join(resp.Header[k], ","))
	}

	if bodyMsg != "" {
		fmt.Printf("\n%s\n", bodyMsg)
	}

	fmta := func(d time.Duration) string {
		return fmt.Sprintf("%7dms", int(d/time.Millisecond))
	}

	fmtb := func(d time.Duration) string {
		return fmt.Sprintf("%-9sms", strconv.Itoa(int(d/time.Millisecond)))
	}

	colorize := func(s string) string {
		v := strings.Split(s, "\n")
		return strings.Join(v, "\n")
	}

	fmt.Println()

	switch url.Scheme {
	case "https":
		fmt.Printf(colorize(HTTPSTemplate),
			fmta(t1.Sub(t0)), // dns lookup
			fmta(t2.Sub(t1)), // tcp connection
			fmta(t3.Sub(t2)), // tls handshake
			fmta(t4.Sub(t3)), // server processing
			fmta(t5.Sub(t4)), // content transfer
			fmtb(t1.Sub(t0)), // namelookup
			fmtb(t2.Sub(t0)), // connect
			fmtb(t3.Sub(t0)), // pretransfer
			fmtb(t4.Sub(t0)), // starttransfer
			fmtb(t5.Sub(t0)), // total
		)
	case "http":
		fmt.Printf(colorize(HTTPTemplate),
			fmta(t1.Sub(t0)), // dns lookup
			fmta(t3.Sub(t1)), // tcp connection
			fmta(t4.Sub(t3)), // server processing
			fmta(t5.Sub(t4)), // content transfer
			fmtb(t1.Sub(t0)), // namelookup
			fmtb(t3.Sub(t0)), // connect
			fmtb(t4.Sub(t0)), // starttransfer
			fmtb(t5.Sub(t0)), // total
		)
	}

	if c.isRedirect(resp) {
		loc, err := resp.Location()
		if err != nil {
			if err == http.ErrNoLocation {
				// 30x but no Location to follow, give up.
				return
			}
			log.Fatalf("unable to follow redirect: %v", err)
		}

		c.redirectsFollowed++
		if c.redirectsFollowed > maxRedirects {
			log.Fatalf("maximum number of redirects (%d) followed", maxRedirects)
		}

		c.visit(loc)
	}
}

// readResponseBody consumes the body of the response.
// readResponseBody returns an informational message about the
// disposition of the response body's contents.
func (c *Command) readResponseBody(req *http.Request, resp *http.Response) string {
	if c.isRedirect(resp) || req.Method == http.MethodHead {
		return ""
	}

	msg := "Body discarded"
	switch c.showBody {
	case true:
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("failed to read response body: %v", err)
		}
		msg = string(data)
	default:
		if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
			log.Fatalf("failed to read response body: %v", err)
		}
	}

	return msg
}

func (c *Command) parseURL(uri string) *url.URL {
	if !strings.Contains(uri, "://") && !strings.HasPrefix(uri, "//") {
		uri = "//" + uri
	}

	url, err := url.Parse(uri)
	if err != nil {
		log.Fatalf("could not parse url %q: %v", uri, err)
	}

	if url.Scheme == "" {
		url.Scheme = "http"
		if !strings.HasSuffix(url.Host, ":80") {
			url.Scheme += "s"
		}
	}
	return url
}

func (c *Command) headerKeyValue(h string) (string, string) {
	i := strings.Index(h, ":")
	if i == -1 {
		log.Fatalf("Header '%s' has invalid format, missing ':'", h)
	}
	return strings.TrimRight(h[:i], " "), strings.TrimLeft(h[i:], " :")
}

func (c *Command) isRedirect(resp *http.Response) bool {
	return resp.StatusCode > 299 && resp.StatusCode < 400
}

func (c *Command) newRequest(method string, url *url.URL, body string) *http.Request {
	req, err := http.NewRequest(method, url.String(), c.createBody(body))
	if err != nil {
		log.Fatalf("unable to create request: %v", err)
	}
	for _, h := range c.headers {
		k, v := c.headerKeyValue(h)
		if strings.EqualFold(k, "host") {
			req.Host = v
			continue
		}
		req.Header.Add(k, v)
	}
	return req
}

func (c *Command) createBody(body string) io.Reader {
	if strings.HasPrefix(body, "@") {
		filename := body[1:]
		f, err := os.Open(filename)
		if err != nil {
			log.Fatalf("failed to open data file %s: %v", filename, err)
		}
		return f
	}
	return strings.NewReader(body)
}

// Aliases are the aliases and name for the command. For instance
// a command can have a long form and short form.
func (c *Command) Aliases() map[string]struct{} {
	return map[string]struct{}{
		"httpstat": struct{}{},
		"hs":       struct{}{},
	}
}

// Description is what is printed in Usage.
func (c *Command) Description() string {
	return "HTTP trace timings"
}
