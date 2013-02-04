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

type ProxyStatus struct {
	ok bool
	proxy string
	errorMessage string
	downloadedUrl string
}

var proxyFetchSemaphore = make(chan int, 1)
var proxyCheckWaitGroup sync.WaitGroup
var proxyStatuses[] ProxyStatus

func (status ProxyStatus) String() string {
	var output string
	if (status.ok) { output += "OK" } else { output += "ERROR" } 
	output += " " + status.proxy
	output += " " + status.downloadedUrl
	if (!status.ok) { output += " " + status.errorMessage }
	return output
}

func checkProxy(proxy string, downloadedUrl string) (success bool, errorMessage string) {
	fmt.Println("Checking:", proxy, downloadedUrl)
	
	proxyUrl, err := url.Parse("http://" + proxy)
    httpClient := &http.Client { Transport: &http.Transport { Proxy: http.ProxyURL(proxyUrl) } }
    response, err := httpClient.Get(downloadedUrl)
	if err != nil { return false, err.Error() }

	body, err := ioutil.ReadAll(response.Body)
	if err != nil { return false, err.Error() }

	bodyString := string(body)

	if strings.Index(bodyString, "<body") < 0 && strings.Index(bodyString, "<head") < 0 {
		return false, "Reveived page is not HTML"
	}

	return true, ""
}

func asyncCheckProxy(proxyInfoChan chan ProxyStatus, proxy string, downloadedUrl string) {
	success, errorMessage := checkProxy(proxy, downloadedUrl)

	var info ProxyStatus
	info.proxy = proxy
	info.ok = success
	info.errorMessage = errorMessage
	info.downloadedUrl = downloadedUrl
	
	proxyInfoChan <- info
}

func checkResults(proxyInfoChan chan ProxyStatus) {
	for {
		status := <- proxyInfoChan
		proxyStatuses = append(proxyStatuses, status)
		fmt.Println(status)
		proxyCheckWaitGroup.Done()
	}
}

func main() {
	var _ = fmt.Print

	runtime.GOMAXPROCS(runtime.NumCPU())

	content, err := ioutil.ReadFile("proxy.short.txt")
	if err != nil {
	    panic(err.Error())
	}
	lines := strings.Split(string(content), "\n")

	var proxies[] string

	for i := 0; i < len(lines); i++ {
		var line = lines[i]
		var tokens = strings.Split(line, " ")
		if len(tokens) <= 0 { continue }
		var proxy = strings.Trim(tokens[0], " \t\n\r")
		if len(proxy) < 7 { continue }

		proxies = append(proxies, proxy)
	}
	
	content, err = ioutil.ReadFile("trackers.short.txt")
	if err != nil {
	    panic(err.Error())
	}
	lines = strings.Split(string(content), "\n")
	
	var trackers[] string

	for i := 0; i < len(lines); i++ {
		var line = lines[i]
		var tokens = strings.Split(line, " ")
		if len(tokens) <= 0 { continue }
		var tracker = strings.Trim(tokens[0], " \t\n\r")
		if len(tracker) < 7 { continue }

		trackers = append(trackers, tracker)
	}

	proxyInfoChan := make(chan ProxyStatus, 10)

	for i := 0; i < len(proxies); i++ {
		proxy := proxies[i]
		for j := 0; j < len(trackers); j++ {
			proxyCheckWaitGroup.Add(1)
			go asyncCheckProxy(proxyInfoChan, proxy, trackers[j])
		}
	}

	go checkResults(proxyInfoChan)

	proxyCheckWaitGroup.Wait()
	
	fmt.Println("======================================");
	for i := 0; i < len(proxyStatuses); i++ {
		fmt.Println(proxyStatuses[i])	
	}
}