package controller

import (
	"fmt"
	"time"

	"github.com/qnap/display-control/internal/config"
	"github.com/qnap/display-control/internal/monitor"
	"github.com/sirupsen/logrus"
)

// SystemController manages the overall QNAP system components
type SystemController struct {
	display      *DisplayController
	led          *LEDController
	usbMonitor   *monitor.USBCopyMonitor
	config       *config.Config
	logger       *logrus.Entry
	buttonHandler ButtonEventHandler
}

// NewSystemController creates a new system controller
func NewSystemController(cfg *config.Config) (*SystemController, error) {
	logger := logrus.WithField("component", "system_controller")

	// Initialize display controller
	display, err := NewDisplayController(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize display controller: %w", err)
	}

	// Initialize LED controller
	led, err := NewLEDController()
	if err != nil {
		logger.WithError(err).Warn("LED controller initialization failed, continuing without LED support")
		led = nil
	}

	// Initialize USB copy monitor
	var usbMonitor *monitor.USBCopyMonitor
	if cfg.USBCopy.IOPort != 0 {
		usbMonitor, err = monitor.NewUSBCopyMonitor(cfg.USBCopy.IOPort)
		if err != nil {
			logger.WithError(err).Warn("USB copy monitor initialization failed, continuing without USB copy support")
			usbMonitor = nil
		}
	}

	sc := &SystemController{
		display:    display,
		led:        led,
		usbMonitor: usbMonitor,
		config:     cfg,
		logger:     logger,
	}

	// Set up button handler for display buttons (ENTER/SELECT)
	display.SetButtonHandler(sc.handleDisplayButtonEvent)

	// Start USB copy button monitoring if available
	if sc.usbMonitor != nil {
		go sc.monitorUSBCopyButton()
	}

	// Initialize system state
	if err := sc.initializeSystem(); err != nil {
		logger.WithError(err).Warn("System initialization partially failed")
	}

	logger.Info("System controller initialized successfully")
	return sc, nil
}

// Close closes the system controller and cleans up resources
func (sc *SystemController) Close() error {
	sc.logger.Info("Closing system controller")

	if sc.usbMonitor != nil {
		if err := sc.usbMonitor.Close(); err != nil {
			sc.logger.WithError(err).Error("Failed to close USB copy monitor")
		}
	}

	if sc.display != nil {
		if err := sc.display.Close(); err != nil {
			sc.logger.WithError(err).Error("Failed to close display controller")
		}
	}

	if sc.led != nil {
		if err := sc.led.Close(); err != nil {
			sc.logger.WithError(err).Error("Failed to close LED controller")
		}
	}

	return nil
}

// GetDisplayController returns the display controller
func (sc *SystemController) GetDisplayController() *DisplayController {
	return sc.display
}

// GetLEDController returns the LED controller
func (sc *SystemController) GetLEDController() *LEDController {
	return sc.led
}

// GetUSBCopyMonitor returns the USB copy monitor
func (sc *SystemController) GetUSBCopyMonitor() *monitor.USBCopyMonitor {
	return sc.usbMonitor
}

// SetButtonHandler sets a unified button handler for all button types
func (sc *SystemController) SetButtonHandler(handler ButtonEventHandler) {
	sc.buttonHandler = handler
}

// initializeSystem sets up the initial system state
func (sc *SystemController) initializeSystem() error {
	if sc.led != nil {
		// Set initial LED states
		sc.led.SetStatusLED(false, true) // Green status LED on
		sc.led.SetLED(USB, false)        // USB LED off
		
		// Turn off all disk LEDs initially
		sc.led.SetDiskLEDs(map[int]bool{
			1: false,
			2: false,
			3: false,
			4: false,
			5: false,
			6: false,
		})
	}

	return nil
}

// handleButtonEvent handles button press events from the display
func (sc *SystemController) handleDisplayButtonEvent(button PanelButton, pressed bool) {
	sc.logger.WithFields(logrus.Fields{
		"button":  button,
		"pressed": pressed,
		"source":  "serial",
	}).Info("Display button event")

	// Forward to unified button handler if set
	if sc.buttonHandler != nil {
		sc.buttonHandler(button, pressed)
		return
	}

	// Default handling if no unified handler is set
	if !pressed {
		return
	}

	switch button {
	case ButtonEnter:
		sc.handleEnterButton()
	case ButtonSelect:
		sc.handleSelectButton()
	}
}

// monitorUSBCopyButton monitors the hardware USB copy button
func (sc *SystemController) monitorUSBCopyButton() {
	sc.logger.Info("Starting USB copy button monitoring")
	
	err := sc.usbMonitor.MonitorButtonPresses(func() {
		sc.logger.WithFields(logrus.Fields{
			"button":  "USB_COPY",
			"pressed": true,
			"source":  "hardware",
		}).Info("USB copy button event")
		
		// Trigger press event
		if sc.buttonHandler != nil {
			sc.buttonHandler(ButtonUSBCopy, true)
			// Add small delay and trigger release
			time.Sleep(100 * time.Millisecond)
			sc.buttonHandler(ButtonUSBCopy, false)
		} else {
			// Default handling
			sc.handleUSBCopyButton()
		}
	})
	
	if err != nil {
		sc.logger.WithError(err).Error("USB copy button monitoring failed")
	}
}

// handleEnterButton handles ENTER button presses
func (sc *SystemController) handleEnterButton() {
	sc.logger.Debug("ENTER button pressed")
	// This will be handled by the menu system
}

// handleSelectButton handles SELECT button presses
func (sc *SystemController) handleSelectButton() {
	sc.logger.Debug("SELECT button pressed")
	// This will be handled by the menu system
}

// handleUSBCopyButton handles USB COPY button presses
func (sc *SystemController) handleUSBCopyButton() {
	sc.logger.Info("USB COPY button pressed")
	
	if sc.led != nil {
		// Flash USB LED to indicate copy operation
		sc.led.SetLED(USB, true)
	}

	// Show copy status on display
	if sc.display != nil {
		sc.display.ShowCopyStatus("Starting...")
	}

	// This will trigger the USB copy functionality
	// Implementation depends on the main application logic
}

// SetDiskActivity sets disk LED activity
func (sc *SystemController) SetDiskActivity(diskNum int, active bool) error {
	if sc.led == nil {
		return nil // LED controller not available
	}

	diskLEDs := map[PanelLED]bool{}
	switch diskNum {
	case 1:
		diskLEDs[Disk1] = active
	case 2:
		diskLEDs[Disk2] = active
	case 3:
		diskLEDs[Disk3] = active
	case 4:
		diskLEDs[Disk4] = active
	case 5:
		diskLEDs[Disk5] = active
	case 6:
		diskLEDs[Disk6] = active
	default:
		return fmt.Errorf("invalid disk number: %d (must be 1-6)", diskNum)
	}

	for led, state := range diskLEDs {
		if err := sc.led.SetLED(led, state); err != nil {
			return fmt.Errorf("failed to set disk %d LED: %w", diskNum, err)
		}
	}

	return nil
}

// FlashDiskLED flashes a disk LED for a specified duration
func (sc *SystemController) FlashDiskLED(diskNum int, duration time.Duration) {
	if sc.led == nil {
		return
	}

	go func() {
		// Turn on LED
		sc.SetDiskActivity(diskNum, true)
		
		// Wait for duration
		time.Sleep(duration)
		
		// Turn off LED
		sc.SetDiskActivity(diskNum, false)
	}()
}

// SetSystemStatus sets the overall system status
func (sc *SystemController) SetSystemStatus(status string, isError bool) error {
	// Update display
	if sc.display != nil {
		if err := sc.display.WriteText(status); err != nil {
			sc.logger.WithError(err).Error("Failed to update display status")
		}
	}

	// Update status LED
	if sc.led != nil {
		if isError {
			sc.led.SetStatusLED(true, false) // Red LED for error
		} else {
			sc.led.SetStatusLED(false, true) // Green LED for normal
		}
	}

	return nil
}

// ShowProgress shows progress on display and optionally flash LEDs
func (sc *SystemController) ShowProgress(percent int, flashDisks bool) error {
	// Update display progress
	if sc.display != nil {
		if err := sc.display.ShowProgress(percent); err != nil {
			return fmt.Errorf("failed to show progress on display: %w", err)
		}
	}

	// Flash disk LEDs based on progress if requested
	if flashDisks && sc.led != nil {
		activeDisk := (percent / 17) + 1 // Each ~17% activates next disk LED (100/6)
		if activeDisk > 6 {
			activeDisk = 6
		}

		for i := 1; i <= 6; i++ {
			sc.SetDiskActivity(i, i <= activeDisk)
		}
	}

	return nil
}
