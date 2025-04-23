package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type PomodoroState int

const (
	Focus PomodoroState = iota
	ShortBreak
	LongBreak
)

func format(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func NewTicker() *time.Ticker {
	// return time.NewTicker(time.Millisecond)
	return time.NewTicker(time.Second)
}

func main() {
	a := app.New()
	w := a.NewWindow("Pomodoro")

	timerLabel := widget.NewLabel("25:00")
	stateLabel := widget.NewLabel("Focus")
	sessionLabel := widget.NewLabel("Session: 1")

	startBtn := widget.NewButton("Start", func() {})

	state := Focus
	session := 1
	remaining := 25 * time.Minute
	running := false
	var ticker *time.Ticker

	pause := false

	update := func() {
		timerLabel.SetText(format(remaining))
		switch state {
		case Focus:
			stateLabel.SetText("Focus")
		case ShortBreak:
			stateLabel.SetText("Short Break")
		case LongBreak:
			stateLabel.SetText("Long Break")
		}
		sessionLabel.SetText(fmt.Sprintf("Session: %d", session))
	}

	startTimer := func() {
		for range ticker.C {
			remaining -= time.Second

			fyne.Do(update)
			if remaining <= 0 {
				ticker.Stop()
				running = false
				fyne.Do(func() {
					startBtn.SetText("Start")
				})
				if state == Focus {
					session++
					if session%4 == 0 {
						state = LongBreak
						remaining = 15 * time.Minute
					} else {
						state = ShortBreak
						remaining = 5 * time.Minute
					}
				} else {
					state = Focus
					remaining = 25 * time.Minute
				}
				fyne.Do(update)
			}
		}
	}

	startBtn.OnTapped = func() {
		if running {
			if pause {
				pause = false
				startBtn.SetText("Pause")
				ticker = NewTicker()
				go startTimer()
			} else {
				ticker.Stop()
				pause = true
				startBtn.SetText("Resume")
				return
			}
		} else {
			running = true
			startBtn.SetText("Pause")
			ticker = NewTicker()
			go startTimer()
		}
	}

	resetBtn := widget.NewButton("Reset", func() {
		if ticker != nil {
			ticker.Stop()
		}
		state = Focus
		session = 1
		remaining = 25 * time.Minute
		running = false
		fyne.Do(update)
	})

	ui := container.NewVBox(
		stateLabel,
		timerLabel,
		startBtn,
		resetBtn,
		sessionLabel,
	)

	fyne.Do(update)
	w.SetContent(ui)
	w.ShowAndRun()
}
