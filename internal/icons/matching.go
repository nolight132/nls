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

var iconMatchers = []iconMatcher{
	{regexp.MustCompile(`(?i)^dockerfile(?:\..*)?$|^\.dockerignore$|^containerfile$`), IconDockerfile},
	{regexp.MustCompile(`(?i)^makefile$|^gnumakefile$|^cmakelists\.txt$|\.(?:mk|make|cmake)$`), IconMakefile},
	{regexp.MustCompile(`(?i)^package-lock\.json$|^yarn\.lock$|^pnpm-lock\.ya?ml$|^composer\.lock$|^gemfile\.lock$|^cargo\.lock$|^mix\.lock$|^poetry\.lock$|^pipfile\.lock$`), IconLockfile},
	{regexp.MustCompile(`(?i)^package\.json$`), IconNode},
	{regexp.MustCompile(`(?i)^\.npmrc$`), IconNPM},
	{regexp.MustCompile(`(?i)^pnpm-workspace\.ya?ml$|^\.yarnrc(?:\.ya?ml)?$`), IconNode},
	{regexp.MustCompile(`(?i)^vite\.config\.[cm]?[jt]s$`), IconVite},
	{regexp.MustCompile(`(?i)^tailwind\.config\.[cm]?[jt]s$`), IconTailwind},
	{regexp.MustCompile(`(?i)^(?:webpack|babel|eslint|prettier|jest|vitest|playwright|cypress|postcss)\.config\.[cm]?[jt]s$|^\.(?:babelrc|eslintrc|prettierrc)(?:\..*)?$`), IconJavaScript},
	{regexp.MustCompile(`(?i)^pyproject\.toml$|^requirements(?:-[\w.-]+)?\.txt$|^pipfile$`), IconPythonPackage},
	{regexp.MustCompile(`(?i)^cargo\.toml$`), IconCargo},
	{regexp.MustCompile(`(?i)^go\.(?:mod|sum|work)$`), IconGoMod},
	{regexp.MustCompile(`(?i)^(?:build|settings)\.gradle(?:\.kts)?$`), IconGradle},
	{regexp.MustCompile(`(?i)^pom\.xml$`), IconMaven},
	{regexp.MustCompile(`(?i)^(?:composer\.json|composer\.lock)$`), IconComposer},
	{regexp.MustCompile(`(?i)^gemfile$`), IconGemfile},
	{regexp.MustCompile(`(?i)^mix\.exs$`), IconMix},
	{regexp.MustCompile(`(?i)^\.gitignore$|^\.gitattributes$|^\.gitmodules$|^codeowners$|^\.gitlab-ci\.ya?ml$`), IconGit},
	{regexp.MustCompile(`(?i)^chart\.ya?ml$|^values\.ya?ml$|(?:^|[-_.])(?:deployment|service|ingress|namespace|configmap|secret|kustomization)\.ya?ml$`), IconKubernetes},
	{regexp.MustCompile(`(?i)^ansible\.cfg$|^playbook\.ya?ml$`), IconAnsible},
	{regexp.MustCompile(`(?i)^license(?:\..*)?$|^copying(?:\..*)?$|^changelog(?:\..*)?$|^readme(?:\..*)?$`), IconText},
	{regexp.MustCompile(`(?i)^\.env(?:\..*)?$`), IconEnvironment},

	{regexp.MustCompile(`(?i)\.(?:c)$`), IconC},
	{regexp.MustCompile(`(?i)\.(?:h)$`), IconC},
	{regexp.MustCompile(`(?i)\.(?:cpp|cc|cxx|c\+\+|hpp|hh|hxx|h\+\+|ipp|tpp)$`), IconCpp},
	{regexp.MustCompile(`(?i)\.(?:m|mm)$`), IconC},
	{regexp.MustCompile(`(?i)\.(?:md|markdown|mdown|mkd)$`), IconMarkdown},
	{regexp.MustCompile(`(?i)\.(?:rs)$`), IconRust},
	{regexp.MustCompile(`(?i)\.(?:go)$`), IconGo},
	{regexp.MustCompile(`(?i)\.(?:py|pyw|pyi)$`), IconPython},
	{regexp.MustCompile(`(?i)\.(?:rb|erb)$`), IconRuby},
	{regexp.MustCompile(`(?i)\.(?:js|mjs|cjs)$`), IconJavaScript},
	{regexp.MustCompile(`(?i)\.(?:ts|mts|cts)$`), IconTypeScript},
	{regexp.MustCompile(`(?i)\.(?:jsx|tsx)$`), IconReact},
	{regexp.MustCompile(`(?i)\.(?:java|kt|kts|scala|sc|groovy|gvy)$`), IconJava},
	{regexp.MustCompile(`(?i)\.(?:php|phtml|blade\.php|hack)$`), IconPHP},
	{regexp.MustCompile(`(?i)\.(?:swift)$`), IconSwift},
	{regexp.MustCompile(`(?i)\.(?:dart)$`), IconDart},
	{regexp.MustCompile(`(?i)\.(?:zig)$`), IconZig},
	{regexp.MustCompile(`(?i)\.(?:lua)$`), IconLua},
	{regexp.MustCompile(`(?i)\.(?:sh|bash|zsh)$`), IconShell},
	{regexp.MustCompile(`(?i)\.(?:fish|cmd|bat)$`), IconTerminal},
	{regexp.MustCompile(`(?i)\.(?:ps1|psm1|psd1)$`), IconPowerShell},
	{regexp.MustCompile(`(?i)\.(?:vim|vimrc|gvimrc)$`), IconVim},
	{regexp.MustCompile(`(?i)\.(?:coffee)$`), IconCoffeeScript},
	{regexp.MustCompile(`(?i)\.(?:pl|pm|t)$`), IconPerl},
	{regexp.MustCompile(`(?i)\.(?:hs|lhs)$`), IconHaskell},
	{regexp.MustCompile(`(?i)\.(?:ex|exs)$`), IconElixir},
	{regexp.MustCompile(`(?i)\.(?:erl|hrl)$`), IconErlang},
	{regexp.MustCompile(`(?i)\.(?:clj|cljs|cljc|edn)$`), IconClojure},
	{regexp.MustCompile(`(?i)\.(?:r|rmd)$`), IconR},
	{regexp.MustCompile(`(?i)\.(?:jl)$`), IconJulia},
	{regexp.MustCompile(`(?i)\.(?:fs|fsi|fsx)$`), IconFSharp},
	{regexp.MustCompile(`(?i)\.(?:cs|csx)$`), IconCSharp},
	{regexp.MustCompile(`(?i)\.(?:vb)$`), IconVisualBasic},

	{regexp.MustCompile(`(?i)\.(?:html|htm|astro)$`), IconHTML},
	{regexp.MustCompile(`(?i)\.(?:css)$`), IconCSS},
	{regexp.MustCompile(`(?i)\.(?:sass|scss)$`), IconSass},
	{regexp.MustCompile(`(?i)\.(?:less)$`), IconLess},
	{regexp.MustCompile(`(?i)\.(?:vue)$`), IconVue},
	{regexp.MustCompile(`(?i)\.(?:svelte)$`), IconReact},
	{regexp.MustCompile(`(?i)\.(?:xml|xsd|xsl|xslt|plist)$`), IconXML},
	{regexp.MustCompile(`(?i)\.(?:ya?ml)$`), IconYAML},
	{regexp.MustCompile(`(?i)\.(?:json|jsonc|json5)$`), IconJSON},
	{regexp.MustCompile(`(?i)\.(?:toml)$`), IconTOML},
	{regexp.MustCompile(`(?i)\.(?:conf|config|cfg|ini|properties|editorconfig)$`), IconConfig},

	{regexp.MustCompile(`(?i)\.(?:tf|tfvars|hcl)$`), IconTerraform},
	{regexp.MustCompile(`(?i)\.(?:nix)$`), IconNix},
	{regexp.MustCompile(`(?i)\.(?:ipynb)$`), IconJupyter},
	{regexp.MustCompile(`(?i)\.(?:csproj|fsproj|vbproj|sln|props|targets)$`), IconDotNet},

	{regexp.MustCompile(`(?i)\.(?:pdf)$`), IconPDF},
	{regexp.MustCompile(`(?i)\.(?:doc|docx|odt|rtf)$`), IconWord},
	{regexp.MustCompile(`(?i)\.(?:xls|xlsx|ods|csv|tsv)$`), IconExcel},
	{regexp.MustCompile(`(?i)\.(?:ppt|pptx|odp)$`), IconPowerPoint},
	{regexp.MustCompile(`(?i)\.(?:svg|svgz)$`), IconSVG},
	{regexp.MustCompile(`(?i)\.(?:png|jpe?g|gif|webp|bmp|ico|tiff?|avif|heic)$`), IconImage},
	{regexp.MustCompile(`(?i)\.(?:mp4|mkv|mov|webm|avi|m4v|wmv|flv)$`), IconVideo},
	{regexp.MustCompile(`(?i)\.(?:mp3|wav|flac|ogg|m4a|aac|opus)$`), IconAudio},
	{regexp.MustCompile(`(?i)\.(?:zip|tar|gz|tgz|bz2|xz|zst|7z|rar|iso|dmg)$`), IconArchive},
	{regexp.MustCompile(`(?i)\.(?:exe|app|out|com)$`), IconExecutable},
	{regexp.MustCompile(`(?i)\.(?:so|dll|dylib|a|lib)$`), IconLibrary},
	{regexp.MustCompile(`(?i)\.(?:ttf|otf|woff2?|eot)$`), IconFont},
	{regexp.MustCompile(`(?i)\.(?:db|sqlite3?|mdb|accdb|bson|rdb)$`), IconDatabase},
	{regexp.MustCompile(`(?i)\.(?:sql|pgsql|mysql)$`), IconSQL},
	{regexp.MustCompile(`(?i)\.(?:bin|dat|dump)$`), IconBinary},
	{regexp.MustCompile(`(?i)\.(?:txt|log)$`), IconText},
}

func MatchIcon(suffix string) string {
	name := strings.TrimSpace(filepath.Base(suffix))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return ""
	}
	for _, matcher := range iconMatchers {
		if matcher.re.MatchString(name) {
			return matcher.icon.Char
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
