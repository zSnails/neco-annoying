package main

import (
	"embed"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"runtime"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/getlantern/systray"
)

var (
	//go:embed assets/favicon.ico
	icon []byte

	//go:embed assets/audio/*
	soundsFs embed.FS

	files []string

	seed = rand.NewSource(time.Now().UnixNano())
	r    = rand.New(seed)
)

func main() {
	audioFolder, _ := soundsFs.ReadDir("assets/audio")

	for _, k := range audioFolder {
		files = append(files, k.Name())
	}

	go func() {
		for {
			audio, err := soundsFs.Open("assets/audio/" + files[r.Intn(len(files))])
			if err != nil {
				panic(err)
			}
			streamer, format, err := mp3.Decode(audio)
			if err != nil {
				panic(err)
			}
			play(streamer, format)
		}
	}()
	systray.Run(onReady, onExit)
}

func play(streamer beep.StreamSeekCloser, format beep.Format) {
	err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		panic(err)
	}
	defer speaker.Close()
	t := time.Duration(r.Intn(15) * int(time.Minute))
	speaker.Play(streamer)
	streamer.Seek(0)
	time.Sleep(t)
	streamer.Close()
}

func onReady() {
	systray.SetIcon(icon)
	systray.SetTitle("Neco arc sound player")
	systray.SetTooltip("Randomly plays neco-arc's sounds over time")
	quitBtn := systray.AddMenuItem("Stop", "Stops the whole app")
	systray.AddSeparator()
	donateBtn := systray.AddMenuItem("Donate", "I appreciate your support")

	go func() {
		for {
			select {
			case <-quitBtn.ClickedCh:
				systray.Quit()
			case <-donateBtn.ClickedCh:
				openbrowser("https://paypal.me/elesneils")
			}
		}
	}()

}

func openbrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func onExit() {
	speaker.Clear()
	speaker.Close()
}
