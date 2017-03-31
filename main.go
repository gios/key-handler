package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	WH_KEYBOARD_LL   = 13
	WM_KEYDOWN       = 256
	MAPVK_VK_TO_CHAR = 2
	MAPVK_VK_TO_VSC  = 0
)

var (
	user32                    = syscall.MustLoadDLL("user32")
	procCallNextHookEx        = user32.MustFindProc("CallNextHookEx")
	procUnhookWindowsHookEx   = user32.MustFindProc("UnhookWindowsHookEx")
	procSetWindowsHookEx      = user32.MustFindProc("SetWindowsHookExW")
	procGetMessage            = user32.MustFindProc("GetMessageW")
	procMapVirtualKey         = user32.MustFindProc("MapVirtualKeyW")
	procToUnicode             = user32.MustFindProc("ToUnicode")
	procGetKeyboardState      = user32.MustFindProc("GetKeyboardState")
	procGetKeyboardLayoutName = user32.MustFindProc("GetKeyboardLayoutNameW")
	procGetKeyboardLayout     = user32.MustFindProc("GetKeyboardLayout")
)

type HOOKPROC func(int, uintptr, uintptr) uintptr

type POINT struct {
	X, Y int32
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type KBDLLHOOKSTRUCT struct {
	VkCode      uintptr
	ScanCode    uintptr
	Flags       uintptr
	Time        uintptr
	DwExtraInfo uintptr
}

func SetWindowsHookEx(idHook int, lpfn HOOKPROC, hMod uintptr, dwThreadId uintptr) uintptr {
	ret, _, _ := procSetWindowsHookEx.Call(
		uintptr(idHook),
		uintptr(syscall.NewCallback(lpfn)),
		uintptr(hMod),
		uintptr(dwThreadId),
	)
	return uintptr(ret)
}

func CallNextHookEx(hhk uintptr, nCode int, wParam uintptr, lParam uintptr) uintptr {
	ret, _, _ := procCallNextHookEx.Call(
		uintptr(hhk),
		uintptr(nCode),
		uintptr(wParam),
		uintptr(lParam),
	)
	return uintptr(ret)
}

func UnhookWindowsHookEx(hhk uintptr) bool {
	ret, _, _ := procUnhookWindowsHookEx.Call(
		uintptr(hhk),
	)
	return ret != 0
}

func LowLevelKeyboardProcess(nCode int, wparam uintptr, lparam uintptr) uintptr {
	var temporaryKeyPtr uintptr
	var keyboardState [256]byte
	var unicodeKey [256]byte
	var keyboardLayoutName [256]byte
	if nCode == 0 && wparam == WM_KEYDOWN {
		key := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lparam))
		sc := MapVirtualKey(key.VkCode, MAPVK_VK_TO_VSC)
		GetKeyboardLayoutName(&keyboardLayoutName)
		fmt.Println(string(keyboardLayoutName[:]))
		fmt.Println(key.VkCode, sc)
		GetKeyboardState(&keyboardState)
		ToUnicode(key.VkCode, uintptr(sc), &keyboardState, &unicodeKey, 256, 0)
		fmt.Println(string(unicodeKey[:]))
	}
	return CallNextHookEx(temporaryKeyPtr, nCode, wparam, lparam)
}

func GetMessage(msg *MSG, hwnd uintptr, msgFilterMin uint32, msgFilterMax uint32) int {
	ret, _, _ := procGetMessage.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax))
	return int(ret)
}

func MapVirtualKey(vkCode uintptr, uMapType uintptr) int {
	ret, _, _ := procMapVirtualKey.Call(
		uintptr(vkCode),
		uintptr(uMapType))
	return int(ret)
}

func ToUnicode(wVirtKey uintptr, wScanCode uintptr, lpKeyState *[256]byte, pwszBuff *[256]byte, cchBuff int, wFlags uint) int {
	ret, _, _ := procToUnicode.Call(
		uintptr(wVirtKey),
		uintptr(wScanCode),
		uintptr(unsafe.Pointer(lpKeyState)),
		uintptr(unsafe.Pointer(pwszBuff)),
		uintptr(cchBuff),
		uintptr(wFlags))
	return int(ret)
}

func GetKeyboardState(lpKeyState *[256]byte) int {
	ret, _, _ := procGetKeyboardState.Call(uintptr(unsafe.Pointer(lpKeyState)))
	return int(ret)
}

func GetKeyboardLayoutName(pwszKLID *[256]byte) int {
	ret, _, _ := procGetKeyboardLayoutName.Call(uintptr(unsafe.Pointer(pwszKLID)))
	return int(ret)
}

func GetKeyboardLayout(idThread uintptr) int {
	ret, _, _ := procGetKeyboardLayout.Call(uintptr(idThread))
	return int(ret)
}

func Start() {
	defer user32.Release()
	var msg MSG
	keyboardHook := SetWindowsHookEx(WH_KEYBOARD_LL, LowLevelKeyboardProcess, 0, 0)
	for GetMessage(&msg, 0, 0, 0) != 0 {
		fmt.Println(msg)
	}
	UnhookWindowsHookEx(keyboardHook)
}

func main() {
	Start()
}
