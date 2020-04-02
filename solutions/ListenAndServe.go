package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// main function: new http server program, listening on port 80,
// to return current Unix time (epoch) when queried
func main() {
	timeHandler := func(w http.ResponseWriter, req *http.Request) {
		currentTime := time.Now().Unix()
		io.WriteString(w, "Current Unix time: "+strconv.Itoa(int(currentTime)))
	}
	fmt.Println("Application connecting to port 8080...")
	http.HandleFunc("/", timeHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
