package hardware

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIOPortAccess(t *testing.T) {
	tests := []struct {
		name        string
		port        uint16
		expectError bool
	}{
		{
			name:        "Valid port",
			port:        0xa05,
			expectError: os.Geteuid() != 0, // Expect error if not root
		},
		{
			name:        "Port 0x80",
			port:        0x80,
			expectError: os.Geteuid() != 0, // Expect error if not root
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io, err := NewIOPortAccess(tt.port)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, io)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, io)
				assert.Equal(t, tt.port, io.port)
				assert.True(t, io.acquired)
				
				// Clean up
				if io != nil {
					io.Close()
				}
			}
		})
	}
}

func TestIOPortAccess_Close(t *testing.T) {
	// Test with mock since we likely don't have root access
	io := &IOPortAccess{
		port:     0xa05,
		acquired: false, // Not actually acquired
	}

	err := io.Close()
	assert.NoError(t, err)
	assert.False(t, io.acquired)

	// Second close should not error
	err = io.Close()
	assert.NoError(t, err)
}

func TestIOPortAccess_ReadByte_NotAcquired(t *testing.T) {
	io := &IOPortAccess{
		port:     0xa05,
		acquired: false,
	}

	_, err := io.ReadByte()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not acquired")
}

func TestIOPortAccess_WriteByte_NotAcquired(t *testing.T) {
	io := &IOPortAccess{
		port:     0xa05,
		acquired: false,
	}

	err := io.WriteByte(0xFF)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not acquired")
}

func TestMockIOPortAccess(t *testing.T) {
	mock := NewMockIOPortAccess(0xa05)
	
	t.Run("Default read value", func(t *testing.T) {
		value, err := mock.ReadByte()
		assert.NoError(t, err)
		assert.Equal(t, byte(0xFF), value)
	})

	t.Run("Set custom read value", func(t *testing.T) {
		mock.SetReadValue(0xAA)
		value, err := mock.ReadByte()
		assert.NoError(t, err)
		assert.Equal(t, byte(0xAA), value)
	})

	t.Run("Set read error", func(t *testing.T) {
		mock.SetReadError(assert.AnError)
		_, err := mock.ReadByte()
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("Write and verify", func(t *testing.T) {
		err := mock.WriteByte(0x55)
		assert.NoError(t, err)
		assert.Equal(t, byte(0x55), mock.GetLastWrittenValue())
	})

	t.Run("Close", func(t *testing.T) {
		err := mock.Close()
		assert.NoError(t, err)
	})
}

func TestIsIOPortAccessAvailable(t *testing.T) {
	available := IsIOPortAccessAvailable()
	
	// The result depends on whether we're running as root
	if os.Geteuid() == 0 {
		// If we're root, it should be available (assuming Linux)
		t.Logf("Running as root, I/O port access available: %v", available)
	} else {
		// If we're not root, it should not be available
		assert.False(t, available, "I/O port access should not be available for non-root users")
	}
}

func TestInbFallback(t *testing.T) {
	// Test the fallback implementation
	value := inbFallback(0x80)
	
	// The fallback returns 0xFF when /dev/port is not accessible
	// or when read fails
	assert.Equal(t, byte(0xFF), value)
}

func TestOutbFallback(t *testing.T) {
	// Test the fallback implementation - should not panic
	assert.NotPanics(t, func() {
		outbFallback(0x80, 0xAA)
	})
}
