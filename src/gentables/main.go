package main

import (
	"context"
	"flag"
	"log"
	"math"
	"os"

	"github.com/jinjor/desktop-audio/src/audio"
	"golang.org/x/sync/errgroup"
)

const numTables = 128
const numSamples = 4096

func main() {
	flag.Parse()
	dir := flag.Arg(0)
	if dir == "" {
		panic("dir is not passed")
	}
	log.SetFlags(log.Lshortfile)

	os.MkdirAll(dir, os.ModePerm)

	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		wts := audio.NewWavetableSet(numTables, numSamples)
		wts.MakeBandLimitedTablesForAllNotes(numSamples, calcPartialSquareAtPhase)
		log.Println("generated square wave")
		err := wts.Save(dir + "/square")
		log.Println("saved square wave")
		return err
	})
	g.Go(func() error {
		wts := audio.NewWavetableSet(numTables, numSamples)
		wts.MakeBandLimitedTablesForAllNotes(numSamples, calcPartialSawAtPhase)
		log.Println("generated saw wave")
		err := wts.Save(dir + "/saw")
		log.Println("saved saw wave")
		return err
	})
	err := g.Wait()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	log.Println("Successfully generated wavetables.")
}

func calcPartialSquareAtPhase(n int, phase float64) float64 {
	if n%2 == 1 {
		x := float64(n)
		return math.Sin(x*phase) / x
	}
	return 0.0
}
func calcPartialSawAtPhase(n int, phase float64) float64 {
	x := float64(n)
	return math.Sin(x*phase) / x
}
