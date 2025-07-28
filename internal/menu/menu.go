package menu

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/qnap/display-control/internal/config"
	"github.com/sirupsen/logrus"
)

// DisplayController interface for menu system
type DisplayController interface {
	WriteTextAt(text string, row, col int) error
	WriteText(text string) error
	ClearDisplay() error
	SetBacklight(on bool) error
}

// MenuSystem manages the menu navigation and display
type MenuSystem struct {
	config         *config.Config
	displayController DisplayController
	currentMenu    *config.MenuItem
	menuStack      []*config.MenuItem
	selectedIndex  int
	menuKeys       []string
	logger         *logrus.Logger
	
	// Output display state
	displayingOutput bool
	outputText       string
	scrollPosition   int
	stopOutputChan   chan bool
}

// NewMenuSystem creates a new menu system
func NewMenuSystem(cfg *config.Config, displayController DisplayController) *MenuSystem {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	ms := &MenuSystem{
		config:           cfg,
		displayController: displayController,
		logger:           logger,
		menuStack:        make([]*config.MenuItem, 0),
		stopOutputChan:   make(chan bool),
	}

	// Start with the main menu
	ms.currentMenu = &cfg.Menu.MainMenu
	ms.updateMenuKeys()

	return ms
}

// Start begins the menu system
func (ms *MenuSystem) Start() error {
	ms.logger.Info("Starting menu system")
	
	// Display the main menu
	if err := ms.displayCurrentMenu(); err != nil {
		return fmt.Errorf("failed to display main menu: %w", err)
	}

	ms.logger.Info("Menu system ready for button events")
	return nil
}

// readButtons is kept for compatibility but button handling is now done externally
func (ms *MenuSystem) readButtons() (string, error) {
	// Button handling is now done through the system controller's unified button events
	return "", nil
}

// handleSelectButton handles the SELECT button press (navigate through options)
func (ms *MenuSystem) handleSelectButton() {
	if len(ms.menuKeys) == 0 {
		return
	}

	ms.selectedIndex = (ms.selectedIndex + 1) % len(ms.menuKeys)
	ms.logger.WithFields(logrus.Fields{
		"selectedIndex": ms.selectedIndex,
		"selectedKey":   ms.menuKeys[ms.selectedIndex],
	}).Debug("SELECT button: moved to next option")
}

// handleEnterButton handles the ENTER button press (select current option)
func (ms *MenuSystem) handleEnterButton() {
	if len(ms.menuKeys) == 0 {
		return
	}

	selectedKey := ms.menuKeys[ms.selectedIndex]
	selectedItem := ms.currentMenu.Items[selectedKey]

	ms.logger.WithFields(logrus.Fields{
		"selectedKey":  selectedKey,
		"selectedItem": selectedItem.Title,
		"type":         selectedItem.Type,
	}).Info("ENTER button: selecting option")

	switch selectedItem.Type {
	case "submenu":
		// Navigate to submenu
		ms.navigateToSubmenu(&selectedItem)
	case "command":
		// Execute system command
		ms.executeCommand(selectedItem.Command)
	case "display_command":
		// Execute display-specific command
		ms.executeDisplayCommand(selectedItem.Command)
	case "back":
		// Go back to previous menu
		ms.navigateBack()
	}
}

// navigateToSubmenu navigates to a submenu
func (ms *MenuSystem) navigateToSubmenu(item *config.MenuItem) {
	// Push current menu to stack
	ms.menuStack = append(ms.menuStack, ms.currentMenu)
	
	// Set new current menu
	ms.currentMenu = item
	ms.selectedIndex = 0
	ms.updateMenuKeys()

	ms.logger.WithField("menu", item.Title).Info("Navigated to submenu")
}

// navigateBack goes back to the previous menu
func (ms *MenuSystem) navigateBack() {
	if len(ms.menuStack) == 0 {
		return // Already at root
	}

	// Pop from stack
	ms.currentMenu = ms.menuStack[len(ms.menuStack)-1]
	ms.menuStack = ms.menuStack[:len(ms.menuStack)-1]
	ms.selectedIndex = 0
	ms.updateMenuKeys()

	ms.logger.Info("Navigated back to previous menu")
}

// executeCommand executes a system command
func (ms *MenuSystem) executeCommand(command string) {
	ms.logger.WithField("command", command).Info("Executing system command")

	// Display "Executing..." message
	if err := ms.displayController.WriteText("Executing...\nPlease wait"); err != nil {
		ms.logger.WithError(err).Error("Failed to display executing message")
	}

	// Execute the command
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		ms.logger.WithError(err).Error("Command execution failed")
		ms.displayScrollingOutput(fmt.Sprintf("Error: %v", err))
	} else {
		ms.logger.Info("Command executed successfully")
		// Clean and prepare output for scrolling display
		cleanOutput := ms.prepareOutputForDisplay(string(output))
		ms.displayScrollingOutput(cleanOutput)
	}
}

// executeDisplayCommand handles QNAP display-specific commands
func (ms *MenuSystem) executeDisplayCommand(command string) {
	ms.logger.WithField("display_command", command).Info("Executing display command")

	switch command {
	case "backlight_on":
		ms.executeBacklightCommand(true)
	case "backlight_off":
		ms.executeBacklightCommand(false)
	default:
		ms.logger.WithField("command", command).Warn("Unknown display command")
		ms.displayScrollingOutput(fmt.Sprintf("Error: Unknown command '%s'", command))
	}
}

// executeBacklightCommand executes backlight control commands
func (ms *MenuSystem) executeBacklightCommand(on bool) {
	status := "off"
	if on {
		status = "on"
	}

	ms.logger.WithField("backlight_on", on).Info("Setting backlight")

	// Use the display controller's backlight method
	if err := ms.displayController.SetBacklight(on); err != nil {
		ms.logger.WithError(err).Error("Failed to set backlight")
		ms.displayScrollingOutput(fmt.Sprintf("Error: Backlight failed - %v", err))
	} else {
		ms.logger.Info("Backlight command sent successfully")
		ms.displayScrollingOutput(fmt.Sprintf("Success: Backlight %s", status))
	}
}

// prepareOutputForDisplay cleans command output for display
func (ms *MenuSystem) prepareOutputForDisplay(output string) string {
	// Remove control characters and excessive whitespace
	output = strings.ReplaceAll(output, "\r", "")
	output = strings.ReplaceAll(output, "\t", " ")
	
	// Split into lines and rejoin with spaces to create one continuous string
	lines := strings.Split(output, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	
	return strings.Join(cleanLines, " ")
}

// displayScrollingOutput displays output with horizontal scrolling
func (ms *MenuSystem) displayScrollingOutput(output string) {
	ms.logger.WithField("output", output).Debug("Starting scrolling output display")
	
	ms.displayingOutput = true
	ms.outputText = output
	ms.scrollPosition = 0
	
	// Start the scrolling display routine
	go ms.scrollOutputRoutine()
}

// scrollOutputRoutine handles the scrolling display of output
func (ms *MenuSystem) scrollOutputRoutine() {
	defer func() {
		ms.displayingOutput = false
		ms.scrollPosition = 0
		// Return to menu display
		if err := ms.displayCurrentMenu(); err != nil {
			ms.logger.WithError(err).Error("Failed to return to menu after output display")
		}
	}()

	displayWidth := ms.config.Display.Width
	outputLen := len(ms.outputText)
	
	// If output fits on display, just show it statically
	if outputLen <= displayWidth {
		if err := ms.displayController.WriteText(ms.outputText + "\nPress any button"); err != nil {
			ms.logger.WithError(err).Error("Failed to display short output")
			return
		}
		
		// Wait for button press
		select {
		case <-ms.stopOutputChan:
			return
		}
	}
	
	// For longer output, implement scrolling
	ticker := time.NewTicker(500 * time.Millisecond) // Scroll every 500ms
	defer ticker.Stop()
	
	for {
		select {
		case <-ms.stopOutputChan:
			return
		case <-ticker.C:
			// Create display window
			line1 := ms.getScrollingWindow(ms.outputText, ms.scrollPosition, displayWidth)
			line2 := "Press any button"
			
			// Display the current window
			if err := ms.displayController.WriteText(line1 + "\n" + line2); err != nil {
				ms.logger.WithError(err).Error("Failed to display scrolling output")
				return
			}
			
			// Advance scroll position
			ms.scrollPosition++
			
			// Reset scroll position when we've scrolled through the entire text
			maxScroll := outputLen - displayWidth + 1
			if maxScroll < 0 {
				maxScroll = 0
			}
			
			if ms.scrollPosition > maxScroll+displayWidth { // Add pause at end
				ms.scrollPosition = 0
			}
		}
	}
}

// getScrollingWindow extracts a window of text for scrolling display
func (ms *MenuSystem) getScrollingWindow(text string, position, width int) string {
	textLen := len(text)
	
	if position >= textLen {
		// We're past the end, show spaces or loop back
		return strings.Repeat(" ", width)
	}
	
	end := position + width
	if end > textLen {
		// Pad with spaces at the end
		window := text[position:]
		padding := width - len(window)
		return window + strings.Repeat(" ", padding)
	}
	
	return text[position:end]
}

// stopOutputDisplay stops the current output display
func (ms *MenuSystem) stopOutputDisplay() {
	if ms.displayingOutput {
		select {
		case ms.stopOutputChan <- true:
		default:
			// Channel might be full, that's okay
		}
	}
}

// updateMenuKeys updates the sorted list of menu keys
func (ms *MenuSystem) updateMenuKeys() {
	ms.menuKeys = make([]string, 0, len(ms.currentMenu.Items))
	
	for key := range ms.currentMenu.Items {
		ms.menuKeys = append(ms.menuKeys, key)
	}
	
	// Sort keys for consistent ordering
	sort.Strings(ms.menuKeys)
	
	// Add "back" option if not at root menu
	if len(ms.menuStack) > 0 {
		ms.menuKeys = append([]string{"back"}, ms.menuKeys...)
	}
	
	// Ensure selected index is valid
	if ms.selectedIndex >= len(ms.menuKeys) {
		ms.selectedIndex = 0
	}
}

// displayCurrentMenu displays the current menu on the LCD
func (ms *MenuSystem) displayCurrentMenu() error {
	if len(ms.menuKeys) == 0 {
		ms.logger.Warn("No menu items available")
		return ms.displayController.WriteText("No Menu Items\n")
	}

	// Get current selected item
	selectedKey := ms.menuKeys[ms.selectedIndex]
	var selectedItem config.MenuItem
	
	if selectedKey == "back" {
		selectedItem = config.MenuItem{
			Title:       "Back",
			Description: "Previous menu",
		}
	} else {
		selectedItem = ms.currentMenu.Items[selectedKey]
	}

	// First line: Menu description or current item title
	line1 := ms.currentMenu.Description
	if line1 == "" {
		line1 = ms.currentMenu.Title
	}
	
	// Second line: Current selection with indicator
	line2 := fmt.Sprintf(">%s", selectedItem.Title)
	
	// Truncate to display width (16 characters)
	if len(line1) > 16 {
		line1 = line1[:13] + "..."
	}
	if len(line2) > 16 {
		line2 = line2[:13] + "..."
	}

	ms.logger.WithFields(logrus.Fields{
		"line1":         line1,
		"line2":         line2,
		"selectedIndex": ms.selectedIndex,
		"selectedKey":   selectedKey,
	}).Info("Displaying menu")

	return ms.displayController.WriteText(line1 + "\n" + line2)
}

// GetCurrentMenuPath returns the current menu path for debugging
func (ms *MenuSystem) GetCurrentMenuPath() []string {
	path := make([]string, 0, len(ms.menuStack)+1)
	
	for _, menu := range ms.menuStack {
		path = append(path, menu.Title)
	}
	
	path = append(path, ms.currentMenu.Title)
	return path
}

// Stop stops the menu system
func (ms *MenuSystem) Stop() {
	ms.logger.Info("Stopping menu system")
	
	// Stop any ongoing output display
	ms.stopOutputDisplay()
	
	// Close the channel to prevent any further operations
	close(ms.stopOutputChan)
}

// HandleSelectButton is a public method to handle SELECT button presses from external sources
func (ms *MenuSystem) HandleSelectButton() {
	// If we're displaying output, stop it and return to menu
	if ms.displayingOutput {
		ms.stopOutputDisplay()
		return
	}
	
	ms.handleSelectButton()
	// Update display after button press
	if err := ms.displayCurrentMenu(); err != nil {
		ms.logger.WithError(err).Warn("Failed to update display after SELECT")
	}
}

// HandleEnterButton is a public method to handle ENTER button presses from external sources
func (ms *MenuSystem) HandleEnterButton() {
	// If we're displaying output, stop it and return to menu
	if ms.displayingOutput {
		ms.stopOutputDisplay()
		return
	}
	
	ms.handleEnterButton()
	// Update display after button press
	if err := ms.displayCurrentMenu(); err != nil {
		ms.logger.WithError(err).Warn("Failed to update display after ENTER")
	}
}

// RefreshDisplay refreshes the current menu display (public method for external use)
func (ms *MenuSystem) RefreshDisplay() error {
	return ms.displayCurrentMenu()
}
