goverview
=========
goverview - Get overview about list of URLs

## Installation

```
go get -u github.com/j3ssie/goverview
```

## Example Commands

```
goverview - Overview about list of URLs - beta v0.1.4 by @j3ssiejjj

Usage:
  goverview [flags]

Flags:
  -c, --concurrency int      Set the concurrency level (default 20)
  -t, --threads int          Set the threads level for do screenshot (default 10)
  -l, --level int            Set level to calculate CheckSum
  -j, --json                 Output as JSON
  -N, --no-output            No output
  -o, --output string        Output Directory (default "out")
  -S, --screenshot string    Summary File for Screenshot (default 'out/screenshot-summary.txt')
  -C, --content string       Summary File for Content (default 'out/content-summary.txt')
  -W, --wordlist string      Wordlists File build from HTTP Content (default 'out/wordlists.txt')
  -B, --burp                 Receive input as base64 burp request
      --sortTag              Sort HTML tag before do checksum
      --skip-words           Skip wordlist builder
  -Q, --skip-screen          Skip screenshot
      --skip-probe           Skip probing for checksum
  -M, --save-response        Save HTTP response
  -L, --redirect             Allow redirect
  -R, --save-redirect        Save redirect URL to overview file too
      --timeout int          HTTP timeout (default 15)
      --retry int            Number of retry
  -H, --headers strings      Custom headers (e.g: -H 'Referer: {{.BaseURL}}') (Multiple -H flags are accepted)
      --a                    Use Absolute path in summary
      --screen-timeout int   screenshot timeout (default 40)
      --height int           Height screenshot
      --width int            Width screenshot
  -v, --verbose              Verbose output
      --debug                Debug output
  -V, --version              Print version
  -h, --help                 help for goverview


Checksum Content Level:
  0 - Only check for src in <script> tag
  1 - Check for all structure of HTML tag + src in <script> tag
  2 - Check for all structure of HTML tag + src in <script> <img> <a> tag
  5 - Entire HTTP response

Examples:
  # Only get summary
  cat list_of_urls.txt | goverview -N -Q -c 50 | tee only-overview.txt

  # Get summary content and store raw response without screenshot
  cat list_of_urls.txt | goverview -v -Q -M -l 2

  # Only do screenshot
  cat list_of_urls.txt | goverview --skip-probe

  # Do full probing and screnshot
  cat list_of_urls.txt | goverview
```

## License

`goverview` is made with ♥  by [@j3ssiejjj](https://twitter.com/j3ssiejjj) and it is released under the MIT license.
