//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type uiStrings struct {
	Open                 string
	Preview              string
	CopyPath             string
	Desktop              string
	Explorer             string
	Refresh              string
	Info                 string
	SearchLabel          string
	StatusPreparing      string
	ColumnName           string
	ColumnPath           string
	ColumnSize           string
	ColumnModified       string
	MenuFile             string
	MenuActions          string
	MenuSettings         string
	MenuHelp             string
	MenuExit             string
	MenuCopyDesktop      string
	MenuShowExplorer     string
	MenuLanguage         string
	MenuEnglish          string
	MenuTurkish          string
	MenuAbout            string
	AboutTitle           string
	AboutMessage         string
	WarnSelect           string
	WarnDesktopMissing   string
	StatusCopiedDesktop  string
	StatusCopiedPath     string
	FileKind             string
	FolderKind           string
	InfoTitle            string
	StatusMinChars       string
	StatusResults        string
	StatusIndex          string
	ScanRunning          string
	ScanReady            string
	PreviewTitle         string
	PreviewBinary        string
	PreviewOpenFailed    string
	CopyFailed           string
	CopyCanceled         string
	ClipboardOpenFailed  string
	ClipboardWriteFailed string
}

var englishUI = uiStrings{
	Open:                 "Open",
	Preview:              "Preview",
	CopyPath:             "Copy Path",
	Desktop:              "Desktop",
	Explorer:             "Explorer",
	Refresh:              "Refresh",
	Info:                 "Info",
	SearchLabel:          "Search:",
	StatusPreparing:      "Indexing...",
	ColumnName:           "Name",
	ColumnPath:           "Path",
	ColumnSize:           "Size",
	ColumnModified:       "Modified",
	MenuFile:             "File",
	MenuActions:          "Actions",
	MenuSettings:         "Settings",
	MenuHelp:             "Help",
	MenuExit:             "Exit",
	MenuCopyDesktop:      "Copy to Desktop",
	MenuShowExplorer:     "Show in Explorer",
	MenuLanguage:         "Language",
	MenuEnglish:          "English",
	MenuTurkish:          "Türkçe",
	MenuAbout:            "About",
	AboutTitle:           "About QBFind",
	AboutMessage:         "About QBFind\n\nSingle-file Go/Win32 file finder.\nCommon user folders are indexed first and extension searches such as .pdf or ext:pdf are supported.",
	WarnSelect:           "Select a result first.",
	WarnDesktopMissing:   "Desktop folder could not be found.",
	StatusCopiedDesktop:  "Selected item was copied to the Desktop.",
	StatusCopiedPath:     "Path copied to clipboard.",
	FileKind:             "File",
	FolderKind:           "Folder",
	InfoTitle:            "Info",
	StatusMinChars:       "Type at least 2 characters. Index: %s items%s",
	StatusResults:        "Showing %s results. Index: %s items%s",
	StatusIndex:          "Index: %s items%s",
	ScanRunning:          " (scanning)",
	ScanReady:            " (ready)",
	PreviewTitle:         "Preview",
	PreviewBinary:        "This file type does not have an in-app text preview. QBFind asked Windows to preview it.",
	PreviewOpenFailed:    "Windows could not preview this file type.",
	CopyFailed:           "Copy failed. Code: %d",
	CopyCanceled:         "Copy was canceled.",
	ClipboardOpenFailed:  "Clipboard could not be opened.",
	ClipboardWriteFailed: "Clipboard could not be written.",
}

var turkishUI = uiStrings{
	Open:                 "Aç",
	Preview:              "Önizle",
	CopyPath:             "Yolu Kopyala",
	Desktop:              "Masaüstü",
	Explorer:             "Explorer",
	Refresh:              "Yenile",
	Info:                 "Bilgi",
	SearchLabel:          "Ara:",
	StatusPreparing:      "İndeks hazırlanıyor...",
	ColumnName:           "Ad",
	ColumnPath:           "Yol",
	ColumnSize:           "Boyut",
	ColumnModified:       "Değişim",
	MenuFile:             "Dosya",
	MenuActions:          "İşlem",
	MenuSettings:         "Ayarlar",
	MenuHelp:             "Yardım",
	MenuExit:             "Çıkış",
	MenuCopyDesktop:      "Masaüstüne Kopyala",
	MenuShowExplorer:     "Explorer'da Göster",
	MenuLanguage:         "Dil",
	MenuEnglish:          "English",
	MenuTurkish:          "Türkçe",
	MenuAbout:            "Hakkında",
	AboutTitle:           "QBFind Hakkında",
	AboutMessage:         "QBFind\n\nTek dosyalı Go/Win32 dosya bulucu.\nSık kullanılan kullanıcı klasörleri önce indekslenir; .pdf veya ext:pdf gibi uzantı aramaları desteklenir.",
	WarnSelect:           "Önce bir sonuç seçin.",
	WarnDesktopMissing:   "Masaüstü klasörü bulunamadı.",
	StatusCopiedDesktop:  "Seçili öğe masaüstüne kopyalandı.",
	StatusCopiedPath:     "Yol panoya kopyalandı.",
	FileKind:             "Dosya",
	FolderKind:           "Klasör",
	InfoTitle:            "Bilgi",
	StatusMinChars:       "En az 2 karakter yazın. İndeks: %s öğe%s",
	StatusResults:        "%s sonuç gösteriliyor. İndeks: %s öğe%s",
	StatusIndex:          "İndeks: %s öğe%s",
	ScanRunning:          " (taranıyor)",
	ScanReady:            " (hazır)",
	PreviewTitle:         "Önizleme",
	PreviewBinary:        "Bu dosya türü için uygulama içi metin önizlemesi yok. QBFind Windows'tan önizleme istedi.",
	PreviewOpenFailed:    "Windows bu dosya türünü önizleyemedi.",
	CopyFailed:           "Kopyalama başarısız oldu. Kod: %d",
	CopyCanceled:         "Kopyalama iptal edildi.",
	ClipboardOpenFailed:  "Pano açılamadı.",
	ClipboardWriteFailed: "Panoya yazılamadı.",
}

func ui() uiStrings {
	if app != nil && app.language == langTurkish {
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
	if app != nil && app.ctx != nil {
		wailsruntime.EventsEmit(app.ctx, "language_changed", app.language)
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
