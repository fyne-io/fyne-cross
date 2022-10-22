package cloud

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/klauspost/compress/zstd"
	"golang.org/x/sync/errgroup"

	archiver "github.com/mholt/archiver/v3"
)

type AWSSession struct {
	s      *session.Session
	bucket string

	m      sync.Mutex
	cancel context.CancelFunc
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return !errors.Is(err, os.ErrNotExist)
	}
	return true
}

func NewAWSSessionFromEnvironment() (*AWSSession, error) {
	return NewAWSSession("", "", os.Getenv("AWS_S3_ENDPOINT"), os.Getenv("AWS_S3_REGION"), os.Getenv("AWS_S3_BUCKET"))
}

func NewAWSSession(akid string, secret string, endpoint string, region string, bucket string) (*AWSSession, error) {
	var cred *credentials.Credentials

	if len(bucket) == 0 {
		return nil, fmt.Errorf("no bucket specified")
	}

	if akid != "" && secret != "" {
		cred = credentials.NewStaticCredentials(akid, secret, "")
	}

	s, err := session.NewSession(
		&aws.Config{
			Endpoint:    aws.String(endpoint),
			Region:      aws.String(region),
			Credentials: cred,
		},
	)
	if err != nil {
		return nil, err
	}

	return &AWSSession{s: s, bucket: bucket, cancel: func() {}}, nil
}

func (a *AWSSession) GetCredentials() (credentials.Value, error) {
	a.m.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.m.Unlock()
	defer a.Cancel()

	return a.s.Config.Credentials.GetWithContext(ctx)
}

func (a *AWSSession) UploadFile(localFile string, s3FilePath string) error {
	file, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer file.Close()

	uploader := s3manager.NewUploader(a.s)

	a.m.Lock()
	ctxt, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.m.Unlock()
	defer a.Cancel()

	_, err = uploader.UploadWithContext(ctxt, &s3manager.UploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(s3FilePath),

		Body: file,
	})

	return err
}

func (a *AWSSession) UploadCompressedDirectory(localDirectoy string, s3FilePath string) error {
	file, err := ioutil.TempFile("", "fyne-cross-s3")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	extension := strings.ToLower(filepath.Ext(s3FilePath))

	var compression archiver.Writer
	var eg errgroup.Group
	var closer io.Closer

	switch extension {
	case ".xz":

		compression = archiver.NewTarXz()
		err := compression.Create(file)
		if err != nil {
			return err
		}
	case ".zstd":
		inZstd, outTar := io.Pipe()
		closer = outTar

		compression = archiver.NewTar()
		err := compression.Create(outTar)
		if err != nil {
			return err
		}

		enc, err := zstd.NewWriter(file)
		if err != nil {
			return err
		}

		eg.Go(func() error {
			_, err := io.Copy(enc, inZstd)
			if err != nil {
				return err
			}

			inZstd.Close()
			enc.Close()
			return nil
		})
	default:
		return fmt.Errorf("unknown extension for %v", s3FilePath)
	}

	base := filepath.Base(localDirectoy)

	err = filepath.Walk(localDirectoy, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		customName := strings.TrimPrefix(path, localDirectoy)
		if customName == path {
			return fmt.Errorf("unexpected path: `%v` triming `%v`", path, localDirectoy)
		}
		customName = filepath.ToSlash(customName)
		if len(customName) == 0 || customName[0] != '/' {
			customName = "/" + customName
		}
		customName = base + customName

		if info.IsDir() {
			return compression.Write(archiver.File{
				FileInfo: archiver.FileInfo{
					FileInfo:   info,
					CustomName: customName,
				},
			})
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		return compression.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   info,
				CustomName: customName,
			},
			ReadCloser: f,
		})
	})
	if err != nil {
		return err
	}

	compression.Close()
	if closer != nil {
		closer.Close()
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	uploader := s3manager.NewUploader(a.s)

	a.m.Lock()
	ctxt, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.m.Unlock()
	defer a.Cancel()

	f, err := os.Open(file.Name())
	if err != nil {
		return err
	}

	_, err = uploader.UploadWithContext(ctxt, &s3manager.UploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(s3FilePath),

		Body: f,
	})
	if err != nil {
		return err
	}
	f.Close()

	return err
}

func (a *AWSSession) DownloadFile(s3FilePath string, localFile string) error {
	f, err := os.Create(localFile)
	if err != nil {
		return err
	}

	downloader := s3manager.NewDownloader(a.s)

	a.m.Lock()
	ctxt, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.m.Unlock()
	defer a.Cancel()

	_, err = downloader.DownloadWithContext(ctxt, f, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(s3FilePath),
	})

	return err
}

func (a *AWSSession) DownloadCompressedDirectory(s3FilePath string, localRootDirectory string) error {
	file, err := ioutil.TempFile("", "fyne-cross-s3")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	a.m.Lock()
	ctxt, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.m.Unlock()
	defer a.Cancel()

	downloader := s3manager.NewDownloader(a.s)
	downloader.Concurrency = 1

	_, err = downloader.DownloadWithContext(ctxt, fakeWriterAt{file}, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(s3FilePath),
	})
	file.Close()
	if err != nil {
		return err
	}

	in, err := os.Open(file.Name())
	if err != nil {
		return err
	}

	extension := strings.ToLower(filepath.Ext(s3FilePath))

	var compression archiver.Reader
	var eg errgroup.Group

	switch extension {
	case ".xz":
		compression = archiver.NewTarXz()
		err := compression.Open(in, 0)
		if err != nil {
			return err
		}
	case ".zstd":
		inTar, outZstd := io.Pipe()

		dec, err := zstd.NewReader(in)
		if err != nil {
			return err
		}
		defer dec.Close()

		eg.Go(func() error {
			// Copy content...
			_, err := io.Copy(outZstd, dec)
			return err
		})

		compression = archiver.NewTar()
		err = compression.Open(inTar, 0)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown extension for %v", s3FilePath)
	}

	for {
		f, err := compression.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = uncompressFile(localRootDirectory, f)
		if err != nil {
			return err
		}
	}

	in.Close()
	return eg.Wait()

}

func (a *AWSSession) Cancel() {
	a.m.Lock()
	defer a.m.Unlock()

	a.cancel()
	a.cancel = func() {}
}

func uncompressFile(localRootDirectory string, f archiver.File) error {
	// be sure to close f before moving on!!
	defer f.Close()

	header := f.Header.(*tar.Header)

	// Do not use strings.Split to split a path as it will generate empty string when given "//"
	splitFn := func(c rune) bool {
		return c == '/'
	}
	paths := strings.FieldsFunc(header.Name, splitFn)
	if len(paths) == 0 {
		if f.Name() != "/" {
			return fmt.Errorf("incorrect path")
		}
		paths = append(paths, "/")
	}

	// Replace top directory in the archive with local path
	paths[0] = localRootDirectory
	localFile := filepath.Join(paths...)
	if f.IsDir() {
		if !Exists(localFile) {
			log.Println("Creating directory:", localFile)
			return os.Mkdir(localFile, f.Mode().Perm())
		}
		return nil
	}

	outFile, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	log.Println(header.Name, "->", localFile)
	_, err = io.Copy(outFile, f)

	return err
}

func (a *AWSSession) GetBucket() string {
	return a.bucket
}

type fakeWriterAt struct {
	w io.Writer
}

func (fw fakeWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	// ignore 'offset' because we forced sequential downloads
	return fw.w.Write(p)
}
