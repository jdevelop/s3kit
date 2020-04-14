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
  cat         cat S3 file(s)
  help        Help about any command
  logs        print S3 Access logs as JSON
  size        calculate size of S3 location

Flags:
  -h, --help          help for s3kit
  -w, --workers int   number of concurrent threads (default 12)

Use "s3kit [command] --help" for more information about a command.
```

### s3kit cat

Often you want to view content of a file on S3, or perhaps *all* of them in a certain path. 
If we're considering data engineering - often these files are represented as compressed CSV files.
There's no easy way to view the file, so basically the `cat` command can identify the file ( by extension ) and uncompress it to `stdout`.

`s3kit cat s3://bucket/path` will print out content of all files under path prefix `s3://bucket/path`

### s3kit logs

Logs command will take the folder/file that contains the access logs for a [static website](https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html) hosted on an S3 bucket.
If logs are configured to be sent to `s3://logs-bucket/accesslog/` - then printing it out is as simple as
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

Getting size of a bucket / path prefix can be achieved with `aws s3 ls --summarize`, however it doesn't have the functionality to collect size of the top-level folders under S3 prefix, which might be useful:

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
