package main

import (
	"embed"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/getlantern/systray"
	"github.com/sirupsen/logrus"
)

var (
	//go:embed assets/favicon.ico
	icon []byte

	//go:embed assets/audio/*
	soundsFs embed.FS

	seed = rand.NewSource(time.Now().UnixNano())
	r    = rand.New(seed)
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	logOut, err := os.Create("output.log")
	if err != nil {
		logrus.Panic(err)
	}

	logrus.SetOutput(logOut)

	audioFolder, _ := soundsFs.ReadDir("assets/audio")
	go func() {
		for {
			idx := r.Intn(len(audioFolder))
			t := time.Duration(r.Intn(15) * int(time.Minute))
			file := audioFolder[idx].Name()
			audio, err := soundsFs.Open("assets/audio/" + file)
			if err != nil {
				logrus.Panic(err)
			}
			streamer, format, err := mp3.Decode(audio)
			if err != nil {
				logrus.Panic(err)
			}
			logrus.Debugf("Playing audio %v", file)
			play(streamer, format)
			time.Sleep(t)
		}
	}()
	systray.Run(onReady, onExit)
}

func play(streamer beep.StreamSeekCloser, format beep.Format) {
	err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		logrus.Panic(err)
	}

	defer speaker.Close()
	defer streamer.Close()
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	<-done
	err = streamer.Err()
	if err != nil {
		logrus.Panic(err)
	}
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
		logrus.Panic(err)
	}
}

func onExit() {
	logrus.Info("Exited program")
}
