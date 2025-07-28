package main

import (
	"fmt"
	"log"
	"time"

	"github.com/qnap/display-control/internal/config"
	"github.com/qnap/display-control/internal/controller"
)

func main() {
	// Create test configuration
	cfg := config.DefaultConfig()
	cfg.SerialPort.Device = "/dev/ttyS1"
	cfg.SerialPort.BaudRate = 1200

	fmt.Println("=== QNAP Button Test & Debug ===")
	
	// Initialize system controller (which includes display, LED, and USB copy monitoring)
	fmt.Println("1. Initializing system controller...")
	
	systemController, err := controller.NewSystemController(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize system controller: %v", err)
	}
	defer systemController.Close()

	// Set up detailed button monitoring using unified handler
	fmt.Println("2. Setting up unified button monitoring...")
	
	buttonEvents := make(chan string, 10)
	
	systemController.SetButtonHandler(func(button controller.PanelButton, pressed bool) {
		buttonName := ""
		switch button {
		case controller.ButtonEnter:
			buttonName = "ENTER"
		case controller.ButtonSelect:
			buttonName = "SELECT"
		case controller.ButtonUSBCopy:
			buttonName = "USB_COPY"
		default:
			buttonName = "UNKNOWN"
		}
		
		eventMsg := fmt.Sprintf("%s=%v", buttonName, pressed)
		fmt.Printf("🔘 Button Event: %s\n", eventMsg)
		
		if pressed {
			buttonEvents <- buttonName
		}
	})

	// Show instructions on display
	displayController := systemController.GetDisplayController()
	displayController.WriteText("Button Test\nPress buttons...")

	// Wait a moment for initialization
	time.Sleep(2 * time.Second)

	// Test each button with detailed feedback
	buttons := []struct {
		name        string
		description string
		timeout     time.Duration
	}{
		{"ENTER", "Press and release the ENTER button", 15 * time.Second},
		{"SELECT", "Press and release the SELECT button", 15 * time.Second},
		{"USB_COPY", "Press and release the USB COPY button", 20 * time.Second},
	}

	for i, btn := range buttons {
		fmt.Printf("\n3.%d Testing %s button...\n", i+1, btn.name)
		fmt.Printf("   %s\n", btn.description)
		
		// Update display
		displayController.WriteText(fmt.Sprintf("%s Test\nPress %s", btn.name, btn.name))
		
		// Wait for button press
		fmt.Printf("   Waiting for %s button press...\n", btn.name)
		
		select {
		case event := <-buttonEvents:
			if event == btn.name {
				fmt.Printf("   ✅ %s button detected successfully!\n", btn.name)
			} else {
				fmt.Printf("   ⚠️  Expected %s but got %s\n", btn.name, event)
			}
		case <-time.After(btn.timeout):
			fmt.Printf("   ❌ No %s button press detected within %v\n", btn.name, btn.timeout)
			
			// Check for any other events that might have been received
			fmt.Printf("   Checking for any button events...\n")
			eventCount := 0
			checkLoop:
			for {
				select {
				case event := <-buttonEvents:
					eventCount++
					fmt.Printf("   📝 Other event received: %s\n", event)
				case <-time.After(1 * time.Second):
					break checkLoop
				}
			}
			
			if eventCount == 0 {
				fmt.Printf("   📭 No button events received at all\n")
				fmt.Printf("   💡 Try checking:\n")
				fmt.Printf("      - Serial port connection (/dev/ttyS1)\n")
				fmt.Printf("      - QNAP button state reporting is enabled\n")
				fmt.Printf("      - Button wiring and hardware\n")
			}
		}
		
		// Clear any remaining events
		for len(buttonEvents) > 0 {
			<-buttonEvents
		}
		
		time.Sleep(1 * time.Second)
	}

	// Final test - listen for any button presses for 30 seconds
	fmt.Println("\n4. Final test - press any buttons for 30 seconds...")
	displayController.WriteText("Final Test\nPress any button")
	
	fmt.Println("   Listening for button events...")
	timeout := time.After(30 * time.Second)
	eventReceived := false
	
	for {
		select {
		case event := <-buttonEvents:
			eventReceived = true
			fmt.Printf("   🎯 Button event: %s\n", event)
		case <-timeout:
			fmt.Println("   ⏰ 30-second test period completed")
			if !eventReceived {
				fmt.Println("   ❌ No button events detected during final test")
				fmt.Println("   🔧 This indicates a button monitoring issue")
			} else {
				fmt.Println("   ✅ Button events were detected!")
			}
			goto testComplete
		}
	}

testComplete:
	displayController.WriteText("Test Complete\nCheck results")
	
	fmt.Println("\n=== Button Test Results ===")
	fmt.Println("✅ Button test completed")
	fmt.Println("📊 Check the output above for:")
	fmt.Println("   - Individual button detection results")
	fmt.Println("   - Serial data reception logs")
	fmt.Println("   - Button event callback execution")
	fmt.Println("   - Hardware troubleshooting tips")
	
	fmt.Println("\n🔧 Troubleshooting:")
	fmt.Println("   - Ensure running with sudo (hardware access required)")
	fmt.Println("   - Check /dev/ttyS1 exists and is accessible")  
	fmt.Println("   - Verify QNAP hardware button connections")
	fmt.Println("   - Check serial communication in system logs")
}
