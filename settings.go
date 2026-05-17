//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"unsafe"
)

func ui() uiStrings {
	if app.language == langTurkish {
		return turkishUI
	}
	return englishUI
}

func buttonText(id uintptr) string {
	s := ui()
	switch id {
	case idBtnOpen:
		return s.Open
	case idBtnPreview:
		return s.Preview
	case idBtnCopyPath:
		return s.CopyPath
	case idBtnDesktop:
		return s.Desktop
	case idBtnExplorer:
		return s.Explorer
	case idBtnRefresh:
		return s.Refresh
	case idBtnInfo:
		return s.Info
	}
	return ""
}

func setLanguage(language string) {
	if language != langTurkish {
		language = langEnglish
	}
	if app.language == language {
		return
	}
	app.language = language
	saveLanguage(language)
	updateLanguageUI()
}

func updateLanguageUI() {
	s := ui()
	if app.hwnd != 0 {
		procSetWindowText.Call(app.hwnd, uintptr(unsafe.Pointer(utf16Ptr(appTitle))))
		createMenu(app.hwnd)
	}
	for i, btn := range app.buttons {
		if i < len(app.buttonIDs) {
			procSetWindowText.Call(btn, uintptr(unsafe.Pointer(utf16Ptr(buttonText(app.buttonIDs[i])))))
		}
	}
	if app.hLabel != 0 {
		procSetWindowText.Call(app.hLabel, uintptr(unsafe.Pointer(utf16Ptr(s.SearchLabel))))
	}
	if app.hList != 0 {
		setListColumn(0, s.ColumnName, 165, LVCFMT_LEFT)
		setListColumn(1, s.ColumnPath, 360, LVCFMT_LEFT)
		setListColumn(2, s.ColumnSize, 80, LVCFMT_RIGHT)
		setListColumn(3, s.ColumnModified, 130, LVCFMT_LEFT)
	}
	layoutControls(app.hwnd)
	if strings.TrimSpace(app.lastQuery) != "" {
		runSearch()
	} else if app.hStatus != 0 {
		setStatus(fmt.Sprintf(s.StatusIndex, formatInt(atomic.LoadInt64(&indexedCount)), scanSuffix()))
	}
}

func loadLanguage() string {
	path := configPath()
	if path == "" {
		return langEnglish
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return langEnglish
	}
	if strings.TrimSpace(string(data)) == langTurkish {
		return langTurkish
	}
	return langEnglish
}

func saveLanguage(language string) {
	path := configPath()
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return
	}
	_ = os.WriteFile(path, []byte(language), 0600)
}

func configPath() string {
	base := os.Getenv("APPDATA")
	if base == "" {
		if dir, err := os.UserConfigDir(); err == nil {
			base = dir
		}
	}
	if base == "" {
		return ""
	}
	return filepath.Join(base, "QBFind", "settings.txt")
}
