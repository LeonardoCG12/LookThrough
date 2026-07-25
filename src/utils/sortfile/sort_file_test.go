package sortfile

import "testing"

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     string
	}{
		{name: "document", fileName: "report.PDF", want: "Documents"},
		{name: "image", fileName: "photo.JpEg", want: "Images"},
		{name: "video", fileName: "movie.mkv", want: "Videos"},
		{name: "programming TypeScript", fileName: "main.ts", want: "Programming"},
		{name: "programming assembly", fileName: "boot.asm", want: "Programming"},
		{name: "compound archive extension", fileName: "backup.tar.gz", want: "Archives"},
		{name: "dot env file", fileName: ".env", want: "Programming"},
		{name: "unknown extension", fileName: "data.unknown", want: "Others"},
		{name: "no extension", fileName: "README", want: "Others"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := GetCategory(test.fileName); got != test.want {
				t.Fatalf("GetCategory(%q) = %q; want %q", test.fileName, got, test.want)
			}
		})
	}
}
