package main

import (
	"fmt"
	"regexp"
	"strings"
	"syscall"
	"unsafe"

	"github.com/TheTitanrain/w32"
	"github.com/atotto/clipboard"
)

func getClipText() (string, error) {
	if w32.OpenClipboard(0) {
		defer w32.CloseClipboard()
		hclip := w32.HGLOBAL(w32.GetClipboardData(w32.CF_UNICODETEXT))
		if hclip == 0 {
			return "", fmt.Errorf("GetClipboardData")
		}

		lpstr := w32.GlobalLock(hclip)
		defer w32.GlobalUnlock(hclip)
		return w32.UTF16PtrToString((*uint16)(lpstr)), nil
	}
	return "", fmt.Errorf("OpenClipboard")
}

func wndProc(hwnd w32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	if msg == w32.WM_CLIPBOARDUPDATE {
		text, err := getClipText()

		if err != nil {
			fmt.Println("error:", err)
			return 0
		}

		w32.RemoveClipboardFormatListener(hwnd)
		w32.EmptyClipboard()

		text = deleteLineBreaks(text)
		clipboard.WriteAll(text)

		w32.AddClipboardFormatListener(hwnd)

		return 0
	}
	return w32.DefWindowProc(hwnd, msg, wParam, lParam)
}

func deleteLineBreaks(text string) string {
	textArray := regexp.MustCompile("\r\n|\r|\n").Split(text, -1)
	var result []string
	for _, i := range textArray {
		i = strings.Replace(i, "..", "〇", -1)
		if strings.HasSuffix(i, "-") {
			i = i[:len(i)-1]
			result = append(result, i)
		} else if strings.HasSuffix(i, ".") {
			i = i + "\n"
			result = append(result, i)
		} else {
			result = append(result, i)
		}
	}
	return strings.Join(result, " ")
}

func main() {
	className := syscall.StringToUTF16Ptr("for clipboard")
	wndClassEx := w32.WNDCLASSEX{
		ClassName: className,
		WndProc:   syscall.NewCallback(wndProc),
	}
	wndClassEx.Size = uint32(unsafe.Sizeof(wndClassEx))
	w32.RegisterClassEx(&wndClassEx)

	hwnd := w32.CreateWindowEx(0, className, className, 0, 0, 0, 0, 0, w32.HWND_MESSAGE, 0, 0, nil)
	w32.AddClipboardFormatListener(hwnd)
	defer w32.RemoveClipboardFormatListener(hwnd)

	msg := w32.MSG{}
	for w32.GetMessage(&msg, 0, 0, 0) > 0 {
		w32.DispatchMessage(&msg)
	}
}
