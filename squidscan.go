package main

import (
        "encoding/base64"
        "fmt"
        "io/ioutil"
        "net"
        "net/http"
        "net/url"
        "strings"
        "sync"
        "time"

        "github.com/cheggaaa/pb/v3"
)

var (
        proxyURL   = "http://192.168.1.3:3128" // adjust proxy IP & port
        numWorkers = 100                       // adjust workers
        numPorts   = 65535                     // adjust ports
        username   = "username"           // replace with actual username
        password   = "password"           // replace with actual password
)

func main() {
        proxyURL, err := url.Parse(proxyURL)
        if err != nil {
                fmt.Printf("Failed to parse proxy URL: %v\n", err)
                return
        }
        transport := &http.Transport{
                Proxy: http.ProxyURL(proxyURL),
                DialContext: (&net.Dialer{
                        Timeout:   3 * time.Second,
                        KeepAlive: 3 * time.Second,
                }).DialContext,
        }
        client := &http.Client{Transport: transport}
        openPorts := make([]int, 0)

        bar := pb.StartNew(numPorts)
        sem := make(chan struct{}, numWorkers)
        var wg sync.WaitGroup
        for port := 1; port <= numPorts; port++ {
                wg.Add(1)
                go func(p int) {
                        defer wg.Done()
                        sem <- struct{}{}
                        defer func() {
                                <-sem
                                bar.Increment()
                        }()

                        address := fmt.Sprintf("127.0.0.1:%d", p)
                        req, err := http.NewRequest("GET", fmt.Sprintf("http://%s", address), nil)
                        if err != nil {
                                return
                        }
                        req.Header.Add("Proxy-Authorization", basicAuth(username, password))

                        r, err := client.Do(req)
                        if err != nil {
                                return
                        }
                        defer r.Body.Close()

                        data, _ := ioutil.ReadAll(r.Body)
                        dataStr := string(data)
                        if strings.Contains(dataStr, "The requested URL could not be retrieved") {
                                return
                        }

                        openPorts = append(openPorts, p)
                        fmt.Printf("Port %d found!\n", p)

                }(port)
        }
        wg.Wait()
        bar.Finish()

        fmt.Println("Open ports:")
        for _, port := range openPorts {
                fmt.Println(port)
        }
}

// basicAuth generates the Base64-encoded string for Basic Authentication.
func basicAuth(username, password string) string {
        auth := username + ":" + password
        return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
