package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"net/http"
	"net/url"
	"runtime"
	"sync"
)

var proxyFetchSemaphore = make(chan int, 1)
var proxyCheckWaitGroup sync.WaitGroup

type ProxyStatus struct {
	ok bool
	proxy string
	errorMessage string
}

func (status ProxyStatus) String() string {
	var output string
	if (status.ok) { output += "OK" } else { output += "ERROR" } 
	output += " " + status.proxy
	if (!status.ok) { output += " " + status.errorMessage }
	return output
}

func checkProxy(proxy string) (success bool, errorMessage string) {
	fmt.Println("Checking:", proxy)
	
	proxyUrl, err := url.Parse("http://" + proxy)
    httpClient := &http.Client { Transport: &http.Transport { Proxy: http.ProxyURL(proxyUrl) } }
    response, err := httpClient.Get("http://stackoverflow.com")
	if err != nil { return false, err.Error() }

	body, err := ioutil.ReadAll(response.Body)
	if err != nil { return false, err.Error() }

	bodyString := string(body)

	var bodyIndex = strings.Index(bodyString, "<body ")
	if bodyIndex < 0 {
		return false, "Reveived page is not HTML"
	}

	return true, ""
}

func asyncCheckProxy(proxyInfoChan chan ProxyStatus, proxy string) {
	success, errorMessage := checkProxy(proxy)

	var info ProxyStatus
	info.proxy = proxy
	info.ok = success
	info.errorMessage = errorMessage
	
	proxyInfoChan <- info

	proxyCheckWaitGroup.Done()
}

func checkResults(proxyInfoChan chan ProxyStatus) {
	for {
		status := <- proxyInfoChan
		fmt.Println(status)
	}
}

func main() {
	var _ = fmt.Print

	runtime.GOMAXPROCS(runtime.NumCPU())

	content, err := ioutil.ReadFile("proxy.txt")
	if err != nil {
	    panic(err.Error())
	}
	lines := strings.Split(string(content), "\n")

	var proxies[] string

	for i := 0; i < len(lines); i++ {
		var line = lines[i]
		var tokens = strings.Split(line, " ")
		if len(tokens) <= 0 { continue }
		var proxy = strings.Trim(tokens[0], " \t")
		if len(proxy) < 7 { continue }

		proxies = append(proxies, proxy)
	}

	proxyInfoChan := make(chan ProxyStatus, 10)

	for i := 0; i < len(proxies); i++ {
		proxy := proxies[i]
		proxyCheckWaitGroup.Add(1)
		go asyncCheckProxy(proxyInfoChan, proxy)
	}

	go checkResults(proxyInfoChan)

	proxyCheckWaitGroup.Wait()
}