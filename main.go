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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/getlantern/systray"
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
	mainApp    = app.NewWithID("neco-annoying")
	w          = mainApp.NewWindow("Neco Annoying")
	volume     = binding.NewFloat()
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

	volPref := mainApp.Preferences().FloatWithFallback("app-volume", 0)
	volume.Set(volPref)
	w.SetIcon(fyne.NewStaticResource("icon", iconPng))
	volumeSlider := widget.NewSliderWithData(-10, 0, volume)
	volumeSlider.Step = 0.01
	volumeSlider.SetValue(0)
	w.SetContent(container.New(layout.NewMaxLayout(), widget.NewLabelWithData(binding.FloatToString(volume)), volumeSlider))
	w.SetCloseIntercept(func() {
		w.Hide()
	})
	w.Resize(fyne.NewSize(300, 150))
	w.SetFixedSize(true)
	go playAudio()
	systray.Register(onReady, nil)
	mainApp.Run()
}

func playAudio() {
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
		if err != nil {
			logrus.Error(err)
			continue
		}
		play(streamer, format)
		time.Sleep(time.Duration(r.Intn(maxTime) * int(time.Minute)))
	}
}

func play(streamer beep.StreamSeekCloser, format beep.Format) {
	vol, err := volume.Get()
	volumeManager := effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   vol,
		Silent:   false,
	}
	volume.AddListener(binding.NewDataListener(func() {
		if err != nil {
			logrus.Error(err)
		}
		speaker.Lock()
		volumeManager.Volume = vol
		speaker.Unlock()
		mainApp.Preferences().SetFloat("app-volume", vol)
	}))
	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		logrus.Error(err)
		return
	}
	done := make(chan bool)
	speaker.Play(beep.Seq(&volumeManager, beep.Callback(func() {
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
	manageButton := systray.AddMenuItem("Show Window", "Opens the audio management window")
	donateBtn := systray.AddMenuItem("Donate", "I appreciate your support")
	systray.AddSeparator()
	quitBtn := systray.AddMenuItem("Stop", "Stops the whole app")
	go func() {
		for {
			select {
			case <-quitBtn.ClickedCh:
				systray.Quit()
				mainApp.Quit()
			case <-donateBtn.ClickedCh:
				// openbrowser()
				u, err := url.Parse("https://paypal.me/elesneils")
				if err != nil {
					logrus.Error(err)
				}
				err = mainApp.OpenURL(u)
				if err != nil {
					logrus.Error(err)
				}
			case <-manageButton.ClickedCh:
				w.Show()
				// w.RequestFocus()
			}
		}
	}()
}
