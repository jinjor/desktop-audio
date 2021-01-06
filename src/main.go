package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/jinjor/desktop-audio/src/audio"
)

const sockFileName = "/tmp/desktop-audio.sock"

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)
	log.Printf("NumCPU: %v\n", runtime.NumCPU())

	commandCh := make(chan []string, 256)
	go audio.Loop(commandCh)

	os.Remove(sockFileName)
	listener, err := net.Listen("unix", sockFileName)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	cleanupSocket := func() {
		listener.Close()
		os.Remove(sockFileName)
	}
	defer cleanupSocket()
	handleSigs(cleanupSocket)
	log.Printf("start listening...\n")

	// hanlde one long connection
	err = handleConn(&listener, commandCh)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func handleConn(listener *net.Listener, commandCh chan []string) error {
	conn, err := (*listener).Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	var line []byte
	for {
		next, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		line = append(line, next...)
		if isPrefix {
			continue
		}
		lineStr := strings.Split(string(line), " ")
		for i, item := range lineStr {
			lineStr[i], err = url.QueryUnescape(item)
			if err != nil {
				return err
			}
		}
		commandCh <- lineStr

		log.Printf("received: %s\n", string(line))
		conn.Write(append(line, "\n"...))
		line = []byte{}
	}
	// max 64KB
	// scanner := bufio.NewScanner(conn)
	// for scanner.Scan() {
	// 	text := scanner.Text()
	// 	log.Printf("received: %s\n", text)
	// 	conn.Write([]byte(text + "\n"))
	// }
	// if err := scanner.Err(); err != nil {
	// 	return err
	// }
	return nil
}

func handleSigs(beforeExit func()) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		beforeExit()
		os.Exit(0)
	}(sigc)
}
