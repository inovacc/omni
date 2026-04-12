# omni

Cross-platform Go-native shell utilities (100+ commands, no exec, pure Go)

## Archive

### bunzip2
Decompress bzip2 files
Flags: `-f/--force` `-k/--keep` `-c/--stdout` `-v/--verbose`

### bzcat
Decompress and print bzip2 files

### bzip2
Decompress bzip2 files
Flags: `-d/--decompress` `-f/--force` `-k/--keep` `-c/--stdout` `-v/--verbose`

### gunzip
Decompress gzip files
Flags: `-f/--force` `-k/--keep` `-c/--stdout` `-v/--verbose`

### gzip
Compress or decompress files
Flags: `-9/--best` `-d/--decompress` `-1/--fast` `-f/--force` `-k/--keep` `-c/--stdout` `-v/--verbose`

### tar
Create, extract, or list archive files
Flags: `-c/--create` `-C/--directory` `-x/--extract` `-f/--file` `-z/--gzip` `--json` `-t/--list` `--strip-components` `-v/--verbose`

### unxz
Decompress xz files
Flags: `-f/--force` `-k/--keep` `-c/--stdout` `-v/--verbose`

### unzip
Extract files from a zip archive
Flags: `-d/--directory` `--json` `-l/--list` `--strip-components` `-v/--verbose`

### xz
Compress or decompress xz files
Flags: `-d/--decompress` `-f/--force` `-k/--keep` `-l/--list` `-c/--stdout` `-v/--verbose`

### xzcat
Decompress and print xz files

### zcat
Decompress and print gzip files

### zip
Package and compress files into a zip archive
Flags: `-C/--directory` `-r/--recursive` `-v/--verbose`

## CodeGen

### scaffold
Code scaffolding utilities
Sub: `cobra` `handler` `mcp` `repository` `test`

## Core

### basename
Strip directory and suffix from file names
Flags: `-s/--suffix`

### cat
Concatenate files and print on the standard output
Flags: `-e/--e` `--json` `-n/--number` `-b/--number-nonblank` `-A/--show-all` `-E/--show-ends` `-v/--show-nonprinting` `-T/--show-tabs` `-s/--squeeze-blank` `-t/--t`

### date
Print the current date and time
Flags: `--iso-8601` `-u/--utc`

### dirname
Strip last component from file name

### echo
Display a line of text
Flags: `-e/--escape` `-E/--no-escape` `-n/--no-newline`

### ls
List directory contents
Flags: `-a/--all` `-A/--almost-all` `-F/--classify` `-d/--directory` `-H/--human-readable` `-i/--inode` `-l/--long` `-U/--no-sort` `-1/--one` `-R/--recursive` `-r/--reverse` `-S/--size` `-t/--time`

### pwd
Print working directory

### readlink
Print resolved symbolic links or canonical file names
Flags: `-f/--canonicalize` `-e/--canonicalize-existing` `-m/--canonicalize-missing` `-n/--no-newline` `-q/--quiet` `-s/--silent` `-v/--verbose` `-z/--zero`

### realpath
Print the resolved path

### tree
Display directory tree structure
Flags: `-a/--all` `--compare` `--date` `-d/--depth` `--detect-moves` `--dirs-only` `--hash` `-i/--ignore` `-j/--json` `--json-stream` `-L/--level` `--max-files` `--max-hash-size` `--no-color` `--no-dir-slash` `--size` `-s/--stats` `-t/--threads`

### yes
Output a string repeatedly until killed

## Data

### dotenv
Load environment variables from .env files
Flags: `-x/--expand` `-e/--export` `-q/--quiet` `-s/--shell`

### jq
Command-line JSON processor
Flags: `-c/--compact-output` `-n/--null-input` `-r/--raw-output` `-s/--slurp` `-S/--sort-keys` `--tab`

### json
JSON utilities (format, minify, validate)
Sub: `fmt` `fromcsv` `fromtoml` `fromxml` `fromyaml` `keys` `minify` `stats` `tocsv` `tostruct` `toxml` `toyaml` `validate`

### yq
Command-line YAML processor
Flags: `-c/--compact-output` `-I/--indent` `-n/--null-input` `-o/--output-format` `-r/--raw-output`

## Database

### bbolt
BoltDB database management
Sub: `buckets` `check` `compact` `create-bucket` `delete` `delete-bucket` `dump` `get` `info` `keys` `page` `pages` `put` `stats`

### sqlite
SQLite database management
Sub: `check` `columns` `dump` `import` `indexes` `query` `schema` `stats` `tables` `vacuum`

## File

### chmod
Change file mode bits
Flags: `-c/--changes` `-R/--recursive` `--reference` `-f/--silent` `-v/--verbose`

### chown
Change file owner and group
Flags: `-c/--changes` `-h/--no-dereference` `--preserve-root` `-R/--recursive` `--reference` `-f/--silent` `-v/--verbose`

### cp
Copy files and directories

### dd
Convert and copy a file

### file
Determine file type
Flags: `-b/--brief` `-i/--mime` `-L/--no-dereference` `-F/--separator`

### find
Search for files in a directory hierarchy
Flags: `--amin` `--atime` `--empty` `--executable` `--iname` `--ipath` `--iregex` `--maxdepth` `--mindepth` `--mmin` `--mtime` `--name` `--not` `--path` `-0/--print0` `--readable` `--regex` `--size` `--type` `--writable`

### ln
Make links between files
Flags: `-b/--backup` `-f/--force` `-n/--no-dereference` `-r/--relative` `-s/--symbolic` `-v/--verbose`

### mkdir
Create directories
Flags: `-p/--parents`

### mv
Move (rename) files

### rm
Remove files or directories
Flags: `-f/--force` `--no-preserve-root` `-r/--recursive`

### rmdir
Remove empty directories
Flags: `--no-preserve-root`

### stat
Display file or file system status

### touch
Update the access and modification times of each FILE to the current time

## Hash/Encoding

### base32
Base32 encode or decode data
Flags: `-d/--decode` `-w/--wrap`

### base58
Base58 encode or decode data (Bitcoin alphabet)
Flags: `-d/--decode`

### base64
Base64 encode or decode data
Flags: `-d/--decode` `-i/--ignore-garbage` `-w/--wrap`

### hash
Compute and check file hashes
Flags: `-a/--algorithm` `-b/--binary` `-c/--check` `--quiet` `-r/--recursive` `--status` `-w/--warn`

### md5sum
Compute and check MD5 message digest
Flags: `-b/--binary` `-c/--check` `--quiet` `--status` `-w/--warn`

### sha256sum
Compute and check SHA256 message digest
Flags: `-b/--binary` `-c/--check` `--quiet` `--status` `-w/--warn`

### sha512sum
Compute and check SHA512 message digest
Flags: `-b/--binary` `-c/--check` `--quiet` `--status` `-w/--warn`

## other

### aws
AWS CLI operations
Sub: `ec2` `iam` `s3` `ssm` `sts`

### banner
Generate ASCII art text banners
Flags: `-f/--font` `-l/--list` `-w/--width`

### brdoc
Brazilian document utilities (CPF, CNPJ)
Sub: `cnpj` `cpf`

### buf
Protocol buffer utilities (lint, format, compile, generate)
Sub: `breaking` `compile` `format` `generate` `lint` `ls-files` `mod`

### case
Text case conversion utilities
Sub: `all` `camel` `constant` `detect` `dot` `kebab` `lower` `pascal` `path` `sentence` `snake` `swap` `title` `toggle` `upper`

### cloud
Cloud profile management
Sub: `profile`

### copy
Alias for cp

### crc32sum
Compute and check CRC32 checksums
Flags: `-b/--binary` `-c/--check` `--quiet` `--status` `-w/--warn`

### crc64sum
Compute and check CRC64 checksums
Flags: `-b/--binary` `-c/--check` `--quiet` `--status` `-w/--warn`

### cron
Parse and explain cron expressions
Flags: `--next` `--validate`

### css
CSS utilities (format, minify, validate)
Flags: `-i/--indent` `--sort-props` `--sort-rules`
Sub: `fmt` `minify` `validate`

### csv
CSV utilities (convert to/from JSON)
Sub: `fromjson` `tojson`

### curl
HTTP client with httpie-like syntax
Flags: `-d/--data` `-f/--form` `-H/--header` `-k/--insecure` `--json` `-L/--location` `-t/--timeout` `-v/--verbose`

### exec
Run external commands with credential pre-flight checks
Flags: `--dry-run` `-f/--force` `--no-prompt` `--strict`

### exist
Check if files, directories, commands, env vars, processes, or ports exist
Sub: `command` `dir` `env` `file` `path` `port` `process`

### for
Loop and execute commands
Sub: `each` `glob` `lines` `range` `split`

### gbc
Git branch clean (alias)
Flags: `--dry-run`

### gh
GitHub CLI shortcuts
Sub: `actions-rerun` `issue-mine` `pr-approve` `pr-checkout` `pr-diff` `repo-clone-org`

### git
Git shortcuts and hacks
Sub: `amend` `blame-line` `branch-clean` `diff-words` `fetch-all` `log-graph` `pull-rebase` `push` `quick-commit` `stash-staged` `status` `undo`

### gops
Display Go process information
Flags: `-j/--json`

### gqc
Git quick commit (alias)
Flags: `-a/--all` `-m/--message`

### hex
Hexadecimal encoding and decoding utilities
Sub: `decode` `encode`

### html
HTML utilities (format, encode, decode)
Sub: `decode` `encode` `fmt` `minify` `validate`

### jwt
JWT (JSON Web Token) utilities
Sub: `decode`

### kconfig
Show kubeconfig info

### kcs
Switch kubectl context

### kdebug
Debug pod with ephemeral container
Flags: `--image` `-n/--namespace`

### kdp
Delete pods by selector
Flags: `--force` `-n/--namespace`

### kdrain
Drain node for maintenance
Flags: `--delete-emptydir` `--ignore-daemonsets`

### keb
Exec into pod with bash
Flags: `-c/--container` `-n/--namespace`

### kga
Get all resources in namespace
Flags: `-A/--all-namespaces` `-n/--namespace`

### kge
Get events sorted by time
Flags: `-A/--all-namespaces` `-n/--namespace`

### klf
Follow pod logs with timestamps
Flags: `-c/--container` `-n/--namespace` `--tail`

### kns
Switch default namespace

### kpf
Quick port forward
Flags: `-n/--namespace`

### krr
Rollout restart deployment
Flags: `-n/--namespace`

### krun
Run a one-off pod
Flags: `--image` `-n/--namespace`

### kscale
Scale deployment
Flags: `-n/--namespace`

### ksuid
Generate K-Sortable Unique IDentifiers
Flags: `-n/--count`

### ktn
Top nodes by resource usage
Flags: `--sort-by`

### ktp
Top pods by resource usage
Flags: `-A/--all-namespaces` `-n/--namespace` `--sort-by`

### kubectl
Kubernetes CLI

### kwp
Watch pods continuously
Flags: `-A/--all-namespaces` `-n/--namespace` `-l/--selector`

### less
View file contents with scrolling
Flags: `-N/--LINE-NUMBERS` `-S/--chop-long-lines` `-i/--ignore-case` `-X/--no-init` `-F/--quit-if-one-screen` `-R/--raw-control-chars`

### loc
Count lines of code by language
Flags: `-e/--exclude` `--hidden`

### lsof
List open files and network connections
Flags: `-c/--command` `-e/--established` `-4/--ipv4` `-6/--ipv6` `-l/--listen` `-i/--network` `-n/--no-headers` `-p/--pid` `--port` `-t/--tcp` `-U/--udp` `-u/--user`

### more
View file contents page by page
Flags: `-n/--line-numbers`

### move
Alias for mv

### nanoid
Generate compact, URL-safe unique IDs
Flags: `-a/--alphabet` `-n/--count` `-l/--length`

### netstat
Display network connections (alias for ss)
Flags: `-a/--all` `-l/--listening` `-n/--numeric` `-p/--processes` `-t/--tcp` `-u/--udp`

### note
Quick note taking to JSON in your Documents folder
Flags: `-n/--limit` `--list`
Sub: `add` `list` `remove`

### path
Path manipulation utilities
Sub: `abs` `clean`

### pgrep
Find processes by name or pattern
Flags: `-c/--count` `-x/--exact` `-f/--full` `-i/--ignore-case` `-n/--newest` `-o/--oldest` `-P/--parent` `-u/--user`

### pipe
Chain omni commands without shell pipes
Flags: `-s/--sep` `--var` `-v/--verbose`

### pipeline
Streaming text processing engine
Flags: `-f/--file` `-v/--verbose`

### pkill
Kill processes by name or pattern
Flags: `-c/--count` `-x/--exact` `-f/--full` `-i/--ignore-case` `-l/--list` `-n/--newest` `-o/--oldest` `-P/--parent` `-s/--signal` `-u/--user` `-v/--verbose`

### printf
Format and print data
Flags: `-n/--no-newline`

### project
Analyze project structure, dependencies, and health
Sub: `deps` `docs` `git` `health` `info`

### remove
Alias for rm

### repo
Repository analysis tools
Sub: `analyze`

### rg
Recursively search for a pattern (ripgrep-style)
Flags: `-A/--after-context` `-B/--before-context` `-b/--byte-offset` `--color` `--colors` `--column` `-C/--context` `-c/--count` `-l/--files-with-matches` `-F/--fixed-strings` `-L/--follow` `-g/--glob` `--hidden` `-i/--ignore-case` `-v/--invert-match` `--json-stream` `-n/--line-number` `-m/--max-count` `--max-depth` `-U/--multiline` `-H/--no-heading` `--no-ignore` `-o/--only-matching` `--passthru` `-q/--quiet` `-r/--replace` `-S/--smart-case` `--stats` `-j/--threads` `--trim` `-t/--type` `-T/--type-not` `-w/--word-regexp`

### snowflake
Generate Twitter Snowflake-style IDs
Flags: `-n/--count` `-w/--worker`

### sql
SQL utilities (format, minify, validate)
Flags: `-i/--indent` `-u/--uppercase`
Sub: `fmt` `minify` `validate`

### ss
Display socket statistics
Flags: `-a/--all` `-e/--extended` `-4/--ipv4` `-6/--ipv6` `-l/--listening` `--no-header` `-n/--numeric` `-p/--processes` `--state` `-s/--summary` `-t/--tcp` `-u/--udp` `-x/--unix`

### tagfixer
Fix and standardize Go struct tags
Flags: `-a/--analyze` `-c/--case` `-d/--dry-run` `--json` `-r/--recursive` `-t/--tags` `-v/--verbose`
Sub: `analyze`

### task
Run tasks defined in Taskfile.yml
Flags: `--allow-external` `-d/--dir` `--dry-run` `-f/--force` `-l/--list` `-s/--silent` `--summary` `-t/--taskfile` `-v/--verbose`

### terraform
Terraform CLI
Sub: `apply` `console` `destroy` `fmt` `get` `graph` `import` `init` `output` `plan` `providers` `refresh` `show` `state` `taint` `test` `untaint` `validate` `version` `workspace`

### testcheck
Check test coverage for Go packages
Flags: `-a/--all` `-s/--summary` `-v/--verbose`

### toml
TOML utilities
Sub: `fmt` `validate`

### top
Display system processes sorted by resource usage
Flags: `--go` `-n/--num` `--sort`

### ulid
Generate Universally Unique Lexicographically Sortable Identifiers
Flags: `-n/--count` `-l/--lower`

### url
URL encoding and decoding utilities
Sub: `decode` `encode`

### vault
HashiCorp Vault CLI
Sub: `delete` `kv` `list` `login` `read` `status` `token` `write`

### video
Download videos from YouTube and other platforms
Flags: `--cookies` `--interactive` `--proxy` `-v/--verbose`
Sub: `auth` `channel` `download` `extractors` `info` `interactive` `list-formats` `search`

### xml
XML utilities (format, validate, convert)
Flags: `-i/--indent` `-m/--minify`
Sub: `fmt` `fromjson` `tojson` `validate`

### xxd
Make a hex dump or reverse it
Flags: `-b/--bits` `-c/--cols` `-g/--groupsize` `-i/--include` `-l/--len` `-p/--plain` `-r/--reverse` `-s/--seek` `-u/--uppercase`

### yaml
YAML utilities
Sub: `fmt` `k8s` `tostruct` `validate`

## Security

### decrypt
Decrypt data using AES-256-GCM
Flags: `-a/--armor` `-b/--base64` `-i/--iterations` `-k/--key-file` `-o/--output` `-p/--password` `-P/--password-file`

### encrypt
Encrypt data using AES-256-GCM
Flags: `-a/--armor` `-b/--base64` `-i/--iterations` `-k/--key-file` `-o/--output` `-p/--password` `-P/--password-file`

### random
Generate random values
Flags: `-c/--charset` `-n/--count` `-l/--length` `--max` `--min` `-s/--separator` `-t/--type`

### uuid
Generate random UUIDs
Flags: `-n/--count` `-x/--no-dashes` `-u/--upper` `-v/--version`

## System

### arch
Print machine architecture

### df
Report file system disk space usage
Flags: `-B/--block-size` `-x/--exclude-type` `-H/--human-readable` `-i/--inodes` `-l/--local` `-P/--portability` `--total` `-t/--type`

### du
Estimate file space usage
Flags: `-a/--all` `--apparent-size` `-B/--block-size` `-b/--bytes` `-H/--human-readable` `-d/--max-depth` `-0/--null` `-x/--one-file-system` `-s/--summarize` `-c/--total`

### env
Print environment variables
Flags: `-i/--ignore-environment` `-0/--null` `-u/--unset`

### free
Display amount of free and used memory in the system
Flags: `-b/--bytes` `-g/--gibibytes` `-H/--human` `-k/--kibibytes` `-m/--mebibytes` `-t/--total` `-w/--wide`

### id
Print user and group information
Flags: `-g/--group` `-G/--groups` `-n/--name` `-r/--real` `-u/--user`

### kill
Send a signal to a process
Flags: `-l/--list` `-s/--signal` `-v/--verbose`

### ps
Report a snapshot of current processes
Flags: `-a/--all` `-f/--full` `--go` `-l/--long` `--no-headers` `-p/--pid` `--sort` `-u/--user`

### uname
Print system information
Flags: `-a/--all` `-i/--hardware-platform` `-s/--kernel-name` `-r/--kernel-release` `-v/--kernel-version` `-m/--machine` `-n/--nodename` `-o/--operating-system` `-p/--processor`

### uptime
Tell how long the system has been running
Flags: `-p/--pretty` `-s/--since`

### which
Locate a command
Flags: `-a/--all`

### whoami
Print effective username

## Text

### awk
Pattern scanning and processing language
Flags: `-v/--assign` `-F/--field-separator`

### cmp
Compare two files byte by byte
Flags: `-n/--bytes` `-i/--ignore-initial` `-b/--print-bytes` `-s/--silent` `-l/--verbose`

### column
Columnate lists
Flags: `-c/--columns` `-x/--fillrows` `-H/--headers` `-J/--json` `-o/--output-separator` `-R/--right` `-s/--separator` `-t/--table`

### comm
Compare two sorted files line by line
Flags: `-1/--1` `-2/--2` `-3/--3` `--check-order` `--nocheck-order` `--output-delimiter` `-z/--zero-terminated`

### cut
Remove sections from each line of files
Flags: `-b/--bytes` `-c/--characters` `--complement` `-d/--delimiter` `-f/--fields` `-s/--only-delimited` `--output-delimiter`

### diff
Compare files line by line
Flags: `-q/--brief` `--color` `-B/--ignore-blank-lines` `-i/--ignore-case` `-b/--ignore-space-change` `--json` `-r/--recursive` `-y/--side-by-side` `--suppress-common-lines` `-u/--unified` `-W/--width`

### egrep
Print lines that match patterns (extended regexp)
Flags: `-C/--context` `-c/--count` `-l/--files-with-matches` `-i/--ignore-case` `-v/--invert-match` `-n/--line-number` `-o/--only-matching` `-q/--quiet`

### fgrep
Print lines that match patterns (fixed strings)
Flags: `-C/--context` `-c/--count` `-l/--files-with-matches` `-i/--ignore-case` `-v/--invert-match` `-n/--line-number` `-o/--only-matching` `-q/--quiet`

### fold
Wrap each input line to fit in specified width
Flags: `-b/--bytes` `-s/--spaces` `-w/--width`

### grep
Print lines that match patterns
Flags: `-A/--after-context` `-B/--before-context` `-C/--context` `-c/--count` `-E/--extended-regexp` `-l/--files-with-matches` `-L/--files-without-match` `-F/--fixed-strings` `-i/--ignore-case` `-v/--invert-match` `-n/--line-number` `-x/--line-regexp` `-m/--max-count` `--no-filename` `-o/--only-matching` `-q/--quiet` `-r/--recursive` `-H/--with-filename` `-w/--word-regexp`

### head
Output the first part of files
Flags: `-c/--bytes` `-n/--lines` `-q/--quiet` `-v/--verbose`

### join
Join lines of two files on a common field
Flags: `-1/--1` `-2/--2` `-a/--a` `-e/--e` `-i/--i` `-t/--t` `-v/--v`

### nl
Number lines of files
Flags: `-b/--body-numbering` `-i/--line-increment` `-n/--number-format` `-s/--number-separator` `-w/--number-width` `-v/--starting-line-number`

### paste
Merge lines of files
Flags: `-d/--delimiters` `-s/--serial` `-z/--zero-terminated`

### rev
Reverse lines characterwise

### sed
Stream editor for filtering and transforming text
Flags: `-e/--expression` `-i/--in-place` `--in-place-suffix` `-n/--quiet` `-r/--r` `-E/--regexp-extended`

### shuf
Generate random permutations
Flags: `-e/--echo` `-n/--head-count` `-i/--input-range` `-r/--repeat` `-z/--zero-terminated`

### sort
Sort lines of text files
Flags: `-c/--check` `-d/--dictionary-order` `-t/--field-separator` `-f/--ignore-case` `-b/--ignore-leading-blanks` `-k/--key` `-n/--numeric-sort` `-o/--output` `-r/--reverse` `-s/--stable` `-u/--unique`

### split
Split a file into pieces
Flags: `-b/--bytes` `-l/--lines` `-d/--numeric-suffixes` `-a/--suffix-length` `--verbose`

### strings
Print the printable strings in files
Flags: `-n/--bytes` `-t/--radix`

### tac
Concatenate and print files in reverse
Flags: `-b/--before` `-r/--regex` `-s/--separator`

### tail
Output the last part of files
Flags: `-c/--bytes` `-f/--follow` `-n/--lines` `-q/--quiet` `--sleep-interval` `-v/--verbose`

### tr
Translate or delete characters
Flags: `-c/--complement` `-d/--delete` `-s/--squeeze-repeats` `-t/--truncate-set1`

### uniq
Report or omit repeated lines
Flags: `-D/--all-repeated` `-w/--check-chars` `-c/--count` `-i/--ignore-case` `-d/--repeated` `-s/--skip-chars` `-f/--skip-fields` `-u/--unique` `-z/--zero-terminated`

### wc
Print newline, word, and byte counts for each file
Flags: `-c/--bytes` `-m/--chars` `-l/--lines` `-L/--max-line-length` `-w/--words`

## Tools

### aicontext
Generate AI context for coding agents
Flags: `-c/--category` `-j/--json` `--no-structure` `-o/--output` `--table`

### cmdtree
Display command tree visualization
Flags: `-b/--brief` `-c/--command` `-v/--verbose`

### lint
Check Taskfiles for portability issues
Flags: `--fix` `-q/--quiet` `--strict`

### logger
Configure omni command logging
Flags: `-d/--disable` `-p/--path` `-s/--status` `-v/--viewer`

## Utilities

### seq
Print a sequence of numbers
Flags: `-w/--equal-width` `-f/--format` `-s/--separator`

### sleep
Delay for a specified amount of time

### time
Time a simple command or give resource usage

### watch
Execute a program periodically, showing output fullscreen
Flags: `-b/--beep` `-g/--chgexit` `-c/--color` `-d/--differences` `-e/--errexit` `-n/--interval` `-t/--no-title` `--only-changes` `-p/--precise`

### xargs
Build and execute command lines from standard input
Flags: `-I/--I` `-d/--delimiter` `-n/--max-args` `-P/--max-procs` `-r/--no-run-if-empty` `-0/--null` `-t/--verbose`

## Structure

- `cmd/` CLI commands (Cobra)
- `internal/cli/` Command implementations
- `testing/` Python black-box tests
- `tests/` Go integration tests
