package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Endpoint /success accessed")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Success")
	})

	http.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Endpoint /fail accessed")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
	})

	http.HandleFunc("/with-body", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Endpoint /with-body accessed")
		body := r.Body
		defer body.Close()
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Received body")
	})

	http.HandleFunc("/with-header", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Endpoint /with-header accessed")
		if r.Header.Get("X-Test-Header") != "" {
			log.Println("X-Test-Header received")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Header received")
		} else {
			log.Println("X-Test-Header missing")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Header missing")
		}
	})

	http.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Endpoint /slow accessed")
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Slow response")
	})

	log.Println("Starting test server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
