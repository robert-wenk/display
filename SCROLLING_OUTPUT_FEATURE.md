# Horizontal Scrolling Output Feature

## Overview
Added horizontal scrolling functionality for command outputs that are longer than the display width (16 characters). The output now continuously scrolls and displays until any button is pressed.

## Key Features

### 1. **Horizontal Scrolling**
- **Auto-scrolling**: Text scrolls every 500ms when longer than display width
- **Continuous loop**: After reaching the end, scrolling restarts from the beginning
- **Smooth display**: Text moves one character position at a time

### 2. **User Interaction**
- **Button termination**: Any button press (ENTER or SELECT) stops output display
- **Return to menu**: After stopping output, system returns to menu navigation
- **Visual feedback**: "Press any button" message on second line

### 3. **Output Processing**
- **Text cleaning**: Removes control characters, excess whitespace, and newlines
- **Line consolidation**: Multiple output lines joined into single scrolling text
- **Padding**: Short text padded with spaces for consistent display

## Technical Implementation

### New MenuSystem Fields
```go
type MenuSystem struct {
    // ... existing fields ...
    
    // Output display state
    displayingOutput bool
    outputText       string
    scrollPosition   int
    stopOutputChan   chan bool
}
```

### Core Methods

#### `displayScrollingOutput(output string)`
- Initiates scrolling display for command output
- Starts background goroutine for scrolling animation
- Sets up termination channel

#### `scrollOutputRoutine()`
- Main scrolling loop running in background goroutine
- Handles both short (static) and long (scrolling) text
- Updates display every 500ms with new scroll position

#### `getScrollingWindow(text, position, width)`
- Extracts display window from full text at given position
- Handles text padding and boundary conditions
- Returns exactly `width` characters for display

#### `stopOutputDisplay()`
- Signals scrolling routine to stop via channel
- Called when any button is pressed during output display

### Button Handler Updates
```go
func (ms *MenuSystem) HandleSelectButton() {
    // If displaying output, stop it and return to menu
    if ms.displayingOutput {
        ms.stopOutputDisplay()
        return
    }
    // ... normal menu navigation ...
}
```

## Display Behavior

### Short Output (≤16 characters)
```
System output text
Press any button
```
- Static display, no scrolling
- Waits for button press

### Long Output (>16 characters)
```
This is a very long command output that needs to scroll horizontally...
Press any button
```
**Scrolling sequence:**
1. `This is a very l` + `Press any button`
2. `his is a very lo` + `Press any button` 
3. `is is a very lon` + `Press any button`
4. ... continues scrolling ...
5. Loops back to beginning

## Command Integration

### System Commands
- **Before**: Truncated output with "..." 
- **After**: Full output with horizontal scrolling
- **Examples**: `uname -a`, `df -h`, `hostname -I`, `ping` results

### Display Commands
- **Backlight control**: Success/error messages scroll if needed
- **Unknown commands**: Error messages display with scrolling

### Error Handling
- **Command failures**: Error messages scroll horizontally
- **Display errors**: Logged with fallback to menu

## Usage Examples

### Long System Info Output
```bash
Command: uname -a
Output: "Linux qnap-ts670 5.4.0-74-generic #83-Ubuntu SMP Sat May 8 02:35:39 UTC 2021 x86_64 x86_64 x86_64 GNU/Linux"
```
**Display scrolls through:**
- `Linux qnap-ts670 ` → `inux qnap-ts670 5` → `nux qnap-ts670 5.` → ...

### Network IP Display
```bash
Command: hostname -I
Output: "192.168.1.100 172.17.0.1 10.0.0.5"
```
**Display scrolls through:**
- `192.168.1.100 172` → `92.168.1.100 172.` → `2.168.1.100 172.1` → ...

### Storage Information
```bash
Command: df -h
Output: "Filesystem Size Used Avail Use% Mounted on /dev/sda1 20G 15G 4.2G 79% /"
```
**Display scrolls showing full filesystem information**

## Configuration

### Timing Settings
- **Scroll speed**: 500ms per character (configurable in code)
- **Loop pause**: Brief pause when reaching end before restart
- **Button response**: Immediate stop on any button press

### Display Settings
- **Width**: Uses `config.Display.Width` (default: 16 characters)
- **Height**: Line 1 for output, Line 2 for instruction
- **Padding**: Spaces added to maintain consistent width

## Benefits

1. **Complete Information**: Users see full command output, not truncated
2. **User Control**: Can stop scrolling and return to menu anytime
3. **Consistent UX**: Same button behavior (SELECT/ENTER) works everywhere
4. **Readable**: Smooth scrolling at comfortable speed
5. **Robust**: Handles various output formats and lengths

## Testing

Test the feature with various commands:
```bash
# Short output (static display)
uname -r

# Medium output (scrolls)
hostname -I

# Long output (extended scrolling)  
uname -a
df -h
cat /proc/cpuinfo | head -1
```

Press SELECT or ENTER during scrolling to verify immediate return to menu.
