package systemdctl

import "strings"

// UnitNameForGameID turns an arbitrary game identifier into a stable, safe
// systemd scope unit name: game-<id>.scope.
func UnitNameForGameID(gameID string) string {
	gameID = strings.TrimSpace(gameID)
	if gameID == "" {
		gameID = "unknown"
	}

	var b strings.Builder
	for _, r := range gameID {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	sanitized := b.String()
	sanitized = strings.Trim(sanitized, "-_")
	if sanitized == "" {
		sanitized = "unknown"
	}
	if len(sanitized) > 80 {
		sanitized = sanitized[:80]
	}
	return "game-" + sanitized + ".scope"
}
