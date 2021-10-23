package main

import (
	"embed"
	"math/rand"
	"os/exec"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/getlantern/systray"
)

var (
	//go:embed assets/favicon.ico
	icon []byte

	//go:embed assets/audio
	sounds embed.FS
)

func main() {

	audio, err := sounds.Open("arc-sound-effect.mp3")
	if err != nil {
		panic(err)
	}
	defer audio.Close()

	streamer, format, err := mp3.Decode(audio)
	if err != nil {
		panic(err)
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		panic(err)
	}
	defer speaker.Clear()

	play(streamer)
	systray.Run(onReady, onExit)
}

func play(streamer beep.StreamSeekCloser) {
	go func() {
		for {
			speaker.Play(streamer)
			streamer.Seek(0)
			time.Sleep(time.Duration(rand.Intn(60) * int(time.Minute)))
		}
	}()
}

func onReady() {
	systray.SetIcon(icon)
	systray.SetTitle("Neco arc sound player")
	systray.SetTooltip("Randomly plays neco-arc's sounds over time")
	quitBtn := systray.AddMenuItem("Stop", "Stops the whole app")
	donateBtn := systray.AddMenuItem("Donate", "I appreciate your support")

	go func() {
		for {
			select {

			case <-quitBtn.ClickedCh:
				systray.Quit()
			case <-donateBtn.ClickedCh:
				cmd := exec.Command("cmd", "/C", "start", "https://paypal.me/elesneils")
				cmd.Start()
			}
		}
	}()

}

func onExit() {

}
