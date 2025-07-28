package monitor

import (
	"testing"
	"time"

	"github.com/qnap/display-control/internal/hardware"
	"github.com/stretchr/testify/assert"
)

func TestNewUSBCopyMonitor(t *testing.T) {
	tests := []struct {
		name        string
		port        uint16
		expectError bool
	}{
		{
			name:        "Valid port",
			port:        0xa05,
			expectError: false, // May error if not root
		},
		{
			name:        "Zero port",
			port:        0x0,
			expectError: false, // May error if not root
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor, err := NewUSBCopyMonitor(tt.port)
			if err != nil {
				// This is expected in test environments without root access
				t.Logf("Monitor creation failed (expected in test environment): %v", err)
				return
			}
			
			assert.NotNil(t, monitor)
			assert.Equal(t, tt.port, monitor.port)
			
			// Clean up
			if monitor != nil {
				monitor.Close()
			}
		})
	}
}

func TestUSBCopyMonitor_Close(t *testing.T) {
	// Create a monitor with mock I/O port
	monitor := NewUSBCopyMonitorWithIOPort(0xa05, hardware.NewMockIOPortAccess(0xa05))

	err := monitor.Close()
	assert.NoError(t, err)
	assert.True(t, monitor.closed)

	// Second close should not error
	err = monitor.Close()
	assert.NoError(t, err)
}

func TestUSBCopyMonitor_IsButtonPressed(t *testing.T) {
	mockIO := hardware.NewMockIOPortAccess(0xa05)
	monitor := NewUSBCopyMonitorWithIOPort(0xa05, mockIO)

	tests := []struct {
		name         string
		portValue    byte
		expected     bool
		shouldError  bool
		closed       bool
	}{
		{
			name:      "Button pressed (bit 0 low)",
			portValue: 0xFE, // 11111110 - bit 0 is 0
			expected:  true,
		},
		{
			name:      "Button not pressed (bit 0 high)",
			portValue: 0xFF, // 11111111 - bit 0 is 1
			expected:  false,
		},
		{
			name:      "Button pressed with other bits",
			portValue: 0xAA, // 10101010 - bit 0 is 0
			expected:  true,
		},
		{
			name:        "Monitor closed",
			closed:      true,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.closed {
				monitor.closed = true
			} else {
				monitor.closed = false
				mockIO.SetReadValue(tt.portValue)
			}

			pressed, err := monitor.IsButtonPressed()
			
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, pressed)
			}
		})
	}
}

func TestUSBCopyMonitor_WaitForButtonPress(t *testing.T) {
	mockIO := hardware.NewMockIOPortAccess(0xa05)
	monitor := NewUSBCopyMonitorWithIOPort(0xa05, mockIO)

	t.Run("Button press detected", func(t *testing.T) {
		// Set button to pressed state
		mockIO.SetReadValue(0xFE)
		
		pressed, err := monitor.WaitForButtonPress(100 * time.Millisecond)
		assert.NoError(t, err)
		assert.True(t, pressed)
	})

	t.Run("Timeout without press", func(t *testing.T) {
		// Set button to not pressed state
		mockIO.SetReadValue(0xFF)
		
		start := time.Now()
		pressed, err := monitor.WaitForButtonPress(50 * time.Millisecond)
		duration := time.Since(start)
		
		assert.NoError(t, err)
		assert.False(t, pressed)
		assert.True(t, duration >= 50*time.Millisecond)
	})

	t.Run("Monitor closed during wait", func(t *testing.T) {
		mockIO.SetReadValue(0xFF)
		
		// Close the monitor after a short delay
		go func() {
			time.Sleep(25 * time.Millisecond)
			monitor.Close()
		}()
		
		pressed, err := monitor.WaitForButtonPress(100 * time.Millisecond)
		assert.Error(t, err)
		assert.False(t, pressed)
	})
}

func TestUSBCopyMonitor_GetButtonState(t *testing.T) {
	mockIO := hardware.NewMockIOPortAccess(0xa05)
	monitor := NewUSBCopyMonitorWithIOPort(0xa05, mockIO)

	t.Run("Consistent pressed state", func(t *testing.T) {
		mockIO.SetReadValue(0xFE) // Button pressed
		
		pressed, err := monitor.GetButtonState()
		assert.NoError(t, err)
		assert.True(t, pressed)
	})

	t.Run("Consistent not pressed state", func(t *testing.T) {
		mockIO.SetReadValue(0xFF) // Button not pressed
		
		pressed, err := monitor.GetButtonState()
		assert.NoError(t, err)
		assert.False(t, pressed)
	})
}
