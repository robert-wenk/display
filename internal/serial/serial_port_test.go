package serial

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSerialPort(t *testing.T) {
	tests := []struct {
		name        string
		device      string
		baudRate    int
		expectError bool
	}{
		{
			name:        "Nonexistent device",
			device:      "/dev/nonexistent",
			baudRate:    9600,
			expectError: true,
		},
		{
			name:        "Invalid baud rate",
			device:      "/dev/ttyUSB0",
			baudRate:    -1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp, err := NewSerialPort(tt.device, tt.baudRate)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, sp)
			} else {
				// In test environment, this will likely fail due to missing hardware
				if err != nil {
					t.Logf("Expected test to pass but got error (likely due to missing hardware): %v", err)
				} else {
					assert.NotNil(t, sp)
					assert.Equal(t, tt.device, sp.config.Name)
					assert.Equal(t, tt.baudRate, sp.config.Baud)
					sp.Close()
				}
			}
		})
	}
}

func TestSerialPort_Close(t *testing.T) {
	sp := &SerialPort{
		port: nil, // Simulate uninitialized port
	}

	err := sp.Close()
	assert.NoError(t, err)
}

func TestSerialPort_Write_Uninitialized(t *testing.T) {
	sp := &SerialPort{
		port: nil,
	}

	err := sp.Write([]byte("test"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestSerialPort_Read_Uninitialized(t *testing.T) {
	sp := &SerialPort{
		port: nil,
	}

	buffer := make([]byte, 10)
	_, err := sp.Read(buffer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestSerialPort_WriteString(t *testing.T) {
	sp := &SerialPort{
		port: nil,
	}

	err := sp.WriteString("test")
	assert.Error(t, err) // Should error because port is nil
}

func TestSerialPort_IsOpen(t *testing.T) {
	sp := &SerialPort{
		port: nil,
	}

	assert.False(t, sp.IsOpen())
}

func TestMockSerialPort(t *testing.T) {
	mock := NewMockSerialPort()

	t.Run("Initial state", func(t *testing.T) {
		assert.True(t, mock.IsOpen())
		assert.Empty(t, mock.GetWrittenData())
	})

	t.Run("Write data", func(t *testing.T) {
		data := []byte("Hello, World!")
		err := mock.Write(data)
		assert.NoError(t, err)
		
		written := mock.GetWrittenData()
		assert.Equal(t, data, written)
	})

	t.Run("Write string", func(t *testing.T) {
		mock.ClearWrittenData()
		text := "Test String"
		err := mock.WriteString(text)
		assert.NoError(t, err)
		
		written := mock.GetWrittenData()
		assert.Equal(t, []byte(text), written)
	})

	t.Run("Read data", func(t *testing.T) {
		testData := []byte("Read Test Data")
		mock.SetReadData(testData)
		
		buffer := make([]byte, len(testData))
		n, err := mock.Read(buffer)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		assert.Equal(t, testData, buffer)
	})

	t.Run("Read partial data", func(t *testing.T) {
		testData := []byte("Partial Read Test")
		mock.SetReadData(testData)
		
		buffer := make([]byte, 7) // Smaller buffer
		n, err := mock.Read(buffer)
		assert.NoError(t, err)
		assert.Equal(t, 7, n)
		assert.Equal(t, testData[:7], buffer)
		
		// Read remaining data
		buffer2 := make([]byte, 10)
		n2, err := mock.Read(buffer2)
		assert.NoError(t, err)
		assert.Equal(t, len(testData)-7, n2)
		assert.Equal(t, testData[7:], buffer2[:n2])
	})

	t.Run("Read with no data", func(t *testing.T) {
		mock.SetReadData([]byte{})
		
		buffer := make([]byte, 10)
		n, err := mock.Read(buffer)
		assert.NoError(t, err)
		assert.Equal(t, 0, n)
	})

	t.Run("Write error", func(t *testing.T) {
		mock.SetWriteError(assert.AnError)
		
		err := mock.Write([]byte("test"))
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("Read error", func(t *testing.T) {
		mock.SetReadError(assert.AnError)
		
		buffer := make([]byte, 10)
		_, err := mock.Read(buffer)
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("Close", func(t *testing.T) {
		err := mock.Close()
		assert.NoError(t, err)
		assert.False(t, mock.IsOpen())
		
		// Operations after close should fail
		err = mock.Write([]byte("test"))
		assert.Error(t, err)
		
		buffer := make([]byte, 10)
		_, err = mock.Read(buffer)
		assert.Error(t, err)
	})

	t.Run("Clear written data", func(t *testing.T) {
		mock2 := NewMockSerialPort()
		mock2.Write([]byte("test data"))
		assert.NotEmpty(t, mock2.GetWrittenData())
		
		mock2.ClearWrittenData()
		assert.Empty(t, mock2.GetWrittenData())
	})
}
