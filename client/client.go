package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"bytes"
	"fmt"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/cookiejar"
)

// Client codeforces client
type Client struct {
	Jar            *cookiejar.Jar `json:"cookies"`
	Handle         string         `json:"handle"`
	HandleOrEmail  string         `json:"handle_or_email"`
	Password       string         `json:"password"`
	Ftaa           string         `json:"ftaa"`
	Bfaa           string         `json:"bfaa"`
	LastSubmission *Info          `json:"last_submission"`
	host           string
	proxy          string
	path           string
	client         *http.Client
}

// Instance global client
var Instance *Client

// Init initialize
func Init(path, host, proxy string) {
	color.Yellow("Initializing client with path: %s, host: %s", path, host)
	
	jar, _ := cookiejar.New(nil)
	c := &Client{
		Jar:            jar,
		LastSubmission: nil,
		path:          path,
		host:          host,
		proxy:         proxy,
		client:        nil,
	}
	
	// Set the global instance immediately
	Instance = c
	color.Yellow("Set global Instance")
	
	if err := c.load(); err != nil {
		color.Red("Failed to load session: %v", err)
		color.Green("Create a new session in %v", path)
	} else {
		color.Green("Loaded existing session")
	}
	
	Proxy := http.ProxyFromEnvironment
	if len(proxy) > 0 {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			color.Red(err.Error())
			color.Green("Use default proxy from environment")
		} else {
			Proxy = http.ProxyURL(proxyURL)
		}
	}
	
	// Create a custom transport with browser-like headers
	transport := &http.Transport{
		Proxy: Proxy,
	}
	
	c.client = &http.Client{
		Jar: c.Jar,
		Transport: transport,
	}

	// Add default headers to all requests
	c.client.Transport = &customTransport{
		base: transport,
		ua: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	}

	if err := c.save(); err != nil {
		color.Red("Failed to save session: %v", err)
	} else {
		color.Green("Saved session successfully")
	}
	
	color.Yellow("Client initialization complete")
}

// Custom transport to add headers
type customTransport struct {
	base *http.Transport
	ua   string
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", t.ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document") 
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	return t.base.RoundTrip(req)
}

// load from path
func (c *Client) load() (err error) {
	file, err := os.Open(c.path)
	if err != nil {
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, c)
}

// save file to path
func (c *Client) save() (err error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err == nil {
		os.MkdirAll(filepath.Dir(c.path), os.ModePerm)
		err = ioutil.WriteFile(c.path, data, 0644)
	}
	if err != nil {
		color.Red("Cannot save session to %v\n%v", c.path, err.Error())
	}
	return
}

// TestConnection attempts to fetch a problem page and returns the response
func (c *Client) TestConnection() error {
	if c == nil {
		return fmt.Errorf("Client is not initialized")
	}

	// Try to fetch a simple problem page
	testURL := c.host + "/contest/1/problem/A"
	resp, err := c.client.Get(testURL)
	if err != nil {
		return fmt.Errorf("Failed to connect: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response: %v", err)
	}

	// Check if we need to login
	if _, err := findHandle(body); err != nil && err.Error() == ErrorNotLogged {
		color.Yellow("Not logged in. Attempting to login...")
		if err := c.Login(); err != nil {
			return fmt.Errorf("Login failed: %v", err)
		}
		// Try the request again after login
		resp, err = c.client.Get(testURL)
		if err != nil {
			return fmt.Errorf("Failed to connect after login: %v", err)
		}
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Failed to read response after login: %v", err)
		}
	}

	// Check if we got the Cloudflare challenge page
	if bytes.Contains(body, []byte("Just a moment...")) || 
	   bytes.Contains(body, []byte("cf-browser-verification")) ||
	   bytes.Contains(body, []byte("challenge-platform")) {
		return fmt.Errorf("Still getting Cloudflare challenge page")
	}

	// Look for typical Codeforces problem indicators
	if !bytes.Contains(body, []byte("problemset")) && 
	   !bytes.Contains(body, []byte("problem-statement")) {
		return fmt.Errorf("Response doesn't look like a Codeforces problem page")
	}

	color.Green("Successfully connected to Codeforces!")
	return nil
}
