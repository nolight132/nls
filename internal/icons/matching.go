package icons

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nolight132/nls/internal/listing"
)

type Icon struct {
	Name string
	Char string
}

var (
	IconC             = Icon{Name: "C", Char: "\ue61e "}
	IconCpp           = Icon{Name: "C++", Char: "\ue61d "}
	IconMarkdown      = Icon{Name: "Markdown", Char: "\ue609 "}
	IconRust          = Icon{Name: "Rust", Char: "\ue7a8 "}
	IconGo            = Icon{Name: "Go", Char: "\U000f07d3 "}
	IconPython        = Icon{Name: "Python", Char: "\ue73c "}
	IconRuby          = Icon{Name: "Ruby", Char: "\ue791 "}
	IconJavaScript    = Icon{Name: "JavaScript", Char: "\ue74e "}
	IconTypeScript    = Icon{Name: "TypeScript", Char: "\ue628 "}
	IconReact         = Icon{Name: "React", Char: "\ue7ba "}
	IconJava          = Icon{Name: "Java", Char: "\ue738 "}
	IconPHP           = Icon{Name: "PHP", Char: "\ue778 "}
	IconSwift         = Icon{Name: "Swift", Char: "\ue77f "}
	IconDart          = Icon{Name: "Dart", Char: "\ue648 "}
	IconZig           = Icon{Name: "Zig", Char: "\ue6a9 "}
	IconLua           = Icon{Name: "Lua", Char: "\ue637 "}
	IconShell         = Icon{Name: "Shell", Char: "\ue691 "}
	IconTerminal      = Icon{Name: "Terminal", Char: "\ue795 "}
	IconPowerShell    = Icon{Name: "PowerShell", Char: "\uebcc "}
	IconVim           = Icon{Name: "Vimscript", Char: "\ue62b "}
	IconCoffeeScript  = Icon{Name: "CoffeeScript", Char: "\ue61b "}
	IconPerl          = Icon{Name: "Perl", Char: "\ue769 "}
	IconHaskell       = Icon{Name: "Haskell", Char: "\ue777 "}
	IconElixir        = Icon{Name: "Elixir", Char: "\ue777 "}
	IconErlang        = Icon{Name: "Erlang", Char: "\ue7b1 "}
	IconClojure       = Icon{Name: "Clojure", Char: "\ue63a "}
	IconR             = Icon{Name: "R", Char: "\ue68a "}
	IconJulia         = Icon{Name: "Julia", Char: "\ue69b "}
	IconFSharp        = Icon{Name: "F#", Char: "\ue7a1 "}
	IconCSharp        = Icon{Name: "C#", Char: "\U000f031b "}
	IconVisualBasic   = Icon{Name: "Visual Basic", Char: "\U000f07c6 "}
	IconHTML          = Icon{Name: "HTML", Char: "\ue736 "}
	IconCSS           = Icon{Name: "CSS", Char: "\ue749 "}
	IconSass          = Icon{Name: "Sass", Char: "\ue74b "}
	IconLess          = Icon{Name: "Less", Char: "\ue758 "}
	IconTailwind      = Icon{Name: "Tailwind", Char: "\ue697 "}
	IconVue           = Icon{Name: "Vue", Char: "\ue6a0 "}
	IconXML           = Icon{Name: "XML", Char: "\ue736 "}
	IconYAML          = Icon{Name: "YAML", Char: "\ue8eb "}
	IconJSON          = Icon{Name: "JSON", Char: "\ue60b "}
	IconTOML          = Icon{Name: "TOML", Char: "\ue6b2 "}
	IconConfig        = Icon{Name: "Config", Char: "\ueb52 "}
	IconEnvironment   = Icon{Name: "Environment", Char: "\uf084 "}
	IconLockfile      = Icon{Name: "Lockfile", Char: "\uf023 "}
	IconText          = Icon{Name: "Text", Char: "\uf0f6 "}
	IconDockerfile    = Icon{Name: "Dockerfile", Char: "\U000f0868 "}
	IconKubernetes    = Icon{Name: "Kubernetes", Char: "\ue6a8 "}
	IconTerraform     = Icon{Name: "Terraform", Char: "\ue69a "}
	IconAnsible       = Icon{Name: "Ansible", Char: "\U000f0462 "}
	IconMakefile      = Icon{Name: "Makefile", Char: "\U000f0c8b "}
	IconNix           = Icon{Name: "Nix", Char: "\ue7c4 "}
	IconGit           = Icon{Name: "Git", Char: "\ue702 "}
	IconNode          = Icon{Name: "Node.js", Char: "\ue718 "}
	IconNPM           = Icon{Name: "npm", Char: "\ue71e "}
	IconVite          = Icon{Name: "Vite", Char: "\ue7ba "}
	IconPythonPackage = Icon{Name: "Python Package", Char: "\ue73c "}
	IconJupyter       = Icon{Name: "Jupyter", Char: "\ue73c "}
	IconCargo         = Icon{Name: "Cargo", Char: "\ue7a8 "}
	IconGoMod         = Icon{Name: "Go Module", Char: "\U000f07d3 "}
	IconGradle        = Icon{Name: "Gradle", Char: "\ue738 "}
	IconMaven         = Icon{Name: "Maven", Char: "\ue738 "}
	IconDotNet        = Icon{Name: ".NET", Char: "\U000f031b "}
	IconComposer      = Icon{Name: "Composer", Char: "\ue778 "}
	IconGemfile       = Icon{Name: "Gemfile", Char: "\ue791 "}
	IconMix           = Icon{Name: "Mix", Char: "\ue777 "}
	IconPDF           = Icon{Name: "PDF", Char: "\uf1c1 "}
	IconWord          = Icon{Name: "Word", Char: "\uf1c2 "}
	IconExcel         = Icon{Name: "Excel", Char: "\uf1c3 "}
	IconPowerPoint    = Icon{Name: "PowerPoint", Char: "\uf1c4 "}
	IconImage         = Icon{Name: "Image", Char: "\uf03e "}
	IconSVG           = Icon{Name: "SVG", Char: "\uf03e "}
	IconVideo         = Icon{Name: "Video", Char: "\uf008 "}
	IconAudio         = Icon{Name: "Audio", Char: "\uf001 "}
	IconArchive       = Icon{Name: "Archive", Char: "\uf410 "}
	IconBinary        = Icon{Name: "Binary", Char: "\uf471 "}
	IconExecutable    = Icon{Name: "Executable", Char: "\uf013 "}
	IconLibrary       = Icon{Name: "Library", Char: "\uf1c0 "}
	IconFont          = Icon{Name: "Font", Char: "\uf031 "}
	IconDatabase      = Icon{Name: "Database", Char: "\U000f061b "}
	IconSQL           = Icon{Name: "SQL", Char: "\ue706 "}
)

type iconMatcher struct {
	re   *regexp.Regexp
	icon Icon
}

type iconGroup struct {
	icon  Icon
	names []string
}

// exactNames map special filenames (lowercased) to icons. They take
// precedence over patternMatchers and extension matches, so a name listed
// here must not also match an earlier pattern with a different icon.
var exactNames = buildNameMap([]iconGroup{
	{IconDockerfile, []string{"dockerfile", ".dockerignore", "containerfile"}},
	{IconMakefile, []string{"makefile", "gnumakefile", "cmakelists.txt"}},
	{IconLockfile, []string{"package-lock.json", "yarn.lock", "pnpm-lock.yaml", "pnpm-lock.yml", "composer.lock", "gemfile.lock", "cargo.lock", "mix.lock", "poetry.lock", "pipfile.lock"}},
	{IconNode, []string{"package.json", "pnpm-workspace.yaml", "pnpm-workspace.yml", ".yarnrc", ".yarnrc.yaml", ".yarnrc.yml"}},
	{IconNPM, []string{".npmrc"}},
	{IconPythonPackage, []string{"pyproject.toml", "pipfile"}},
	{IconCargo, []string{"cargo.toml"}},
	{IconGoMod, []string{"go.mod", "go.sum", "go.work"}},
	{IconGradle, []string{"build.gradle", "settings.gradle", "build.gradle.kts", "settings.gradle.kts"}},
	{IconMaven, []string{"pom.xml"}},
	{IconComposer, []string{"composer.json"}},
	{IconGemfile, []string{"gemfile"}},
	{IconMix, []string{"mix.exs"}},
	{IconGit, []string{".gitignore", ".gitattributes", ".gitmodules", "codeowners", ".gitlab-ci.yaml", ".gitlab-ci.yml"}},
	{IconKubernetes, []string{"chart.yaml", "chart.yml", "values.yaml", "values.yml"}},
	{IconAnsible, []string{"ansible.cfg", "playbook.yaml", "playbook.yml"}},
})

// patternMatchers cover the special filenames that are not fixed strings.
// Checked in order after exactNames, before the extension lookup.
var patternMatchers = []iconMatcher{
	{regexp.MustCompile(`(?i)^dockerfile\.`), IconDockerfile},
	{regexp.MustCompile(`(?i)^vite\.config\.[cm]?[jt]s$`), IconVite},
	{regexp.MustCompile(`(?i)^tailwind\.config\.[cm]?[jt]s$`), IconTailwind},
	{regexp.MustCompile(`(?i)^(?:webpack|babel|eslint|prettier|jest|vitest|playwright|cypress|postcss)\.config\.[cm]?[jt]s$|^\.(?:babelrc|eslintrc|prettierrc)(?:\..*)?$`), IconJavaScript},
	{regexp.MustCompile(`(?i)^requirements(?:-[\w.-]+)?\.txt$`), IconPythonPackage},
	{regexp.MustCompile(`(?i)(?:^|[-_.])(?:deployment|service|ingress|namespace|configmap|secret|kustomization)\.ya?ml$`), IconKubernetes},
	{regexp.MustCompile(`(?i)^(?:license|copying|changelog|readme)(?:\..*)?$`), IconText},
	{regexp.MustCompile(`(?i)^\.env(?:\..*)?$`), IconEnvironment},
}

// extNames map filename extensions (lowercased, without the dot) to icons.
var extNames = buildNameMap([]iconGroup{
	{IconMakefile, []string{"mk", "make", "cmake"}},
	{IconC, []string{"c", "h", "m", "mm"}},
	{IconCpp, []string{"cpp", "cc", "cxx", "c++", "hpp", "hh", "hxx", "h++", "ipp", "tpp"}},
	{IconMarkdown, []string{"md", "markdown", "mdown", "mkd"}},
	{IconRust, []string{"rs"}},
	{IconGo, []string{"go"}},
	{IconPython, []string{"py", "pyw", "pyi"}},
	{IconRuby, []string{"rb", "erb"}},
	{IconJavaScript, []string{"js", "mjs", "cjs"}},
	{IconTypeScript, []string{"ts", "mts", "cts"}},
	{IconReact, []string{"jsx", "tsx", "svelte"}},
	{IconJava, []string{"java", "kt", "kts", "scala", "sc", "groovy", "gvy"}},
	{IconPHP, []string{"php", "phtml", "hack"}},
	{IconSwift, []string{"swift"}},
	{IconDart, []string{"dart"}},
	{IconZig, []string{"zig"}},
	{IconLua, []string{"lua"}},
	{IconShell, []string{"sh", "bash", "zsh"}},
	{IconTerminal, []string{"fish", "cmd", "bat"}},
	{IconPowerShell, []string{"ps1", "psm1", "psd1"}},
	{IconVim, []string{"vim", "vimrc", "gvimrc"}},
	{IconCoffeeScript, []string{"coffee"}},
	{IconPerl, []string{"pl", "pm", "t"}},
	{IconHaskell, []string{"hs", "lhs"}},
	{IconElixir, []string{"ex", "exs"}},
	{IconErlang, []string{"erl", "hrl"}},
	{IconClojure, []string{"clj", "cljs", "cljc", "edn"}},
	{IconR, []string{"r", "rmd"}},
	{IconJulia, []string{"jl"}},
	{IconFSharp, []string{"fs", "fsi", "fsx"}},
	{IconCSharp, []string{"cs", "csx"}},
	{IconVisualBasic, []string{"vb"}},
	{IconHTML, []string{"html", "htm", "astro"}},
	{IconCSS, []string{"css"}},
	{IconSass, []string{"sass", "scss"}},
	{IconLess, []string{"less"}},
	{IconVue, []string{"vue"}},
	{IconXML, []string{"xml", "xsd", "xsl", "xslt", "plist"}},
	{IconYAML, []string{"yaml", "yml"}},
	{IconJSON, []string{"json", "jsonc", "json5"}},
	{IconTOML, []string{"toml"}},
	{IconConfig, []string{"conf", "config", "cfg", "ini", "properties", "editorconfig"}},
	{IconTerraform, []string{"tf", "tfvars", "hcl"}},
	{IconNix, []string{"nix"}},
	{IconJupyter, []string{"ipynb"}},
	{IconDotNet, []string{"csproj", "fsproj", "vbproj", "sln", "props", "targets"}},
	{IconPDF, []string{"pdf"}},
	{IconWord, []string{"doc", "docx", "odt", "rtf"}},
	{IconExcel, []string{"xls", "xlsx", "ods", "csv", "tsv"}},
	{IconPowerPoint, []string{"ppt", "pptx", "odp"}},
	{IconSVG, []string{"svg", "svgz"}},
	{IconImage, []string{"png", "jpg", "jpeg", "gif", "webp", "bmp", "ico", "tif", "tiff", "avif", "heic"}},
	{IconVideo, []string{"mp4", "mkv", "mov", "webm", "avi", "m4v", "wmv", "flv"}},
	{IconAudio, []string{"mp3", "wav", "flac", "ogg", "m4a", "aac", "opus"}},
	{IconArchive, []string{"zip", "tar", "gz", "tgz", "bz2", "xz", "zst", "7z", "rar", "iso", "dmg"}},
	{IconExecutable, []string{"exe", "app", "out", "com"}},
	{IconLibrary, []string{"so", "dll", "dylib", "a", "lib"}},
	{IconFont, []string{"ttf", "otf", "woff", "woff2", "eot"}},
	{IconDatabase, []string{"db", "sqlite", "sqlite3", "mdb", "accdb", "bson", "rdb"}},
	{IconSQL, []string{"sql", "pgsql", "mysql"}},
	{IconBinary, []string{"bin", "dat", "dump"}},
	{IconText, []string{"txt", "log"}},
})

// buildNameMap resolves duplicate keys first-wins, matching the priority
// order of the original top-to-bottom matcher list.
func buildNameMap(groups []iconGroup) map[string]Icon {
	m := make(map[string]Icon, 4*len(groups))
	for _, g := range groups {
		for _, name := range g.names {
			if _, ok := m[name]; !ok {
				m[name] = g.icon
			}
		}
	}
	return m
}

func MatchIcon(suffix string) string {
	name := strings.TrimSpace(filepath.Base(suffix))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return ""
	}
	lower := strings.ToLower(name)
	if icon, ok := exactNames[lower]; ok {
		return icon.Char
	}
	for _, matcher := range patternMatchers {
		if matcher.re.MatchString(name) {
			return matcher.icon.Char
		}
	}
	if i := strings.LastIndexByte(lower, '.'); i >= 0 && i+1 < len(lower) {
		if icon, ok := extNames[lower[i+1:]]; ok {
			return icon.Char
		}
	}
	return ""
}

func basicIcon(kind listing.Kind) string {
	switch kind {
	case listing.KindDirectory:
		return "\uf07b "
	case listing.KindSymlink:
		return "\uf0c1 "
	case listing.KindExecutable:
		return "\uf013 "
	default:
		return "\uf15b "
	}
}
