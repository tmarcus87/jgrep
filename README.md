jgrep
=====

jgrep means json-grep

## Usage

```bash
$ zcat gzip/foo.gz
{ "field": { "subField": "foo" } }
$ zcat gzip/ba.gz
{ "field": { "subField": "bar" } }
{ "field": { "subField": "bazz" } }

# Command
$ jgrep -g -e field.subField=ba.* gzip/*.gz
{ "field": { "subField": "bar" } }
{ "field": { "subField": "bazz" } } 

# Pipe
$ cat gzip/*.gz | jgrep -g -e field.subField=ba.*
{ "field": { "subField": "bar" } }
{ "field": { "subField": "bazz" } }
```

## Options
```bash
$ jgrep -h
Usage:
  jgrep [OPTIONS] PATTERN [PATH]

Application Options:
  -v, --invert-match
  -e, --regexp
  -g, --gzip

Help Options:
  -h, --help          Show this help message
```