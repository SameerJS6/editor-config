#Requires AutoHotkey v2.0

; Double Shift to Toggle CapsLock Script
; Only triggers when Shift is pressed and released quickly twice
; without any other keys being pressed in between

; Global variables
lastShiftUpTime := 0
doubleTapThreshold := 300  ; Time in ms between taps
shiftHeldTime := 0
otherKeyUsed := false
waitingForSecondTap := false

; Track when any other key is pressed to avoid false triggers
; Letters
~*a:: ResetDoubleTap()
~*b:: ResetDoubleTap()
~*c::ResetDoubleTap()
~*d:: ResetDoubleTap()
~*e::ResetDoubleTap()
~*f:: ResetDoubleTap()
~*g::ResetDoubleTap()
~*h::ResetDoubleTap()
~*i::ResetDoubleTap()
~*j:: ResetDoubleTap()
~*k::ResetDoubleTap()
~*l:: ResetDoubleTap()
~*m::ResetDoubleTap()
~*n:: ResetDoubleTap()
~*o:: ResetDoubleTap()
~*p:: ResetDoubleTap()
~*q:: ResetDoubleTap()
~*r:: ResetDoubleTap()
~*s:: ResetDoubleTap()
~*t:: ResetDoubleTap()
~*u:: ResetDoubleTap()
~*v:: ResetDoubleTap()
~*w:: ResetDoubleTap()
~*x:: ResetDoubleTap()
~*y:: ResetDoubleTap()
~*z::ResetDoubleTap()

; Numbers
~*1:: ResetDoubleTap()
~*2:: ResetDoubleTap()
~*3:: ResetDoubleTap()
~*4:: ResetDoubleTap()
~*5:: ResetDoubleTap()
~*6:: ResetDoubleTap()
~*7:: ResetDoubleTap()
~*8:: ResetDoubleTap()
~*9:: ResetDoubleTap()
~*0:: ResetDoubleTap()

; Common keys
~*Space:: ResetDoubleTap()
~*Enter:: ResetDoubleTap()
~*Tab:: ResetDoubleTap()
~*Backspace:: ResetDoubleTap()
~*Delete:: ResetDoubleTap()
~*Insert:: ResetDoubleTap()
~*Home:: ResetDoubleTap()
~*End:: ResetDoubleTap()
~*PgUp:: ResetDoubleTap()
~*PgDn:: ResetDoubleTap()
~*Up:: ResetDoubleTap()
~*Down:: ResetDoubleTap()
~*Left:: ResetDoubleTap()
~*Right::ResetDoubleTap()
~*Esc:: ResetDoubleTap()

; Symbols and punctuation (using scan codes for reliability)
~*SC027:: ResetDoubleTap()  ; ; (semicolon)
~*SC028:: ResetDoubleTap()  ; ' (apostrophe/quote)
~*SC033:: ResetDoubleTap()  ; , (comma)
~*SC034:: ResetDoubleTap()  ; . (period)
~*SC035:: ResetDoubleTap()  ; / (forward slash)
~*SC01A:: ResetDoubleTap()  ; [ (left bracket)
~*SC01B:: ResetDoubleTap()  ; ] (right bracket)
~*SC02B:: ResetDoubleTap()  ; \ (backslash)
~*SC00C:: ResetDoubleTap()  ; - (minus/hyphen)
~*SC00D:: ResetDoubleTap()  ; = (equals)
~*SC029:: ResetDoubleTap()  ; ` (backtick/grave)

; Function keys
~*F1:: ResetDoubleTap()
~*F2:: ResetDoubleTap()
~*F3:: ResetDoubleTap()
~*F4:: ResetDoubleTap()
~*F5:: ResetDoubleTap()
~*F6:: ResetDoubleTap()
~*F7:: ResetDoubleTap()
~*F8::ResetDoubleTap()
~*F9:: ResetDoubleTap()
~*F10:: ResetDoubleTap()
~*F11:: ResetDoubleTap()
~*F12:: ResetDoubleTap()

; Modifier keys (Ctrl, Alt, Win)
~*Ctrl:: ResetDoubleTap()
~*LCtrl::  ResetDoubleTap()
~*RCtrl:: ResetDoubleTap()
~*Alt:: ResetDoubleTap()
~*LAlt:: ResetDoubleTap()
~*RAlt:: ResetDoubleTap()
~*LWin:: ResetDoubleTap()
~*RWin:: ResetDoubleTap()
~*AppsKey:: ResetDoubleTap()

; Numpad keys
~*Numpad0:: ResetDoubleTap()
~*Numpad1:: ResetDoubleTap()
~*Numpad2:: ResetDoubleTap()
~*Numpad3:: ResetDoubleTap()
~*Numpad4:: ResetDoubleTap()
~*Numpad5:: ResetDoubleTap()
~*Numpad6:: ResetDoubleTap()
~*Numpad7:: ResetDoubleTap()
~*Numpad8:: ResetDoubleTap()
~*Numpad9:: ResetDoubleTap()
~*NumpadDot:: ResetDoubleTap()
~*NumpadAdd:: ResetDoubleTap()
~*NumpadSub:: ResetDoubleTap()
~*NumpadMult:: ResetDoubleTap()
~*NumpadDiv:: ResetDoubleTap()
~*NumpadEnter:: ResetDoubleTap()

; Handle Left Shift
~LShift::
{
    global shiftHeldTime, otherKeyUsed
    shiftHeldTime := A_TickCount
    otherKeyUsed := false
}

~LShift Up::
{
    global lastShiftUpTime, doubleTapThreshold, shiftHeldTime, otherKeyUsed, waitingForSecondTap
    
    now := A_TickCount
    heldDuration := now - shiftHeldTime
    
    ; Only consider it a potential double-tap if:
    ; 1. Shift was held for less than 200ms (quick tap)
    ; 2. No other keys were used while holding Shift
    if (heldDuration < 200 && !otherKeyUsed) {
        if (waitingForSecondTap && (now - lastShiftUpTime <= doubleTapThreshold)) {
            ; Double tap detected - toggle CapsLock
            SetCapsLockState(!GetKeyState("CapsLock", "T"))
            waitingForSecondTap := false
            lastShiftUpTime := 0
        } else {
            ; First tap - start waiting for second
            waitingForSecondTap := true
            lastShiftUpTime := now
            ; Set timer to reset if no second tap comes
            SetTimer(() => (waitingForSecondTap := false), -doubleTapThreshold)
        }
    } else {
        ; Reset if held too long or other keys were used
        waitingForSecondTap := false
        lastShiftUpTime := 0
    }
}

; Handle Right Shift (same logic)
~RShift::
{
    global shiftHeldTime, otherKeyUsed
    shiftHeldTime := A_TickCount
    otherKeyUsed := false
}

~RShift Up::
{
    global lastShiftUpTime, doubleTapThreshold, shiftHeldTime, otherKeyUsed, waitingForSecondTap
    
    now := A_TickCount
    heldDuration := now - shiftHeldTime
    
    if (heldDuration < 200 && !otherKeyUsed) {
        if (waitingForSecondTap && (now - lastShiftUpTime <= doubleTapThreshold)) {
            SetCapsLockState(!GetKeyState("CapsLock", "T"))
            waitingForSecondTap := false
            lastShiftUpTime := 0
        } else {
            waitingForSecondTap := true
            lastShiftUpTime := now
            SetTimer(() => (waitingForSecondTap := false), -doubleTapThreshold)
        }
    } else {
        waitingForSecondTap := false
        lastShiftUpTime := 0
    }
}

; Function to reset double-tap detection when other keys are used
ResetDoubleTap() {
    global otherKeyUsed, waitingForSecondTap
    if (GetKeyState("LShift", "P") || GetKeyState("RShift", "P")) {
        otherKeyUsed := true
    }
    waitingForSecondTap := false
}

; Optional: Show startup notification
ToolTip("Double-tap Shift to toggle CapsLock", 0, 0)
SetTimer(() => ToolTip(), -2000)