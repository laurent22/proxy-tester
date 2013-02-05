package main

import (
	"fmt"
	"runtime"
	"net/http"
	"io/ioutil"
	"strings"
)

func main() {
	var _ = fmt.Print
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	proxyListTemplateUrl := "http://www.samair.ru/proxy/proxy-%02v.htm"
	
	pageIndex := 1
	for {
		proxyListUrl := fmt.Sprintf(proxyListTemplateUrl, pageIndex)
		fmt.Println("# Getting", proxyListUrl)
		response, err := http.Get(proxyListUrl)
		if err != nil {
			// Most likely, all the pages have been downloaded
			fmt.Println("# ", err.Error())
			break
		}
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("# ", err.Error())
			continue
		}
		bodyString := string(body)
		
		openingString := "<strong>Country</strong></a>"
		closingString := "<div style=\"width: 780px;"
		openingIndex := strings.Index(bodyString, openingString)
		if openingIndex < 0 {
			fmt.Println("# Couldn't find proxy list (missing opening string):", proxyListUrl)
			break
		}
		closingIndex := strings.Index(bodyString, closingString)
		if closingIndex < 0 {
			fmt.Println("# Couldn't find proxy list (missing closing string):", proxyListUrl)
			break
		}
		openingIndex += len(openingString)
		
		proxyListString := strings.Trim(bodyString[openingIndex:closingIndex], " \n\r\t")
		
		fmt.Println(proxyListString)
		
		pageIndex++
	}
}
