package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage:
  aws-s3 [-access-key-id=<...> -secret-access-key=<...>] -region=<...> -bucket=<...> <command>`)
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
  Each flag has a corresponding environment variable which will be read if set.
  If both environment variable and flag have been set, the flag overrides the
  environment variable.`)
		fmt.Fprintln(os.Stderr, `
  Commands:
    ls       [<prefix>]              Lists bucket objects, optionally filtered by the given prefix
    download <object> [<object>...]  Downloads given object(s) from the bucket
    upload   <file> [<file>...]      Uploads given file(s) to the bucket
    read     <object>                Reads a bucket object to stdout
    write    <object>                Writes a bucket object from stdin
    rm       <object> [<object>...]  Deletes the given object(s) from the bucket`)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, `
AWS Access Keys:
  https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys`)
		fmt.Fprintln(os.Stderr)
	}

	accessKeyID := flag.String("access-key-id", "", "AWS access key ID (ACCESS_KEY_ID)")
	secretAccessKey := flag.String("secret-access-key", "", "AWS secret access key (SECRET_ACCESS_KEY)")
	region := flag.String("region", "", "AWS region (REGION)")
	bucket := flag.String("bucket", "", "bucket name (BUCKET)")

	flag.Parse()

	if *accessKeyID == "" {
		*accessKeyID = os.Getenv("ACCESS_KEY_ID")
	}

	if *secretAccessKey == "" {
		*secretAccessKey = os.Getenv("SECRET_ACCESS_KEY")
	}

	if *region == "" {
		*region = os.Getenv("REGION")
	}

	if *bucket == "" {
		*bucket = os.Getenv("BUCKET")
	}

	var missing []string

	if *region == "" {
		missing = append(missing, "region")
	}

	if *bucket == "" {
		missing = append(missing, "bucket")
	}

	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "missing flag: %s\n\n", strings.Join(missing, ", "))
		flag.Usage()
		os.Exit(1)
	}

	if *accessKeyID != "" || *secretAccessKey != "" {
		switch {
		case *accessKeyID != "" && *secretAccessKey == "":
			fmt.Fprint(os.Stderr, "-access-key-id must be used with -secret-access-key")
			os.Exit(1)
		case *accessKeyID == "" && *secretAccessKey != "":
			fmt.Fprint(os.Stderr, "-secret-access-key must be used with -access-key-id")
			os.Exit(1)
		}
	}

	var opts []func(*S3Client)

	if *accessKeyID != "" && *secretAccessKey != "" {
		opts = append(opts, WithCredentials(*accessKeyID, *secretAccessKey))
	}

	client := NewS3Client(*region, opts...)

	ctx, cancelCtx := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancelCtx()

	switch cmd := flag.Arg(0); cmd {
	case "ls":
		prefix := flag.Arg(1)
		ctx, cancelCtx = context.WithTimeout(ctx, 10*time.Minute)
		defer cancelCtx()
		cmdLs(ctx, client, *bucket, prefix)
	case "download":
		keys := flag.Args()[1:]
		if len(keys) == 0 {
			fmt.Fprint(os.Stderr, "missing object name\n\n")
			flag.Usage()
			os.Exit(1)
		}
		cmdDownload(ctx, client, *bucket, keys)
	case "upload":
		keys := flag.Args()[1:]
		if len(keys) == 0 {
			fmt.Fprint(os.Stderr, "missing file name\n\n")
			flag.Usage()
			os.Exit(1)
		}
		cmdUpload(ctx, client, *bucket, keys)
	case "read":
		key := flag.Arg(1)
		if key == "" {
			fmt.Fprint(os.Stderr, "missing object name\n\n")
			flag.Usage()
			os.Exit(1)
		}
		cmdRead(ctx, client, *bucket, key)
	case "write":
		key := flag.Arg(1)
		if key == "" {
			fmt.Fprint(os.Stderr, "missing object name\n\n")
			flag.Usage()
			os.Exit(1)
		}
		cmdWrite(ctx, client, *bucket, key)
	case "rm":
		keys := flag.Args()[1:]
		if len(keys) == 0 {
			fmt.Fprint(os.Stderr, "missing object name\n\n")
			flag.Usage()
			os.Exit(1)
		}
		cmdRm(ctx, client, *bucket, keys)
	case "":
		fmt.Fprint(os.Stderr, "missing command\n\n")
		flag.Usage()
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		flag.Usage()
		os.Exit(1)
	}
}

func cmdLs(ctx context.Context, client *S3Client, bucket, prefix string) {
	keys, err := client.List(ctx, bucket, prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing objects: %v\n", err)
		os.Exit(1)
	}

	for _, key := range keys {
		fmt.Println(key)
	}
}

func cmdDownload(ctx context.Context, client *S3Client, bucket string, keys []string) {
	var count int
	printDownloaded := func() {
		fmt.Fprintf(os.Stderr, "Downloaded %d object(s)\n", count)
	}
	defer printDownloaded()

	for _, key := range keys {
		if err := func() error {
			r, err := client.Open(ctx, bucket, key)
			if err != nil {
				return fmt.Errorf("error opening object: %v", err)
			}
			defer r.Close()

			f, err := os.Create(key)
			if err != nil {
				return fmt.Errorf("error opening file for writing: %v", err)
			}
			defer f.Close()

			if _, err := io.Copy(f, r); err != nil {
				return fmt.Errorf("error downloading object: %v", err)
			}

			count++

			return nil
		}(); err != nil {
			fmt.Fprintf(os.Stderr, "[%s] %v\n", key, err)
		}
	}
}

func cmdUpload(ctx context.Context, client *S3Client, bucket string, paths []string) {
	var count int
	printUploaded := func() {
		fmt.Fprintf(os.Stderr, "Uploaded %d file(s)\n", count)
	}
	defer printUploaded()

	for _, p := range paths {
		if err := func() error {
			f, err := os.Open(p)
			if err != nil {
				return fmt.Errorf("error opening file: %v", err)
			}
			defer f.Close()

			key := path.Base(p)

			if err := client.Upload(ctx, bucket, key, f); err != nil {
				return fmt.Errorf("error uploading file: %v", err)
			}

			count++

			return nil
		}(); err != nil {
			fmt.Fprintf(os.Stderr, "[%s] %v\n", p, err)
			printUploaded()
			os.Exit(1)
		}
	}
}

func cmdRead(ctx context.Context, client *S3Client, bucket, key string) {
	r, err := client.Open(ctx, bucket, key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] error opening object: %v\n", key, err)
		os.Exit(1)
	}
	defer r.Close()

	if _, err := io.Copy(os.Stdout, r); err != nil {
		fmt.Fprintf(os.Stderr, "[%s] error reading object: %v\n", key, err)
		os.Exit(1)
	}
}

func cmdWrite(ctx context.Context, client *S3Client, bucket, key string) {
	if err := client.Upload(ctx, bucket, key, bufio.NewReader(os.Stdin)); err != nil {
		fmt.Fprintf(os.Stderr, "[%s] error writing object: %v\n", key, err)
		os.Exit(1)
	}
}

func cmdRm(ctx context.Context, client *S3Client, bucket string, keys []string) {
	var count int
	printDeleted := func() {
		fmt.Fprintf(os.Stderr, "Deleted %d object(s)\n", count)
	}
	defer printDeleted()

	for _, key := range keys {
		if err := func() error {
			if err := client.Delete(ctx, bucket, key); err != nil {
				return fmt.Errorf("error deleting object: %v", err)
			}

			count++

			return nil
		}(); err != nil {
			fmt.Fprintf(os.Stderr, "[%s] %v\n", key, err)
		}
	}
}
