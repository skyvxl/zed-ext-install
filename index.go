package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// ExtensionIndex represents the Zed extensions/index.json structure.
type ExtensionIndex struct {
	Extensions map[string]ExtensionIndexEntry     `json:"extensions"`
	Themes     map[string]ExtensionIndexThemeEntry `json:"themes"`
	IconThemes map[string]interface{}              `json:"icon_themes"`
	Languages  map[string]ExtensionIndexLangEntry  `json:"languages"`
}

type ExtensionIndexEntry struct {
	Manifest ExtensionManifest `json:"manifest"`
	Dev      bool              `json:"dev"`
}

type ExtensionManifest struct {
	ID          string            `json:"id" toml:"id"`
	Name        string            `json:"name" toml:"name"`
	Version     string            `json:"version" toml:"version"`
	Description string            `json:"description" toml:"description"`
	Authors     []string          `json:"authors" toml:"authors"`
	Repository  string            `json:"repository" toml:"repository"`
	Lib         *ManifestLib      `json:"lib" toml:"lib"`
	Themes      []string          `json:"themes" toml:"themes"`
	IconThemes  []string          `json:"icon_themes" toml:"icon_themes"`
	Languages   []string          `json:"languages" toml:"languages"`
	Grammars    map[string]interface{} `json:"grammars" toml:"grammars"`
}

type ManifestLib struct {
	Kind    string `json:"kind,omitempty" toml:"kind"`
	Version string `json:"version,omitempty" toml:"version"`
}

type ExtensionIndexThemeEntry struct {
	Extension string `json:"extension"`
	Path      string `json:"path"`
}

type ExtensionIndexLangEntry struct {
	Extension string `json:"extension"`
	Path      string `json:"path"`
}

// LoadIndex reads the existing index.json or returns an empty index.
func LoadIndex(paths *ZedPaths) (*ExtensionIndex, error) {
	data, err := os.ReadFile(paths.Index)
	if err != nil {
		if os.IsNotExist(err) {
			return newEmptyIndex(), nil
		}
		return nil, fmt.Errorf("read index: %w", err)
	}

	var idx ExtensionIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}

	if idx.Extensions == nil {
		idx.Extensions = make(map[string]ExtensionIndexEntry)
	}
	if idx.Themes == nil {
		idx.Themes = make(map[string]ExtensionIndexThemeEntry)
	}
	if idx.IconThemes == nil {
		idx.IconThemes = make(map[string]interface{})
	}
	if idx.Languages == nil {
		idx.Languages = make(map[string]ExtensionIndexLangEntry)
	}

	return &idx, nil
}

func newEmptyIndex() *ExtensionIndex {
	return &ExtensionIndex{
		Extensions: make(map[string]ExtensionIndexEntry),
		Themes:     make(map[string]ExtensionIndexThemeEntry),
		IconThemes: make(map[string]interface{}),
		Languages:  make(map[string]ExtensionIndexLangEntry),
	}
}

// SaveIndex writes the index.json file.
func SaveIndex(paths *ZedPaths, idx *ExtensionIndex) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	if err := os.WriteFile(paths.Index, data, 0644); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return nil
}

// UpdateIndexForExtension reads extension.toml from the installed extension
// and updates the index accordingly.
func UpdateIndexForExtension(extID string, paths *ZedPaths, idx *ExtensionIndex) error {
	extDir := filepath.Join(paths.Installed, extID)
	tomlPath := filepath.Join(extDir, "extension.toml")

	data, err := os.ReadFile(tomlPath)
	if err != nil {
		return fmt.Errorf("read extension.toml: %w", err)
	}

	var manifest ExtensionManifest
	if err := toml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("parse extension.toml: %w", err)
	}

	// Add extension entry
	idx.Extensions[extID] = ExtensionIndexEntry{
		Manifest: manifest,
		Dev:      false,
	}

	// Register languages
	if manifest.Languages != nil {
		for _, langDir := range manifest.Languages {
			langName := filepath.Base(langDir)
			// Try to read config.toml for the actual language name
			configPath := filepath.Join(extDir, langDir, "config.toml")
			if name := readLanguageName(configPath); name != "" {
				langName = name
			}
			idx.Languages[langName] = ExtensionIndexLangEntry{
				Extension: extID,
				Path:      langDir,
			}
		}
	} else {
		// Auto-detect languages from filesystem
		langsDir := filepath.Join(extDir, "languages")
		entries, err := os.ReadDir(langsDir)
		if err == nil {
			for _, e := range entries {
				if e.IsDir() {
					langPath := filepath.Join("languages", e.Name())
					langName := e.Name()
					configPath := filepath.Join(langsDir, e.Name(), "config.toml")
					if name := readLanguageName(configPath); name != "" {
						langName = name
					}
					idx.Languages[langName] = ExtensionIndexLangEntry{
						Extension: extID,
						Path:      langPath,
					}
				}
			}
		}
	}

	// Register themes
	if manifest.Themes != nil {
		for _, themePath := range manifest.Themes {
			themeName := strings.TrimSuffix(filepath.Base(themePath), filepath.Ext(themePath))
			idx.Themes[themeName] = ExtensionIndexThemeEntry{
				Extension: extID,
				Path:      themePath,
			}
		}
	} else {
		// Auto-detect themes
		themesDir := filepath.Join(extDir, "themes")
		entries, err := os.ReadDir(themesDir)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
					themeName := strings.TrimSuffix(e.Name(), ".json")
					idx.Themes[themeName] = ExtensionIndexThemeEntry{
						Extension: extID,
						Path:      filepath.Join("themes", e.Name()),
					}
				}
			}
		}
	}

	return nil
}

// RemoveFromIndex removes an extension and its resources from the index.
func RemoveFromIndex(extID string, idx *ExtensionIndex) {
	delete(idx.Extensions, extID)

	for name, entry := range idx.Themes {
		if entry.Extension == extID {
			delete(idx.Themes, name)
		}
	}
	for name, entry := range idx.Languages {
		if entry.Extension == extID {
			delete(idx.Languages, name)
		}
	}
}

func readLanguageName(configPath string) string {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}
	var config struct {
		Name string `toml:"name"`
	}
	if err := toml.Unmarshal(data, &config); err != nil {
		return ""
	}
	return config.Name
}
