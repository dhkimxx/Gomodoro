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

const (
	LABEL_START  = "Start"
	LABEL_PAUSE  = "Pause"
	LABEL_RESUME = "Resume"
	LABEL_RESET  = "Reset"
)

type PomodoroApp struct {
	app          fyne.App
	window       fyne.Window
	timerLabel   *widget.Label
	stateLabel   *widget.Label
	sessionLabel *widget.Label
	startBtn     *widget.Button
	resetBtn     *widget.Button
	state        PomodoroState
	session      int
	remaining    time.Duration
	running      bool
	pause        bool
	ticker       *time.Ticker
}

func NewPomodoroApp() *PomodoroApp {
	a := app.New()
	w := a.NewWindow("Pomodoro")

	p := &PomodoroApp{
		app:          a,
		window:       w,
		timerLabel:   widget.NewLabel("25:00"),
		stateLabel:   widget.NewLabel("Focus"),
		sessionLabel: widget.NewLabel("Session: 1"),
		startBtn:     widget.NewButton(LABEL_START, nil),
		resetBtn:     widget.NewButton(LABEL_RESET, nil),
		state:        Focus,
		session:      1,
		remaining:    25 * time.Minute,
	}

	p.startBtn.OnTapped = p.onStartTapped
	p.resetBtn.OnTapped = p.onResetTapped
	p.setupUI()

	return p
}

func (p *PomodoroApp) setupUI() {
	ui := container.NewVBox(
		p.stateLabel,
		p.timerLabel,
		p.startBtn,
		p.resetBtn,
		p.sessionLabel,
	)
	p.window.SetContent(ui)
}

func (p *PomodoroApp) updateUI() {
	fyne.Do(func() {
		p.timerLabel.SetText(formatTimer(p.remaining))
		switch p.state {
		case Focus:
			p.stateLabel.SetText("Focus")
		case ShortBreak:
			p.stateLabel.SetText("Short Break")
		case LongBreak:
			p.stateLabel.SetText("Long Break")
		}
		p.sessionLabel.SetText(fmt.Sprintf("Session: %d", p.session))
	})
}

func (p *PomodoroApp) onStartTapped() {
	if p.running {
		if p.pause {
			p.pause = false
			p.startBtn.SetText(LABEL_PAUSE)
			p.ticker = time.NewTicker(time.Second)
			go p.startTimer()
		} else {
			p.ticker.Stop()
			p.pause = true
			p.startBtn.SetText(LABEL_RESUME)
		}
	} else {
		p.running = true
		p.startBtn.SetText(LABEL_PAUSE)
		p.ticker = time.NewTicker(time.Second)
		go p.startTimer()
	}
}

func (p *PomodoroApp) onResetTapped() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
	p.state = Focus
	p.session = 1
	p.remaining = 25 * time.Minute
	p.running = false
	p.pause = false
	p.startBtn.SetText(LABEL_START)
	p.updateUI()
}

func (p *PomodoroApp) startTimer() {
	for range p.ticker.C {
		p.remaining -= time.Second
		p.updateUI()
		if p.remaining <= 0 {
			p.ticker.Stop()
			p.running = false
			p.startBtn.SetText(LABEL_START)
			if p.state == Focus {
				p.session++
				if p.session%4 == 0 {
					p.state = LongBreak
					p.remaining = 15 * time.Minute
				} else {
					p.state = ShortBreak
					p.remaining = 5 * time.Minute
				}
			} else {
				p.state = Focus
				p.remaining = 25 * time.Minute
			}
			p.updateUI()
		}
	}
}

func formatTimer(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func main() {
	p := NewPomodoroApp()
	p.window.ShowAndRun()
}
