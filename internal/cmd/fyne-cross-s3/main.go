package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fyne-io/fyne-cross/internal/cloud"
	"github.com/urfave/cli/v2"
)

func main() {
	var endpoint string
	var region string
	var bucket string
	var akid string
	var secret string

	app := &cli.App{
		Name:  "fyne-cross-s3",
		Usage: "Upload and download to S3 bucket specified in the environment",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "aws-endpoint",
				Aliases:     []string{"e"},
				Usage:       "AWS endpoint to connect to (can be used to connect to non AWS S3 services)",
				EnvVars:     []string{"AWS_S3_ENDPOINT"},
				Destination: &endpoint,
			},
			&cli.StringFlag{
				Name:        "aws-region",
				Aliases:     []string{"r"},
				Usage:       "AWS region to connect to",
				EnvVars:     []string{"AWS_S3_REGION"},
				Destination: &region,
			},
			&cli.StringFlag{
				Name:        "aws-bucket",
				Aliases:     []string{"b"},
				Usage:       "AWS bucket to store data into",
				EnvVars:     []string{"AWS_S3_BUCKET"},
				Destination: &bucket,
			},
			&cli.StringFlag{
				Name:        "aws-secret",
				Aliases:     []string{"s"},
				Usage:       "AWS secret to use to establish S3 connection",
				Destination: &secret,
			},
			&cli.StringFlag{
				Name:        "aws-AKID",
				Aliases:     []string{"a"},
				Usage:       "AWS Access Key ID to use to establish S3 connection",
				Destination: &akid,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "upload-directory",
				Usage: "Upload specified directory as an archive to the specified destination in S3 bucket",
				Action: func(c *cli.Context) error {
					if c.Args().Len() != 2 {
						return fmt.Errorf("directory to archive and destination should be specified")
					}

					log.Println("Connecting to AWS")
					aws, err := cloud.NewAWSSession(akid, secret, endpoint, region, bucket)
					if err != nil {
						return err
					}

					log.Println("Uploading directory", c.Args().Get(0), "to", c.Args().Get(1))
					err = aws.UploadCompressedDirectory(c.Args().Get(0), c.Args().Get(1))
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "upload-file",
				Usage: "Upload specified file to S3 bucket",
				Action: func(c *cli.Context) error {
					if c.Args().Len() != 2 {
						return fmt.Errorf("file to upload and destination should be specified")
					}

					log.Println("Connecting to AWS")
					aws, err := cloud.NewAWSSession(akid, secret, endpoint, region, bucket)
					if err != nil {
						return err
					}

					log.Println("Uploading file", c.Args().Get(0), "to", c.Args().Get(1))
					err = aws.UploadFile(c.Args().Get(0), c.Args().Get(1))
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "download-directory",
				Usage: "Download archive from specified S3 bucket to be expanded in a specified directory",
				Action: func(c *cli.Context) error {
					if c.Args().Len() != 2 {
						return fmt.Errorf("archive to download and destination should be specified")
					}

					log.Println("Connecting to AWS")
					aws, err := cloud.NewAWSSession(akid, secret, endpoint, region, bucket)
					if err != nil {
						return err
					}

					log.Println("Download", c.Args().Get(0), "to directory", c.Args().Get(1))
					err = aws.DownloadCompressedDirectory(c.Args().Get(0), c.Args().Get(1))
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "download-file",
				Usage: "Download specified file from S3 bucket to be deposited at specified local destination",
				Action: func(c *cli.Context) error {
					if c.Args().Len() != 2 {
						return fmt.Errorf("file to upload and destination should be specified")
					}

					log.Println("Connecting to AWS")
					aws, err := cloud.NewAWSSession(akid, secret, endpoint, region, bucket)
					if err != nil {
						return err
					}

					log.Println("Downloading file from", c.Args().Get(0), "to", c.Args().Get(1))
					err = aws.DownloadFile(c.Args().Get(0), c.Args().Get(1))
					if err != nil {
						return err
					}

					return nil
				},
			},
		}}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
