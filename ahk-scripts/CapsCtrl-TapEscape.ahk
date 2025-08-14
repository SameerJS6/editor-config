#Requires AutoHotkey v2.0

; Disable CapsLock toggle functionality
SetCapsLockState("AlwaysOff")

; Global variable to track if other keys were pressed
otherKeyPressed := false

; Handle CapsLock press
*CapsLock::
{
    global otherKeyPressed
    otherKeyPressed := false
    
    ; Start monitoring for other key presses
    SetTimer(CheckForOtherKeys, 10)
    
    ; Send Ctrl down immediately for responsiveness
    Send("{Ctrl down}")
}

; Handle CapsLock release  
*CapsLock Up::
{
    global otherKeyPressed
    
    ; Stop monitoring for other keys
    SetTimer(CheckForOtherKeys, 0)
    
    ; Release Ctrl
    Send("{Ctrl up}")
    
    ; If no other key was pressed, send Escape
    if (!otherKeyPressed) {
        Send("{Esc}")
    }
}

; Function to check for other key presses
CheckForOtherKeys() {
    global otherKeyPressed
    
    ; Check if CapsLock is still held
    if (!GetKeyState("CapsLock", "P")) {
        SetTimer(CheckForOtherKeys, 0)
        return
    }
    
    ; List of keys to monitor (using proper key names)
    keysToCheck := [
        "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
        "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
        "1", "2", "3", "4", "5", "6", "7", "8", "9", "0",
        "Space", "Enter", "Tab", "Backspace", "Delete", "Insert", "Home", "End",
        "PgUp", "PgDn", "Up", "Down", "Left", "Right",
        "F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12",
        "Shift", "Alt", "LWin", "RWin", "AppsKey",
        "SC027", "SC028", "SC033", "SC034", "SC035", "SC01A", "SC01B", "SC02B", "SC00C", "SC00D",
        "PrintScreen", "ScrollLock", "Pause"
    ]
    
    ; Check if any monitored key is pressed
    for key in keysToCheck {
        if (GetKeyState(key, "P")) {
            otherKeyPressed := true
            return
        }
    }
}

; Optional: Show startup notification
ToolTip("CapsLock: Hold=Ctrl, Tap=Esc", 0, 0)
SetTimer(() => ToolTip(), -2000)