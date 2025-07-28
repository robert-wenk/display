package menu

import (
	"testing"

	"github.com/qnap/display-control/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMenuSystem(t *testing.T) {
	cfg := config.DefaultConfig()
	mockDisplay := NewMockDisplayController()
	
	ms := NewMenuSystem(cfg, mockDisplay)
	
	assert.NotNil(t, ms)
	assert.Equal(t, cfg, ms.config)
	assert.NotNil(t, ms.currentMenu)
	assert.Equal(t, "Main Menu", ms.currentMenu.Title)
}

func TestMenuNavigation(t *testing.T) {
	cfg := config.DefaultConfig()
	mockDisplay := NewMockDisplayController()
	
	ms := NewMenuSystem(cfg, mockDisplay)
	
	// Test initial state
	assert.Equal(t, 0, ms.selectedIndex)
	assert.Greater(t, len(ms.menuKeys), 0)
	
	// Test SELECT button (move to next option)
	initialIndex := ms.selectedIndex
	ms.handleSelectButton()
	expectedIndex := (initialIndex + 1) % len(ms.menuKeys)
	assert.Equal(t, expectedIndex, ms.selectedIndex)
}

func TestMenuPath(t *testing.T) {
	cfg := config.DefaultConfig()
	mockDisplay := NewMockDisplayController()
	
	ms := NewMenuSystem(cfg, mockDisplay)
	
	// Test initial path
	path := ms.GetCurrentMenuPath()
	assert.Equal(t, []string{"Main Menu"}, path)
	
	// Navigate to submenu (if network submenu exists)
	if networkItem, exists := ms.currentMenu.Items["network"]; exists && networkItem.Type == "submenu" {
		ms.navigateToSubmenu(&networkItem)
		
		path = ms.GetCurrentMenuPath()
		assert.Equal(t, []string{"Main Menu", "Network"}, path)
		
		// Test navigation back
		ms.navigateBack()
		path = ms.GetCurrentMenuPath()
		assert.Equal(t, []string{"Main Menu"}, path)
	}
}

func TestDisplayCurrentMenu(t *testing.T) {
	cfg := config.DefaultConfig()
	mockDisplay := NewMockDisplayController()
	
	ms := NewMenuSystem(cfg, mockDisplay)
	
	// Test displaying current menu
	err := ms.displayCurrentMenu()
	require.NoError(t, err)
	
	// Check that display was called
	assert.Contains(t, mockDisplay.Calls, "WriteText")
	assert.NotEmpty(t, mockDisplay.LastText)
}

func TestButtonReading(t *testing.T) {
	cfg := config.DefaultConfig()
	mockDisplay := NewMockDisplayController()
	
	ms := NewMenuSystem(cfg, mockDisplay)
	
	// Test no button press (simplified since button handling moved to display controller)
	button, err := ms.readButtons()
	require.NoError(t, err)
	assert.Equal(t, "", button)
}
