package error

import (
	"fmt"
	"runtime"
)

// ErrorType represents different categories of errors
type ErrorType int

const (
	// ErrorTypeUnknown represents an unknown error
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeSerialPort represents serial port related errors
	ErrorTypeSerialPort
	// ErrorTypeIOPort represents I/O port access errors
	ErrorTypeIOPort
	// ErrorTypeDisplay represents display controller errors
	ErrorTypeDisplay
	// ErrorTypeUSBMonitor represents USB copy monitor errors
	ErrorTypeUSBMonitor
	// ErrorTypeConfig represents configuration errors
	ErrorTypeConfig
	// ErrorTypePermission represents permission/privilege errors
	ErrorTypePermission
	// ErrorTypeHardware represents hardware access errors
	ErrorTypeHardware
)

// String returns the string representation of the error type
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeSerialPort:
		return "SerialPort"
	case ErrorTypeIOPort:
		return "IOPort"
	case ErrorTypeDisplay:
		return "Display"
	case ErrorTypeUSBMonitor:
		return "USBMonitor"
	case ErrorTypeConfig:
		return "Config"
	case ErrorTypePermission:
		return "Permission"
	case ErrorTypeHardware:
		return "Hardware"
	default:
		return "Unknown"
	}
}

// QNAPError represents a structured error with context
type QNAPError struct {
	Type      ErrorType
	Message   string
	Cause     error
	Context   map[string]interface{}
	File      string
	Line      int
	Function  string
}

// Error implements the error interface
func (e *QNAPError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *QNAPError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target
func (e *QNAPError) Is(target error) bool {
	if qnapErr, ok := target.(*QNAPError); ok {
		return e.Type == qnapErr.Type
	}
	return false
}

// WithContext adds context information to the error
func (e *QNAPError) WithContext(key string, value interface{}) *QNAPError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// GetContext retrieves context information from the error
func (e *QNAPError) GetContext(key string) (interface{}, bool) {
	if e.Context == nil {
		return nil, false
	}
	value, exists := e.Context[key]
	return value, exists
}

// NewError creates a new QNAP error with caller information
func NewError(errType ErrorType, message string) *QNAPError {
	pc, file, line, _ := runtime.Caller(1)
	function := runtime.FuncForPC(pc).Name()

	return &QNAPError{
		Type:     errType,
		Message:  message,
		File:     file,
		Line:     line,
		Function: function,
	}
}

// WrapError wraps an existing error with QNAP error context
func WrapError(errType ErrorType, message string, cause error) *QNAPError {
	pc, file, line, _ := runtime.Caller(1)
	function := runtime.FuncForPC(pc).Name()

	return &QNAPError{
		Type:     errType,
		Message:  message,
		Cause:    cause,
		File:     file,
		Line:     line,
		Function: function,
	}
}

// Convenience functions for common error types

// NewSerialPortError creates a new serial port error
func NewSerialPortError(message string) *QNAPError {
	return NewError(ErrorTypeSerialPort, message)
}

// WrapSerialPortError wraps an existing error as a serial port error
func WrapSerialPortError(message string, cause error) *QNAPError {
	return WrapError(ErrorTypeSerialPort, message, cause)
}

// NewIOPortError creates a new I/O port error
func NewIOPortError(message string) *QNAPError {
	return NewError(ErrorTypeIOPort, message)
}

// WrapIOPortError wraps an existing error as an I/O port error
func WrapIOPortError(message string, cause error) *QNAPError {
	return WrapError(ErrorTypeIOPort, message, cause)
}

// NewDisplayError creates a new display error
func NewDisplayError(message string) *QNAPError {
	return NewError(ErrorTypeDisplay, message)
}

// WrapDisplayError wraps an existing error as a display error
func WrapDisplayError(message string, cause error) *QNAPError {
	return WrapError(ErrorTypeDisplay, message, cause)
}

// NewUSBMonitorError creates a new USB monitor error
func NewUSBMonitorError(message string) *QNAPError {
	return NewError(ErrorTypeUSBMonitor, message)
}

// WrapUSBMonitorError wraps an existing error as a USB monitor error
func WrapUSBMonitorError(message string, cause error) *QNAPError {
	return WrapError(ErrorTypeUSBMonitor, message, cause)
}

// NewConfigError creates a new configuration error
func NewConfigError(message string) *QNAPError {
	return NewError(ErrorTypeConfig, message)
}

// WrapConfigError wraps an existing error as a configuration error
func WrapConfigError(message string, cause error) *QNAPError {
	return WrapError(ErrorTypeConfig, message, cause)
}

// NewPermissionError creates a new permission error
func NewPermissionError(message string) *QNAPError {
	return NewError(ErrorTypePermission, message)
}

// WrapPermissionError wraps an existing error as a permission error
func WrapPermissionError(message string, cause error) *QNAPError {
	return WrapError(ErrorTypePermission, message, cause)
}

// NewHardwareError creates a new hardware error
func NewHardwareError(message string) *QNAPError {
	return NewError(ErrorTypeHardware, message)
}

// WrapHardwareError wraps an existing error as a hardware error
func WrapHardwareError(message string, cause error) *QNAPError {
	return WrapError(ErrorTypeHardware, message, cause)
}

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, errType ErrorType) bool {
	if qnapErr, ok := err.(*QNAPError); ok {
		return qnapErr.Type == errType
	}
	return false
}

// GetErrorType returns the error type if it's a QNAP error
func GetErrorType(err error) (ErrorType, bool) {
	if qnapErr, ok := err.(*QNAPError); ok {
		return qnapErr.Type, true
	}
	return ErrorTypeUnknown, false
}
