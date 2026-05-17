//go:build windows

package main

import (
	"path/filepath"
	"sort"
	"strings"
)

type searchQuery struct {
	tokens []string
	exts   []string
	phrase string
}

type rankedEntry struct {
	entry fileEntry
	score int
}

func runSearch() {
	// Wails search is handled reactively by exposed Search method in app.go
}

func refreshFromIndex() {
	// Wails indexing updates are sent via status_update events in startup
}

func searchIndex(query string, limit int) []fileEntry {
	parsed := parseSearchQuery(query)
	if len(parsed.tokens) == 0 && len(parsed.exts) == 0 {
		return nil
	}

	matches := make([]rankedEntry, 0, limit)

	app.mu.RLock()
	defer app.mu.RUnlock()

	for _, e := range app.entries {
		if !matchesQuery(e, parsed) {
			continue
		}
		matches = append(matches, rankedEntry{entry: e, score: scoreEntry(e, parsed)})
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score < matches[j].score
		}
		return matches[i].entry.LowerPath < matches[j].entry.LowerPath
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}
	results := make([]fileEntry, len(matches))
	for i, match := range matches {
		results[i] = match.entry
	}
	return results
}

func parseSearchQuery(query string) searchQuery {
	rawTokens := strings.Fields(searchFold(query))
	parsed := searchQuery{}
	for _, token := range rawTokens {
		if ext, ok := parseExtensionToken(token); ok {
			parsed.exts = append(parsed.exts, ext)
			continue
		}
		parsed.tokens = append(parsed.tokens, token)
	}
	parsed.phrase = strings.Join(parsed.tokens, " ")
	return parsed
}

func parseExtensionToken(token string) (string, bool) {
	switch {
	case strings.HasPrefix(token, "extension:"):
		token = strings.TrimPrefix(token, "extension:")
	case strings.HasPrefix(token, "ext:"):
		token = strings.TrimPrefix(token, "ext:")
	case strings.HasPrefix(token, "*."):
		token = strings.TrimPrefix(token, "*.")
	case strings.HasPrefix(token, ".") && len(token) > 1:
		token = strings.TrimPrefix(token, ".")
	default:
		return "", false
	}
	token = strings.Trim(token, ". ")
	if token == "" || strings.ContainsAny(token, `\/:*?"<>|`) {
		return "", false
	}
	return token, true
}

func matchesQuery(e fileEntry, query searchQuery) bool {
	if len(query.exts) > 0 {
		ext := strings.TrimPrefix(e.LowerExt, ".")
		if ext == "" {
			return false
		}
		matchedExt := false
		for _, wanted := range query.exts {
			if ext == wanted {
				matchedExt = true
				break
			}
		}
		if !matchedExt {
			return false
		}
	}
	if len(query.tokens) == 0 {
		return true
	}
	return matchesAll(e.LowerPath, query.tokens)
}

func scoreEntry(e fileEntry, query searchQuery) int {
	score := e.Priority*1000 + pathDepth(e.Path)*2
	if e.IsDir {
		score -= 80
	}
	ext := strings.TrimPrefix(e.LowerExt, ".")
	switch ext {
	case "lnk", "url", "exe", "appref-ms":
		score -= 60
	case "dll", "sys", "tmp", "log", "cache", "dat", "pak", "manifest":
		score += 180
	}
	if query.phrase != "" {
		stem := strings.TrimSuffix(e.LowerName, e.LowerExt)
		switch {
		case e.LowerName == query.phrase || stem == query.phrase:
			score -= 320
		case strings.HasPrefix(stem, query.phrase) || strings.HasPrefix(e.LowerName, query.phrase):
			score -= 220
		case strings.Contains(stem, query.phrase) || strings.Contains(e.LowerName, query.phrase):
			score -= 130
		case matchesAll(e.LowerName, query.tokens):
			score -= 90
		}
	}
	for _, wanted := range query.exts {
		if ext == wanted {
			score -= 120
			break
		}
	}
	if strings.Contains(e.LowerPath, `\setup\`) || strings.Contains(e.LowerPath, `\installer\`) {
		score += 120
	}
	return score
}

func pathDepth(path string) int {
	cleaned := filepath.Clean(path)
	if cleaned == "." || cleaned == string(osPathSeparator()) {
		return 0
	}
	return strings.Count(cleaned, string(osPathSeparator()))
}

func matchesAll(s string, tokens []string) bool {
	for _, token := range tokens {
		if !strings.Contains(s, token) {
			return false
		}
	}
	return true
}

func searchFold(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case 'ı', 'İ':
			r = 'i'
		case '\u0307':
			continue
		case 'ç':
			r = 'c'
		case 'ğ':
			r = 'g'
		case 'ö':
			r = 'o'
		case 'ş':
			r = 's'
		case 'ü':
			r = 'u'
		}
		b.WriteRune(r)
	}
	return b.String()
}

func osPathSeparator() byte {
	return '\\'
}
