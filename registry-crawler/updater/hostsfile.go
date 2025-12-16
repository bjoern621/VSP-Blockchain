package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// buildSeedHostsFile generates a CoreDNS hosts file content.
// The file maps IP addresses to DNS names in the configured zone.
// Format per line: <ip> <name>.<zone>. <name>.<namespace>.<zone>.
func buildSeedHostsFile(seedServiceName, namespace, zone string, ipStrings []string, source string) (string, error) {
	zone = strings.TrimSpace(zone)
	if zone == "" {
		return "", fmt.Errorf("seed dns zone is empty")
	}
	zone = strings.TrimSuffix(zone, ".")

	baseName := fmt.Sprintf("%s.%s", seedServiceName, zone)
	namespacedName := fmt.Sprintf("%s.%s.%s", seedServiceName, namespace, zone)

	lines := []string{
		"# Managed by registry-crawler. One line per IP.",
		fmt.Sprintf("# generated_at=%s source=%s", time.Now().UTC().Format(time.RFC3339Nano), strings.TrimSpace(source)),
	}
	for _, ip := range ipStrings {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s %s. %s.", ip, baseName, namespacedName))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n"), nil
}

// writeFileAtomically writes data to a file atomically using a temp file and rename.
// This prevents partial writes from corrupting the file.
func writeFileAtomically(path string, data []byte) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("empty path")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
