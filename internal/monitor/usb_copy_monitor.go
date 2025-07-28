package monitor

import (
	"fmt"
	"sync"
	"time"

	"github.com/qnap/display-control/internal/hardware"
	"github.com/sirupsen/logrus"
)

// USBCopyMonitor monitors the USB copy button
type USBCopyMonitor struct {
	ioPort     IOPortReader
	port       uint16
	lastState  bool
	mutex      sync.RWMutex
	logger     *logrus.Entry
	closed     bool
	closeChan  chan struct{}
}

// IOPortReader interface for I/O port access
type IOPortReader interface {
	ReadByte() (byte, error)
	Close() error
}

// NewUSBCopyMonitor creates a new USB copy button monitor
func NewUSBCopyMonitor(port uint16) (*USBCopyMonitor, error) {
	logger := logrus.WithField("component", "usb_copy_monitor")

	ioPort, err := hardware.NewIOPortAccess(port)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize I/O port access: %w", err)
	}

	monitor := &USBCopyMonitor{
		ioPort:    ioPort,
		port:      port,
		lastState: false,
		logger:    logger,
		closeChan: make(chan struct{}),
	}

	logger.WithField("port", fmt.Sprintf("0x%x", port)).Info("USB copy monitor initialized")
	return monitor, nil
}

// NewUSBCopyMonitorWithIOPort creates a monitor with a custom IOPortReader (for testing)
func NewUSBCopyMonitorWithIOPort(port uint16, ioPort IOPortReader) *USBCopyMonitor {
	logger := logrus.WithField("component", "usb_copy_monitor")

	monitor := &USBCopyMonitor{
		ioPort:    ioPort,
		port:      port,
		lastState: false,
		logger:    logger,
		closeChan: make(chan struct{}),
	}

	logger.WithField("port", fmt.Sprintf("0x%x", port)).Info("USB copy monitor initialized")
	return monitor
}

// Close closes the USB copy monitor and cleans up resources
func (m *USBCopyMonitor) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return nil
	}

	m.logger.Info("Closing USB copy monitor")
	m.closed = true
	close(m.closeChan)

	if m.ioPort != nil {
		return m.ioPort.Close()
	}

	return nil
}

// IsButtonPressed checks if the USB copy button is currently pressed
func (m *USBCopyMonitor) IsButtonPressed() (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.closed {
		return false, fmt.Errorf("monitor is closed")
	}

	value, err := m.ioPort.ReadByte()
	if err != nil {
		return false, fmt.Errorf("failed to read I/O port: %w", err)
	}

	// Button is pressed when bit 0 is low (assuming active-low button)
	pressed := (value & 0x01) == 0

	m.logger.WithFields(logrus.Fields{
		"port_value": fmt.Sprintf("0x%02x", value),
		"pressed":    pressed,
	}).Trace("Button state check")

	return pressed, nil
}

// WaitForButtonPress waits for a button press event with timeout
func (m *USBCopyMonitor) WaitForButtonPress(timeout time.Duration) (bool, error) {
	m.logger.WithField("timeout", timeout).Debug("Waiting for button press")

	startTime := time.Now()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-m.closeChan:
			return false, fmt.Errorf("monitor closed")
		case <-ticker.C:
			pressed, err := m.IsButtonPressed()
			if err != nil {
				return false, err
			}

			if pressed {
				m.logger.Info("Button press detected")
				return true, nil
			}

			if time.Since(startTime) > timeout {
				return false, nil // Timeout, no press detected
			}
		}
	}
}

// MonitorButtonPresses continuously monitors for button press events
func (m *USBCopyMonitor) MonitorButtonPresses(callback func()) error {
	m.logger.Info("Starting button press monitoring")

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var lastPressed bool

	for {
		select {
		case <-m.closeChan:
			m.logger.Info("Button monitoring stopped")
			return nil
		case <-ticker.C:
			pressed, err := m.IsButtonPressed()
			if err != nil {
				m.logger.WithError(err).Error("Error checking button state")
				continue
			}

			// Detect rising edge (button press)
			if pressed && !lastPressed {
				m.logger.Info("Button press event detected")
				if callback != nil {
					go callback() // Run callback in goroutine to avoid blocking
				}
			}

			lastPressed = pressed
		}
	}
}

// GetButtonState returns the current button state with debouncing
func (m *USBCopyMonitor) GetButtonState() (bool, error) {
	const debounceTime = 20 * time.Millisecond
	const sampleCount = 3

	var pressedCount int

	for i := 0; i < sampleCount; i++ {
		pressed, err := m.IsButtonPressed()
		if err != nil {
			return false, err
		}

		if pressed {
			pressedCount++
		}

		if i < sampleCount-1 {
			time.Sleep(debounceTime / sampleCount)
		}
	}

	// Button is considered pressed if majority of samples indicate pressed
	debounced := pressedCount > sampleCount/2

	m.logger.WithFields(logrus.Fields{
		"pressed_samples": pressedCount,
		"total_samples":   sampleCount,
		"debounced_state": debounced,
	}).Trace("Debounced button state")

	return debounced, nil
}

// StartBackgroundMonitoring starts monitoring button presses in the background
func (m *USBCopyMonitor) StartBackgroundMonitoring(pressChan chan<- bool) error {
	m.logger.Info("Starting background button monitoring")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				m.logger.WithField("panic", r).Error("Panic in background monitoring")
			}
		}()

		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		var lastPressed bool

		for {
			select {
			case <-m.closeChan:
				m.logger.Info("Background monitoring stopped")
				return
			case <-ticker.C:
				pressed, err := m.GetButtonState()
				if err != nil {
					m.logger.WithError(err).Error("Error getting button state")
					continue
				}

				// Detect rising edge (button press)
				if pressed && !lastPressed {
					m.logger.Info("Background: Button press detected")
					select {
					case pressChan <- true:
					default:
						m.logger.Warn("Press channel full, dropping event")
					}
				}

				lastPressed = pressed
			}
		}
	}()

	return nil
}
