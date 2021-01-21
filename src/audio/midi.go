package audio

import (
	"context"
	"log"

	"gitlab.com/gomidi/rtmididrv"
)

// ListenToMidiIn ...
func ListenToMidiIn(ctx context.Context) <-chan []byte {
	ch := make(chan []byte, 65536)
	go func() {
		drv, err := rtmididrv.New()
		if err != nil {
			log.Printf("failed to initialize MIDI driver: %v\n", err)
			return
		}
		defer func() {
			err := drv.Close()
			if err != nil {
				log.Printf("failed to close MIDI driver: %v\n", err)
			}
		}()
		ins, err := drv.Ins()
		if err != nil {
			log.Printf("failed to get MIDI IN: %v\n", err)
			return
		}
		log.Printf("MIDI IN: %v\n", ins)

		if len(ins) == 0 {
			log.Println("WARN: MIDI IN not fonud")
			return
		}
		in := ins[0]
		if err := in.Open(); err != nil {
			log.Printf("failed to open MIDI IN: %v\n", err)
			return
		}
		log.Println("opened " + in.String())
		defer func() {
			err := in.Close()
			if err != nil {
				log.Printf("failed to close MIDI IN: %v\n", err)
			}
		}()
		log.Println("start listening MIDI IN...")
		if err := in.SetListener(func(data []byte, deltaMicroseconds int64) {
			ch <- data
		}); err != nil {
			log.Println("failed to set listener: " + err.Error())
		}
		defer func() {
			log.Println("stop listening MIDI IN...")
			err := in.StopListening()
			if err != nil {
				log.Printf("failed to stop listening: %v\n", err)
			}
		}()
		defer close(ch)
		<-ctx.Done()
	}()
	return ch
}
