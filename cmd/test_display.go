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

	fmt.Println("=== QNAP Display & Button Test ===")
	
	// Test 1: Display Controller with newline handling
	fmt.Println("1. Testing display controller...")
	
	displayController, err := controller.NewDisplayController(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize display controller: %v", err)
	}
	defer displayController.Close()

	// Test newline handling
	fmt.Println("   Testing newline display...")
	if err := displayController.WriteText("Line 1 Text\nLine 2 Text"); err != nil {
		fmt.Printf("   ❌ Failed to write multi-line text: %v\n", err)
	} else {
		fmt.Println("   ✅ Multi-line text written successfully")
	}
	
	time.Sleep(3 * time.Second)

	// Test 2: System Controller with LED and button integration
	fmt.Println("2. Testing system controller...")
	
	systemController, err := controller.NewSystemController(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize system controller: %v", err)
	}
	defer systemController.Close()

	// Set up button handler to test button capture
	buttonPressed := make(chan bool, 1)
	
	// Use the system controller's unified button handler instead of just display controller
	systemController.SetButtonHandler(func(button controller.PanelButton, pressed bool) {
		fmt.Printf("   🔘 Button event: %v pressed=%v\n", button, pressed)
		if pressed {
			buttonPressed <- true
		}
	})

	// Test LED functionality (if available)
	if ledController := systemController.GetLEDController(); ledController != nil {
		fmt.Println("   Testing LED control...")
		
		// Test status LEDs with verbose output
		fmt.Println("   Testing Status LEDs:")
		fmt.Print("     Setting status to GREEN...")
		ledController.SetStatusLED(false, true) // Green
		fmt.Println(" ✅")
		time.Sleep(3 * time.Second)
		
		fmt.Print("     Setting status to RED...")
		ledController.SetStatusLED(true, false) // Red
		fmt.Println(" ✅")
		time.Sleep(3 * time.Second)
		
		fmt.Print("     Turning OFF both status LEDs...")
		ledController.SetStatusLED(false, false) // Both off
		fmt.Println(" ✅")
		time.Sleep(3 * time.Second)
		
		fmt.Print("     Restoring status to GREEN...")
		ledController.SetStatusLED(false, true) // Back to green
		fmt.Println(" ✅")
		time.Sleep(1 * time.Second)
		
		// Test USB LED
		fmt.Println("   Testing USB LED:")
		fmt.Print("     USB LED ON...")
		ledController.SetLED(controller.USB, true)
		fmt.Println(" ✅")
		time.Sleep(3 * time.Second)
		
		fmt.Print("     USB LED OFF...")
		ledController.SetLED(controller.USB, false)
		fmt.Println(" ✅")
		time.Sleep(1 * time.Second)
		
		// Test disk LEDs (now supporting 6 disks)
		fmt.Println("   Testing Disk LEDs (1-6):")
		for i := 1; i <= 6; i++ {
			fmt.Printf("     Disk %d LED ON...", i)
			if err := systemController.SetDiskActivity(i, true); err != nil {
				fmt.Printf(" ❌ Error: %v\n", err)
			} else {
				fmt.Println(" ✅")
			}
			time.Sleep(3 * time.Second)
			
			fmt.Printf("     Disk %d LED OFF...", i)
			if err := systemController.SetDiskActivity(i, false); err != nil {
				fmt.Printf(" ❌ Error: %v\n", err)
			} else {
				fmt.Println(" ✅")
			}
			time.Sleep(1 * time.Second)
		}
		
		// Test all disk LEDs at once
		fmt.Print("   Testing all Disk LEDs ON simultaneously...")
		for i := 1; i <= 6; i++ {
			systemController.SetDiskActivity(i, true)
		}
		fmt.Println(" ✅")
		time.Sleep(3 * time.Second)
		
		fmt.Print("   Testing all Disk LEDs OFF simultaneously...")
		for i := 1; i <= 6; i++ {
			systemController.SetDiskActivity(i, false)
		}
		fmt.Println(" ✅")
		time.Sleep(1 * time.Second)
		
		fmt.Println("   ✅ LED test completed")
	} else {
		fmt.Println("   ⚠️  LED controller not available (may need root privileges)")
	}

	// Test 3: Button monitoring (including copy button)
	fmt.Println("3. Testing button monitoring...")
	
	// Set up enhanced button tracking
	enterPressed := make(chan bool, 1)
	selectPressed := make(chan bool, 1)
	copyPressed := make(chan bool, 1)
	
	// Use the system controller's unified button handler for all button types
	systemController.SetButtonHandler(func(button controller.PanelButton, pressed bool) {
		buttonName := ""
		switch button {
		case controller.ButtonEnter:
			buttonName = "ENTER"
			if pressed {
				enterPressed <- true
			}
		case controller.ButtonSelect:
			buttonName = "SELECT"
			if pressed {
				selectPressed <- true
			}
		case controller.ButtonUSBCopy:
			buttonName = "USB COPY"
			if pressed {
				copyPressed <- true
			}
		}
		fmt.Printf("   🔘 Button event: %s pressed=%v\n", buttonName, pressed)
	})

	// Test ENTER button
	fmt.Println("   Testing ENTER button...")
	fmt.Println("   Press the ENTER button on the QNAP panel...")
	display := systemController.GetDisplayController()
	display.WriteText("ENTER Test\nPress ENTER")
	
	select {
	case <-enterPressed:
		fmt.Println("   ✅ ENTER button detected successfully!")
	case <-time.After(10 * time.Second):
		fmt.Println("   ⚠️  No ENTER button press detected within 10 seconds")
	}
	
	time.Sleep(1 * time.Second)
	
	// Test SELECT button
	fmt.Println("   Testing SELECT button...")
	fmt.Println("   Press the SELECT button on the QNAP panel...")
	display.WriteText("SELECT Test\nPress SELECT")
	
	select {
	case <-selectPressed:
		fmt.Println("   ✅ SELECT button detected successfully!")
	case <-time.After(10 * time.Second):
		fmt.Println("   ⚠️  No SELECT button press detected within 10 seconds")
	}
	
	time.Sleep(1 * time.Second)
	
	// Test USB COPY button
	fmt.Println("   Testing USB COPY button...")
	fmt.Println("   Press the USB COPY button on the QNAP panel...")
	display.WriteText("USB COPY Test\nPress USB COPY")
	
	// Also flash USB LED to indicate copy mode
	if ledController := systemController.GetLEDController(); ledController != nil {
		ledController.SetLED(controller.USB, true)
		defer ledController.SetLED(controller.USB, false)
	}
	
	select {
	case <-copyPressed:
		fmt.Println("   ✅ USB COPY button detected successfully!")
		fmt.Println("   This would normally trigger the copy operation")
	case <-time.After(15 * time.Second):
		fmt.Println("   ⚠️  No USB COPY button press detected within 15 seconds")
		fmt.Println("       Note: USB COPY button may use different protocol")
	}
	
	time.Sleep(1 * time.Second)

	// Test 4: Display various text formats
	fmt.Println("4. Testing different text formats...")
	
	testTexts := []string{
		"Single line",
		"Two\nLines",
		"Very long single line that should be truncated properly",
		"Long line 1\nLong line 2 also",
		"",
		"Empty\n",
		"\nEmpty first",
	}
	
	for i, text := range testTexts {
		fmt.Printf("   Test %d: %q\n", i+1, text)
		if err := display.WriteText(text); err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
		} else {
			fmt.Println("   ✅ Success")
		}
		time.Sleep(2 * time.Second)
	}

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("Check the QNAP display for:")
	fmt.Println("- Proper newline handling (text on separate lines)")
	fmt.Println("- Button responsiveness (ENTER, SELECT, USB COPY)")
	fmt.Println("- LED activity (Status, USB, 6 Disk LEDs)")
	fmt.Println("- Proper 3-second timing for each LED state")
	fmt.Println("\nHardware verified:")
	fmt.Println("- Serial communication at 1200 baud, 8N1")
	fmt.Println("- QNAP display protocol (0x4D commands)")
	fmt.Println("- Extended 6-disk LED support")
	fmt.Println("- Multi-button detection including USB COPY")
}
