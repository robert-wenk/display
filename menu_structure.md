# QNAP Display Control - Menu Structure & USB Copy

## Main Menu: "QNAP Control"

The menu system now includes the following options:

### 1. System Info
- **Type**: Command
- **Function**: Shows system information (`uname -a`)

### 2. Network
- **Type**: Submenu
- **Options**:
  - **Show IP**: Display IP address (`hostname -I`)
  - **Ping Test**: Test network connectivity (`ping -c 1 8.8.8.8`)
  - **← Back**: Return to main menu

### 3. Display ✨ *NEW*
- **Type**: Submenu  
- **Options**:
  - **Backlight On**: Turn display backlight on (display_command)
  - **Backlight Off**: Turn display backlight off (display_command)
  - **← Back**: Return to main menu

### 4. Storage
- **Type**: Command
- **Function**: Show storage information (`df -h`)

### 5. Reboot
- **Type**: Command
- **Function**: Restart system (`systemctl reboot`)

## USB Copy Button ✨ *NEW*

### Hardware Integration
- **Button Detection**: Real-time I/O port monitoring (default: 0xa05)
- **Polling Rate**: 50ms intervals for responsive detection
- **Background Operation**: Non-blocking execution in separate goroutine

### Copy Operation Flow
1. **Button Press**: USB copy button is pressed
2. **Display Update**: Shows "Copy in progress" on line 1
3. **Command Execution**: Runs configurable copy command
4. **Progress Display**: Shows status/output on line 2
5. **Completion**: Shows result for 3 seconds
6. **Menu Return**: Automatically returns to current menu

### Configuration
```json
{
  "usb_copy": {
    "io_port": 2565,
    "poll_interval_ms": 50,
    "enabled": true,
    "command": "cp -r /media/usb/* /share/USB_Copy/ && sync"
  }
}
```

### Copy Command Examples
- **Basic Copy**: `"cp -r /media/usb/* /share/USB_Copy/"`
- **With Sync**: `"cp -r /media/usb/* /share/USB_Copy/ && sync"`
- **RSYNC**: `"rsync -av /media/usb/ /share/USB_Copy/"`
- **With Logging**: `"cp -r /media/usb/* /share/USB_Copy/ 2>&1 | tee /var/log/usb_copy.log"`

## Navigation

- **SELECT Button**: Cycle through menu options
- **ENTER Button**: Select/execute current option
- **USB Copy Button**: Execute copy operation (hardware button)
- **Back Options**: Available in all submenus to return to parent menu

## Command Types

### 1. **command** - System Commands
- Executes shell commands via `sh -c`
- Shows "Executing..." message during execution
- Displays command output or error messages
- Used for: system info, network commands, storage info, reboot

### 2. **display_command** - Hardware Display Commands
- Dedicated command type for QNAP display control
- Direct hardware communication via serial port
- No shell command execution
- Clean separation of concerns
- Used for: backlight control, future display features

### 3. **submenu** - Navigation
- Navigates to submenu with additional options
- Maintains menu stack for back navigation

### 4. **back** - Navigation
- Returns to previous menu level
- Available in all submenus

## Architecture Benefits

✅ **Clean Separation**: System commands and display commands are handled separately
✅ **Extensible**: Easy to add new display commands without affecting system command logic  
✅ **Type Safety**: Clear command types prevent confusion
✅ **Maintainable**: Each command type has its own execution path
✅ **Testable**: Isolated command handlers for better testing
✅ **Hardware Integration**: USB copy button provides instant file copying functionality
✅ **Configurable**: Copy command fully customizable via config.json
