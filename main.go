package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// for test only

func main() {
	const (
		endpoint   = "/random"
		configFile = "ibmcos.yml"
	)

	imageServer := ImageServer{}
	err := imageServer.Configure(configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = imageServer.FetchFileKeys()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		u, _, e := imageServer.GetRandomUrl(time.Minute)
		if e != nil {
			w.WriteHeader(500)
		}
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Location", u)
		w.WriteHeader(307)
	})
	err = http.ListenAndServe("0.0.0.0:8000", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
}
