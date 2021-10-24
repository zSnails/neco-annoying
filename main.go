package main

import (
	"embed"
	"flag"
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

	r = rand.New(rand.NewSource(time.Now().UnixNano()))

	outputFile string
	maxTime    int
)

func init() {
	flag.StringVar(&outputFile, "output", "output.log", "set the output file")
	flag.IntVar(&maxTime, "max-time", 15, "set the max time between each audio")
	flag.Parse()
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logOut, err := os.OpenFile(outputFile, os.O_APPEND|os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.SetOutput(logOut)
	audioFolder, _ := soundsFs.ReadDir("assets/audio")
	go func() {
		for {
			file := audioFolder[r.Intn(len(audioFolder))].Name()
			audio, err := soundsFs.Open("assets/audio/" + file)
			if err != nil {
				logrus.Error(err)
				return
			}
			streamer, format, err := mp3.Decode(audio)
			if err != nil {
				logrus.Error(err)
				return
			}
			logrus.Infof("Playing audio %v", file)
			play(streamer, format)
			time.Sleep(time.Duration(r.Intn(maxTime) * int(time.Minute)))
		}
	}()
	systray.Run(onReady, nil)
}

func play(streamer beep.StreamSeekCloser, format beep.Format) {
	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		logrus.Error(err)
		return
	}
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	<-done
	if err := streamer.Err(); err != nil {
		logrus.Error(err)
		return
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
		logrus.Error(err)
	}
}
