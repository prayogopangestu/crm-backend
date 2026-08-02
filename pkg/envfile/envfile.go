// Package envfile memuat variabel dari file .env ke environment proses.
// Variabel yang sudah di-set lebih dulu (mis. dari Docker/deploy) tidak ditimpa,
// sehingga fallback ke configs/config.yaml tetap berfungsi saat tidak ada .env.
package envfile

import (
	"os"
	"strings"
)

// Load membaca path dan mengisi environment dengan key=value dari file.
// File yang tidak ada diabaikan (bukan error), karena ini bersifat opsional.
func Load(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}
