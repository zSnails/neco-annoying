package main

import (
	"embed"
	"flag"
	"math/rand"
	"net/url"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/sirupsen/logrus"
)

var (
	//go:embed assets/favicon.ico
	icon []byte

	//go:embed assets/neco.png
	iconPng []byte

	//go:embed assets/audio/*
	soundsFs embed.FS

	r = rand.New(rand.NewSource(time.Now().UnixNano()))

	outputFile string
	maxTime    int
	mApp       = app.NewWithID("Neco-Annoying")
	mWindow    = mApp.NewWindow("Neco Arc Player")
	volume     = binding.NewFloat()
	iconRes    = fyne.NewStaticResource("icon", iconPng)
)

func init() {
	flag.StringVar(&outputFile, "output", "output.log", "set the output file")
	flag.IntVar(&maxTime, "max-time", 15, "set the max time between each audio")
	flag.Parse()
}

func main() {
	logrus.SetLevel(logrus.ErrorLevel)
	logOut, err := os.OpenFile(outputFile, os.O_APPEND|os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.SetOutput(logOut)

	volPref := mApp.Preferences().FloatWithFallback("app-volume", 0)

	volume.Set(volPref)

	volumeSlider := widget.NewSliderWithData(-10, 0, volume)
	volumeSlider.Step = 0.01
	volumeSlider.SetValue(0)

	mWindow.SetIcon(iconRes)

	mWindow.Resize(fyne.NewSize(300, 150))
	mWindow.SetFixedSize(true)
	mWindow.SetContent(container.New(layout.NewMaxLayout(), widget.NewLabelWithData(binding.FloatToString(volume)), volumeSlider))
	mWindow.SetCloseIntercept(func() {
		mWindow.Hide()
	})

	if desk, ok := mApp.(desktop.App); ok {
		menu := fyne.NewMenu("Neco Arc Random Sounds", fyne.NewMenuItem("Manage volume", func() {
			mWindow.Show()
			mWindow.RequestFocus()
		}), fyne.NewMenuItem("Donate", func() {
			u, err := url.Parse("https://paypal.me/elesneils")
			if err != nil {
				logrus.Error(err)
			}
			err = mApp.OpenURL(u)
			if err != nil {
				logrus.Error(err)
			}
		}))
		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(iconRes)
	}

	go playAudio()

	mApp.Run()
}

func playAudio() {
	vol, err := volume.Get()
	if err != nil {
		logrus.Fatal(err)
	}

	volumeManager := effects.Volume{
		Base:   2,
		Volume: vol,
		Silent: false,
	}

	volume.AddListener(binding.NewDataListener(func() {
		vol, err := volume.Get()
		if err != nil {
			logrus.Fatal(err)
		}
		speaker.Lock()
		volumeManager.Volume = vol
		speaker.Unlock()
		mApp.Preferences().SetFloat("app-volume", vol)
	}))

	audioFolder, _ := soundsFs.ReadDir("assets/audio")
	for {
		file := audioFolder[r.Intn(len(audioFolder))].Name()
		audio, err := soundsFs.Open("assets/audio/" + file)
		if err != nil {
			logrus.Error(err)
			continue
		}
		streamer, format, err := mp3.Decode(audio)
		if err != nil {
			logrus.Error(err)
			continue
		}
		logrus.Infof("Playing audio %v", file)

		if err := streamer.Err(); err != nil {
			logrus.Error(err)
			continue
		}

		volumeManager.Streamer = streamer
		play(volumeManager, streamer, format)
		time.Sleep(time.Duration(r.Intn(maxTime) * int(time.Minute)))
	}

}

func play(manager effects.Volume, streamer beep.StreamSeekCloser, format beep.Format) {

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		logrus.Error(err)
		return
	}
	done := make(chan bool)
	speaker.Play(beep.Seq(&manager, beep.Callback(func() {
		done <- true
	})))
	<-done
}
