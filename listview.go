//go:build windows

package main

import (
	"strings"
	"unsafe"
)

func clearList() {
	procSendMessage.Call(app.hList, WM_SETREDRAW, 0, 0)
	procSendMessage.Call(app.hList, LVM_DELETEALLITEMS, 0, 0)
	procSendMessage.Call(app.hList, WM_SETREDRAW, 1, 0)
	procInvalidateRect.Call(app.hList, 0, 1)
}

func fillList(results []fileEntry, keepPath string) {
	procSendMessage.Call(app.hList, WM_SETREDRAW, 0, 0)
	procSendMessage.Call(app.hList, LVM_DELETEALLITEMS, 0, 0)
	selectedIndex := -1
	for i, e := range results {
		insertListRow(i, e)
		if keepPath != "" && strings.EqualFold(e.Path, keepPath) {
			selectedIndex = i
		}
	}
	if selectedIndex >= 0 {
		selectListRow(selectedIndex)
	}
	procSendMessage.Call(app.hList, WM_SETREDRAW, 1, 0)
	procInvalidateRect.Call(app.hList, 0, 1)
}

func insertListRow(i int, e fileEntry) {
	item := LVITEM{
		Mask:    LVIF_TEXT | LVIF_PARAM,
		IItem:   int32(i),
		PszText: utf16Ptr(e.Name),
		LParam:  uintptr(i),
	}
	procSendMessage.Call(app.hList, LVM_INSERTITEMW, 0, uintptr(unsafe.Pointer(&item)))
	setListText(i, 1, e.Path)
	setListText(i, 2, formatEntrySize(e))
	setListText(i, 3, e.ModTime.Format("2006-01-02 15:04"))
}

func setListText(row int, col int, text string) {
	item := LVITEM{
		ISubItem: int32(col),
		PszText:  utf16Ptr(text),
	}
	procSendMessage.Call(app.hList, LVM_SETITEMTEXTW, uintptr(row), uintptr(unsafe.Pointer(&item)))
}

func selectedEntry() (fileEntry, bool) {
	ret, _, _ := procSendMessage.Call(app.hList, LVM_GETNEXTITEM, ^uintptr(0), LVNI_SELECTED)
	idx := int32(ret)
	if idx < 0 || int(idx) >= len(app.results) {
		return fileEntry{}, false
	}
	return app.results[idx], true
}

func selectedPath() string {
	e, ok := selectedEntry()
	if !ok {
		return ""
	}
	return e.Path
}

func selectListRow(row int) {
	if row < 0 {
		return
	}
	clear := LVITEM{
		Mask:      LVIF_STATE,
		StateMask: LVIS_SELECTED | LVIS_FOCUSED,
	}
	procSendMessage.Call(app.hList, LVM_SETITEMSTATE, ^uintptr(0), uintptr(unsafe.Pointer(&clear)))
	item := LVITEM{
		Mask:      LVIF_STATE,
		State:     LVIS_SELECTED | LVIS_FOCUSED,
		StateMask: LVIS_SELECTED | LVIS_FOCUSED,
	}
	procSendMessage.Call(app.hList, LVM_SETITEMSTATE, uintptr(row), uintptr(unsafe.Pointer(&item)))
	procSendMessage.Call(app.hList, LVM_ENSUREVISIBLE, uintptr(row), 0)
}

func selectListItemAtScreen(x, y int32) bool {
	if app.hList == 0 {
		return false
	}
	pt := POINT{X: x, Y: y}
	procScreenToClient.Call(app.hList, uintptr(unsafe.Pointer(&pt)))
	hit := LVHITTESTINFO{Pt: pt}
	ret, _, _ := procSendMessage.Call(app.hList, LVM_HITTEST, 0, uintptr(unsafe.Pointer(&hit)))
	if int32(ret) < 0 || hit.IItem < 0 || int(hit.IItem) >= len(app.results) {
		return false
	}
	selectListRow(int(hit.IItem))
	return true
}

func showContextMenu(x, y int32) {
	if _, ok := selectedEntry(); !ok {
		return
	}
	s := ui()
	menu, _, _ := procCreatePopupMenu.Call()
	appendMenu(menu, MF_STRING, idBtnOpen, s.Open)
	appendMenu(menu, MF_STRING, idBtnPreview, s.Preview)
	appendMenu(menu, MF_STRING, idBtnCopyPath, s.CopyPath)
	appendMenu(menu, MF_SEPARATOR, 0, "")
	appendMenu(menu, MF_STRING, idBtnExplorer, s.MenuShowExplorer)
	appendMenu(menu, MF_STRING, idBtnDesktop, s.MenuCopyDesktop)
	appendMenu(menu, MF_SEPARATOR, 0, "")
	appendMenu(menu, MF_STRING, idBtnInfo, s.Info)
	procTrackPopupMenu.Call(menu, TPM_RIGHTBUTTON, uintptr(x), uintptr(y), 0, app.hwnd, 0)
	procDestroyMenu.Call(menu)
}
