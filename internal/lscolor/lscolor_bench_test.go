package lscolor

import (
	"testing"

	"github.com/nolight132/nls/internal/listing"
)

// benchLSColors mirrors a typical GNU dircolors database (~70 wildcard rules).
const benchLSColors = "rs=0:di=01;34:ln=01;36:mh=00:pi=40;33:so=01;35:do=01;35:" +
	"bd=40;33;01:cd=40;33;01:or=40;31;01:mi=00:su=37;41:sg=30;43:ca=00:" +
	"tw=30;42:ow=34;42:st=37;44:ex=01;32:" +
	"*.tar=01;31:*.tgz=01;31:*.arc=01;31:*.arj=01;31:*.taz=01;31:*.lha=01;31:" +
	"*.lz4=01;31:*.lzh=01;31:*.lzma=01;31:*.tlz=01;31:*.txz=01;31:*.tzo=01;31:" +
	"*.t7z=01;31:*.zip=01;31:*.z=01;31:*.dz=01;31:*.gz=01;31:*.lrz=01;31:" +
	"*.lz=01;31:*.lzo=01;31:*.xz=01;31:*.zst=01;31:*.tzst=01;31:*.bz2=01;31:" +
	"*.bz=01;31:*.tbz=01;31:*.tbz2=01;31:*.tz=01;31:*.deb=01;31:*.rpm=01;31:" +
	"*.jar=01;31:*.war=01;31:*.ear=01;31:*.sar=01;31:*.rar=01;31:*.alz=01;31:" +
	"*.ace=01;31:*.zoo=01;31:*.cpio=01;31:*.7z=01;31:*.rz=01;31:*.cab=01;31:" +
	"*.jpg=01;35:*.jpeg=01;35:*.mjpg=01;35:*.mjpeg=01;35:*.gif=01;35:*.bmp=01;35:" +
	"*.pbm=01;35:*.pgm=01;35:*.ppm=01;35:*.tga=01;35:*.xbm=01;35:*.xpm=01;35:" +
	"*.tif=01;35:*.tiff=01;35:*.png=01;35:*.svg=01;35:*.svgz=01;35:*.mng=01;35:" +
	"*.pcx=01;35:*.mov=01;35:*.mpg=01;35:*.mpeg=01;35:*.m2v=01;35:*.mkv=01;35:" +
	"*.webm=01;35:*.ogm=01;35:*.mp4=01;35:*.m4v=01;35:*.mp4v=01;35:*.vob=01;35:" +
	"*.aac=00;36:*.au=00;36:*.flac=00;36:*.m4a=00;36:*.mid=00;36:*.midi=00;36:" +
	"*.mka=00;36:*.mp3=00;36:*.mpc=00;36:*.ogg=00;36:*.ra=00;36:*.wav=00;36"

var benchColorizeNames = []string{
	"archive.tar", "photo.png", "song.mp3", "video.mkv", "backup.zip",
	"main.go", "notes.txt", "README.md", "noextension", "deep.file.name.gz",
}

func BenchmarkStylerNew(b *testing.B) {
	b.Setenv("LS_COLORS", benchLSColors)
	for b.Loop() {
		New()
	}
}

func BenchmarkColorize(b *testing.B) {
	b.Setenv("LS_COLORS", benchLSColors)
	s := New()
	for b.Loop() {
		for _, name := range benchColorizeNames {
			s.Colorize(name, listing.KindFile)
		}
	}
}
