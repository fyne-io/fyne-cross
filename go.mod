module github.com/fyne-io/fyne-cross

go 1.14

require (
	github.com/BurntSushi/toml v1.1.0
	github.com/Kodeworks/golang-image-ico v0.0.0-20141118225523-73f0f4cfade9
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/fyne-io/fyne-cross/internal/cloud v0.0.0
	github.com/golang/snappy v0.0.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.14 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/urfave/cli/v2 v2.11.1
	golang.org/x/mod v0.7.0
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f
)

replace github.com/fyne-io/fyne-cross/internal/cloud v0.0.0 => ./internal/cloud
