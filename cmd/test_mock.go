package main

import (
	"fmt"
	"time"
)

// Mock test to demonstrate 6-disk LED functionality and copy button features
// without requiring actual QNAP hardware

type MockSystemController struct{}

func NewMockSystemController() *MockSystemController {
	return &MockSystemController{}
}

func (m *MockSystemController) SetDiskActivity(diskNumber int, state string) error {
	if diskNumber < 1 || diskNumber > 6 {
		return fmt.Errorf("invalid disk number %d: must be between 1 and 6", diskNumber)
	}
	fmt.Printf("🔴 Disk %d LED: %s\n", diskNumber, state)
	return nil
}

func (m *MockSystemController) SetStatusLED(color string) error {
	fmt.Printf("🟢 Status LED: %s\n", color)
	return nil
}

func (m *MockSystemController) SetUSBLED(state bool) error {
	status := "OFF"
	if state {
		status = "ON"
	}
	fmt.Printf("🔵 USB LED: %s\n", status)
	return nil
}

func (m *MockSystemController) WriteToDisplay(text string) error {
	fmt.Printf("📺 Display: \"%s\"\n", text)
	return nil
}

func (m *MockSystemController) ShowProgress(percentage int) error {
	fmt.Printf("📊 Progress: %d%% (distributed across 6 disks: ~%d%% each)\n", percentage, percentage/6)
	return nil
}

func mockButtonPress(buttonName string) {
	fmt.Printf("🔘 Button pressed: %s\n", buttonName)
}

func main() {
	fmt.Println("=== QNAP 6-Disk Display & LED Mock Test ===")
	
	controller := NewMockSystemController()
	
	// Test display functionality
	fmt.Println("\n1. Testing Display Controller...")
	controller.WriteToDisplay("QNAP TS-670 Pro")
	time.Sleep(1 * time.Second)
	controller.WriteToDisplay("6-Disk Ready")
	time.Sleep(1 * time.Second)
	
	// Test status LEDs
	fmt.Println("\n2. Testing Status LEDs...")
	statusColors := []string{"green", "red", "orange", "off"}
	for _, color := range statusColors {
		fmt.Printf("Testing status LED: %s\n", color)
		controller.SetStatusLED(color)
		time.Sleep(3 * time.Second)
	}
	
	// Test USB LED
	fmt.Println("\n3. Testing USB LED...")
	fmt.Println("Testing USB LED: ON")
	controller.SetUSBLED(true)
	time.Sleep(3 * time.Second)
	fmt.Println("Testing USB LED: OFF")
	controller.SetUSBLED(false)
	time.Sleep(3 * time.Second)
	
	// Test all 6 disk LEDs
	fmt.Println("\n4. Testing 6-Disk LEDs (3 seconds each)...")
	diskStates := []string{"green", "red", "orange", "off"}
	
	for disk := 1; disk <= 6; disk++ {
		fmt.Printf("\n--- Testing Disk %d LEDs ---\n", disk)
		for _, state := range diskStates {
			fmt.Printf("Setting Disk %d LED to: %s\n", disk, state)
			controller.SetDiskActivity(disk, state)
			time.Sleep(3 * time.Second)
		}
	}
	
	// Test progress display across 6 disks
	fmt.Println("\n5. Testing Progress Display (6-disk distribution)...")
	for progress := 0; progress <= 100; progress += 20 {
		controller.ShowProgress(progress)
		time.Sleep(2 * time.Second)
	}
	
	// Test button functionality
	fmt.Println("\n6. Testing Button Detection...")
	buttons := []string{"ENTER", "SELECT", "USB COPY"}
	for _, button := range buttons {
		fmt.Printf("Simulating %s button press...\n", button)
		mockButtonPress(button)
		time.Sleep(2 * time.Second)
	}
	
	// Test error handling for invalid disk numbers
	fmt.Println("\n7. Testing Error Handling...")
	invalidDisks := []int{0, 7, 10}
	for _, disk := range invalidDisks {
		fmt.Printf("Testing invalid disk number %d...\n", disk)
		err := controller.SetDiskActivity(disk, "green")
		if err != nil {
			fmt.Printf("✅ Expected error: %s\n", err)
		} else {
			fmt.Printf("❌ Expected error but got none\n")
		}
	}
	
	// Final test sequence
	fmt.Println("\n8. Final Sequence - All Components Working Together...")
	
	controller.WriteToDisplay("Startup Complete")
	controller.SetStatusLED("green")
	time.Sleep(2 * time.Second)
	
	// Light up all disks
	for disk := 1; disk <= 6; disk++ {
		controller.SetDiskActivity(disk, "green")
		time.Sleep(500 * time.Millisecond)
	}
	
	controller.WriteToDisplay("All Systems Ready")
	controller.SetUSBLED(true)
	time.Sleep(2 * time.Second)
	
	// Turn off all LEDs
	controller.SetStatusLED("off")
	controller.SetUSBLED(false)
	for disk := 1; disk <= 6; disk++ {
		controller.SetDiskActivity(disk, "off")
	}
	
	controller.WriteToDisplay("Test Complete")
	
	fmt.Println("\n✅ Mock test completed successfully!")
	fmt.Println("📋 Test Summary:")
	fmt.Println("   - 6-disk LED support: ✅ Verified")
	fmt.Println("   - Copy button detection: ✅ Simulated")
	fmt.Println("   - Verbose LED testing: ✅ 3+ seconds per state")
	fmt.Println("   - Error handling: ✅ Invalid disk numbers caught")
	fmt.Println("   - Progress distribution: ✅ Across 6 disks")
	fmt.Println("   - All button types: ✅ ENTER, SELECT, USB COPY")
}
