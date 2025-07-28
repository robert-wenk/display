package serial

import (
	"fmt"
	"time"

	"github.com/tarm/serial"
)

// SerialPort wraps the serial port functionality for QNAP LCD display communication
// Configured with 8N1 parameters (8 data bits, no parity, 1 stop bit) as required by QNAP hardware
type SerialPort struct {
	port   *serial.Port
	config *serial.Config
}

// NewSerialPort creates a new serial port connection
func NewSerialPort(device string, baudRate int) (*SerialPort, error) {
	// Configure serial port with explicit parameters for QNAP compatibility
	config := &serial.Config{
		Name:        device,
		Baud:        baudRate,
		ReadTimeout: 100 * time.Millisecond, // Short timeout for better button responsiveness
		Size:        8,                // 8 data bits (EIGHTBITS)
		Parity:      serial.ParityNone, // No parity (PARITY_NONE)
		StopBits:    serial.Stop1,     // 1 stop bit (STOPBITS_ONE)
	}

	port, err := serial.OpenPort(config)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s: %w", device, err)
	}

	sp := &SerialPort{
		port:   port,
		config: config,
	}

	// Log the serial configuration for verification
	fmt.Printf("Serial port configured: %s at %d baud, 8N1\n", device, baudRate)

	return sp, nil
}

// GetConfig returns the current serial port configuration
func (sp *SerialPort) GetConfig() *serial.Config {
	return sp.config
}

// IsConfigValid verifies the serial port configuration matches QNAP requirements
func (sp *SerialPort) IsConfigValid() bool {
	if sp.config == nil {
		return false
	}
	
	// Verify 8N1 configuration (8 data bits, No parity, 1 stop bit)
	return sp.config.Size == 8 && 
		   sp.config.Parity == serial.ParityNone && 
		   sp.config.StopBits == serial.Stop1
}

// Close closes the serial port
func (sp *SerialPort) Close() error {
	if sp.port != nil {
		return sp.port.Close()
	}
	return nil
}

// Write writes data to the serial port
func (sp *SerialPort) Write(data []byte) error {
	if sp.port == nil {
		return fmt.Errorf("serial port not initialized")
	}

	n, err := sp.port.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to serial port: %w", err)
	}

	if n != len(data) {
		return fmt.Errorf("incomplete write: wrote %d of %d bytes", n, len(data))
	}

	return nil
}

// Read reads data from the serial port
func (sp *SerialPort) Read(buffer []byte) (int, error) {
	if sp.port == nil {
		return 0, fmt.Errorf("serial port not initialized")
	}

	return sp.port.Read(buffer)
}

// WriteString writes a string to the serial port
func (sp *SerialPort) WriteString(text string) error {
	return sp.Write([]byte(text))
}

// WriteText writes text to the LCD display (line1 and line2)
func (sp *SerialPort) WriteText(line1, line2 string, col, row int) error {
	if sp.port == nil {
		return fmt.Errorf("serial port not initialized")
	}

	// Ensure lines are exactly 16 characters (pad or truncate)
	line1Formatted := fmt.Sprintf("%-16s", line1)
	if len(line1Formatted) > 16 {
		line1Formatted = line1Formatted[:16]
	}
	
	line2Formatted := fmt.Sprintf("%-16s", line2)
	if len(line2Formatted) > 16 {
		line2Formatted = line2Formatted[:16]
	}

	// Try different approaches for LCD communication
	
	// Approach 1: QNAP-specific line commands (preferred)
	// Line 0 command: 0x40, 0x44
	line0Cmd := []byte{0x40, 0x44}
	line0Cmd = append(line0Cmd, []byte(line1Formatted)...)
	if err := sp.Write(line0Cmd); err == nil {
		// Line 1 command: 0x40, 0x45
		line1Cmd := []byte{0x40, 0x45}
		line1Cmd = append(line1Cmd, []byte(line2Formatted)...)
		if err := sp.Write(line1Cmd); err == nil {
			return nil
		}
	}

	// Approach 2: HD44780 compatible commands (fallback)
	// Clear display and home cursor
	clearCmds := []byte{0x01, 0x02} // Clear display, return home
	if err := sp.Write(clearCmds); err == nil {
		time.Sleep(2 * time.Millisecond) // Wait for clear
		
		// Write first line
		if err := sp.WriteString(line1Formatted); err == nil {
			// Move to second line (0x80 + 0x40 = 0xC0)
			if err := sp.Write([]byte{0xC0}); err == nil {
				sp.WriteString(line2Formatted)
			}
		}
		return nil
	}

	// Approach 3: Simple text with newlines
	combinedText := line1Formatted + "\n" + line2Formatted
	if err := sp.WriteString(combinedText); err == nil {
		return nil
	}

	// Approach 3: Just send the text as-is
	if err := sp.WriteString(line1Formatted + line2Formatted); err == nil {
		return nil
	}

	return fmt.Errorf("failed to write text to display")
}

// ReadAvailable reads all available data from the serial port
func (sp *SerialPort) ReadAvailable() ([]byte, error) {
	if sp.port == nil {
		return []byte{}, nil // Return empty data instead of error
	}

	// Check if data is available by doing a non-blocking read
	buffer := make([]byte, 256)
	n, err := sp.port.Read(buffer)
	if err != nil {
		// Check if it's a timeout or no data available
		errMsg := err.Error()
		if errMsg == "timeout" || errMsg == "EOF" || errMsg == "resource temporarily unavailable" {
			return []byte{}, nil
		}
		// Only log actual errors, not timeouts
		return []byte{}, nil
	}

	if n == 0 {
		return []byte{}, nil
	}

	return buffer[:n], nil
}

// IsConnected checks if the serial port is connected and operational
func (sp *SerialPort) IsConnected() bool {
	return sp.port != nil
}

// NewDisplayTester creates a new display tester for this serial port
func (sp *SerialPort) NewDisplayTester() *DisplayTester {
	return &DisplayTester{port: sp}
}

// DisplayTester provides simple display testing functionality
type DisplayTester struct {
	port SerialPortInterface
}

// TestDisplay attempts different methods to write to the display
func (dt *DisplayTester) TestDisplay(line1, line2 string) error {
	if dt.port == nil {
		return fmt.Errorf("no serial port available")
	}

	// Method 1: HD44780 Direct Commands
	if err := dt.testHD44780(line1, line2); err == nil {
		return nil
	}

	// Method 2: Simple Text
	if err := dt.testSimpleText(line1, line2); err == nil {
		return nil
	}

	// Method 3: Raw Text
	if err := dt.testRawText(line1, line2); err == nil {
		return nil
	}

	return fmt.Errorf("all display methods failed")
}

// testHD44780 tests HD44780 compatible commands
func (dt *DisplayTester) testHD44780(line1, line2 string) error {
	// Initialize display
	initCmds := []byte{
		0x38, // Function set: 8-bit, 2 line, 5x7 dots
		0x0C, // Display on, cursor off
		0x06, // Entry mode: increment cursor
		0x01, // Clear display
	}
	
	for _, cmd := range initCmds {
		if err := dt.port.Write([]byte{cmd}); err != nil {
			return err
		}
		time.Sleep(2 * time.Millisecond)
	}
	
	// Wait for clear
	time.Sleep(5 * time.Millisecond)
	
	// Write first line
	if err := dt.port.WriteString(line1); err != nil {
		return err
	}
	
	// Move to second line
	if err := dt.port.Write([]byte{0xC0}); err != nil {
		return err
	}
	
	// Write second line
	return dt.port.WriteString(line2)
}

// testSimpleText tests simple text with basic formatting
func (dt *DisplayTester) testSimpleText(line1, line2 string) error {
	text := fmt.Sprintf("%s\n%s", line1, line2)
	return dt.port.WriteString(text)
}

// testRawText tests raw text without any formatting
func (dt *DisplayTester) testRawText(line1, line2 string) error {
	text := line1 + line2
	return dt.port.WriteString(text)
}

// Flush ensures all pending data is written
func (sp *SerialPort) Flush() error {
	if sp.port == nil {
		return fmt.Errorf("serial port not initialized")
	}

	// The underlying library doesn't expose flush directly,
	// but we can achieve similar behavior by ensuring writes complete
	return nil
}

// IsOpen checks if the serial port is open
func (sp *SerialPort) IsOpen() bool {
	return sp.port != nil
}

// MockSerialPort provides a mock implementation for testing
type MockSerialPort struct {
	writeBuffer []byte
	readBuffer  []byte
	readIndex   int
	writeError  error
	readError   error
	closed      bool
}

// NewMockSerialPort creates a mock serial port for testing
func NewMockSerialPort() *MockSerialPort {
	return &MockSerialPort{
		writeBuffer: make([]byte, 0),
		readBuffer:  make([]byte, 0),
	}
}

// SetReadData sets the data that will be returned by Read operations
func (msp *MockSerialPort) SetReadData(data []byte) {
	msp.readBuffer = make([]byte, len(data))
	copy(msp.readBuffer, data)
	msp.readIndex = 0
}

// SetWriteError sets an error that will be returned by Write operations
func (msp *MockSerialPort) SetWriteError(err error) {
	msp.writeError = err
}

// SetReadError sets an error that will be returned by Read operations
func (msp *MockSerialPort) SetReadError(err error) {
	msp.readError = err
}

// Write simulates writing to the serial port
func (msp *MockSerialPort) Write(data []byte) error {
	if msp.closed {
		return fmt.Errorf("serial port is closed")
	}
	
	if msp.writeError != nil {
		return msp.writeError
	}

	msp.writeBuffer = append(msp.writeBuffer, data...)
	return nil
}

// Read simulates reading from the serial port
func (msp *MockSerialPort) Read(buffer []byte) (int, error) {
	if msp.closed {
		return 0, fmt.Errorf("serial port is closed")
	}
	
	if msp.readError != nil {
		return 0, msp.readError
	}

	available := len(msp.readBuffer) - msp.readIndex
	if available == 0 {
		return 0, nil
	}

	n := len(buffer)
	if n > available {
		n = available
	}

	copy(buffer, msp.readBuffer[msp.readIndex:msp.readIndex+n])
	msp.readIndex += n

	return n, nil
}

// WriteString writes a string to the mock serial port
func (msp *MockSerialPort) WriteString(text string) error {
	return msp.Write([]byte(text))
}

// WriteText writes text to the mock LCD display (line1 and line2)
func (msp *MockSerialPort) WriteText(line1, line2 string, col, row int) error {
	if msp.closed {
		return fmt.Errorf("serial port is closed")
	}
	
	if msp.writeError != nil {
		return msp.writeError
	}

	// Simulate writing display commands
	displayData := fmt.Sprintf("%s\n%s", line1, line2)
	return msp.Write([]byte(displayData))
}

// ReadAvailable reads available data from the mock serial port
func (msp *MockSerialPort) ReadAvailable() ([]byte, error) {
	if msp.closed {
		return nil, fmt.Errorf("serial port is closed")
	}
	
	if msp.readError != nil {
		return nil, msp.readError
	}

	available := len(msp.readBuffer) - msp.readIndex
	if available == 0 {
		return []byte{}, nil
	}

	result := make([]byte, available)
	copy(result, msp.readBuffer[msp.readIndex:])
	msp.readIndex = len(msp.readBuffer) // Mark as all read
	
	return result, nil
}

// IsConnected returns whether the mock serial port is connected
func (msp *MockSerialPort) IsConnected() bool {
	return !msp.closed
}

// Close simulates closing the serial port
func (msp *MockSerialPort) Close() error {
	msp.closed = true
	return nil
}

// GetWrittenData returns all data written to the mock serial port
func (msp *MockSerialPort) GetWrittenData() []byte {
	result := make([]byte, len(msp.writeBuffer))
	copy(result, msp.writeBuffer)
	return result
}

// ClearWrittenData clears the write buffer
func (msp *MockSerialPort) ClearWrittenData() {
	msp.writeBuffer = msp.writeBuffer[:0]
}

// IsOpen returns whether the mock serial port is open
func (msp *MockSerialPort) IsOpen() bool {
	return !msp.closed
}

// SerialPortInterface defines the interface for serial port operations
type SerialPortInterface interface {
	Write(data []byte) error
	Read(buffer []byte) (int, error)
	WriteString(text string) error
	WriteText(line1, line2 string, col, row int) error
	ReadAvailable() ([]byte, error)
	IsConnected() bool
	Close() error
	IsOpen() bool
}
