package menu

import "strings"

// MockDisplayController is a mock implementation for testing
type MockDisplayController struct {
	LastText     string
	LastRow      int
	LastCol      int
	BacklightOn  bool
	WriteTextErr error
	WriteAtErr   error
	ClearErr     error
	BacklightErr error
	Calls        []string
}

// NewMockDisplayController creates a new mock display controller
func NewMockDisplayController() *MockDisplayController {
	return &MockDisplayController{
		Calls: make([]string, 0),
	}
}

// WriteTextAt mocks writing text at a specific position
func (m *MockDisplayController) WriteTextAt(text string, row, col int) error {
	m.Calls = append(m.Calls, "WriteTextAt")
	m.LastText = text
	m.LastRow = row
	m.LastCol = col
	return m.WriteAtErr
}

// WriteText mocks writing text to the display
func (m *MockDisplayController) WriteText(text string) error {
	m.Calls = append(m.Calls, "WriteText")
	m.LastText = text
	// Parse multi-line text for testing
	lines := strings.Split(text, "\n")
	if len(lines) > 0 {
		m.LastText = lines[0]
	}
	return m.WriteTextErr
}

// ClearDisplay mocks clearing the display
func (m *MockDisplayController) ClearDisplay() error {
	m.Calls = append(m.Calls, "ClearDisplay")
	m.LastText = ""
	return m.ClearErr
}

// SetBacklight mocks setting the backlight
func (m *MockDisplayController) SetBacklight(on bool) error {
	m.Calls = append(m.Calls, "SetBacklight")
	m.BacklightOn = on
	return m.BacklightErr
}

// Reset resets the mock state
func (m *MockDisplayController) Reset() {
	m.LastText = ""
	m.LastRow = 0
	m.LastCol = 0
	m.BacklightOn = false
	m.WriteTextErr = nil
	m.WriteAtErr = nil
	m.ClearErr = nil
	m.BacklightErr = nil
	m.Calls = make([]string, 0)
}
