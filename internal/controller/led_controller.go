package controller

import (
	"fmt"
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
)

// PanelLED represents the available QNAP panel LEDs
type PanelLED int

const (
	StatusGreen PanelLED = iota
	StatusRed
	USB
	Disk1
	Disk2
	Disk3
	Disk4
	Disk5
	Disk6
)

// LEDController manages QNAP panel LEDs using hardware I/O ports
type LEDController struct {
	logger    *logrus.Entry
	portPerms bool
}

const (
	regPort   = 0xa05
	valuePort = 0xa06
	portCount = 2
)

// Port configuration for different LED groups
type portConfig struct {
	register byte
	leds     map[PanelLED]byte // LED -> bit position
}

var (
	statusLEDPort = portConfig{
		register: 0x91,
		leds: map[PanelLED]byte{
			StatusGreen: 2,
			StatusRed:   3,
		},
	}

	diskLEDPort = portConfig{
		register: 0x81,
		leds: map[PanelLED]byte{
			Disk1: 0,
			Disk2: 1,
			Disk3: 2,
			Disk4: 3,
			Disk5: 4, // Extended to use more bits on same port
			Disk6: 5, // Extended to use more bits on same port
		},
	}

	usbLEDPort = portConfig{
		register: 0xE1,
		leds: map[PanelLED]byte{
			USB: 7,
		},
	}
)

// NewLEDController creates a new LED controller
func NewLEDController() (*LEDController, error) {
	logger := logrus.WithField("component", "led_controller")

	lc := &LEDController{
		logger: logger,
	}

	// Try to get I/O port permissions
	if err := lc.requestPortPermissions(); err != nil {
		logger.WithError(err).Warn("Failed to get I/O port permissions, LED control will be disabled")
		return lc, nil // Return controller but mark as non-functional
	}

	logger.Info("LED controller initialized with I/O port access")
	return lc, nil
}

// requestPortPermissions requests access to the hardware I/O ports
func (lc *LEDController) requestPortPermissions() error {
	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("LED control requires root privileges")
	}

	// Request I/O port permissions using ioperm syscall
	// ioperm(from, num, turn_on)
	_, _, errno := syscall.Syscall(syscall.SYS_IOPERM, regPort, portCount, 1)
	if errno != 0 {
		return fmt.Errorf("ioperm failed: %v", errno)
	}

	lc.portPerms = true
	return nil
}

// Close releases I/O port permissions
func (lc *LEDController) Close() error {
	if lc.portPerms {
		// Release I/O port permissions
		syscall.Syscall(syscall.SYS_IOPERM, regPort, portCount, 0)
		lc.portPerms = false
	}
	return nil
}

// SetLED controls a specific LED
func (lc *LEDController) SetLED(led PanelLED, on bool) error {
	if !lc.portPerms {
		lc.logger.Debug("I/O port permissions not available, skipping LED control")
		return nil
	}

	lc.logger.WithFields(logrus.Fields{
		"led": led,
		"on":  on,
	}).Debug("Setting LED state")

	// Determine which port configuration to use
	var port portConfig
	var found bool

	if bit, exists := statusLEDPort.leds[led]; exists {
		port = statusLEDPort
		port.leds = map[PanelLED]byte{led: bit}
		found = true
	} else if bit, exists := diskLEDPort.leds[led]; exists {
		port = diskLEDPort
		port.leds = map[PanelLED]byte{led: bit}
		found = true
	} else if bit, exists := usbLEDPort.leds[led]; exists {
		port = usbLEDPort
		port.leds = map[PanelLED]byte{led: bit}
		found = true
	}

	if !found {
		return fmt.Errorf("unknown LED: %v", led)
	}

	return lc.updatePortLEDs(port, map[PanelLED]bool{led: on})
}

// SetDiskLEDs controls all disk LEDs at once
func (lc *LEDController) SetDiskLEDs(states map[int]bool) error {
	if !lc.portPerms {
		lc.logger.Debug("I/O port permissions not available, skipping LED control")
		return nil
	}

	ledStates := make(map[PanelLED]bool)
	diskLEDs := []PanelLED{Disk1, Disk2, Disk3, Disk4, Disk5, Disk6}

	for diskNum, state := range states {
		if diskNum >= 1 && diskNum <= 6 {
			ledStates[diskLEDs[diskNum-1]] = state
		}
	}

	if len(ledStates) == 0 {
		return nil
	}

	return lc.updatePortLEDs(diskLEDPort, ledStates)
}

// SetStatusLED controls the status LED (green or red)
func (lc *LEDController) SetStatusLED(red bool, green bool) error {
	if !lc.portPerms {
		lc.logger.Debug("I/O port permissions not available, skipping LED control")
		return nil
	}

	ledStates := map[PanelLED]bool{
		StatusRed:   red,
		StatusGreen: green,
	}

	return lc.updatePortLEDs(statusLEDPort, ledStates)
}

// updatePortLEDs updates the LED states for a specific port
func (lc *LEDController) updatePortLEDs(port portConfig, newStates map[PanelLED]bool) error {
	// Read current port state
	currentMask, err := lc.readPort(port.register)
	if err != nil {
		return fmt.Errorf("failed to read port 0x%x: %w", port.register, err)
	}

	// Apply changes (note: QNAP LEDs are inverted - set bit means OFF)
	mask := currentMask
	for led, state := range newStates {
		if bit, exists := port.leds[led]; exists {
			if state {
				mask &^= (1 << bit) // Clear bit to turn LED ON
			} else {
				mask |= (1 << bit) // Set bit to turn LED OFF
			}
		}
	}

	// Write new state if changed
	if mask != currentMask {
		if err := lc.writePort(port.register, mask); err != nil {
			return fmt.Errorf("failed to write port 0x%x: %w", port.register, err)
		}
		lc.logger.WithFields(logrus.Fields{
			"port":     fmt.Sprintf("0x%x", port.register),
			"old_mask": fmt.Sprintf("0x%x", currentMask),
			"new_mask": fmt.Sprintf("0x%x", mask),
		}).Debug("Updated LED port")
	}

	return nil
}

// readPort reads the current state of a hardware port
func (lc *LEDController) readPort(register byte) (byte, error) {
	// Set register
	if err := lc.outb(register, regPort); err != nil {
		return 0, err
	}

	// Read value
	return lc.inb(valuePort)
}

// writePort writes a value to a hardware port
func (lc *LEDController) writePort(register byte, value byte) error {
	// Set register
	if err := lc.outb(register, regPort); err != nil {
		return err
	}

	// Write value
	return lc.outb(value, valuePort)
}

// outb writes a byte to an I/O port using syscall
func (lc *LEDController) outb(value byte, port uint16) error {
	// On Linux, we can use /dev/port for I/O port access
	file, err := os.OpenFile("/dev/port", os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open /dev/port: %w", err)
	}
	defer file.Close()

	// Seek to the port address
	if _, err := file.Seek(int64(port), 0); err != nil {
		return fmt.Errorf("failed to seek to port %x: %w", port, err)
	}

	// Write the value
	if _, err := file.Write([]byte{value}); err != nil {
		return fmt.Errorf("failed to write to port %x: %w", port, err)
	}

	return nil
}

// inb reads a byte from an I/O port using syscall
func (lc *LEDController) inb(port uint16) (byte, error) {
	// On Linux, we can use /dev/port for I/O port access
	file, err := os.OpenFile("/dev/port", os.O_RDONLY, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to open /dev/port: %w", err)
	}
	defer file.Close()

	// Seek to the port address
	if _, err := file.Seek(int64(port), 0); err != nil {
		return 0, fmt.Errorf("failed to seek to port %x: %w", port, err)
	}

	// Read the value
	buffer := make([]byte, 1)
	if _, err := file.Read(buffer); err != nil {
		return 0, fmt.Errorf("failed to read from port %x: %w", port, err)
	}

	return buffer[0], nil
}

// GetLEDStates returns the current state of all LEDs
func (lc *LEDController) GetLEDStates() (map[PanelLED]bool, error) {
	if !lc.portPerms {
		return make(map[PanelLED]bool), nil
	}

	states := make(map[PanelLED]bool)

	// Read status LEDs
	if mask, err := lc.readPort(statusLEDPort.register); err == nil {
		for led, bit := range statusLEDPort.leds {
			states[led] = (mask & (1 << bit)) == 0 // Inverted logic
		}
	}

	// Read disk LEDs
	if mask, err := lc.readPort(diskLEDPort.register); err == nil {
		for led, bit := range diskLEDPort.leds {
			states[led] = (mask & (1 << bit)) == 0 // Inverted logic
		}
	}

	// Read USB LED
	if mask, err := lc.readPort(usbLEDPort.register); err == nil {
		for led, bit := range usbLEDPort.leds {
			states[led] = (mask & (1 << bit)) == 0 // Inverted logic
		}
	}

	return states, nil
}
