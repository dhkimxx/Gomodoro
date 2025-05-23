package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/gen2brain/beeep"
)

// 개발 모드 플래그
var devMode bool

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

const (
	FOCUS_TIME       = 25
	SHORT_BREAK_TIME = 5
	LONG_BREAK_TIME  = 15
)

type PomodoroState int

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

	// 개발 모드일 때는 시간을 초 단위로 설정
	focusTime := FOCUS_TIME
	timeUnit := time.Minute

	if devMode {
		timeUnit = time.Second
		w.SetTitle("Pomodoro (Dev Mode)")
	}

	p := &PomodoroApp{
		app:          a,
		window:       w,
		timerLabel:   widget.NewLabel(fmt.Sprintf("%02d:00", focusTime)),
		stateLabel:   widget.NewLabel("Focus"),
		sessionLabel: widget.NewLabel("Session: 1"),
		startBtn:     widget.NewButton(LABEL_START, nil),
		resetBtn:     widget.NewButton(LABEL_RESET, nil),
		state:        Focus,
		session:      1,
		remaining:    time.Duration(focusTime) * timeUnit,
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

func (p *PomodoroApp) playNotification(message string) {
	// 시스템 알림 표시 (크로스 플랫폼)
	var err error
	switch runtime.GOOS {
	case "windows":
		err = beeep.Notify("Gomodoro", message, "")
	case "darwin":
		err = beeep.Notify("Gomodoro", message, "")
	default: // linux and others
		err = beeep.Notify("Gomodoro", message, "")
	}

	if err != nil {
		fmt.Printf("알림 오류 (%s): %v\n", runtime.GOOS, err)
		// 알림 실패시 다이얼로그로 표시
		fyne.Do(func() {
			dialog.ShowInformation("알림", message, p.window)
		})
	}

	// 시스템 알림음 재생 (크로스 플랫폼)
	if err := beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration); err != nil {
		fmt.Printf("알림음 오류 (%s): %v\n", runtime.GOOS, err)
	}
}

func (p *PomodoroApp) startTimer() {
	timeUnit := time.Minute
	if devMode {
		timeUnit = time.Second
	}

	for range p.ticker.C {
		p.remaining -= time.Second
		p.updateUI()

		if p.remaining <= 0 {
			p.ticker.Stop()

			fyne.Do(func() {
				p.running = false
				p.startBtn.SetText(LABEL_START)
			})

			// Show notification with sound
			var message string
			if p.state == Focus {
				message = "집중 시간이 끝났습니다. 휴식 시간을 가지세요!"
			} else {
				message = "휴식 시간이 끝났습니다. 다시 집중할 시간입니다!"
			}

			p.playNotification(message)

			if p.state == Focus {
				p.session++
				if p.session%4 == 0 {
					p.state = LongBreak
					p.remaining = time.Duration(LONG_BREAK_TIME) * timeUnit
				} else {
					p.state = ShortBreak
					p.remaining = time.Duration(SHORT_BREAK_TIME) * timeUnit
				}
			} else {
				p.state = Focus
				p.remaining = time.Duration(FOCUS_TIME) * timeUnit
			}
			p.updateUI()
		}
	}
}

func (p *PomodoroApp) onStartTapped() {
	if p.running {
		if p.pause {
			p.pause = false
			fyne.Do(func() {
				p.startBtn.SetText(LABEL_PAUSE)
			})
			p.ticker = time.NewTicker(time.Second)
			go p.startTimer()
		} else {
			p.ticker.Stop()
			p.pause = true
			fyne.Do(func() {
				p.startBtn.SetText(LABEL_RESUME)
			})
		}
	} else {
		p.running = true
		fyne.Do(func() {
			p.startBtn.SetText(LABEL_PAUSE)
		})
		p.ticker = time.NewTicker(time.Second)
		go p.startTimer()
	}
}

func (p *PomodoroApp) onResetTapped() {
	if p.ticker != nil {
		p.ticker.Stop()
	}

	timeUnit := time.Minute
	if devMode {
		timeUnit = time.Second
	}

	p.state = Focus
	p.session = 1
	p.remaining = time.Duration(FOCUS_TIME) * timeUnit
	p.running = false
	p.pause = false

	fyne.Do(func() {
		p.startBtn.SetText(LABEL_START)
		p.updateUI()
	})
}

func formatTimer(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func main() {
	// 개발 모드 플래그 정의
	flag.BoolVar(&devMode, "dev", false, "Enable development mode (1 minute = 1 second)")
	flag.Parse()

	p := NewPomodoroApp()
	p.window.ShowAndRun()
}
