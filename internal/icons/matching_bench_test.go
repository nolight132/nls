package icons

import "testing"

func BenchmarkMatchIcon(b *testing.B) {
	cases := []struct {
		name  string
		files []string
	}{
		{"special", []string{"Dockerfile", "Makefile", "package.json", "go.mod", "README.md"}},
		{"code", []string{"main.go", "lib.rs", "index.ts", "app.jsx", "styles.css"}},
		{"late", []string{"notes.txt", "server.log", "photo.png", "backup.zip", "video.mp4"}},
		{"unmatched", []string{"noextension", "weird.xyz", "data.foo", "UPPER.QUX", "file"}},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				for _, f := range tc.files {
					MatchIcon(f)
				}
			}
		})
	}
}
