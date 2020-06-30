package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/tidwall/gjson"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"regexp"
	"strings"
)

type Options struct {
	InvertMatch  bool `short:"v" long:"invert-match" description:""`
	EnableRegexp bool `short:"e" long:"regexp"       description:""`
	EnableGZip   bool `short:"g" long:"gzip"         description:""`
}

func logFatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}c

func logError(err error) {
	fmt.Fprintln(os.Stderr, err)
}

func logDebugf(format string, a ...interface{}) {
	if _, ok := os.LookupEnv("DEBUG"); ok {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format, a...)
	}
}

func out(in string) {
	fmt.Fprintln(os.Stdout, in)
}

func main() {
	opts := Options{}
	parser := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	parser.Name = os.Args[0]
	parser.Usage = "[OPTIONS] PATTERN [PATH]"

	args, err := parser.Parse()
	if err != nil {
		logFatal(err)
	}

	logDebugf("%+v => %+v\n", opts, args)

	m, err := newMatcher(opts.EnableRegexp, args[0])
	if err != nil {
		logFatal(err)
	}

	if terminal.IsTerminal(0) {
		// Not pipe
		if len(args) == 1 {
			parser.WriteHelp(os.Stderr)
			os.Exit(1)
		}

		for _, p := range args[1:] {
			fi, err := os.Stat(p)
			if os.IsNotExist(err) {
				logFatal(fmt.Errorf("file(%s) is not exists", p))
			}
			if fi.Size() == 0 {
				continue
			}

			f, err := os.Open(p)
			if err != nil {
				logFatal(fmt.Errorf("failed to open '%s' : %w", p, err))
			}
			defer f.Close()
			scanAndGrep(gzipFilter(opts.EnableGZip, f), m)
		}
	} else {
		// Pipe
		scanAndGrep(gzipFilter(opts.EnableGZip, os.Stdin), m)
	}
	_ = m
}

func gzipFilter(useGZip bool, reader io.Reader) io.Reader {
	if useGZip {
		gr, err := gzip.NewReader(reader)
		if err != nil {
			logFatal(err)
		}
		return gr
	}
	return reader
}

func newMatcher(isRegExp bool, pattern string) (matcher, error) {
	f, v, err := parsePattern(pattern)
	if err != nil {
		return nil, err
	}

	if isRegExp {
		p, err := regexp.Compile(v)
		if err != nil {
			return nil, fmt.Errorf("failed to compile regexp pattern : %w", err)
		}
		return &regexpMatcher{
			field:   f,
			pattern: p,
		}, nil
	} else {
		return &simpleMatcher{
			field: f,
			value: v,
		}, nil
	}
}

func parsePattern(pattern string) (field, value string, err error) {
	token := strings.Builder{}

	isEscaped := false
	r := strings.NewReader(pattern)
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", "", fmt.Errorf("failed to read : %w", err)
		}

		if isEscaped {
			token.WriteByte(b)
		} else {
			if b == '\\' {
				isEscaped = true
			} else if b == '=' {
				field = token.String()
				token = strings.Builder{}
			} else {
				token.WriteByte(b)
			}
		}
	}

	value = token.String()

	if field == "" || value == "" {
		return "", "", fmt.Errorf("invalid pattern : '%s' [f=%v, v=%v]", pattern, field == "", value == "")
	}

	return
}

func scanAndGrep(r io.Reader, m matcher) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if t := scanner.Text(); m.Match(t) {
			out(t)
		}
	}
}

type matcher interface {
	Match(in string) bool
}

type simpleMatcher struct {
	field string
	value string
}

func (m *simpleMatcher) Match(in string) bool {
	return strings.Contains(gjson.Parse(in).Get(m.field).String(), m.value)
}

type regexpMatcher struct {
	field   string
	pattern *regexp.Regexp
}

func (m *regexpMatcher) Match(in string) bool {
	return m.pattern.MatchString(gjson.Parse(in).Get(m.field).String())
}
