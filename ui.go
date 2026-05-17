//go:build windows

package main

import (
	"unsafe"
)

func wndProc(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		createMenu(hwnd)
		createControls(hwnd)
		layoutControls(hwnd)
		startIndexing(true)
		procSetTimer.Call(hwnd, timerRefresh, 700, 0)
		return 0
	case WM_SIZE:
		layoutControls(hwnd)
		return 0
	case WM_COMMAND:
		id := loword(wParam)
		code := hiword(wParam)
		switch id {
		case idSearch:
			if code == EN_CHANGE {
				procSetTimer.Call(hwnd, timerSearch, 80, 0)
			}
		case idBtnOpen:
			openSelected()
		case idBtnPreview:
			previewSelected()
		case idBtnCopyPath:
			copySelectedPath()
		case idBtnDesktop:
			copySelectedToDesktop()
		case idBtnExplorer:
			showSelectedInExplorer()
		case idBtnRefresh:
			startIndexing(true)
			runSearch()
		case idBtnInfo:
			showSelectedInfo()
		case idMenuExit:
			procSendMessage.Call(hwnd, WM_CLOSE, 0, 0)
		case idMenuAbout:
			s := ui()
			info(s.AboutTitle, s.AboutMessage)
		case idLangEnglish:
			setLanguage(langEnglish)
		case idLangTurkish:
			setLanguage(langTurkish)
		}
		return 0
	case WM_TIMER:
		switch wParam {
		case timerSearch:
			procKillTimer.Call(hwnd, timerSearch)
			runSearch()
		case timerRefresh:
			refreshFromIndex()
		}
		return 0
	case WM_NOTIFY:
		hdr := (*NMHDR)(unsafe.Pointer(lParam))
		if hdr.HwndFrom == app.hList {
			switch hdr.Code {
			case NM_DBLCLK:
				openSelected()
				return 0
			case NM_RCLICK:
				var pt POINT
				procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
				selectListItemAtScreen(pt.X, pt.Y)
				showContextMenu(pt.X, pt.Y)
				return 0
			case LVN_BEGINDRAG:
				nm := (*NMLISTVIEW)(unsafe.Pointer(lParam))
				if nm.IItem >= 0 && int(nm.IItem) < len(app.results) {
					doDragFiles([]string{app.results[nm.IItem].Path})
				}
				return 0
			}
		}
	case WM_CONTEXTMENU:
		if wParam == app.hList {
			x := xFromLParam(lParam)
			y := yFromLParam(lParam)
			if x == -1 && y == -1 {
				var pt POINT
				procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
				x, y = pt.X, pt.Y
				showContextMenu(x, y)
			}
			return 0
		}
	case WM_CLOSE:
		atomicAddInt64(&scanID, 1)
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func createControls(hwnd uintptr) {
	hFont, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)
	app.hFont = hFont

	buttonIDs := []uintptr{
		idBtnOpen,
		idBtnPreview,
		idBtnCopyPath,
		idBtnDesktop,
		idBtnExplorer,
		idBtnRefresh,
		idBtnInfo,
	}
	app.buttonIDs = buttonIDs
	app.buttons = make([]uintptr, 0, len(buttonIDs))
	for _, id := range buttonIDs {
		app.buttons = append(app.buttons, createButton(hwnd, id, buttonText(id)))
	}

	s := ui()
	app.hLabel = createStatic(hwnd, 0, s.SearchLabel)
	app.hSearch = createWindow("EDIT", "", WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|ES_AUTOHSCROLL, 0, 0, 0, 0, hwnd, idSearch)
	app.hList = createWindow("SysListView32", "", WS_CHILD|WS_VISIBLE|WS_BORDER|WS_TABSTOP|WS_CLIPSIBLINGS|LVS_REPORT|LVS_SINGLESEL|LVS_SHOWSELALWAYS, 0, 0, 0, 0, hwnd, idList)
	app.hStatus = createWindow("STATIC", s.StatusPreparing, WS_CHILD|WS_VISIBLE, 0, 0, 0, 0, hwnd, idStatus)

	setFont(app.hSearch)
	setFont(app.hList)
	setFont(app.hStatus)

	procSendMessage.Call(app.hList, LVM_SETEXTENDEDLISTVIEWSTYLE, 0, LVS_EX_FULLROWSELECT|LVS_EX_GRIDLINES|LVS_EX_DOUBLEBUFFER)
	addListColumn(0, s.ColumnName, 165, LVCFMT_LEFT)
	addListColumn(1, s.ColumnPath, 360, LVCFMT_LEFT)
	addListColumn(2, s.ColumnSize, 80, LVCFMT_RIGHT)
	addListColumn(3, s.ColumnModified, 130, LVCFMT_LEFT)
}

func createMenu(hwnd uintptr) {
	s := ui()
	mainMenu, _, _ := procCreateMenu.Call()
	fileMenu, _, _ := procCreatePopupMenu.Call()
	actionMenu, _, _ := procCreatePopupMenu.Call()
	settingsMenu, _, _ := procCreatePopupMenu.Call()
	languageMenu, _, _ := procCreatePopupMenu.Call()
	helpMenu, _, _ := procCreatePopupMenu.Call()

	appendMenu(fileMenu, MF_STRING, idBtnOpen, s.Open)
	appendMenu(fileMenu, MF_STRING, idBtnPreview, s.Preview)
	appendMenu(fileMenu, MF_STRING, idBtnCopyPath, s.CopyPath)
	appendMenu(fileMenu, MF_STRING, idBtnExplorer, s.MenuShowExplorer)
	appendMenu(fileMenu, MF_SEPARATOR, 0, "")
	appendMenu(fileMenu, MF_STRING, idMenuExit, s.MenuExit)

	appendMenu(actionMenu, MF_STRING, idBtnDesktop, s.MenuCopyDesktop)
	appendMenu(actionMenu, MF_STRING, idBtnRefresh, s.Refresh)
	appendMenu(actionMenu, MF_STRING, idBtnInfo, s.Info)

	enFlag := uintptr(MF_STRING)
	trFlag := uintptr(MF_STRING)
	if app.language == langTurkish {
		trFlag |= MF_CHECKED
	} else {
		enFlag |= MF_CHECKED
	}
	appendMenu(languageMenu, enFlag, idLangEnglish, s.MenuEnglish)
	appendMenu(languageMenu, trFlag, idLangTurkish, s.MenuTurkish)
	appendMenu(settingsMenu, MF_POPUP, languageMenu, s.MenuLanguage)

	appendMenu(helpMenu, MF_STRING, idMenuAbout, s.MenuAbout)

	appendMenu(mainMenu, MF_POPUP, fileMenu, s.MenuFile)
	appendMenu(mainMenu, MF_POPUP, actionMenu, s.MenuActions)
	appendMenu(mainMenu, MF_POPUP, settingsMenu, s.MenuSettings)
	appendMenu(mainMenu, MF_POPUP, helpMenu, s.MenuHelp)
	procSetMenu.Call(hwnd, mainMenu)
	procDrawMenuBar.Call(hwnd)
}

func appendMenu(menu uintptr, flags uintptr, id uintptr, text string) {
	var textPtr uintptr
	if text != "" {
		textPtr = uintptr(unsafe.Pointer(utf16Ptr(text)))
	}
	procAppendMenu.Call(menu, flags, id, textPtr)
}

func createButton(parent uintptr, id uintptr, text string) uintptr {
	h := createWindow("BUTTON", text, WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 0, 0, 0, 0, parent, id)
	setFont(h)
	return h
}

func createStatic(parent uintptr, id uintptr, text string) uintptr {
	h := createWindow("STATIC", text, WS_CHILD|WS_VISIBLE, 0, 0, 0, 0, parent, id)
	setFont(h)
	return h
}

func createWindow(className, title string, style uintptr, x, y, w, h int32, parent uintptr, id uintptr) uintptr {
	hwnd, _, _ := procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(className))),
		uintptr(unsafe.Pointer(utf16Ptr(title))),
		style,
		uintptr(x),
		uintptr(y),
		uintptr(w),
		uintptr(h),
		parent,
		id,
		0,
		0,
	)
	return hwnd
}

func setFont(hwnd uintptr) {
	if hwnd != 0 && app.hFont != 0 {
		procSendMessage.Call(hwnd, WM_SETFONT, app.hFont, 1)
	}
}

func layoutControls(hwnd uintptr) {
	if hwnd == 0 || app.hList == 0 {
		return
	}
	var rc RECT
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top
	if w < 260 {
		w = 260
	}
	if h < 180 {
		h = 180
	}
	margin := int32(6)
	buttonY := int32(6)
	buttonH := int32(26)
	buttonW := int32(96)
	gap := int32(4)
	buttonsPerRow := int((w - margin*2 + gap) / (buttonW + gap))
	if buttonsPerRow < 1 {
		buttonsPerRow = 1
	}

	for i, btn := range app.buttons {
		row := int32(i / buttonsPerRow)
		col := int32(i % buttonsPerRow)
		x := margin + col*(buttonW+gap)
		y := buttonY + row*(buttonH+gap)
		procMoveWindow.Call(btn, uintptr(x), uintptr(y), uintptr(buttonW), uintptr(buttonH), 1)
	}

	rows := int32(0)
	if len(app.buttons) > 0 {
		rows = int32((len(app.buttons) + buttonsPerRow - 1) / buttonsPerRow)
	}
	searchY := buttonY + rows*buttonH + (rows-1)*gap + 8
	if rows == 0 {
		searchY = buttonY
	}
	labelW := int32(56)
	procMoveWindow.Call(app.hLabel, uintptr(margin), uintptr(searchY+4), uintptr(labelW), uintptr(20), 1)
	searchW := w - margin*2 - labelW
	if searchW < 80 {
		searchW = 80
	}
	procMoveWindow.Call(app.hSearch, uintptr(margin+labelW), uintptr(searchY), uintptr(searchW), uintptr(24), 1)

	statusH := int32(22)
	listY := searchY + 30
	listW := w - margin*2
	listH := h - listY - statusH - margin
	if listW < 120 {
		listW = 120
	}
	if listH < 40 {
		listH = 40
	}
	procMoveWindow.Call(app.hList, uintptr(margin), uintptr(listY), uintptr(listW), uintptr(listH), 1)
	procMoveWindow.Call(app.hStatus, uintptr(margin), uintptr(h-statusH), uintptr(w-margin*2), uintptr(statusH), 1)
}

func addListColumn(index int32, title string, width int32, fmt int32) {
	col := LVCOLUMN{
		Mask:     LVCF_TEXT | LVCF_WIDTH | LVCF_FMT | LVCF_SUBITEM,
		Fmt:      fmt,
		Cx:       width,
		PszText:  utf16Ptr(title),
		ISubItem: index,
	}
	procSendMessage.Call(app.hList, LVM_INSERTCOLUMNW, uintptr(index), uintptr(unsafe.Pointer(&col)))
}

func setListColumn(index int32, title string, width int32, fmt int32) {
	col := LVCOLUMN{
		Mask:     LVCF_TEXT | LVCF_WIDTH | LVCF_FMT | LVCF_SUBITEM,
		Fmt:      fmt,
		Cx:       width,
		PszText:  utf16Ptr(title),
		ISubItem: index,
	}
	procSendMessage.Call(app.hList, LVM_SETCOLUMNW, uintptr(index), uintptr(unsafe.Pointer(&col)))
}

func setStatus(text string) {
	if app.hStatus != 0 {
		procSetWindowText.Call(app.hStatus, uintptr(unsafe.Pointer(utf16Ptr(text))))
	}
}

func info(title, msg string) {
	procMessageBox.Call(app.hwnd, uintptr(unsafe.Pointer(utf16Ptr(msg))), uintptr(unsafe.Pointer(utf16Ptr(title))), MB_OK|MB_ICONINFORMATION)
}

func warn(msg string) {
	procMessageBox.Call(app.hwnd, uintptr(unsafe.Pointer(utf16Ptr(msg))), uintptr(unsafe.Pointer(utf16Ptr(appTitle))), MB_OK|MB_ICONWARNING)
}
