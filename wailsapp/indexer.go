//go:build windows

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

func startIndexing(reset bool) {
	id := atomic.AddInt64(&scanID, 1)
	atomic.StoreInt32(&scanning, 1)
	if reset {
		app.mu.Lock()
		app.entries = nil
		app.mu.Unlock()
		atomic.StoreInt64(&indexedCount, 0)
		app.lastIndexedShown = 0
	}
	go scanComputer(id)
}

func scanComputer(id int64) {
	roots := logicalDrives()
	if len(roots) == 0 {
		atomic.StoreInt32(&scanning, 0)
		return
	}
	quickRoots := commonScanRoots()
	quickRootSet := make(map[string]struct{}, len(quickRoots))
	for _, root := range quickRoots {
		quickRootSet[pathKey(root)] = struct{}{}
	}

	tasks := make(chan string, 256)
	var taskWG sync.WaitGroup
	var workerWG sync.WaitGroup
	workers := runtime.NumCPU() * 2
	if workers < 2 {
		workers = 2
	}
	if workers > 16 {
		workers = 16
	}

	for i := 0; i < workers; i++ {
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			batch := make([]fileEntry, 0, 512)
			for root := range tasks {
				walkTree(id, root, &batch, quickRootSet)
				flushBatch(id, &batch)
				taskWG.Done()
			}
		}()
	}

	rootBatch := make([]fileEntry, 0, 256)
	for _, root := range quickRoots {
		if atomic.LoadInt64(&scanID) != id {
			break
		}
		taskWG.Add(1)
		tasks <- root
	}
	for _, root := range roots {
		if atomic.LoadInt64(&scanID) != id {
			break
		}
		enqueueRoot(id, root, tasks, &taskWG, &rootBatch)
		flushBatch(id, &rootBatch)
	}

	taskWG.Wait()
	close(tasks)
	workerWG.Wait()

	if atomic.LoadInt64(&scanID) == id {
		atomic.StoreInt32(&scanning, 0)
	}
}

func enqueueRoot(id int64, root string, tasks chan<- string, taskWG *sync.WaitGroup, batch *[]fileEntry) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, de := range entries {
		if atomic.LoadInt64(&scanID) != id {
			return
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		full := filepath.Join(root, de.Name())
		isDir := info.IsDir()
		*batch = append(*batch, makeEntry(full, de.Name(), info, isDir))
		if len(*batch) >= 512 {
			flushBatch(id, batch)
		}
		if isDir && !isReparse(info) {
			taskWG.Add(1)
			tasks <- full
		}
	}
}

func walkTree(id int64, root string, batch *[]fileEntry, quickRootSet map[string]struct{}) {
	stack := []string{root}
	rootKey := pathKey(root)
	for len(stack) > 0 {
		if atomic.LoadInt64(&scanID) != id {
			return
		}
		dir := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if key := pathKey(dir); key != rootKey {
			if _, ok := quickRootSet[key]; ok {
				continue
			}
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, de := range entries {
			if atomic.LoadInt64(&scanID) != id {
				return
			}
			info, err := de.Info()
			if err != nil {
				continue
			}
			full := filepath.Join(dir, de.Name())
			isDir := info.IsDir()
			*batch = append(*batch, makeEntry(full, de.Name(), info, isDir))
			if len(*batch) >= 512 {
				flushBatch(id, batch)
			}
			if isDir && !isReparse(info) {
				stack = append(stack, full)
			}
		}
	}
}

func makeEntry(path string, name string, info fs.FileInfo, isDir bool) fileEntry {
	lowerPath := searchFold(path)
	lowerExt := searchFold(filepath.Ext(name))
	return fileEntry{
		Path:      path,
		Name:      name,
		LowerPath: lowerPath,
		LowerName: searchFold(name),
		LowerExt:  lowerExt,
		Size:      info.Size(),
		ModTime:   info.ModTime(),
		IsDir:     isDir,
		Priority:  entryPriority(path),
	}
}

func flushBatch(id int64, batch *[]fileEntry) {
	if len(*batch) == 0 || atomic.LoadInt64(&scanID) != id {
		*batch = (*batch)[:0]
		return
	}
	app.mu.Lock()
	app.entries = append(app.entries, (*batch)...)
	app.mu.Unlock()
	atomic.AddInt64(&indexedCount, int64(len(*batch)))
	*batch = (*batch)[:0]
}

func logicalDrives() []string {
	mask, _, _ := procGetLogicalDrives.Call()
	var roots []string
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		root := fmt.Sprintf("%c:\\", 'A'+rune(i))
		typ, _, _ := procGetDriveType.Call(uintptr(unsafePointer(utf16Ptr(root))))
		switch typ {
		case DRIVE_FIXED, DRIVE_REMOVABLE, DRIVE_REMOTE, DRIVE_RAMDISK:
			roots = append(roots, root)
		}
	}
	return roots
}

func initPathPriority() {
	priorityRootKeys = existingPathKeys(commonUserFolders())
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		userRootKeys = existingPathKeys([]string{home})
	}
	noisy := []string{
		os.Getenv("SystemRoot"),
		os.Getenv("ProgramFiles"),
		os.Getenv("ProgramFiles(x86)"),
		os.Getenv("ProgramW6432"),
		os.Getenv("ProgramData"),
		os.Getenv("TEMP"),
		os.Getenv("TMP"),
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		noisy = append(noisy, filepath.Join(home, "AppData"))
	}
	noisyRootKeys = existingPathKeys(noisy)
}

func commonScanRoots() []string {
	return existingDirs(commonUserFolders())
}

func commonUserFolders() []string {
	var paths []string
	addPath := func(path string) {
		if path != "" {
			paths = append(paths, path)
		}
	}
	addPath(desktopPath())
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		appendKnownFolderNames(&paths, home)
	}
	if oneDrive := os.Getenv("OneDrive"); oneDrive != "" {
		appendKnownFolderNames(&paths, oneDrive)
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		addPath(filepath.Join(appData, "Microsoft", "Windows", "Start Menu"))
	}
	if programData := os.Getenv("ProgramData"); programData != "" {
		addPath(filepath.Join(programData, "Microsoft", "Windows", "Start Menu"))
	}
	return paths
}

func appendKnownFolderNames(paths *[]string, base string) {
	names := []string{
		"Desktop",
		"Masaüstü",
		"Downloads",
		"İndirilenler",
		"Documents",
		"Belgeler",
		"Pictures",
		"Resimler",
		"Music",
		"Müzik",
		"Videos",
		"Videolar",
	}
	for _, name := range names {
		*paths = append(*paths, filepath.Join(base, name))
	}
}

func existingDirs(paths []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, path := range paths {
		key := pathKey(path)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, path)
	}
	return out
}

func existingPathKeys(paths []string) []string {
	dirs := existingDirs(paths)
	keys := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		keys = append(keys, pathKey(dir))
	}
	return keys
}

func entryPriority(path string) int {
	key := pathKey(path)
	if key == "" {
		return 3
	}
	if isUnderAnyKey(key, priorityRootKeys) {
		return 0
	}
	if isUnderAnyKey(key, noisyRootKeys) || hasNoisySegment(key) {
		return 5
	}
	if isUnderAnyKey(key, userRootKeys) {
		return 1
	}
	return 3
}

func hasNoisySegment(key string) bool {
	noisySegments := []string{
		`\windows\`,
		`\program files\`,
		`\program files (x86)\`,
		`\programdata\`,
		`\appdata\`,
		`\node_modules\`,
		`\.git\`,
		`\system volume information\`,
		`\$recycle.bin\`,
	}
	for _, segment := range noisySegments {
		if strings.Contains(key, segment) {
			return true
		}
	}
	return false
}

func isUnderAnyKey(key string, roots []string) bool {
	for _, root := range roots {
		if sameOrUnderKey(key, root) {
			return true
		}
	}
	return false
}

func sameOrUnderKey(key, root string) bool {
	if key == "" || root == "" {
		return false
	}
	if key == root {
		return true
	}
	if strings.HasSuffix(root, `\`) || strings.HasSuffix(root, `/`) {
		return strings.HasPrefix(key, root)
	}
	return strings.HasPrefix(key, root+string(filepath.Separator))
}

func pathKey(path string) string {
	if path == "" {
		return ""
	}
	return strings.ToLower(filepath.Clean(path))
}

func unsafePointer(p *uint16) unsafe.Pointer {
	return unsafe.Pointer(p)
}
