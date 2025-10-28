package hotkey

import "golang.design/x/hotkey"

// ConflictInfo represents information about a known shortcut conflict
type ConflictInfo struct {
	Name        string
	Description string
	Modifiers   []hotkey.Modifier
	Key         hotkey.Key
}

// knownConflicts contains a list of known macOS shortcuts that might conflict
var knownConflicts = []ConflictInfo{
	{
		Name:        "Spotlight",
		Description: "macOS Spotlight search",
		Modifiers:   []hotkey.Modifier{hotkey.ModCmd},
		Key:         hotkey.KeySpace,
	},
	{
		Name:        "Alfred",
		Description: "Alfred launcher (common default)",
		Modifiers:   []hotkey.Modifier{hotkey.ModCmd},
		Key:         hotkey.KeySpace,
	},
	{
		Name:        "Raycast",
		Description: "Raycast launcher (common default)",
		Modifiers:   []hotkey.Modifier{hotkey.ModCmd},
		Key:         hotkey.KeySpace,
	},
	{
		Name:        "IME Switch",
		Description: "Input method editor switch",
		Modifiers:   []hotkey.Modifier{hotkey.ModCmd},
		Key:         hotkey.KeySpace,
	},
	{
		Name:        "Force Quit",
		Description: "macOS Force Quit",
		Modifiers:   []hotkey.Modifier{hotkey.ModCmd, hotkey.ModOption},
		Key:         hotkey.KeyEscape,
	},
}

// CheckConflicts checks if the given hotkey conflicts with known system shortcuts
func CheckConflicts(modifiers []hotkey.Modifier, key hotkey.Key) []ConflictInfo {
	var conflicts []ConflictInfo

	for _, known := range knownConflicts {
		if hotkeyMatches(modifiers, key, known.Modifiers, known.Key) {
			conflicts = append(conflicts, known)
		}
	}

	return conflicts
}

// hotkeyMatches checks if two hotkey combinations are identical
func hotkeyMatches(mods1 []hotkey.Modifier, key1 hotkey.Key, mods2 []hotkey.Modifier, key2 hotkey.Key) bool {
	if key1 != key2 {
		return false
	}

	if len(mods1) != len(mods2) {
		return false
	}

	// Create maps for comparison
	modMap1 := make(map[hotkey.Modifier]bool)
	modMap2 := make(map[hotkey.Modifier]bool)

	for _, mod := range mods1 {
		modMap1[mod] = true
	}

	for _, mod := range mods2 {
		modMap2[mod] = true
	}

	// Check if all modifiers match
	for mod := range modMap1 {
		if !modMap2[mod] {
			return false
		}
	}

	return true
}

// FormatHotkey returns a human-readable string representation of the hotkey
func FormatHotkey(modifiers []hotkey.Modifier, key hotkey.Key) string {
	result := ""

	for _, mod := range modifiers {
		switch mod {
		case hotkey.ModCtrl:
			result += "⌃"
		case hotkey.ModShift:
			result += "⇧"
		case hotkey.ModOption:
			result += "⌥"
		case hotkey.ModCmd:
			result += "⌘"
		}
	}

	result += keyToString(key)
	return result
}

// keyToString converts a hotkey.Key to a display string
func keyToString(key hotkey.Key) string {
	// Map common keys to their display names
	keyMap := map[hotkey.Key]string{
		hotkey.KeySpace:  "Space",
		hotkey.KeyEscape: "Esc",
		hotkey.KeyReturn: "Return",
		hotkey.KeyTab:    "Tab",
		hotkey.KeyDelete: "Delete",
	}

	if name, ok := keyMap[key]; ok {
		return name
	}

	// For letter keys (A-Z)
	if key >= hotkey.KeyA && key <= hotkey.KeyZ {
		return string(rune('A' + int(key-hotkey.KeyA)))
	}

	// For number keys (0-9)
	if key >= hotkey.Key0 && key <= hotkey.Key9 {
		return string(rune('0' + int(key-hotkey.Key0)))
	}

	return "Unknown"
}
