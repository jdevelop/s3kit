## Purpose

Simplify getting some basic statistics about content of AWS S3 buckets, that are missing or not very convenient to use in AWS CLI.
Could be quite useful for a fellow data engineer.

## Build

`go install`

## Usage

```
AWS S3 command line toolkit

Usage:
  s3kit [command]

Available Commands:
  cat         Print content of S3 file(s) to stdout
  compliance  Add compliance lock
  govern      Add/remove governance lock
  help        Help about any command
  hold        Add/remove legal hold
  logs        Print S3 Access logs as JSON
  size        Calculate size of S3 location

Flags:
      --debug         print debug messages
  -h, --help          help for s3kit
      --quiet         print warnings and errors
  -w, --workers int   number of concurrent threads (default 12)

Use "s3kit [command] --help" for more information about a command.
```

### s3kit cat

Often you want to view content of a file on S3, or perhaps *all* of them in a certain path. 
If we're considering data engineering - often these files are represented as compressed CSV files.
There's no easy way to view the file, so basically the `cat` command can identify the file ( by extension ) and uncompress it to `stdout`.

```
Print content of S3 file(s) to stdout

Usage:
  s3kit cat s3://bucket/key1 s3://bucket/prefix/ ... [flags]

Flags:
  -h, --help   help for cat

Global Flags:
  -w, --workers int   number of concurrent threads (default 12)
```

#### Example
`s3kit cat s3://bucket/path` will print out content of all files under path prefix `s3://bucket/path`



### s3kit logs

Logs command will take the folder/file that contains the access logs for a [static website](https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html) hosted on an S3 bucket.

```
Print S3 Access logs as JSON

Usage:
  s3kit logs s3://bucket/key1 s3://bucket/prefix/ ... [flags]

Flags:
  -e, --end Date     end date ( YYYY-MM-DD ) (default 2020-04-18 20:00:00 -0400 EDT)
  -h, --help         help for logs
  -s, --start Date   start date ( YYYY-MM-DD ) (default 0001-01-01 00:00:00 +0000 UTC)

Global Flags:
  -w, --workers int   number of concurrent threads (default 12)

```

#### Example
```
s3kit logs s3://logs-bucket/accesslog/
```

It will print JSON objects ( one per line ) as:
```
{
  "time": "2020-04-13T13:27:30Z",
  "remote_ip": "103.249.100.196",
  "operation": "WEBSITE.GET.OBJECT",
  "key": "web/wp-includes/wlwmanifest.xml",
  "request_uri": "GET /web/wp-includes/wlwmanifest.xml HTTP/1.1",
  "http_status": 403,
  "bytes_sent": 303,
  "object_size": 0,
  "flight_time": 59000000,
  "turnaround_time": 0,
  "referer": "-",
  "user_agent": "-"
}
```

With `jq` it is quite easy to get top IP addresses that visited the website:
```
logs s3://logs-bucket/accesslog/ | grep "WEBSITE.GET.OBJECT" | jq ".remote_ip" | sort | uniq -c | sort -n
```

### s3kit size

Getting size of a bucket / path prefix can be achieved with `aws s3 ls --summarize`, however it doesn't have the functionality to collect size of the top-level folders under S3 prefix, which might be useful.

```
Calculate size of S3 location

Usage:
  s3kit size  s3://bucket/key1 s3://bucket/prefix/ ... [flags]

Flags:
  -g, --group   group sizes by top-level folders
  -h, --help    help for size
      --json    output as JSON array
      --raw     raw numbers, no human-formatted size

Global Flags:
  -w, --workers int   number of concurrent threads (default 12)
```

#### Example

```
s3kit size s3://dataeng-data/ -g
+--------------------------------+-------+--------+
|              PATH              | COUNT |  SIZE  |
+--------------------------------+-------+--------+
|  s3://dataeng-data/categories/ |    17 | 62 kB  |
|  s3://dataeng-data/meetups/    |    18 | 15 MB  |
|  s3://dataeng-data/members/    | 96599 | 8.2 GB |
+--------------------------------+-------+--------+
|             TOTAL:             | 96634 | 8.2 GB |
+--------------------------------+-------+--------+
```

It is also possible to get JSON output:
```
[
  {
    "Path": "s3://dataeng-data/categories/",
    "Count": 17,
    "Size": 61684
  },
  {
    "Path": "s3://dataeng-data/meetups/",
    "Count": 18,
    "Size": 15408356
  },
  {
    "Path": "s3://dataeng-data/members/",
    "Count": 96599,
    "Size": 8202309787
  }
]
```

## s3kit compliance
Adds the [compliance lock](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock.html) to a given object identified by a prefix and applicable to all versions of the object(s), latest version of the object(s) or specific version of the object(s).

```
Add compliance lock

Usage:
  s3kit compliance s3://bucket/key1 s3://bucket/prefix/ ... [flags]

Flags:
      --all               Apply to all versions of object(s)
      --expire duration   compliance lock duration (1m, 1h etc)
  -h, --help              help for compliance
      --latest            Apply to latest version of object(s) (default true)
      --version string    Apply to a specific version

Global Flags:
  -w, --workers int   number of concurrent threads (default 12)
```

This operation will explicitly ask for confirmation prior to applying to the object version, because there's no way to revert the compliance lock:

```
compliance s3://demolocal123/go.mod --expire 10s
Locking s3://bucket/key version .oNOa6ZVNDTUcKHaaaECJUmdA9XaBTI4 expires 2020-04-18 15:31:35, proceed? (y/N):
```

The default answer is **NO**.

## s3kit hold add / rm

Adds or removes the [legal hold](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock.html) to the object(s) found by given S3 prefix(es).

```
Add/remove legal hold for given object(s)

Usage:
  s3kit hold add s3://bucket/key1 s3://bucket/prefix/ ... [flags]

Flags:
  -h, --help   help for add

Global Flags:
      --all              Apply to all versions of object(s)
      --latest           Apply to latest version of object(s) (default true)
      --version string   Apply to a specific version
  -w, --workers int      number of concurrent threads (default 12)
```

## s3kit govern add / rm

Adds or removes the [governance retention lock](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock.html) to the object(s) found by given S3 prefix(es).

```
Add/remove governance lock for given object(s)

Usage:
  s3kit govern add s3://bucket/key1 s3://bucket/prefix/ ... [flags]

Flags:
      --expire duration   governance lock duration (1m, 1h etc)
  -h, --help              help for add

Global Flags:
      --all              Apply to all versions of object(s)
      --latest           Apply to latest version of object(s) (default true)
      --version string   Apply to a specific version
  -w, --workers int      number of concurrent threads (default 12)
```
