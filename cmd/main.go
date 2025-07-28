package main

import (
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/qnap/display-control/internal/config"
	"github.com/qnap/display-control/internal/controller"
	"github.com/qnap/display-control/internal/menu"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configFile = flag.String("config", "/etc/qnap-display/config.json", "Path to configuration file")
	port       = flag.String("port", "/dev/ttyS1", "Serial port device")
	baudRate   = flag.Int("baud", 1200, "Serial port baud rate")
	verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	daemon     = flag.Bool("daemon", false, "Run as daemon")
)

// executeCopyCommand executes the USB copy command and shows progress
func executeCopyCommand(cfg *config.Config, systemController *controller.SystemController, menuSystem *menu.MenuSystem) {
	logrus.Info("Starting USB copy operation")
	
	displayController := systemController.GetDisplayController()
	
	// Show "Copy in progress" on first line
	if err := displayController.WriteTextAt("Copy in progress", 0, 0); err != nil {
		logrus.WithError(err).Error("Failed to show copy progress")
		return
	}
	
	// Clear second line initially
	if err := displayController.WriteTextAt("Starting...", 1, 0); err != nil {
		logrus.WithError(err).Error("Failed to clear second line")
	}
	
	// Flash disk LEDs to indicate activity
	if ledController := systemController.GetLEDController(); ledController != nil {
		ledController.SetLED(controller.USB, true)
		defer ledController.SetLED(controller.USB, false)
	}
	
	// Execute the copy command
	cmd := exec.Command("sh", "-c", cfg.USBCopy.Command)
	output, err := cmd.CombinedOutput()
	
	var statusLine string
	if err != nil {
		logrus.WithError(err).Error("Copy command failed")
		statusLine = "Copy failed"
	} else {
		logrus.Info("Copy command completed successfully")
		statusLine = "Copy complete"
		
		// Show truncated output if available
		if len(output) > 0 {
			outputStr := strings.TrimSpace(string(output))
			if len(outputStr) > 16 {
				statusLine = outputStr[:13] + "..."
			} else if len(outputStr) > 0 {
				statusLine = outputStr
			}
		}
	}
	
	// Show result on second line
	if err := displayController.WriteTextAt(statusLine, 1, 0); err != nil {
		logrus.WithError(err).Error("Failed to show copy result")
	}
	
	// Wait 3 seconds to show the result
	time.Sleep(3 * time.Second)
	
	// Return to menu system if it's running
	if menuSystem != nil {
		logrus.Info("Returning to menu system")
		// Refresh the menu display
		if err := menuSystem.RefreshDisplay(); err != nil {
			logrus.WithError(err).Error("Failed to refresh menu display")
		}
	} else {
		// Clear display if no menu system
		if err := displayController.ClearDisplay(); err != nil {
			logrus.WithError(err).Error("Failed to clear display")
		}
	}
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "qnap-display-control",
		Short: "QNAP Display Control with USB Copy Button Support",
		Long:  "A high-performance display controller for QNAP devices with USB copy button monitoring",
		Run:   runMain,
	}

	rootCmd.Flags().StringVarP(configFile, "config", "c", "/etc/qnap-display/config.json", "Configuration file path")
	rootCmd.Flags().StringVarP(port, "port", "p", "/dev/ttyS1", "Serial port device")
	rootCmd.Flags().IntVarP(baudRate, "baud", "b", 1200, "Serial port baud rate")
	rootCmd.Flags().BoolVarP(verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().BoolVarP(daemon, "daemon", "d", false, "Run as daemon")

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func runMain(cmd *cobra.Command, args []string) {
	// Configure logging
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.Info("Starting QNAP Display Control Service")

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		logrus.WithError(err).Warn("Failed to load config file, using defaults")
		cfg = config.DefaultConfig()
	}

	// Override config with command line flags
	if *port != "/dev/ttyS1" {
		cfg.SerialPort.Device = *port
	}
	if *baudRate != 1200 {
		cfg.SerialPort.BaudRate = *baudRate
	}

	// Initialize system controller (includes display and LED controllers)
	systemController, err := controller.NewSystemController(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize system controller")
	}
	defer systemController.Close()

	displayController := systemController.GetDisplayController()

	// Test display communication first
	if err := displayController.WriteText("QNAP Starting\nPlease wait..."); err != nil {
		logrus.WithError(err).Warn("Display test failed, but continuing")
	} else {
		logrus.Info("Display communication working")
		time.Sleep(2 * time.Second) // Show startup message
	}

	// Initialize menu system if enabled
	var menuSystem *menu.MenuSystem
	if cfg.Menu.Enabled {
		menuSystem = menu.NewMenuSystem(cfg, displayController)
		if err := menuSystem.Start(); err != nil {
			logrus.WithError(err).Error("Failed to start menu system")
			// Fallback to simple display
			if err := displayController.WriteText("Menu Failed\nBasic Mode"); err != nil {
				logrus.WithError(err).Error("Failed to display fallback message")
			}
		} else {
			logrus.Info("Menu system started successfully")
		}
		defer menuSystem.Stop()
	} else {
		// Show default message if menu is disabled
		if err := displayController.WriteText(cfg.Display.DefaultText + "\nMenu Disabled"); err != nil {
			logrus.WithError(err).Error("Failed to display default message")
		}
	}

	// Set up unified button handler for the system controller
	systemController.SetButtonHandler(func(button controller.PanelButton, pressed bool) {
		if !pressed {
			return // Only handle button press events, not releases
		}

		logrus.WithField("button", button).Info("Button event received")

		switch button {
		case controller.ButtonEnter:
			if menuSystem != nil {
				menuSystem.HandleEnterButton()
			}
		case controller.ButtonSelect:
			if menuSystem != nil {
				menuSystem.HandleSelectButton()
			}
		case controller.ButtonUSBCopy:
			logrus.Info("USB Copy button pressed")
			// Execute copy command in a goroutine to avoid blocking
			go executeCopyCommand(cfg, systemController, menuSystem)
		}
	})

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main event loop
	logrus.Info("QNAP Display Control Service started successfully")
	
	// Wait for shutdown signal
	sig := <-sigChan
	logrus.WithField("signal", sig).Info("Received shutdown signal")
}
