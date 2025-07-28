package config

import (
	"encoding/json"
	"os"
)

// Config represents the application configuration
type Config struct {
	SerialPort SerialPortConfig `json:"serial_port"`
	USBCopy    USBCopyConfig    `json:"usb_copy"`
	Display    DisplayConfig    `json:"display"`
	Logging    LoggingConfig    `json:"logging"`
	Menu       MenuConfig       `json:"menu"`
}

// SerialPortConfig contains serial port settings
type SerialPortConfig struct {
	Device   string `json:"device"`
	BaudRate int    `json:"baud_rate"`
	Timeout  int    `json:"timeout_ms"`
}

// USBCopyConfig contains USB copy button settings
type USBCopyConfig struct {
	IOPort      uint16 `json:"io_port"`
	PollInterval int    `json:"poll_interval_ms"`
	Enabled     bool   `json:"enabled"`
	Command     string `json:"command"`
}

// DisplayConfig contains display settings
type DisplayConfig struct {
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	BacklightPin int    `json:"backlight_pin"`
	Contrast     int    `json:"contrast"`
	DefaultText  string `json:"default_text"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level    string `json:"level"`
	File     string `json:"file"`
	MaxSize  int    `json:"max_size_mb"`
	MaxAge   int    `json:"max_age_days"`
	Compress bool   `json:"compress"`
}

// MenuConfig contains menu system configuration
type MenuConfig struct {
	Enabled     bool       `json:"enabled"`
	MainMenu    MenuItem   `json:"main_menu"`
	ButtonDelay int        `json:"button_delay_ms"`
}

// MenuItem represents a single menu item
type MenuItem struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Type        string            `json:"type"` // "submenu", "command", "display_command", or "back"
	Command     string            `json:"command,omitempty"`
	Items       map[string]MenuItem `json:"items,omitempty"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		SerialPort: SerialPortConfig{
			Device:   "/dev/ttyS1",
			BaudRate: 1200,
			Timeout:  1000,
		},
		USBCopy: USBCopyConfig{
			IOPort:      0xa05,
			PollInterval: 50,
			Enabled:     true,
			Command:     "TIMESTAMP=$(date +%Y%m%d%H%M%S) && mkdir -p /mnt/pool/Multimedia/usb-copy$TIMESTAMP && cp -r /media/usb/* /mnt/pool/Multimedia/usb-copy$TIMESTAMP/ && sync && sleep 10",
		},
		Display: DisplayConfig{
			Width:        16,
			Height:       2,
			BacklightPin: -1,
			Contrast:     128,
			DefaultText:  "QNAP Ready",
		},
		Logging: LoggingConfig{
			Level:    "info",
			File:     "",
			MaxSize:  10,
			MaxAge:   30,
			Compress: true,
		},
		Menu: MenuConfig{
			Enabled:     true,
			ButtonDelay: 200,
			MainMenu: MenuItem{
				Title:       "Main Menu",
				Description: "QNAP Control",
				Type:        "submenu",
				Items: map[string]MenuItem{
					"system": {
						Title:       "System Info",
						Description: "Show system information",
						Type:        "command",
						Command:     "uname -a",
					},
					"network": {
						Title:       "Network",
						Description: "Network configuration",
						Type:        "submenu",
						Items: map[string]MenuItem{
							"ip": {
								Title:       "Show IP",
								Description: "Display IP address",
								Type:        "command",
								Command:     "hostname -I",
							},
							"ping": {
								Title:       "Ping Test",
								Description: "Test network connectivity",
								Type:        "command",
								Command:     "ping -c 1 8.8.8.8",
							},
							"back": {
								Title:       "← Back",
								Description: "Return to main menu",
								Type:        "back",
								Command:     "",
							},
						},
					},
					"display": {
						Title:       "Display",
						Description: "Display settings",
						Type:        "submenu",
						Items: map[string]MenuItem{
							"backlight_on": {
								Title:       "Backlight On",
								Description: "Turn backlight on",
								Type:        "display_command",
								Command:     "backlight_on",
							},
							"backlight_off": {
								Title:       "Backlight Off",
								Description: "Turn backlight off",
								Type:        "display_command",
								Command:     "backlight_off",
							},
							"back": {
								Title:       "← Back",
								Description: "Return to main menu",
								Type:        "back",
								Command:     "",
							},
						},
					},
					"storage": {
						Title:       "Storage",
						Description: "Storage information",
						Type:        "command",
						Command:     "df -h",
					},
					"reboot": {
						Title:       "Reboot",
						Description: "Restart system",
						Type:        "command",
						Command:     "systemctl reboot",
					},
				},
			},
		},
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves configuration to a JSON file
func (c *Config) SaveConfig(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
