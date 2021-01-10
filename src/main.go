package main

import (
	"bufio"
	"context"
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
	"golang.org/x/sync/errgroup"
)

const sockFileName = "/tmp/desktop-audio.sock"

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)
	log.Printf("NumCPU: %v\n", runtime.NumCPU())

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	audio, err := audio.NewAudio()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	defer audio.Close()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer func() {
		signal.Stop(signalCh)
		cancel()
	}()
	go func() {
		sig := <-signalCh
		log.Printf("Caught signal %s: shutting down...\n", sig)
		cancel()
	}()
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return audio.Start(ctx)
	})
	g.Go(func() error {
		return receiveCommands(ctx, audio.CommandCh)
	})
	if err := g.Wait(); err != nil {
		log.Fatalf("error: %v\n", err)
	}
	log.Println("main() ended.")
}

func receiveCommands(ctx context.Context, commandCh chan<- []string) error {
	os.Remove(sockFileName)
	listener, err := new(net.ListenConfig).Listen(ctx, "unix", sockFileName)
	if err != nil {
		return err
	}
	defer func() {
		log.Println("Closeing IPC...")
		listener.Close()
		os.Remove(sockFileName)
	}()
	log.Printf("start listening...\n")
	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	var line []byte
loop:
	for {
		select {
		case <-ctx.Done():
			log.Println("Connection interrupted")
			break loop
		default:
		}
		next, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			break loop
		}
		if err != nil {
			return err
		}
		line = append(line, next...)
		if isPrefix {
			continue
		}
		command, err := parseCommand(string(line))
		if err != nil {
			return err
		}
		commandCh <- command

		log.Printf("received: %s\n", string(line))
		conn.Write(append(line, "\n"...))
		line = []byte{}
	}
	log.Println("receiveCommands() ended.")
	return nil
}

func parseCommand(line string) ([]string, error) {
	lineStr := strings.Split(line, " ")
	for i, item := range lineStr {
		escaped, err := url.QueryUnescape(item)
		if err != nil {
			return nil, err
		}
		lineStr[i] = escaped
	}
	return lineStr, nil
}
