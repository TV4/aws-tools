# aws-s3

[![Build Status](https://github.com/TV4/aws-tools/workflows/build/badge.svg)](https://github.com/TV4/aws-tools/actions?query=workflow%3Abuild)
[![Go Report Card](https://goreportcard.com/badge/github.com/TV4/aws-tools)](https://goreportcard.com/report/github.com/TV4/aws-tools)
[![License MIT](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat)](https://github.com/TV4/aws-tools#license)

`aws-s3` reads and writes from/to
[Simple Storage Service](https://aws.amazon.com/s3/).

## Installation
```
go install github.com/TV4/aws-tools/cmd/aws-s3@latest
```

## Usage
```
aws-s3 [-access-key-id=<...> -secret-access-key=<...>] -region=<...> -bucket=<...> <command>

  -access-key-id string
        AWS access key ID (ACCESS_KEY_ID)
  -bucket string
        bucket name (BUCKET)
  -region string
        AWS region (REGION)
  -secret-access-key string
        AWS secret access key (SECRET_ACCESS_KEY)

  Each flag has a corresponding environment variable which will be read if set.
  If both environment variable and flag have been set, the flag overrides the
  environment variable.

Commands:
  ls       [<prefix>]              Lists bucket objects, optionally filtered by the given prefix
  download <object> [<object>...]  Downloads given object(s) from the bucket
  upload   <file> [<file>...]      Uploads given file(s) to the bucket
  read     <object>                Reads a bucket object to stdout
  write    <object>                Writes a bucket object from stdin
  rm       <object> [<object>...]  Deletes the given object(s) from the bucket


AWS Access Keys:
  https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys
```

## License
Copyright (c) 2020 TV4

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
