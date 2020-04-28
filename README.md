goverview
=========
goverview - Get overview about list of URLs

## Installation

```
go get -u github.com/j3ssie/goverview
```

## Example Commands

```
goverview - Get overview about list of URLs - beta v0.1 by @j3ssiejjj

Usage:
cat <input_file> | goverview [options]

Flags:
  -c, --concurrency int     Set the concurrency level (default 30)
  -t, --threads int         Set the threads level for do screenshot (default 10)
  -l, --level int           Set level to calculate CheckSum
  -o, --output string       Output Directory (default "out")
  -S, --screenshot string   Summary File for Screenshot (default 'out/content-summary.txt')
  -C, --content string      Summary File for Content (default 'out/screenshot-summary.txt')
  -Q, --skip-screen         Skip screenshot
      --skip-probe          Skip probing for checksum
  -M, --save-response       Save HTTP response
      --a                   Use Absolute path in summary
  -R, --redirect            Allow redirect
      --timeout int         screenshot timeout (default 10)
      --retry int           Number of retry
      --height int          Height screenshot
      --width int           Width screenshot
  -v, --verbose             Verbose output
      --debug               Debug output
  -V, --version             Check version
  -h, --help                help for goverview

Checksum Content Level
  0 - Only check for src in <script> tag
  1 - Check for all structure of HTML tag + src in <script> tag
  2 - Check for all structure of HTML tag + src in <script> <img> <a> tag
  5 - Entire HTTP response

Examples:
  cat list_of_urls.txt | goverview -l 1
  cat list_of_urls.txt | goverview -v -Q -l 1
  cat list_of_urls.txt | goverview -v -Q -M -l 2
```

## License

`goverview` is made with ♥  by [@j3ssiejjj](https://twitter.com/j3ssiejjj) and it is released under the MIT license.
