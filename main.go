package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/planerist/wordstats/stats"
	"net"
	"net/http"
	"os"
	"strconv"
)

const (
	tcp_listen_addr  = "localhost:9000"
	http_listen_addr = "localhost:8000"
)

func main() {
	stats := stats.NewStats()

	go listenTcp(stats)
	listenHttp(stats)
}

func listenTcp(stats *stats.Stats) {
	l, err := net.Listen("tcp", tcp_listen_addr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleTcp(conn, stats)
	}
}

func listenHttp(stats *stats.Stats) {
	http.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleHttp(w, r, stats)
	}))
	err := http.ListenAndServe(http_listen_addr, nil)
	if err != nil {
		fmt.Println("HTTP listener error: ", err.Error())
		os.Exit(1)
	}
}

func handleTcp(conn net.Conn, stats *stats.Stats) {
	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		stats.AppendWord(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading:", err)
	}
}

type JsonResult struct {
	Words []string `json:"top_words"`
}

func handleHttp(w http.ResponseWriter, r *http.Request, stats *stats.Stats) {
	nq := r.URL.Query().Get("N")
	if len(nq) == 0 {
		writeHttpError(w, http.StatusBadRequest)
		return
	}

	var (
		num int
		err error
	)
	if num, err = strconv.Atoi(nq); err != nil {
		writeHttpError(w, http.StatusBadRequest)
		return
	}

	responseCh := make(chan []string)
	stats.GetStats(responseCh, num)
	response := <-responseCh

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JsonResult{Words: response})
}

func writeHttpError(w http.ResponseWriter, errorCode int) {
	w.Header().Set("Content-Type", "application/text; charset=UTF-8")
	w.WriteHeader(errorCode)
	w.Write([]byte((fmt.Sprintf("%d %s", errorCode, http.StatusText(errorCode)))))
}
