package benchmark

import (
	"testing"
	"time"

	"github.com/qnap/display-control/internal/hardware"
	"github.com/qnap/display-control/internal/monitor"
)

// BenchmarkIOPortAccess benchmarks I/O port access operations
func BenchmarkIOPortAccess(b *testing.B) {
	mockIO := hardware.NewMockIOPortAccess(0xa05)
	defer mockIO.Close()

	b.Run("ReadByte", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := mockIO.ReadByte()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("WriteByte", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := mockIO.WriteByte(0xAA)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkUSBMonitor benchmarks USB copy monitor operations
func BenchmarkUSBMonitor(b *testing.B) {
	mockIO := hardware.NewMockIOPortAccess(0xa05)
	
	usbMonitor := &monitor.USBCopyMonitor{
		// Initialize with mock for benchmarking
	}
	defer func() {
		if usbMonitor != nil {
			usbMonitor.Close()
		}
		mockIO.Close()
	}()

	b.Run("IsButtonPressed", func(b *testing.B) {
		// This benchmark would need a properly initialized monitor
		// For now, we'll benchmark the mock directly
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := mockIO.ReadByte()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ButtonStateChange", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate button state changes
			if i%2 == 0 {
				mockIO.SetReadValue(0xFE) // Pressed
			} else {
				mockIO.SetReadValue(0xFF) // Not pressed
			}
			
			_, err := mockIO.ReadByte()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkSerialPort benchmarks serial port operations using mock
func BenchmarkSerialPort(b *testing.B) {
	// We'll use a simple benchmark since we can't assume real hardware
	data := []byte("Test display message")
	
	b.Run("WriteOperations", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			// In a real benchmark, this would write to a mock serial port
			// For now, we'll just measure the overhead of the operation
			_ = append([]byte(nil), data...)
		}
	})
}

// BenchmarkDisplayController benchmarks display controller operations
func BenchmarkDisplayController(b *testing.B) {
	// Benchmark display operations without actual hardware
	
	b.Run("TextFormatting", func(b *testing.B) {
		text := "USB Copy Status"
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			// Simulate text formatting operations
			formatted := make([]byte, len(text)+10)
			copy(formatted[2:], []byte(text))
			formatted[0] = 0xFE
			formatted[1] = 0x80
		}
	})

	b.Run("ProgressCalculation", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			percent := i % 101 // 0-100%
			barWidth := 14
			filled := (percent * barWidth) / 100
			
			// Simulate progress bar generation
			progressBar := make([]byte, barWidth+2)
			progressBar[0] = '['
			for j := 1; j <= barWidth; j++ {
				if j-1 < filled {
					progressBar[j] = '='
				} else {
					progressBar[j] = ' '
				}
			}
			progressBar[barWidth+1] = ']'
		}
	})
}

// BenchmarkConcurrentOperations benchmarks concurrent access patterns
func BenchmarkConcurrentOperations(b *testing.B) {
	mockIO := hardware.NewMockIOPortAccess(0xa05)
	defer mockIO.Close()

	b.Run("ConcurrentReads", func(b *testing.B) {
		b.SetParallelism(4)
		b.ResetTimer()
		
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := mockIO.ReadByte()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("ConcurrentWrites", func(b *testing.B) {
		b.SetParallelism(4)
		b.ResetTimer()
		
		b.RunParallel(func(pb *testing.PB) {
			value := byte(0xAA)
			for pb.Next() {
				err := mockIO.WriteByte(value)
				if err != nil {
					b.Fatal(err)
				}
				value++ // Vary the value
			}
		})
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("ByteSliceAllocation", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			// Simulate typical buffer allocations
			buffer := make([]byte, 32)
			_ = buffer
		}
	})

	b.Run("StringToByteConversion", func(b *testing.B) {
		text := "QNAP Display Control"
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			bytes := []byte(text)
			_ = bytes
		}
	})
}

// BenchmarkTimingOperations benchmarks timing-sensitive operations
func BenchmarkTimingOperations(b *testing.B) {
	b.Run("PollingLoop", func(b *testing.B) {
		mockIO := hardware.NewMockIOPortAccess(0xa05)
		defer mockIO.Close()
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			// Simulate a single poll operation
			_, err := mockIO.ReadByte()
			if err != nil {
				b.Fatal(err)
			}
			
			// Simulate minimal sleep (would be time.Sleep in real code)
			if i%1000 == 0 {
				time.Sleep(time.Nanosecond)
			}
		}
	})

	b.Run("ButtonDebouncing", func(b *testing.B) {
		mockIO := hardware.NewMockIOPortAccess(0xa05)
		defer mockIO.Close()
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			// Simulate debouncing algorithm
			samples := 3
			pressedCount := 0
			
			for j := 0; j < samples; j++ {
				value, err := mockIO.ReadByte()
				if err != nil {
					b.Fatal(err)
				}
				
				if (value & 0x01) == 0 {
					pressedCount++
				}
			}
			
			_ = pressedCount > samples/2
		}
	})
}
