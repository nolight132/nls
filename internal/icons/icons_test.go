package icons

import (
	"testing"

	"github.com/nolight132/nls/internal/listing"
)

func TestResolveNoIconsFlagWins(t *testing.T) {
	t.Setenv("NLS_ICONS", "1")
	t.Setenv("NERD_FONT", "1")
	if got := Resolve(true, true, true); got != SetNone {
		t.Fatalf("no-icons flag should win, got %v", got)
	}
}

func TestResolveEnvOverridesConfigOff(t *testing.T) {
	t.Setenv("NLS_ICONS", "1")
	t.Setenv("NERD_FONT", "1")
	if got := Resolve(false, false, true); got != SetNerd {
		t.Fatalf("env on + config off should enable, got %v", got)
	}
}

func TestResolveConfigEnablesWhenEnvUnset(t *testing.T) {
	t.Setenv("NLS_ICONS", "")
	t.Setenv("NERD_FONT", "1")
	if got := Resolve(false, true, true); got != SetNerd {
		t.Fatalf("config on + env unset should enable, got %v", got)
	}
}

func TestResolveDefaultsOff(t *testing.T) {
	t.Setenv("NLS_ICONS", "")
	t.Setenv("NERD_FONT", "1")
	if got := Resolve(false, false, true); got != SetNone {
		t.Fatalf("everything off should yield SetNone, got %v", got)
	}
}

func TestResolveNeedsNerdFont(t *testing.T) {
	t.Setenv("NLS_ICONS", "")
	t.Setenv("NERD_FONT", "0")
	t.Setenv("NLS_NERD_FONT", "0")
	t.Setenv("TERM", "")
	t.Setenv("FONT", "")
	t.Setenv("FONTFACE", "")
	t.Setenv("FONT_FAMILY", "")
	if got := Resolve(false, true, true); got != SetNone {
		t.Fatalf("no nerd font available should yield SetNone, got %v", got)
	}
}

func TestResolveCanDisableSpecialIcons(t *testing.T) {
	t.Setenv("NLS_ICONS", "")
	t.Setenv("NERD_FONT", "1")
	if got := Resolve(false, true, false); got != SetNerdBasic {
		t.Fatalf("special icons disabled should yield basic nerd set, got %v", got)
	}
}

func TestMatchIconTypes(t *testing.T) {
	tests := []struct {
		name string
		file string
		icon Icon
	}{
		{"c", "main.c", IconC},
		{"cpp", "main.cpp", IconCpp},
		{"markdown", "guide.md", IconMarkdown},
		{"rust", "lib.rs", IconRust},
		{"go", "server.go", IconGo},
		{"python", "script.py", IconPython},
		{"ruby", "task.rb", IconRuby},
		{"javascript", "app.js", IconJavaScript},
		{"typescript", "app.ts", IconTypeScript},
		{"react", "component.tsx", IconReact},
		{"java", "Main.java", IconJava},
		{"php", "index.php", IconPHP},
		{"swift", "App.swift", IconSwift},
		{"dart", "widget.dart", IconDart},
		{"zig", "build.zig", IconZig},
		{"lua", "init.lua", IconLua},
		{"shell", "install.sh", IconShell},
		{"terminal", "run.fish", IconTerminal},
		{"powershell", "profile.ps1", IconPowerShell},
		{"vim", "plugin.vim", IconVim},
		{"coffeescript", "cake.coffee", IconCoffeeScript},
		{"perl", "script.pl", IconPerl},
		{"haskell", "Main.hs", IconHaskell},
		{"elixir", "app.ex", IconElixir},
		{"erlang", "server.erl", IconErlang},
		{"clojure", "core.clj", IconClojure},
		{"r", "plot.r", IconR},
		{"julia", "model.jl", IconJulia},
		{"fsharp", "Program.fs", IconFSharp},
		{"csharp", "Program.cs", IconCSharp},
		{"visual-basic", "Module.vb", IconVisualBasic},
		{"html", "index.html", IconHTML},
		{"css", "style.css", IconCSS},
		{"sass", "style.scss", IconSass},
		{"less", "style.less", IconLess},
		{"tailwind", "tailwind.config.js", IconTailwind},
		{"vue", "App.vue", IconVue},
		{"xml", "schema.xml", IconXML},
		{"yaml", "data.yaml", IconYAML},
		{"json", "data.json", IconJSON},
		{"toml", "config.toml", IconTOML},
		{"config", "app.ini", IconConfig},
		{"environment", ".env.local", IconEnvironment},
		{"lockfile", "yarn.lock", IconLockfile},
		{"text", "notes.txt", IconText},
		{"dockerfile", "Dockerfile", IconDockerfile},
		{"kubernetes", "deployment.yaml", IconKubernetes},
		{"terraform", "main.tf", IconTerraform},
		{"ansible", "ansible.cfg", IconAnsible},
		{"makefile", "Makefile", IconMakefile},
		{"nix", "flake.nix", IconNix},
		{"git", ".gitignore", IconGit},
		{"node", "package.json", IconNode},
		{"npm", ".npmrc", IconNPM},
		{"vite", "vite.config.ts", IconVite},
		{"python-package", "pyproject.toml", IconPythonPackage},
		{"jupyter", "notebook.ipynb", IconJupyter},
		{"cargo", "Cargo.toml", IconCargo},
		{"go-mod", "go.mod", IconGoMod},
		{"gradle", "build.gradle", IconGradle},
		{"maven", "pom.xml", IconMaven},
		{"dotnet", "app.csproj", IconDotNet},
		{"composer", "composer.json", IconComposer},
		{"gemfile", "Gemfile", IconGemfile},
		{"mix", "mix.exs", IconMix},
		{"pdf", "manual.pdf", IconPDF},
		{"word", "draft.docx", IconWord},
		{"excel", "sheet.xlsx", IconExcel},
		{"powerpoint", "deck.pptx", IconPowerPoint},
		{"image", "photo.png", IconImage},
		{"svg", "logo.svg", IconSVG},
		{"video", "movie.mp4", IconVideo},
		{"audio", "song.mp3", IconAudio},
		{"archive", "bundle.zip", IconArchive},
		{"binary", "blob.bin", IconBinary},
		{"executable", "tool.exe", IconExecutable},
		{"library", "libsqlite.so", IconLibrary},
		{"font", "display.ttf", IconFont},
		{"database", "app.sqlite", IconDatabase},
		{"sql", "query.sql", IconSQL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchIcon(tt.file); got != tt.icon.Char {
				t.Fatalf("MatchIcon(%q) = %q, want %q (%s)", tt.file, got, tt.icon.Char, tt.icon.Name)
			}
		})
	}
}

func TestMatchIconSpecialIcons(t *testing.T) {
	tests := []struct {
		name string
		file string
		icon Icon
	}{
		{"env uses key", ".env.local", IconEnvironment},
		{"font uses font glyph", "display.ttf", IconFont},
		{"lockfile uses lock", "package-lock.json", IconLockfile},
		{"config uses settings", "app.ini", IconConfig},
		{"yaml uses yaml glyph", "data.yaml", IconYAML},
		{"text uses readable file", "notes.txt", IconText},
		{"go module uses go", "go.mod", IconGoMod},
		{"dockerfile uses docker", "Dockerfile", IconDockerfile},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchIcon(tt.file); got != tt.icon.Char {
				t.Fatalf("MatchIcon(%q) = %q, want %q (%s)", tt.file, got, tt.icon.Char, tt.icon.Name)
			}
		})
	}
}

func TestForHonorsSpecialIconSet(t *testing.T) {
	entry := listing.Entry{Name: ".env", Kind: listing.KindFile}
	if got := For(entry, SetNerd); got != IconEnvironment.Char {
		t.Fatalf("special set icon = %q, want %q", got, IconEnvironment.Char)
	}
	if got := For(entry, SetNerdBasic); got != basicIcon(listing.KindFile) {
		t.Fatalf("basic set icon = %q, want file fallback", got)
	}
	if got := For(entry, SetNone); got != "" {
		t.Fatalf("no icon set = %q, want empty", got)
	}
}

func TestMatchIconUsesBaseName(t *testing.T) {
	if got := MatchIcon("/tmp/project/main.go"); got != IconGo.Char {
		t.Fatalf("MatchIcon should match against the base name, got %q, want %q", got, IconGo.Char)
	}
}

func TestMatchIconUnknown(t *testing.T) {
	if got := MatchIcon("unknown-file"); got != "" {
		t.Fatalf("unknown file should not match an icon, got %q", got)
	}
}
