# goshell Command Reference

## Core Commands

### ls - List directory contents
```bash
goshell ls [OPTION]... [FILE]...
  -l, --long       use long listing format
  -a, --all        show hidden files
  -1               one file per line
  -h, --human      human-readable sizes
  -R, --recursive  list subdirectories recursively
  -S               sort by size
  -t               sort by modification time
  -r, --reverse    reverse order
  --json           output as JSON
```

### pwd - Print working directory
```bash
goshell pwd
  -L    print logical path (default)
  -P    print physical path (resolve symlinks)
```

### cat - Concatenate and print files
```bash
goshell cat [OPTION]... [FILE]...
  -n, --number     number all output lines
  -b               number non-empty lines
  -s               squeeze blank lines
  -E               display $ at end of lines
```

### date - Print current date and time
```bash
goshell date [OPTION]...
  -u, --utc        print UTC time
  -R, --rfc-email  RFC 5322 format
  -I, --iso-8601   ISO 8601 format
  --format=FMT     custom format (Go time format)
```

---

## File Operations

### cp - Copy files and directories
```bash
goshell cp [OPTION]... SOURCE DEST
  -r, -R, --recursive  copy directories recursively
  -f, --force          overwrite without prompting
  -n, --no-clobber     don't overwrite existing
  -v, --verbose        explain what is being done
```

### mv - Move/rename files
```bash
goshell mv [OPTION]... SOURCE DEST
  -f, --force       overwrite without prompting
  -n, --no-clobber  don't overwrite existing
  -v, --verbose     explain what is being done
```

### rm - Remove files or directories
```bash
goshell rm [OPTION]... FILE...
  -r, -R, --recursive  remove directories recursively
  -f, --force          ignore nonexistent files
  -v, --verbose        explain what is being done
```

### mkdir - Create directories
```bash
goshell mkdir [OPTION]... DIRECTORY...
  -p, --parents  create parent directories as needed
  -m, --mode     set permission mode
```

### ln - Create links
```bash
goshell ln [OPTION]... TARGET LINK_NAME
  -s, --symbolic  create symbolic link
  -f, --force     remove existing destination
  -v, --verbose   print name of each linked file
```

### chmod - Change file permissions
```bash
goshell chmod [OPTION]... MODE FILE...
  -R, --recursive  change files recursively
  -v, --verbose    explain what is being done
  # MODE: octal (755) or symbolic (u+x, go-w)
```

### chown - Change file ownership
```bash
goshell chown [OPTION]... OWNER[:GROUP] FILE...
  -R, --recursive  operate recursively
  -v, --verbose    explain what is being done
```

### touch - Update timestamps
```bash
goshell touch [OPTION]... FILE...
  -a             change access time only
  -m             change modification time only
  -c, --no-create  don't create files
```

### stat - Display file status
```bash
goshell stat [OPTION]... FILE...
  --json  output as JSON
```

---

## Text Processing

### grep - Search for patterns
```bash
goshell grep [OPTION]... PATTERN [FILE]...
  -i, --ignore-case    case insensitive
  -v, --invert-match   select non-matching lines
  -n, --line-number    print line numbers
  -c, --count          count matching lines
  -l, --files-with-matches  print filenames only
  -r, --recursive      search directories
  -E, --extended-regexp  extended regex
  -F, --fixed-strings  literal strings
  -w, --word-regexp    match whole words
  -A NUM               print NUM lines after match
  -B NUM               print NUM lines before match
  -C NUM               print NUM lines context
```

### sed - Stream editor
```bash
goshell sed [OPTION]... SCRIPT [FILE]...
  -e SCRIPT    add script to commands
  -i[SUFFIX]   edit files in place
  -n           suppress automatic printing
  -E, -r       extended regular expressions
  # Supported: s/pattern/replace/flags, d, p, q
```

### awk - Pattern scanning
```bash
goshell awk [OPTION]... 'PROGRAM' [FILE]...
  -F FS        input field separator
  -v VAR=VAL   set variable
  # Supported: $0-$N fields, BEGIN/END, /regex/, print, NF
```

### head - Output first lines
```bash
goshell head [OPTION]... [FILE]...
  -n, --lines=N  print first N lines (default 10)
  -c, --bytes=N  print first N bytes
  -q             never print headers
```

### tail - Output last lines
```bash
goshell tail [OPTION]... [FILE]...
  -n, --lines=N  print last N lines (default 10)
  -c, --bytes=N  print last N bytes
  -f, --follow   output appended data
  -q             never print headers
```

### sort - Sort lines
```bash
goshell sort [OPTION]... [FILE]...
  -r, --reverse     reverse order
  -n, --numeric     numeric sort
  -u, --unique      output unique lines only
  -f, --ignore-case case insensitive
  -t SEP            field separator
  -k KEY            sort by key
```

### uniq - Filter duplicate lines
```bash
goshell uniq [OPTION]... [INPUT [OUTPUT]]
  -c, --count   prefix lines by occurrence count
  -d, --repeated  only print duplicates
  -u, --unique    only print unique lines
  -i             case insensitive
```

### wc - Word/line/byte count
```bash
goshell wc [OPTION]... [FILE]...
  -l, --lines  print line count
  -w, --words  print word count
  -c, --bytes  print byte count
  -m, --chars  print character count
```

### cut - Extract fields
```bash
goshell cut [OPTION]... [FILE]...
  -b LIST       select bytes
  -c LIST       select characters
  -d DELIM      use DELIM as delimiter
  -f LIST       select fields
  -s            only lines with delimiter
  --complement  complement selection
```

### tr - Translate characters
```bash
goshell tr [OPTION]... SET1 [SET2]
  -c, --complement  complement SET1
  -d, --delete      delete characters in SET1
  -s, --squeeze     squeeze repeated characters
  # Supports: [:alpha:], [:digit:], [:upper:], etc.
```

### nl - Number lines
```bash
goshell nl [OPTION]... [FILE]...
  -b STYLE    body numbering style (a=all, t=nonempty, n=none)
  -n FORMAT   line number format (ln, rn, rz)
  -w WIDTH    line number width
  -s SEP      separator after number
```

### column - Columnate lists
```bash
goshell column [OPTION]... [FILE]...
  -t              table mode
  -s SEP          input delimiter
  -o SEP          output delimiter
  -c WIDTH        output width
  -x              fill rows before columns
```

### fold - Wrap lines
```bash
goshell fold [OPTION]... [FILE]...
  -w, --width=N  wrap at N characters (default 80)
  -b, --bytes    count bytes instead of columns
  -s, --spaces   break at spaces
```

### join - Join files on field
```bash
goshell join [OPTION]... FILE1 FILE2
  -1 FIELD  join on FIELD of file 1
  -2 FIELD  join on FIELD of file 2
  -t CHAR   field separator
  -i        ignore case
```

---

## System Information

### env - Print environment
```bash
goshell env [OPTION]... [NAME=VALUE]... [COMMAND]
  -i, --ignore-environment  start with empty environment
  -u, --unset=NAME          remove variable from environment
  -0, --null                end lines with NUL
```

### whoami - Print current user
```bash
goshell whoami
```

### id - Print user/group IDs
```bash
goshell id [OPTION]... [USER]
  -u, --user   print user ID
  -g, --group  print group ID
  -G, --groups print all group IDs
  -n, --name   print names instead of IDs
```

### uname - Print system info
```bash
goshell uname [OPTION]...
  -a, --all     print all information
  -s, --kernel-name    kernel name
  -n, --nodename       network hostname
  -r, --kernel-release kernel release
  -v, --kernel-version kernel version
  -m, --machine        hardware name
  -o, --operating-system  operating system
```

### uptime - Show system uptime
```bash
goshell uptime [OPTION]...
  -p, --pretty  human-readable format
  -s, --since   system up since time
```

### free - Display memory info
```bash
goshell free [OPTION]...
  -b, --bytes      output in bytes
  -k, --kibibytes  output in kibibytes
  -m, --mebibytes  output in mebibytes
  -g, --gibibytes  output in gibibytes
  -H, --human      human-readable
  -t, --total      show total
```

### df - Show disk usage
```bash
goshell df [OPTION]... [FILE]...
  -h, --human-readable  human-readable sizes
  -H, --si              powers of 1000
  -i, --inodes          show inode info
  -P, --portability     POSIX output format
  --total               show total
```

### du - Estimate file space
```bash
goshell du [OPTION]... [FILE]...
  -h, --human-readable  human-readable sizes
  -s, --summarize       display only total
  -a, --all             show all files
  -c, --total           produce grand total
  -d, --max-depth=N     limit depth
```

### ps - List processes
```bash
goshell ps [OPTION]...
  -a, --all        show all processes
  -f, --full       full format
  -l, --long       long format
  -u USER          show user's processes
  -p PID           show specific process
  --sort=COL       sort by column
  --no-headers     don't print headers
```

### kill - Send signals
```bash
goshell kill [OPTION]... PID...
  -s, --signal=SIG  specify signal
  -l, --list        list signal names
  -v, --verbose     report successful signals
```

---

## Archive & Compression

### tar - Create/extract archives
```bash
goshell tar [OPTION]... [FILE]...
  -c, --create   create archive
  -x, --extract  extract archive
  -t, --list     list contents
  -f FILE        archive file
  -v, --verbose  verbose output
  -z, --gzip     gzip compression
  -C DIR         change directory
  --strip-components=N  strip path components
```

### zip - Create zip archive
```bash
goshell zip [OPTION]... ZIPFILE FILE...
  -v, --verbose  verbose output
  -r             recurse directories
```

### unzip - Extract zip archive
```bash
goshell unzip [OPTION]... ZIPFILE
  -l, --list      list contents
  -v, --verbose   verbose output
  -d DIR          extract to directory
```

---

## Hash & Encoding

### hash - Compute file hashes
```bash
goshell hash [OPTION]... [FILE]...
  -a, --algorithm ALG  md5, sha1, sha256, sha512
  -c, --check          verify checksums
  -r, --recursive      hash recursively
```

### sha256sum / sha512sum / md5sum
```bash
goshell sha256sum [OPTION]... [FILE]...
  -c, --check  read checksums and check
  -b, --binary binary mode
  --quiet      don't print OK
  --status     silent, use exit code
```

### base64 - Base64 encode/decode
```bash
goshell base64 [OPTION]... [FILE]
  -d, --decode  decode data
  -w N          wrap at N characters
```

### base32 - Base32 encode/decode
```bash
goshell base32 [OPTION]... [FILE]
  -d, --decode  decode data
```

### base58 - Base58 encode/decode
```bash
goshell base58 [OPTION]... [FILE]
  -d, --decode  decode data
```

---

## Data Processing

### jq - JSON processor
```bash
goshell jq [OPTION]... FILTER [FILE]...
  -r, --raw-output   output raw strings
  -c, --compact      compact output
  -s, --slurp        read all into array
  -n, --null-input   don't read input
  # Filters: ., .field, .[n], .[], keys, length, type, |
```

### yq - YAML processor
```bash
goshell yq [OPTION]... FILTER [FILE]...
  -r, --raw-output   output raw strings
  -o json/yaml       output format
  -c                 compact JSON output
  # Same filter syntax as jq
```

### dotenv - Parse .env files
```bash
goshell dotenv [OPTION]... [FILE]...
  -e, --export  output as export statements
  -q, --quiet   suppress warnings
  -x, --expand  expand variables
```

---

## Security & Random

### encrypt - AES-256-GCM encryption
```bash
goshell encrypt [OPTION]... [FILE]
  -p, --password STRING  encryption password
  -P, --password-file    read password from file
  -o, --output FILE      output file
  -a, --armor            base64 output
  -i, --iterations N     PBKDF2 iterations
```

### decrypt - AES-256-GCM decryption
```bash
goshell decrypt [OPTION]... [FILE]
  -p, --password STRING  decryption password
  -P, --password-file    read password from file
  -o, --output FILE      output file
  -a, --armor            base64 input
```

### uuid - Generate UUIDs
```bash
goshell uuid [OPTION]...
  -n, --count N    generate N UUIDs
  -u, --upper      uppercase output
  -x, --no-dashes  without dashes
```

### random - Generate random values
```bash
goshell random [OPTION]...
  -n, --count N    number of values
  -l, --length N   string length (default 16)
  -t, --type TYPE  int, float, string, hex, password, bytes
  --min N          minimum for integers
  --max N          maximum for integers
  -c, --charset    custom character set
```

---

## Flow Control

### xargs - Build arguments
```bash
goshell xargs [OPTION]... [COMMAND]
  -0, --null       NUL-separated input
  -d DELIM         custom delimiter
  -n MAX           max arguments per command
  -P MAX           parallel processes
  -I REPLACE       replace string
  -t, --verbose    print commands
```

### watch - Execute repeatedly
```bash
goshell watch [OPTION]... COMMAND
  -n, --interval=N  seconds between runs
  -d, --differences highlight changes
  -t, --no-title    turn off header
```

### yes - Output repeatedly
```bash
goshell yes [STRING]...
  # Output STRING repeatedly until killed
```

---

## Command Tree

```
goshell
├── Core:        ls, pwd, cat, date, dirname, basename, realpath
├── File:        cp, mv, rm, mkdir, rmdir, touch, stat, ln, readlink, chmod, chown
├── Text:        grep, egrep, fgrep, head, tail, sort, uniq, wc, cut, tr,
│                nl, paste, tac, column, fold, join, sed, awk
├── System:      env, whoami, id, uname, uptime, free, df, du, ps, kill, time
├── Flow:        xargs, watch, yes, nohup
├── Archive:     tar, zip, unzip
├── Hash:        hash, sha256sum, sha512sum, md5sum
├── Encoding:    base64, base32, base58
├── Data:        jq, yq, dotenv
└── Security:    encrypt, decrypt, uuid, random
```
