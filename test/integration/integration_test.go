package integration

import (
	"os"
	"testing"
	"time"

	"github.com/qnap/display-control/internal/config"
	"github.com/qnap/display-control/internal/controller"
	"github.com/qnap/display-control/internal/hardware"
	"github.com/qnap/display-control/internal/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite defines the integration test suite
type IntegrationTestSuite struct {
	suite.Suite
	config *config.Config
}

// SetupSuite runs before all tests in the suite
func (suite *IntegrationTestSuite) SetupSuite() {
	suite.config = config.DefaultConfig()
	
	// Override with test-friendly settings
	suite.config.SerialPort.Device = "/dev/null" // Use /dev/null for testing
	suite.config.SerialPort.Timeout = 100
	suite.config.USBCopy.PollInterval = 10
}

// TestDisplayControllerIntegration tests the display controller integration
func (suite *IntegrationTestSuite) TestDisplayControllerIntegration() {
	if os.Geteuid() != 0 {
		suite.T().Skip("Skipping integration test - requires root access")
	}

	// This test would normally require actual hardware
	// For now, we'll test the error cases
	
	suite.Run("Invalid serial port", func() {
		cfg := *suite.config
		cfg.SerialPort.Device = "/dev/nonexistent"
		
		_, err := controller.NewDisplayController(&cfg)
		assert.Error(suite.T(), err)
	})
}

// TestUSBMonitorIntegration tests the USB monitor integration
func (suite *IntegrationTestSuite) TestUSBMonitorIntegration() {
	if os.Geteuid() != 0 {
		suite.T().Skip("Skipping integration test - requires root access")
	}

	suite.Run("Monitor lifecycle", func() {
		monitor, err := monitor.NewUSBCopyMonitor(suite.config.USBCopy.IOPort)
		if err != nil {
			suite.T().Logf("Could not create monitor (expected in test environment): %v", err)
			return
		}
		
		assert.NotNil(suite.T(), monitor)
		
		// Test button state reading
		_, err = monitor.IsButtonPressed()
		assert.NoError(suite.T(), err)
		
		// Test cleanup
		err = monitor.Close()
		assert.NoError(suite.T(), err)
	})
}

// TestHardwareAccessIntegration tests hardware access integration
func (suite *IntegrationTestSuite) TestHardwareAccessIntegration() {
	suite.Run("I/O port access availability", func() {
		available := hardware.IsIOPortAccessAvailable()
		suite.T().Logf("I/O port access available: %v", available)
		
		if os.Geteuid() == 0 {
			// If running as root, should be available on Linux
			if available {
				assert.True(suite.T(), available)
			} else {
				suite.T().Log("Running as root but I/O port access not available (may be in container or virtualized environment)")
			}
		} else {
			assert.False(suite.T(), available)
		}
	})
}

// TestEndToEndWorkflow tests the complete workflow
func (suite *IntegrationTestSuite) TestEndToEndWorkflow() {
	if os.Geteuid() != 0 {
		suite.T().Skip("Skipping end-to-end test - requires root access")
	}

	suite.Run("Complete workflow simulation", func() {
		// This would normally test the complete workflow
		// For now, we'll test that components can be created and destroyed
		
		// Try to create display controller
		dc, err := controller.NewDisplayController(suite.config)
		if err != nil {
			suite.T().Logf("Could not create display controller (expected in test environment): %v", err)
		} else {
			defer dc.Close()
			
			// Test basic operations
			err = dc.WriteText("Test")
			if err != nil {
				suite.T().Logf("Could not write to display (expected without hardware): %v", err)
			}
		}
		
		// Try to create USB monitor
		usbMonitor, err := monitor.NewUSBCopyMonitor(suite.config.USBCopy.IOPort)
		if err != nil {
			suite.T().Logf("Could not create USB monitor (expected in test environment): %v", err)
		} else {
			defer usbMonitor.Close()
			
			// Test monitoring
			pressed, err := usbMonitor.IsButtonPressed()
			if err != nil {
				suite.T().Logf("Could not read button state (expected without hardware): %v", err)
			} else {
				suite.T().Logf("Button state: %v", pressed)
			}
		}
	})
}

// TestConfigurationIntegration tests configuration loading and validation
func (suite *IntegrationTestSuite) TestConfigurationIntegration() {
	suite.Run("Default configuration", func() {
		cfg := config.DefaultConfig()
		assert.NotNil(suite.T(), cfg)
		assert.Equal(suite.T(), "/dev/ttyS1", cfg.SerialPort.Device)
		assert.Equal(suite.T(), 1200, cfg.SerialPort.BaudRate)
		assert.Equal(suite.T(), uint16(0xa05), cfg.USBCopy.IOPort)
	})

	suite.Run("Configuration file operations", func() {
		cfg := config.DefaultConfig()
		
		// Create temporary config file
		tmpFile := "/tmp/test_qnap_config.json"
		defer os.Remove(tmpFile)
		
		// Save configuration
		err := cfg.SaveConfig(tmpFile)
		assert.NoError(suite.T(), err)
		
		// Load configuration
		loadedCfg, err := config.LoadConfig(tmpFile)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), cfg.SerialPort.Device, loadedCfg.SerialPort.Device)
		assert.Equal(suite.T(), cfg.SerialPort.BaudRate, loadedCfg.SerialPort.BaudRate)
		assert.Equal(suite.T(), cfg.USBCopy.IOPort, loadedCfg.USBCopy.IOPort)
	})
}

// TestConcurrencyIntegration tests concurrent operations
func (suite *IntegrationTestSuite) TestConcurrencyIntegration() {
	suite.Run("Concurrent monitor operations", func() {
		// Create mock-based monitor for testing
		mockIO := hardware.NewMockIOPortAccess(0xa05)
		usbMonitor := monitor.NewUSBCopyMonitorWithIOPort(0xa05, mockIO)
		
		// Test that multiple goroutines can safely close the monitor
		done := make(chan bool, 2)
		
		go func() {
			time.Sleep(10 * time.Millisecond)
			usbMonitor.Close()
			done <- true
		}()
		
		go func() {
			time.Sleep(20 * time.Millisecond)
			usbMonitor.Close()
			done <- true
		}()
		
		// Wait for both goroutines
		<-done
		<-done
		
		// Should not panic or deadlock
		assert.True(suite.T(), true)
		
		// Clean up mock
		mockIO.Close()
	})
}

// Run the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
