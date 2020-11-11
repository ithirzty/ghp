package main

func generateServer() string {
	return `
package main

import (
	----importpackages----
)

type configStruct struct {
	Port             int
	enableSSL        bool
	PortSSL          int
	CertFullChainSSL string
	CertPrivKeySSL   string
	Index            string
}

var config configStruct = configStruct{80, false, 443, "", "", "index.html"}
var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}
var netClient = &http.Client{
	Timeout:   time.Second * 5,
	Transport: netTransport,
}

func main() {
	configFile, err := ioutil.ReadFile("./config.json")
	if err != nil {
		configJSONformat, _ := json.Marshal(config)
		ioutil.WriteFile("./config.json", configJSONformat, 0644)
		fmt.Println("Generated config. Restart for changes to take effect.")
	}
	json.Unmarshal(configFile, &config)
	serverRunning := make(chan bool)
	http.HandleFunc("/", server)
	go func() {
		fmt.Println("launched")
		if config.enableSSL == true {
			http.ListenAndServeTLS(":"+strconv.Itoa(config.PortSSL), config.CertFullChainSSL, config.CertPrivKeySSL, nil)
		} else {
			http.ListenAndServe(":"+strconv.Itoa(config.Port), nil)
		}
	}()
	<-serverRunning
}

func server(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path[1:] {
	default:
		file, _ := ioutil.ReadFile(r.URL.Path[1:])
		w.Write(file)
`
}
