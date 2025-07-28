package serial

import (
	"fmt"
	"time"
)

// DisplayTester provides simple display testing functionality
type DisplayTester struct {
	port SerialPortInterface
}

// NewDisplayTester creates a new display tester
func NewDisplayTester(port SerialPortInterface) *DisplayTester {
	return &DisplayTester{port: port}
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
