package main

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	langEnglish = "en"
	langTurkish = "tr"
)

var (
	app *App

	indexedCount int64
	scanning     int32
	scanID       int64

	priorityRootKeys []string
	userRootKeys     []string
	noisyRootKeys    []string
)

type App struct {
	ctx      context.Context
	language string

	mu               sync.RWMutex
	entries          []fileEntry
	lastIndexedShown int64
}

type fileEntry struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	LowerPath string    `json:"lowerPath"`
	LowerName string    `json:"lowerName"`
	LowerExt  string    `json:"lowerExt"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"modTime"`
	IsDir     bool      `json:"isDir"`
	Priority  int       `json:"priority"`
}

type PreviewResult struct {
	Success bool   `json:"success"`
	Text    string `json:"text"`
	IsText  bool   `json:"isText"`
	Size    int64  `json:"size"`
	Name    string `json:"name"`
}

func NewApp() *App {
	a := &App{
		language: langEnglish,
	}
	app = a
	return a
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.language = loadLanguage()
	initPathPriority()

	// Periodic status emitter to notify frontend about indexing count & scanning state
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			if a.ctx == nil {
				break
			}
			count := atomic.LoadInt64(&indexedCount)
			isScanning := atomic.LoadInt32(&scanning) != 0

			wailsruntime.EventsEmit(a.ctx, "status_update", map[string]interface{}{
				"count":    count,
				"scanning": isScanning,
				"suffix":   scanSuffix(),
			})
		}
	}()

	// Start scanning logical drives on application startup
	startIndexing(true)
}

func (a *App) Search(query string) []fileEntry {
	return searchIndex(query, maxResults)
}

func (a *App) OpenFile(path string) {
	shellExecute(0, "open", path, "", "")
}

func (a *App) ShowInExplorer(path string) {
	args := `/select,"` + path + `"`
	shellExecute(0, "open", "explorer.exe", args, "")
}

func (a *App) CopyPath(path string) string {
	if err := setClipboardText(path); err != nil {
		return err.Error()
	}
	return ui().StatusCopiedPath
}

func (a *App) CopyToDesktop(path string) string {
	desktop := desktopPath()
	if desktop == "" {
		return ui().WarnDesktopMissing
	}
	if err := shellCopyTo(path, desktop); err != nil {
		return err.Error()
	}
	return ui().StatusCopiedDesktop
}

func (a *App) GetPreview(path string, size int64, name string) PreviewResult {
	text, ok := getTextPreview(path, size, name)
	if !ok {
		return PreviewResult{
			Success: true,
			Text:    ui().PreviewBinary,
			IsText:  false,
			Size:    size,
			Name:    name,
		}
	}
	return PreviewResult{
		Success: true,
		Text:    text,
		IsText:  true,
		Size:    size,
		Name:    name,
	}
}

func (a *App) StartScan(reset bool) {
	startIndexing(reset)
}

func (a *App) GetLanguage() string {
	return a.language
}

func (a *App) SetLanguage(lang string) {
	setLanguage(lang)
}

func (a *App) GetAboutMessage() map[string]string {
	s := ui()
	return map[string]string{
		"title":   s.AboutTitle,
		"message": s.AboutMessage,
	}
}
