package sortfile

import (
	"path/filepath"
	"strings"
)

var extensionMap map[string]string

func init() {
	extensionMap = make(map[string]string)

	rules := map[string][]string{
		"Documents": {
			"pdf", "docx", "doc", "xlsx", "xls", "pptx", "ppt", "txt", "csv", "odt", "ods", "odp", "rtf", "tex", "abw", "gnumeric", "pages", "numbers", "wps", "wpd", "odf", "gdoc", "gsheet", "gslides",
		},
		"Ebooks_Comics": {
			"epub", "mobi", "azw3", "azw", "fb2", "djvu", "cbz", "cbr", "ibooks",
		},
		"Images": {
			"jpg", "jpeg", "png", "gif", "svg", "webp", "bmp", "ico", "icns", "heic", "heif", "avif", "tiff", "tif", "eps", "raw", "cr2", "cr3", "nef", "arw", "dng", "orf", "rw2", "pef", "raf", "psd", "psb", "ai", "xcf", "kra", "clip", "cdr", "pdn", "sketch", "xd", "afphoto", "afdesign", "dwg", "dxf", "odg", "fig", "svgz", "tga", "exr", "hdr",
		},
		"Videos": {
			"mp4", "mkv", "avi", "mov", "wmv", "flv", "webm", "mpeg", "mpg", "m4v", "3gp", "vob", "ogv", "qt", "ts", "mts", "m2ts", "asf", "rm", "rmvb", "divx", "mxf", "rcv",
		},
		"Audio": {
			"mp3", "wav", "flac", "ogg", "m4a", "aac", "wma", "alac", "opus", "mp2", "aif", "aiff", "ape", "wv", "tta", "dsf", "dff", "mid", "midi", "kar", "sf2", "m3u", "pls", "m3u8",
		},
		"Archives": {
			"zip", "tar", "gz", "xz", "7z", "rar", "bz2", "tgz", "z", "lzma", "lz", "sz", "zst", "cab", "arj", "lzh", "ace", "uue", "sit", "sitx", "pea",
		},
		"Disk_Images": {
			"iso", "dmg", "vcd", "img", "cue", "toast", "nrg", "mdf", "mds",
		},
		"Virtual_Machines": {
			"vmdk", "vdi", "vhd", "vhdx", "ova", "ovf", "qcow2", "pvm",
		},
		"Executables_Installers": {
			"exe", "msi", "bat", "cmd", "com", "scr", "reg", "deb", "rpm", "apk", "sh", "run", "appimage", "flatpak", "snap", "aar", "pkg", "app", "ipa", "mpkg", "gadget", "wsf", "vbs", "bin",
		},
		"Programming": {
			"rs", "go", "cpp", "c", "h", "hpp", "cc", "cxx", "hh", "s", "asm", "html", "htm", "css", "scss", "sass", "js", "jsx", "ts", "tsx", "wasm", "svelte", "vue", "py", "pyw", "java", "class", "jar", "cs", "rb", "pl", "php", "swift", "kt", "kts", "dart", "scala", "clj", "json", "yaml", "yml", "xml", "toml", "ini", "conf", "env", "dockerfile", "lock", "bash", "zsh", "ps1", "fish", "awk", "sed", "md", "rst", "ipynb", "rmd", "lua", "m", "f", "f90", "vb", "r", "vba", "gd",
		},
		"Databases": {
			"sql", "sqlite", "db", "db3", "mdb", "accdb", "psql", "bson", "rdb", "ibd", "frm",
		},
		"Design_3D_CAD": {
			"step", "stp", "igs", "iges", "stl", "obj", "fbx", "gcode", "blend", "skp", "3ds", "c4d", "sldprt", "sldasm", "ipt", "iam", "dwf", "sat", "max", "ma", "mb", "gltf", "glb", "ply", "prt", "asm",
		},
		"Fonts": {
			"ttf", "otf", "woff", "woff2", "eot", "fon", "ttc", "pfa", "pfb",
		},
		"Electronics_Automation": {
			"ino", "hex", "sch", "brd", "kicad_pcb", "kicad_sch", "gbr", "pho", "edf", "dsn", "pdsprj", "circ", "plc", "l5x", "awl", "scl", "ap14", "ap15", "ap16", "ap17", "ap18",
		},
		"Audio_Production": {
			"als", "flp", "rpp", "logic", "cpr", "ptx", "sfz", "nki", "aup", "fxp", "fst", "vst", "vst3",
		},
		"Cryptography_Security": {
			"pem", "crt", "cer", "der", "p12", "pfx", "asc", "gpg", "sig", "key", "keystore", "jks", "kdbx",
		},
		"Subtitles": {
			"srt", "vtt", "ass", "ssa", "sub", "sbv", "mpsub",
		},
		"Logs": {
			"log", "err", "trace", "panic",
		},
		"GIS_Mapping": {
			"shp", "geojson", "kml", "kmz", "gpx", "osm",
		},
	}

	for category, extensions := range rules {
		for _, ext := range extensions {
			extensionMap["."+ext] = category
		}
	}
}

func GetCategory(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))

	if category, exists := extensionMap[ext]; exists {
		return category
	}

	return "Others"
}
