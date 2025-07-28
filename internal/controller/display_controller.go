package controller

import (
	"fmt"
	"strings"
	"time"

	"github.com/qnap/display-control/internal/config"
	"github.com/qnap/display-control/internal/serial"
	"github.com/sirupsen/logrus"
)

// PanelButton represents available QNAP panel buttons
type PanelButton int

const (
	ButtonEnter PanelButton = iota
	ButtonSelect
	ButtonUSBCopy
)

// ButtonEventHandler is a callback function for button events
type ButtonEventHandler func(button PanelButton, pressed bool)

// DisplayController manages the LCD display
type DisplayController struct {
	serialPort       *serial.SerialPort
	config          *config.Config
	logger          *logrus.Entry
	buttonHandler   ButtonEventHandler
	lastButtonState map[PanelButton]bool
}

// NewDisplayController creates a new display controller
func NewDisplayController(cfg *config.Config) (*DisplayController, error) {
	logger := logrus.WithField("component", "display_controller")

	serialPort, err := serial.NewSerialPort(cfg.SerialPort.Device, cfg.SerialPort.BaudRate)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize serial port: %w", err)
	}

	dc := &DisplayController{
		serialPort:      serialPort,
		config:         cfg,
		logger:         logger,
		lastButtonState: make(map[PanelButton]bool),
	}

	// Initialize display
	if err := dc.initializeDisplay(); err != nil {
		serialPort.Close()
		return nil, fmt.Errorf("failed to initialize display: %w", err)
	}

	// Verify serial configuration
	if !serialPort.IsConfigValid() {
		logger.Warn("Serial port configuration may not be optimal for QNAP display")
	} else {
		logger.Debug("Serial port configured with 8N1 (8 data bits, no parity, 1 stop bit)")
	}

	// Start button monitoring in background
	go dc.monitorButtons()

	logger.Info("Display controller initialized successfully")
	return dc, nil
}

// Close closes the display controller and cleans up resources
func (dc *DisplayController) Close() error {
	dc.logger.Info("Closing display controller")
	if dc.serialPort != nil {
		return dc.serialPort.Close()
	}
	return nil
}

// initializeDisplay sets up the LCD display
func (dc *DisplayController) initializeDisplay() error {
	dc.logger.Debug("Initializing QNAP LCD display")

	// Based on qnapctl reference: enable button state reporting
	// Send the command to enable button state reporting first
	buttonStateCmd := []byte{0x4D, 0x06}
	if err := dc.serialPort.Write(buttonStateCmd); err != nil {
		dc.logger.WithError(err).Warn("Failed to enable button state reporting")
	} else {
		dc.logger.Info("Button state reporting enabled successfully")
	}

	// Give the controller time to process the command
	time.Sleep(100 * time.Millisecond)

	// Turn on backlight using correct QNAP protocol
	if err := dc.SetBacklight(true); err != nil {
		dc.logger.WithError(err).Debug("Failed to turn on backlight")
	}

	// Clear both lines using correct QNAP protocol
	if err := dc.WriteTextAt("", 0, 0); err != nil {
		dc.logger.WithError(err).Warn("Failed to clear line 0")
	}
	if err := dc.WriteTextAt("", 1, 0); err != nil {
		dc.logger.WithError(err).Warn("Failed to clear line 1")
	}

	// Show default text if specified
	if dc.config.Display.DefaultText != "" {
		if err := dc.WriteText(dc.config.Display.DefaultText); err != nil {
			dc.logger.WithError(err).Warn("Failed to write default text")
		}
	} else {
		// Show a test message to confirm display is working
		if err := dc.WriteText("QNAP Display\nReady"); err != nil {
			dc.logger.WithError(err).Warn("Failed to write test message")
		}
	}

	return nil
}

// WriteText writes text to the display
func (dc *DisplayController) WriteText(text string) error {
	dc.logger.WithField("text", text).Debug("Writing text to display")

	// Split text by newlines first, then handle line wrapping
	lines := strings.Split(text, "\n")
	
	// Ensure we have exactly 2 lines for the 2-line display
	displayLines := make([]string, 2)
	
	// Handle the lines
	if len(lines) >= 1 {
		displayLines[0] = lines[0]
	}
	if len(lines) >= 2 {
		displayLines[1] = lines[1]
	}
	
	// Truncate lines that are too long
	for i := range displayLines {
		if len(displayLines[i]) > 16 {
			displayLines[i] = displayLines[i][:16]
		}
	}

	// Write each line using the QNAP line command format
	for i, line := range displayLines {
		if err := dc.WriteTextAt(line, i, 0); err != nil {
			return fmt.Errorf("failed to write line %d: %w", i, err)
		}
	}

	return nil
}

// WriteTextAt writes text at a specific position
func (dc *DisplayController) WriteTextAt(text string, row, col int) error {
	dc.logger.WithFields(logrus.Fields{
		"text": text,
		"row":  row,
		"col":  col,
	}).Debug("Writing text at position")

	// Validate row (0 or 1 for 2-line display)
	if row < 0 || row > 1 {
		return fmt.Errorf("invalid row: %d. Must be 0 or 1", row)
	}

	const LCD_CHARS_PER_LINE = 16
	
	// Truncate and pad text to fit LCD width
	displayText := text
	if len(displayText) > LCD_CHARS_PER_LINE {
		displayText = displayText[:LCD_CHARS_PER_LINE]
	}
	// Pad with spaces to fill the line
	for len(displayText) < LCD_CHARS_PER_LINE {
		displayText += " "
	}

	// Use correct QNAP protocol: 0x4D, 0x0C, line, 0x10, followed by 16 characters
	// This is the verified protocol from qnapctl reference implementation
	command := []byte{0x4D, 0x0C, byte(row), 0x10}
	command = append(command, []byte(displayText)...)

	if err := dc.serialPort.Write(command); err != nil {
		dc.logger.WithError(err).WithField("line", row).Warn("Failed to write text using QNAP protocol")
		return err
	}

	dc.logger.WithField("line", row).Debug("Text written using QNAP protocol")
	return nil
}

// ClearDisplay clears the entire display
func (dc *DisplayController) ClearDisplay() error {
	dc.logger.Debug("Clearing display")

	// Clear both lines by writing empty text to each line
	if err := dc.WriteTextAt("", 0, 0); err != nil {
		return fmt.Errorf("failed to clear line 0: %w", err)
	}
	
	if err := dc.WriteTextAt("", 1, 0); err != nil {
		return fmt.Errorf("failed to clear line 1: %w", err)
	}

	return nil
}

// SetBacklight controls the display backlight (if supported)
func (dc *DisplayController) SetBacklight(on bool) error {
	dc.logger.WithField("on", on).Debug("Setting backlight")

	// Use correct QNAP protocol: 0x4D, 0x5E, on/off
	// This is the verified protocol from qnapctl reference implementation
	var cmd []byte
	if on {
		cmd = []byte{0x4D, 0x5E, 0x01} // Backlight on
	} else {
		cmd = []byte{0x4D, 0x5E, 0x00} // Backlight off
	}

	if err := dc.serialPort.Write(cmd); err != nil {
		return fmt.Errorf("failed to set backlight: %w", err)
	}

	return nil
}

// ShowCopyStatus displays copy operation status
func (dc *DisplayController) ShowCopyStatus(status string) error {
	dc.logger.WithField("status", status).Info("Showing copy status")

	// Line 1: "USB Copy"
	if err := dc.WriteTextAt("USB Copy", 0, 0); err != nil {
		return err
	}

	// Line 2: Status message
	if err := dc.WriteTextAt(status, 1, 0); err != nil {
		return err
	}

	return nil
}

// ShowProgress displays a progress bar (simplified)
func (dc *DisplayController) ShowProgress(percent int) error {
	dc.logger.WithField("percent", percent).Debug("Showing progress")

	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	// Calculate progress bar width (for 16 character display)
	barWidth := 14 // Leave space for [ ]
	filled := (percent * barWidth) / 100

	progressBar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			progressBar += "="
		} else {
			progressBar += " "
		}
	}
	progressBar += "]"

	// Show progress on second line using QNAP line command
	if err := dc.WriteTextAt(progressBar, 1, 0); err != nil {
		return err
	}

	return nil
}

// SetButtonHandler sets the callback function for button events
func (dc *DisplayController) SetButtonHandler(handler ButtonEventHandler) {
	dc.logger.Info("Button handler set")
	dc.buttonHandler = handler
}

// RequestButtonState manually requests current button state from the QNAP controller
func (dc *DisplayController) RequestButtonState() error {
	// Send button state request command
	buttonStateRequestCmd := []byte{0x4D, 0x05}
	if err := dc.serialPort.Write(buttonStateRequestCmd); err != nil {
		return fmt.Errorf("failed to request button state: %w", err)
	}
	
	dc.logger.Debug("Button state request sent")
	return nil
}

// monitorButtons monitors button presses in the background
func (dc *DisplayController) monitorButtons() {
	dc.logger.Info("Starting button monitoring")

	// Buffer to accumulate partial messages
	messageBuffer := make([]byte, 0, 32)
	
	// Timer for periodic button state requests
	buttonRequestTicker := time.NewTicker(500 * time.Millisecond)
	defer buttonRequestTicker.Stop()

	for {
		select {
		case <-buttonRequestTicker.C:
			// Periodically request button state to ensure we get updates
			if err := dc.RequestButtonState(); err != nil {
				dc.logger.WithError(err).Debug("Failed to request button state")
			}
			
		default:
			// Use ReadAvailable for non-blocking read
			data, err := dc.serialPort.ReadAvailable()
			if err != nil {
				dc.logger.WithError(err).Debug("Error reading button data")
				time.Sleep(50 * time.Millisecond)
				continue
			}

			if len(data) == 0 {
				time.Sleep(50 * time.Millisecond) // Poll every 50ms when no data
				continue
			}

			// Append new data to buffer
			messageBuffer = append(messageBuffer, data...)

			// Log received data only at debug level to reduce noise
			dc.logger.WithFields(logrus.Fields{
				"length":     len(data),
				"hex":        fmt.Sprintf("% 02x", data),
				"ascii":      fmt.Sprintf("%q", data),
				"buffer_len": len(messageBuffer),
				"buffer_hex": fmt.Sprintf("% 02x", messageBuffer),
			}).Debug("Received serial data")

			// Process complete messages in buffer
			dc.processMessageBuffer(&messageBuffer)
			
			time.Sleep(10 * time.Millisecond) // Small delay between reads
		}
	}
}

// processMessageBuffer processes accumulated data for complete button messages
func (dc *DisplayController) processMessageBuffer(buffer *[]byte) {
	for len(*buffer) >= 4 {
		// Look for standard button message: 0x53, 0x05, 0x00, button_state
		if (*buffer)[0] == 0x53 && (*buffer)[1] == 0x05 && (*buffer)[2] == 0x00 {
			buttonState := (*buffer)[3]
			dc.logger.WithField("button_state", fmt.Sprintf("0x%02x", buttonState)).Info("Parsing button state")
			dc.parseButtonState(buttonState)
			
			// Remove processed message from buffer
			*buffer = (*buffer)[4:]
			continue
		}
		
		// Look for alternative button message formats
		if (*buffer)[0] == 0x4D {
			// QNAP command response - might contain button info
			if len(*buffer) >= 3 {
				dc.logger.WithField("qnap_response", fmt.Sprintf("% 02x", (*buffer)[:3])).Debug("QNAP response received")
				// Remove this message
				*buffer = (*buffer)[3:]
				continue
			}
		}
		
		// Look for copy button specific message (may use different protocol)
		if (*buffer)[0] == 0x55 || (*buffer)[0] == 0x43 { // 'U' or 'C' for USB/Copy
			if len(*buffer) >= 2 {
				dc.logger.WithField("copy_message", fmt.Sprintf("% 02x", (*buffer)[:2])).Info("Potential copy button message")
				// Parse as copy button press
				dc.triggerButtonEvent(ButtonUSBCopy, true)
				time.Sleep(100 * time.Millisecond) // Debounce
				dc.triggerButtonEvent(ButtonUSBCopy, false)
				*buffer = (*buffer)[2:]
				continue
			}
		}
		
		// If we don't recognize the message, remove first byte and try again
		dc.logger.WithField("unknown_byte", fmt.Sprintf("0x%02x", (*buffer)[0])).Debug("Unknown message byte, discarding")
		*buffer = (*buffer)[1:]
		
		// Prevent buffer from growing too large
		if len(*buffer) > 16 {
			dc.logger.Warn("Message buffer too large, clearing")
			*buffer = (*buffer)[:0]
			break
		}
	}
}

// parseButtonState parses the button state byte and triggers events
func (dc *DisplayController) parseButtonState(state byte) {
	// Based on qnapctl reference, button bits are:
	// Bit 0 (0x01): ENTER button (inverted logic - 0 = pressed)
	// Bit 1 (0x02): SELECT button (inverted logic - 0 = pressed)  
	// Bit 2 (0x04): USB COPY button (may use different logic)
	
	const (
		buttonEnterBit  = 0x01
		buttonSelectBit = 0x02
		buttonUSBCopyBit = 0x04
	)

	// QNAP uses inverted logic for ENTER and SELECT buttons (0 = pressed)
	enterPressed := (state & buttonEnterBit) == 0
	selectPressed := (state & buttonSelectBit) == 0
	
	// USB copy button may use normal logic (1 = pressed) - test both
	usbCopyPressed := (state & buttonUSBCopyBit) != 0

	dc.logger.WithFields(logrus.Fields{
		"state_hex":      fmt.Sprintf("0x%02x", state),
		"state_binary":   fmt.Sprintf("%08b", state),
		"enter_pressed":  enterPressed,
		"select_pressed": selectPressed,
		"copy_pressed":   usbCopyPressed,
	}).Debug("Button state analysis")

	// Check for state changes and trigger events
	if dc.checkButtonStateChange(ButtonEnter, enterPressed) {
		dc.triggerButtonEvent(ButtonEnter, enterPressed)
	}

	if dc.checkButtonStateChange(ButtonSelect, selectPressed) {
		dc.triggerButtonEvent(ButtonSelect, selectPressed)
	}

	if dc.checkButtonStateChange(ButtonUSBCopy, usbCopyPressed) {
		dc.triggerButtonEvent(ButtonUSBCopy, usbCopyPressed)
	}
}

// checkButtonStateChange checks if a button state has changed
func (dc *DisplayController) checkButtonStateChange(button PanelButton, pressed bool) bool {
	lastState, exists := dc.lastButtonState[button]
	if !exists || lastState != pressed {
		dc.lastButtonState[button] = pressed
		return true
	}
	return false
}

// triggerButtonEvent triggers a button event if handler is set
func (dc *DisplayController) triggerButtonEvent(button PanelButton, pressed bool) {
	buttonName := ""
	switch button {
	case ButtonEnter:
		buttonName = "ENTER"
	case ButtonSelect:
		buttonName = "SELECT"
	case ButtonUSBCopy:
		buttonName = "USB_COPY"
	default:
		buttonName = "UNKNOWN"
	}

	dc.logger.WithFields(logrus.Fields{
		"button":      buttonName,
		"button_id":   int(button),
		"pressed":     pressed,
		"has_handler": dc.buttonHandler != nil,
	}).Info("Button event triggered")

	if dc.buttonHandler != nil {
		// Call handler in a separate goroutine to prevent blocking
		go func() {
			defer func() {
				if r := recover(); r != nil {
					dc.logger.WithField("panic", r).Error("Button handler panicked")
				}
			}()
			dc.buttonHandler(button, pressed)
		}()
	} else {
		dc.logger.Warn("No button handler set - button event ignored")
	}
}
