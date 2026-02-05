# omni - AI Context Document

## Overview

Cross-platform, Go-native replacement for common shell utilities designed for Taskfile, CI/CD, and enterprise environments.

### Design Principles

- No exec: Never spawns external processes
- Pure Go: Standard library first, minimal dependencies
- Cross-platform: Linux, macOS, Windows support
- Library + CLI: All commands usable as Go packages
- Safe defaults: Destructive operations require explicit flags
- Testable: io.Writer pattern for all output

### Key Features

- 100+ commands implemented in pure Go
- JSON output mode for all commands
- Structured logging with OpenTelemetry support
- No external dependencies at runtime
- Consistent flag conventions across commands

## Command Categories

### Archive

Compression and archive management utilities

Commands: `bunzip2`, `bzcat`, `bzip2`, `gunzip`, `gzip`, `tar`, `unxz`, `unzip`, `xz`, `xzcat`, `zcat`, `zip`

### Code Generation

Code scaffolding and generation tools

Commands: `generate`

### Core

Essential file system navigation and basic I/O operations

Commands: `basename`, `cat`, `date`, `dirname`, `echo`, `ls`, `pwd`, `readlink`, `realpath`, `tree`, `yes`

### Data Processing

JSON, YAML, and structured data manipulation

Commands: `dotenv`, `jq`, `json`, `yq`

### Database

Embedded database operations (SQLite, BoltDB)

Commands: `bbolt`, `sqlite`

### File Operations

File manipulation, permissions, and management commands

Commands: `chmod`, `chown`, `cp`, `dd`, `file`, `find`, `ln`, `mkdir`, `mv`, `rm`, `rmdir`, `stat`, `touch`

### Hash & Encoding

Cryptographic hashes and encoding/decoding tools

Commands: `base32`, `base58`, `base64`, `hash`, `md5sum`, `sha256sum`, `sha512sum`

### Other

Commands: `aws`, `brdoc`, `buf`, `case`, `cloud`, `copy`, `cron`, `css`, `csv`, `curl`, `for`, `gbc`, `git`, `gops`, `gqc`, `hex`, `html`, `jwt`, `kconfig`, `kcs`, `kdebug`, `kdp`, `kdrain`, `keb`, `kga`, `kge`, `klf`, `kns`, `kpf`, `krr`, `krun`, `kscale`, `ksuid`, `ktn`, `ktp`, `kubectl`, `kwp`, `less`, `loc`, `more`, `move`, `nanoid`, `pipe`, `pipeline`, `printf`, `remove`, `rg`, `snowflake`, `sql`, `tagfixer`, `task`, `terraform`, `testcheck`, `toml`, `top`, `ulid`, `url`, `xml`, `xxd`, `yaml`

### Security

Encryption, random generation, and security utilities

Commands: `decrypt`, `encrypt`, `random`, `uuid`

### System Info

System information, process management, and environment

Commands: `arch`, `df`, `du`, `env`, `free`, `id`, `kill`, `ps`, `uname`, `uptime`, `which`, `whoami`

### Text Processing

Text transformation, filtering, and analysis tools

Commands: `awk`, `cmp`, `column`, `comm`, `cut`, `diff`, `egrep`, `fgrep`, `fold`, `grep`, `head`, `join`, `nl`, `paste`, `rev`, `sed`, `shuf`, `sort`, `split`, `strings`, `tac`, `tail`, `tr`, `uniq`, `wc`

### Tooling

Development and introspection tools

Commands: `aicontext`, `cmdtree`, `lint`, `logger`

### Utilities

General-purpose helper utilities

Commands: `seq`, `sleep`, `time`, `watch`, `xargs`

## Complete Command Reference

### aicontext

**Category:** Tooling

**Usage:** `omni aicontext [flags]`

**Description:** Generate AI-optimized context documentation

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --category | string | - | filter to specific category |
| --compact | bool | false | omit examples and long descriptions |
| --json | bool | false | output as structured JSON |
| -o, --output | string | - | write to file instead of stdout |

---

### arch

**Category:** System Info

**Usage:** `omni arch [flags]`

**Description:** Print machine architecture

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### awk

**Category:** Text Processing

**Usage:** `omni awk [OPTION]... 'program' [FILE]... [flags]`

**Description:** Pattern scanning and processing language

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -v, --assign | stringSlice | [] | assign value to variable (var=value) |
| -F, --field-separator | string | - | use FS for the input field separator |

---

### aws

**Category:** Other

**Usage:** `omni aws`

**Description:** AWS CLI operations

**Subcommands:** `ec2`, `iam`, `s3`, `ssm`, `sts`

---

### aws ec2

**Category:** Other

**Usage:** `omni aws ec2`

**Description:** AWS EC2 operations

**Subcommands:** `describe-instances`, `describe-security-groups`, `describe-vpcs`, `start-instances`, `stop-instances`

---

### aws ec2 describe-instances

**Category:** Other

**Usage:** `omni aws ec2 describe-instances [flags]`

**Description:** Describe EC2 instances

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --filters | stringSlice | [] | filters in format 'Name=value,Values=v1,v2' |
| --instance-ids | stringSlice | [] | instance IDs |
| --max-results | int32 | 0 | maximum number of results |

---

### aws ec2 describe-security-groups

**Category:** Other

**Usage:** `omni aws ec2 describe-security-groups [flags]`

**Description:** Describe security groups

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --filters | stringSlice | [] | filters |
| --group-ids | stringSlice | [] | security group IDs |

---

### aws ec2 describe-vpcs

**Category:** Other

**Usage:** `omni aws ec2 describe-vpcs [flags]`

**Description:** Describe VPCs

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --filters | stringSlice | [] | filters |
| --vpc-ids | stringSlice | [] | VPC IDs |

---

### aws ec2 start-instances

**Category:** Other

**Usage:** `omni aws ec2 start-instances [flags]`

**Description:** Start EC2 instances

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --instance-ids | stringSlice | [] | instance IDs (required) |

---

### aws ec2 stop-instances

**Category:** Other

**Usage:** `omni aws ec2 stop-instances [flags]`

**Description:** Stop EC2 instances

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --force | bool | false | force stop without graceful shutdown |
| --instance-ids | stringSlice | [] | instance IDs (required) |

---

### aws iam

**Category:** Other

**Usage:** `omni aws iam`

**Description:** AWS IAM operations

**Subcommands:** `get-policy`, `get-role`, `get-user`, `list-policies`, `list-roles`

---

### aws iam get-policy

**Category:** Other

**Usage:** `omni aws iam get-policy [flags]`

**Description:** Get IAM policy information

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --policy-arn | string | - | policy ARN (required) |

---

### aws iam get-role

**Category:** Other

**Usage:** `omni aws iam get-role [flags]`

**Description:** Get IAM role information

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --role-name | string | - | role name (required) |

---

### aws iam get-user

**Category:** Other

**Usage:** `omni aws iam get-user [flags]`

**Description:** Get IAM user information

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --user-name | string | - | user name (optional, defaults to current user) |

---

### aws iam list-policies

**Category:** Other

**Usage:** `omni aws iam list-policies [flags]`

**Description:** List IAM policies

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --max-items | int32 | 0 | maximum number of items |
| --only-attached | bool | false | only show attached policies |
| --path-prefix | string | - | path prefix filter |
| --scope | string | All | scope: All, AWS, Local |

---

### aws iam list-roles

**Category:** Other

**Usage:** `omni aws iam list-roles [flags]`

**Description:** List IAM roles

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --max-items | int32 | 0 | maximum number of items |
| --path-prefix | string | - | path prefix filter |

---

### aws s3

**Category:** Other

**Usage:** `omni aws s3`

**Description:** AWS S3 operations

**Subcommands:** `cp`, `ls`, `mb`, `presign`, `rb`, `rm`

---

### aws s3 cp

**Category:** File Operations

**Usage:** `omni aws s3 cp <SOURCE> <DESTINATION> [flags]`

**Description:** Copy files to/from S3

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dryrun | bool | false | display operations without executing |
| --quiet | bool | false | suppress output |
| --recursive | bool | false | copy recursively |

---

### aws s3 ls

**Category:** Core

**Usage:** `omni aws s3 ls [S3_URI] [flags]`

**Description:** List S3 objects or buckets

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --human-readable | bool | false | display file sizes in human-readable format |
| --recursive | bool | false | list recursively |
| --summarize | bool | false | display summary information |

---

### aws s3 mb

**Category:** Other

**Usage:** `omni aws s3 mb <S3_URI>`

**Description:** Create an S3 bucket

---

### aws s3 presign

**Category:** Other

**Usage:** `omni aws s3 presign <S3_URI> [flags]`

**Description:** Generate a presigned URL

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --expires-in | int | 900 | URL expiration time in seconds (default 15 minutes) |

---

### aws s3 rb

**Category:** Other

**Usage:** `omni aws s3 rb <S3_URI> [flags]`

**Description:** Remove an S3 bucket

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --force | bool | false | delete all objects before removing bucket |

---

### aws s3 rm

**Category:** File Operations

**Usage:** `omni aws s3 rm <S3_URI> [flags]`

**Description:** Remove S3 objects

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dryrun | bool | false | display operations without executing |
| --quiet | bool | false | suppress output |
| --recursive | bool | false | delete recursively |

---

### aws ssm

**Category:** Other

**Usage:** `omni aws ssm`

**Description:** AWS SSM Parameter Store operations

**Subcommands:** `delete-parameter`, `get-parameter`, `get-parameters`, `get-parameters-by-path`, `put-parameter`

---

### aws ssm delete-parameter

**Category:** Other

**Usage:** `omni aws ssm delete-parameter [flags]`

**Description:** Delete a parameter

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --name | string | - | parameter name (required) |

---

### aws ssm get-parameter

**Category:** Other

**Usage:** `omni aws ssm get-parameter [flags]`

**Description:** Get a parameter value

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --name | string | - | parameter name (required) |
| --with-decryption | bool | false | decrypt SecureString values |

---

### aws ssm get-parameters

**Category:** Other

**Usage:** `omni aws ssm get-parameters [flags]`

**Description:** Get multiple parameters

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --names | stringSlice | [] | parameter names (required) |
| --with-decryption | bool | false | decrypt SecureString values |

---

### aws ssm get-parameters-by-path

**Category:** Other

**Usage:** `omni aws ssm get-parameters-by-path [flags]`

**Description:** Get parameters by path

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --max-results | int32 | 10 | maximum results per page |
| --path | string | - | parameter path (required) |
| --recursive | bool | false | include nested parameters |
| --with-decryption | bool | false | decrypt SecureString values |

---

### aws ssm put-parameter

**Category:** Other

**Usage:** `omni aws ssm put-parameter [flags]`

**Description:** Create or update a parameter

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --description | string | - | parameter description |
| --key-id | string | - | KMS key for SecureString |
| --name | string | - | parameter name (required) |
| --overwrite | bool | false | overwrite existing parameter |
| --type | string | String | parameter type: String, StringList, SecureString |
| --value | string | - | parameter value (required) |

---

### aws sts

**Category:** Other

**Usage:** `omni aws sts`

**Description:** AWS STS operations

**Subcommands:** `assume-role`, `get-caller-identity`

---

### aws sts assume-role

**Category:** Other

**Usage:** `omni aws sts assume-role [flags]`

**Description:** Assume an IAM role

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --duration-seconds | int32 | 3600 | duration of the session in seconds |
| --external-id | string | - | external ID for cross-account access |
| --role-arn | string | - | ARN of the role to assume (required) |
| --role-session-name | string | - | session name (required) |

---

### aws sts get-caller-identity

**Category:** Other

**Usage:** `omni aws sts get-caller-identity`

**Description:** Get details about the IAM identity calling the API

---

### base32

**Category:** Hash & Encoding

**Usage:** `omni base32 [OPTION]... [FILE] [flags]`

**Description:** Base32 encode or decode data

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --decode | bool | false | decode data |
| -w, --wrap | int | 76 | wrap encoded lines after N characters (0 = no wrap) |

---

### base58

**Category:** Hash & Encoding

**Usage:** `omni base58 [OPTION]... [FILE] [flags]`

**Description:** Base58 encode or decode data (Bitcoin alphabet)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --decode | bool | false | decode data |

---

### base64

**Category:** Hash & Encoding

**Usage:** `omni base64 [OPTION]... [FILE] [flags]`

**Description:** Base64 encode or decode data

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --decode | bool | false | decode data |
| -i, --ignore-garbage | bool | false | ignore non-alphabet characters when decoding |
| -w, --wrap | int | 76 | wrap encoded lines after N characters (0 = no wrap) |

---

### basename

**Category:** Core

**Usage:** `omni basename NAME [SUFFIX] [flags]`

**Description:** Strip directory and suffix from file names

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |
| -s, --suffix | string | - | remove a trailing SUFFIX |

---

### bbolt

**Category:** Database

**Usage:** `omni bbolt`

**Description:** BoltDB database management

**Subcommands:** `buckets`, `check`, `compact`, `create-bucket`, `delete`, `delete-bucket`, `dump`, `get`, `info`, `keys`, `page`, `pages`, `put`, `stats`

---

### bbolt buckets

**Category:** Other

**Usage:** `omni bbolt buckets <database>`

**Description:** List all buckets in the database

---

### bbolt check

**Category:** Other

**Usage:** `omni bbolt check <database>`

**Description:** Verify database integrity

---

### bbolt compact

**Category:** Other

**Usage:** `omni bbolt compact <source> <destination>`

**Description:** Compact database to a new file

---

### bbolt create-bucket

**Category:** Other

**Usage:** `omni bbolt create-bucket <database> <bucket>`

**Description:** Create a new bucket

---

### bbolt delete

**Category:** Other

**Usage:** `omni bbolt delete <database> <bucket> <key>`

**Description:** Delete a key from a bucket

---

### bbolt delete-bucket

**Category:** Other

**Usage:** `omni bbolt delete-bucket <database> <bucket>`

**Description:** Delete a bucket

---

### bbolt dump

**Category:** Other

**Usage:** `omni bbolt dump <database> <bucket> [flags]`

**Description:** Dump all keys and values in a bucket

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --hex | bool | false | display values in hexadecimal |
| --prefix | string | - | filter keys by prefix |

---

### bbolt get

**Category:** Other

**Usage:** `omni bbolt get <database> <bucket> <key> [flags]`

**Description:** Get value for a key

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --hex | bool | false | display value in hexadecimal |

---

### bbolt info

**Category:** Other

**Usage:** `omni bbolt info <database>`

**Description:** Display database information

---

### bbolt keys

**Category:** Other

**Usage:** `omni bbolt keys <database> <bucket> [flags]`

**Description:** List keys in a bucket

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --prefix | string | - | filter keys by prefix |

---

### bbolt page

**Category:** Other

**Usage:** `omni bbolt page <database> <page-id>`

**Description:** Hex dump of a specific page

---

### bbolt pages

**Category:** Other

**Usage:** `omni bbolt pages <database>`

**Description:** List database pages

---

### bbolt put

**Category:** Other

**Usage:** `omni bbolt put <database> <bucket> <key> <value>`

**Description:** Store a key-value pair

---

### bbolt stats

**Category:** Other

**Usage:** `omni bbolt stats <database>`

**Description:** Display database statistics

---

### brdoc

**Category:** Other

**Usage:** `omni brdoc`

**Description:** Brazilian document utilities (CPF, CNPJ)

**Subcommands:** `cnpj`, `cpf`

---

### brdoc cnpj

**Category:** Other

**Usage:** `omni brdoc cnpj [CNPJ...] [flags]`

**Description:** CNPJ operations (generate, validate, format)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | number of CNPJs to generate |
| -f, --format | bool | false | format CNPJ(s) |
| -g, --generate | bool | false | generate valid CNPJ(s) |
| --json | bool | false | output as JSON |
| -l, --legacy | bool | false | generate numeric-only CNPJ |
| -v, --validate | bool | false | validate CNPJ(s) |

---

### brdoc cpf

**Category:** Other

**Usage:** `omni brdoc cpf [CPF...] [flags]`

**Description:** CPF operations (generate, validate, format)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | number of CPFs to generate |
| -f, --format | bool | false | format CPF(s) |
| -g, --generate | bool | false | generate valid CPF(s) |
| --json | bool | false | output as JSON |
| -v, --validate | bool | false | validate CPF(s) |

---

### buf

**Category:** Other

**Usage:** `omni buf`

**Description:** Protocol buffer utilities (lint, format, compile, generate)

**Subcommands:** `breaking`, `compile`, `format`, `generate`, `lint`, `ls-files`, `mod`

---

### buf breaking

**Category:** Other

**Usage:** `omni buf breaking [DIR] [flags]`

**Description:** Check for breaking changes

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --against | string | - | source to compare against (required) |
| --error-format | string | text | output format (text, json, github-actions) |
| --exclude-imports | bool | false | don't check imported files |
| --exclude-path | stringSlice | [] | paths to exclude |

---

### buf compile

**Category:** Other

**Usage:** `omni buf compile [DIR] [flags]`

**Description:** Compile proto files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --error-format | string | text | output format (text, json, github-actions) |
| --exclude-path | stringSlice | [] | paths to exclude |
| -o, --output | string | - | output file path |

---

### buf format

**Category:** Other

**Usage:** `omni buf format [DIR] [flags]`

**Description:** Format proto files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --diff | bool | false | display diff |
| --exit-code | bool | false | exit with non-zero if files unformatted |
| -w, --write | bool | false | rewrite files in place |

---

### buf generate

**Category:** Code Generation

**Usage:** `omni buf generate [DIR] [flags]`

**Description:** Generate code from proto files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --include-imports | bool | false | include imported files |
| -o, --output | string | - | base output directory |
| --template | string | - | alternate buf.gen.yaml location |

---

### buf lint

**Category:** Tooling

**Usage:** `omni buf lint [DIR] [flags]`

**Description:** Lint proto files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --config | string | - | custom config file path |
| --error-format | string | text | output format (text, json, github-actions) |
| --exclude-path | stringSlice | [] | paths to exclude |

---

### buf ls-files

**Category:** Other

**Usage:** `omni buf ls-files [DIR]`

**Description:** List proto files in the module

---

### buf mod

**Category:** Other

**Usage:** `omni buf mod`

**Description:** Module management commands

**Subcommands:** `init`, `update`

---

### buf mod init

**Category:** Other

**Usage:** `omni buf mod init [NAME] [flags]`

**Description:** Initialize a new buf module

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dir | string | . | directory to initialize |

---

### buf mod update

**Category:** Other

**Usage:** `omni buf mod update`

**Description:** Update dependencies

---

### bunzip2

**Category:** Archive

**Usage:** `omni bunzip2 [OPTION]... [FILE]... [flags]`

**Description:** Decompress bzip2 files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -f, --force | bool | false | force overwrite |
| -k, --keep | bool | false | keep original files |
| -c, --stdout | bool | false | write to stdout |
| -v, --verbose | bool | false | verbose mode |

---

### bzcat

**Category:** Archive

**Usage:** `omni bzcat [FILE]...`

**Description:** Decompress and print bzip2 files

---

### bzip2

**Category:** Archive

**Usage:** `omni bzip2 [OPTION]... [FILE]... [flags]`

**Description:** Decompress bzip2 files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --decompress | bool | false | decompress |
| -f, --force | bool | false | force overwrite |
| -k, --keep | bool | false | keep original files |
| -c, --stdout | bool | false | write to stdout |
| -v, --verbose | bool | false | verbose mode |

---

### case

**Category:** Other

**Usage:** `omni case`

**Description:** Text case conversion utilities

**Subcommands:** `all`, `camel`, `constant`, `detect`, `dot`, `kebab`, `lower`, `pascal`, `path`, `sentence`, `snake`, `swap`, `title`, `toggle`, `upper`

---

### case all

**Category:** Other

**Usage:** `omni case all [TEXT...] [flags]`

**Description:** Show all case conversions

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case camel

**Category:** Other

**Usage:** `omni case camel [TEXT...] [flags]`

**Description:** Convert to camelCase

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case constant

**Category:** Other

**Usage:** `omni case constant [TEXT...] [flags]`

**Description:** Convert to CONSTANT_CASE

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case detect

**Category:** Other

**Usage:** `omni case detect [TEXT...] [flags]`

**Description:** Detect the case type of text

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case dot

**Category:** Other

**Usage:** `omni case dot [TEXT...] [flags]`

**Description:** Convert to dot.case

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case kebab

**Category:** Other

**Usage:** `omni case kebab [TEXT...] [flags]`

**Description:** Convert to kebab-case

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case lower

**Category:** Other

**Usage:** `omni case lower [TEXT...] [flags]`

**Description:** Convert to lowercase

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case pascal

**Category:** Other

**Usage:** `omni case pascal [TEXT...] [flags]`

**Description:** Convert to PascalCase

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case path

**Category:** Other

**Usage:** `omni case path [TEXT...] [flags]`

**Description:** Convert to path/case

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case sentence

**Category:** Other

**Usage:** `omni case sentence [TEXT...] [flags]`

**Description:** Convert to Sentence case

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case snake

**Category:** Other

**Usage:** `omni case snake [TEXT...] [flags]`

**Description:** Convert to snake_case

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case swap

**Category:** Other

**Usage:** `omni case swap [TEXT...] [flags]`

**Description:** Swap case of each character

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case title

**Category:** Other

**Usage:** `omni case title [TEXT...] [flags]`

**Description:** Convert to Title Case

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case toggle

**Category:** Other

**Usage:** `omni case toggle [TEXT...] [flags]`

**Description:** Toggle first character's case

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### case upper

**Category:** Other

**Usage:** `omni case upper [TEXT...] [flags]`

**Description:** Convert to UPPERCASE

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### cat

**Category:** Core

**Usage:** `omni cat [file...] [flags]`

**Description:** Concatenate files and print on the standard output

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -e, --e | bool | false | equivalent to -vE |
| --json | bool | false | output as JSON array of lines |
| -n, --number | bool | false | number all output lines |
| -b, --number-nonblank | bool | false | number nonempty output lines, overrides -n |
| -A, --show-all | bool | false | equivalent to -vET |
| -E, --show-ends | bool | false | display $ at end of each line |
| -v, --show-nonprinting | bool | false | use ^ and M- notation, except for LFD and TAB |
| -T, --show-tabs | bool | false | display TAB characters as ^I |
| -s, --squeeze-blank | bool | false | suppress repeated empty output lines |
| -t, --t | bool | false | equivalent to -vT |

---

### chmod

**Category:** File Operations

**Usage:** `omni chmod [OPTION]... MODE[,MODE]... FILE... [flags]`

**Description:** Change file mode bits

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --changes | bool | false | like verbose but report only when a change is made |
| -R, --recursive | bool | false | change files and directories recursively |
| --reference | string | - | use RFILE's mode instead of MODE values |
| -f, --silent | bool | false | suppress most error messages |
| -v, --verbose | bool | false | output a diagnostic for every file processed |

---

### chown

**Category:** File Operations

**Usage:** `omni chown [OPTION]... OWNER[:GROUP] FILE... [flags]`

**Description:** Change file owner and group

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --changes | bool | false | like verbose but report only when a change is made |
| -h, --no-dereference | bool | false | affect symbolic links instead of referenced file |
| --preserve-root | bool | false | fail to operate recursively on '/' |
| -R, --recursive | bool | false | operate on files and directories recursively |
| --reference | string | - | use RFILE's owner and group |
| -f, --silent | bool | false | suppress most error messages |
| -v, --verbose | bool | false | output a diagnostic for every file processed |

---

### cloud

**Category:** Other

**Usage:** `omni cloud`

**Description:** Cloud profile management

**Subcommands:** `profile`

---

### cloud profile

**Category:** Other

**Usage:** `omni cloud profile`

**Description:** Manage cloud profiles

**Subcommands:** `add`, `delete`, `import`, `list`, `show`, `use`

---

### cloud profile add

**Category:** Other

**Usage:** `omni cloud profile add <name> [flags]`

**Description:** Add a new cloud profile

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --access-key-id | string | - | AWS Access Key ID |
| --account-id | string | - | Account/Subscription ID |
| --client-id | string | - | Azure Client ID |
| --client-secret | string | - | Azure Client Secret |
| --default | bool | false | Set as default profile |
| --key-file | string | - | Path to GCP service account JSON file |
| -p, --provider | string | - | Cloud provider (aws, azure, gcp) (required) |
| --region | string | - | Default region for the profile |
| --role-arn | string | - | IAM Role ARN (AWS only) |
| --secret-access-key | string | - | AWS Secret Access Key |
| --session-token | string | - | AWS Session Token (optional) |
| --subscription-id | string | - | Azure Subscription ID |
| --tenant-id | string | - | Azure Tenant ID |

---

### cloud profile delete

**Category:** Other

**Usage:** `omni cloud profile delete <name> [flags]`

**Description:** Delete a cloud profile

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -f, --force | bool | false | Skip confirmation |
| -p, --provider | string | - | Provider (required) |

---

### cloud profile import

**Category:** Other

**Usage:** `omni cloud profile import [name] [flags]`

**Description:** Import credentials from existing cloud CLI

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --default | bool | false | Set as default profile after import |
| --list | bool | false | List available profiles/credentials to import |
| -p, --provider | string | - | Cloud provider (aws, azure, gcp) (required) |
| -s, --source | string | - | Source profile/file to import from |

---

### cloud profile list

**Category:** Other

**Usage:** `omni cloud profile list [flags]`

**Description:** List cloud profiles

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -p, --provider | string | - | Filter by provider (aws, azure, gcp) |

---

### cloud profile show

**Category:** Other

**Usage:** `omni cloud profile show <name> [flags]`

**Description:** Show profile details

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -p, --provider | string | - | Provider (defaults to aws) |

---

### cloud profile use

**Category:** Other

**Usage:** `omni cloud profile use <name> [flags]`

**Description:** Set a profile as default

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -p, --provider | string | - | Provider (defaults to aws) |

---

### cmdtree

**Category:** Tooling

**Usage:** `omni cmdtree [flags]`

**Description:** Display command tree visualization

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --brief | bool | false | Show compact tree with short descriptions only |
| -c, --command | string | - | Show details for a specific command only |
| --json | bool | false | Output in JSON format |
| -v, --verbose | bool | true | Show full details for all commands (default) |

---

### cmp

**Category:** Text Processing

**Usage:** `omni cmp [OPTION]... FILE1 FILE2 [flags]`

**Description:** Compare two files byte by byte

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --bytes | int64 | 0 | compare at most LIMIT bytes |
| -i, --ignore-initial | int64 | 0 | skip first SKIP bytes |
| --json | bool | false | output as JSON |
| -b, --print-bytes | bool | false | print differing bytes |
| -s, --silent | bool | false | suppress all output |
| -l, --verbose | bool | false | output byte numbers and values |

---

### column

**Category:** Text Processing

**Usage:** `omni column [OPTION]... [FILE]... [flags]`

**Description:** Columnate lists

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --columns | int | 80 | output width in characters |
| -x, --fillrows | bool | false | fill rows before columns |
| -H, --headers | string | - | column headers (comma-separated) |
| -J, --json | bool | false | output as JSON |
| -o, --output-separator | string | - | output separator for table mode |
| -R, --right | bool | false | right-align columns |
| -s, --separator | string | - | delimiter characters for -t option |
| -t, --table | bool | false | determine column count based on input |

---

### comm

**Category:** Text Processing

**Usage:** `omni comm [OPTION]... FILE1 FILE2 [flags]`

**Description:** Compare two sorted files line by line

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -1, --1 | bool | false | suppress column 1 |
| -2, --2 | bool | false | suppress column 2 |
| -3, --3 | bool | false | suppress column 3 |
| --check-order | bool | false | check input is sorted |
| --json | bool | false | output as JSON |
| --nocheck-order | bool | false | do not check input order |
| --output-delimiter | string | - | use STR as output delimiter |
| -z, --zero-terminated | bool | false | line delimiter is NUL |

---

### copy

**Category:** Other

**Usage:** `omni copy`

**Description:** Alias for cp

---

### cp

**Category:** File Operations

**Usage:** `omni cp [source...] [destination]`

**Description:** Copy files and directories

---

### cron

**Category:** Other

**Usage:** `omni cron EXPRESSION [flags]`

**Description:** Parse and explain cron expressions

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |
| --next | int | 0 | show next N scheduled runs |
| --validate | bool | false | only validate the expression |

---

### css

**Category:** Other

**Usage:** `omni css [FILE] [flags]`

**Description:** CSS utilities (format, minify, validate)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| --sort-props | bool | false | sort properties alphabetically |
| --sort-rules | bool | false | sort selectors alphabetically |

**Subcommands:** `fmt`, `minify`, `validate`

---

### css fmt

**Category:** Other

**Usage:** `omni css fmt [FILE] [flags]`

**Description:** Format/beautify CSS

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| --sort-props | bool | false | sort properties alphabetically |
| --sort-rules | bool | false | sort selectors alphabetically |

---

### css minify

**Category:** Other

**Usage:** `omni css minify [FILE]`

**Description:** Minify CSS

---

### css validate

**Category:** Other

**Usage:** `omni css validate [FILE] [flags]`

**Description:** Validate CSS syntax

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### csv

**Category:** Other

**Usage:** `omni csv`

**Description:** CSV utilities (convert to/from JSON)

**Subcommands:** `fromjson`, `tojson`

---

### csv fromjson

**Category:** Other

**Usage:** `omni csv fromjson [FILE] [flags]`

**Description:** Convert JSON array to CSV

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --delimiter | string | , | field delimiter |
| --no-header | bool | false | don't include header row |
| --no-quotes | bool | false | don't quote fields |

---

### csv tojson

**Category:** Other

**Usage:** `omni csv tojson [FILE] [flags]`

**Description:** Convert CSV to JSON array

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --array | bool | false | always output as array |
| -d, --delimiter | string | , | field delimiter |
| --no-header | bool | false | first row is data, not headers |

---

### curl

**Category:** Other

**Usage:** `omni curl [METHOD] URL [ITEM...] [flags]`

**Description:** HTTP client with httpie-like syntax

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --data | string | - | request body data |
| -f, --form | bool | false | send as form data instead of JSON |
| -H, --header | stringArray | [] | custom header (can be used multiple times) |
| -k, --insecure | bool | false | skip TLS verification |
| --json | bool | false | output response as structured JSON |
| -L, --location | bool | true | follow redirects |
| -t, --timeout | int | 30 | request timeout in seconds |
| -v, --verbose | bool | false | show request/response details |

---

### cut

**Category:** Text Processing

**Usage:** `omni cut [OPTION]... [FILE]... [flags]`

**Description:** Remove sections from each line of files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --bytes | string | - | select only these bytes |
| -c, --characters | string | - | select only these characters |
| --complement | bool | false | complement the set of selected bytes, characters or fields |
| -d, --delimiter | string | - | use DELIM instead of TAB for field delimiter |
| -f, --fields | string | - | select only these fields |
| --json | bool | false | output as JSON |
| -s, --only-delimited | bool | false | do not print lines not containing delimiters |
| --output-delimiter | string | - | use STRING as the output delimiter |

---

### date

**Category:** Core

**Usage:** `omni date [+FORMAT] [flags]`

**Description:** Print the current date and time

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --iso-8601 | bool | false | output date/time in ISO 8601 format |
| --json | bool | false | output as JSON |
| -u, --utc | bool | false | print Coordinated Universal Time (UTC) |

---

### dd

**Category:** File Operations

**Usage:** `omni dd [OPERAND]...`

**Description:** Convert and copy a file

---

### decrypt

**Category:** Security

**Usage:** `omni decrypt [OPTION]... [FILE] [flags]`

**Description:** Decrypt data using AES-256-GCM

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --armor | bool | false | input is ASCII armored (base64) |
| -b, --base64 | bool | false | input is base64 (same as -a) |
| -i, --iterations | int | 100000 | PBKDF2 iterations |
| -k, --key-file | string | - | use key file for decryption |
| -o, --output | string | - | write output to file |
| -p, --password | string | - | password for decryption |
| -P, --password-file | string | - | read password from file |

---

### df

**Category:** System Info

**Usage:** `omni df [OPTION]... [FILE]... [flags]`

**Description:** Report file system disk space usage

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -B, --block-size | int64 | 0 | scale sizes by SIZE before printing them |
| -x, --exclude-type | string | - | exclude file systems of type TYPE |
| -H, --human-readable | bool | false | print sizes in human readable format |
| -i, --inodes | bool | false | list inode information instead of block usage |
| --json | bool | false | output as JSON |
| -l, --local | bool | false | limit listing to local file systems |
| -P, --portability | bool | false | use the POSIX output format |
| --total | bool | false | produce a grand total |
| -t, --type | string | - | limit listing to file systems of type TYPE |

---

### diff

**Category:** Text Processing

**Usage:** `omni diff [OPTION]... FILE1 FILE2 [flags]`

**Description:** Compare files line by line

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -q, --brief | bool | false | report only when files differ |
| --color | bool | false | colorize the output |
| -B, --ignore-blank-lines | bool | false | ignore changes where lines are all blank |
| -i, --ignore-case | bool | false | ignore case differences |
| -b, --ignore-space-change | bool | false | ignore changes in amount of white space |
| --json | bool | false | compare as JSON files |
| -r, --recursive | bool | false | recursively compare subdirectories |
| -y, --side-by-side | bool | false | output in two columns |
| --suppress-common-lines | bool | false | do not output common lines in side-by-side |
| -u, --unified | int | 3 | output NUM lines of unified context |
| -W, --width | int | 130 | output at most NUM columns |

---

### dirname

**Category:** Core

**Usage:** `omni dirname [path...] [flags]`

**Description:** Strip last component from file name

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### dotenv

**Category:** Data Processing

**Usage:** `omni dotenv [OPTION]... [FILE]... [flags]`

**Description:** Load environment variables from .env files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -x, --expand | bool | false | expand variables in values |
| -e, --export | bool | false | output as export statements |
| -q, --quiet | bool | false | suppress warnings |

---

### du

**Category:** System Info

**Usage:** `omni du [OPTION]... [FILE]... [flags]`

**Description:** Estimate file space usage

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | false | write counts for all files, not just directories |
| --apparent-size | bool | false | print apparent sizes, rather than disk usage |
| -B, --block-size | int64 | 0 | scale sizes by SIZE before printing them |
| -b, --bytes | bool | false | equivalent to --apparent-size --block-size=1 |
| -H, --human-readable | bool | false | print sizes in human readable format |
| --json | bool | false | output as JSON |
| -d, --max-depth | int | 0 | print total for directory only if N or fewer levels deep |
| -0, --null | bool | false | end each output line with NUL, not newline |
| -x, --one-file-system | bool | false | skip directories on different file systems |
| -s, --summarize | bool | false | display only a total for each argument |
| -c, --total | bool | false | produce a grand total |

---

### echo

**Category:** Core

**Usage:** `omni echo [STRING]... [flags]`

**Description:** Display a line of text

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -e, --escape | bool | false | enable interpretation of backslash escapes |
| --json | bool | false | output as JSON |
| -E, --no-escape | bool | false | disable interpretation of backslash escapes (default) |
| -n, --no-newline | bool | false | do not output the trailing newline |

---

### egrep

**Category:** Text Processing

**Usage:** `omni egrep [options] PATTERN [FILE...] [flags]`

**Description:** Print lines that match patterns (extended regexp)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -C, --context | int | 0 | print NUM lines of context |
| -c, --count | bool | false | only print a count of matching lines |
| -l, --files-with-matches | bool | false | only print FILE names with matches |
| -i, --ignore-case | bool | false | ignore case distinctions |
| -v, --invert-match | bool | false | select non-matching lines |
| -n, --line-number | bool | false | prefix each line with line number |
| -o, --only-matching | bool | false | show only matched parts |
| -q, --quiet | bool | false | suppress all normal output |

---

### encrypt

**Category:** Security

**Usage:** `omni encrypt [OPTION]... [FILE] [flags]`

**Description:** Encrypt data using AES-256-GCM

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --armor | bool | false | ASCII armor (base64) output |
| -b, --base64 | bool | false | base64 output (same as -a) |
| -i, --iterations | int | 100000 | PBKDF2 iterations |
| -k, --key-file | string | - | use key file for encryption |
| -o, --output | string | - | write output to file |
| -p, --password | string | - | password for encryption |
| -P, --password-file | string | - | read password from file |

---

### env

**Category:** System Info

**Usage:** `omni env [NAME...] [flags]`

**Description:** Print environment variables

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --ignore-environment | bool | false | start with an empty environment |
| --json | bool | false | output in JSON format |
| -0, --null | bool | false | end each output line with NUL, not newline |
| -u, --unset | string | - | remove variable from the environment |

---

### fgrep

**Category:** Text Processing

**Usage:** `omni fgrep [options] PATTERN [FILE...] [flags]`

**Description:** Print lines that match patterns (fixed strings)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -C, --context | int | 0 | print NUM lines of context |
| -c, --count | bool | false | only print a count of matching lines |
| -l, --files-with-matches | bool | false | only print FILE names with matches |
| -i, --ignore-case | bool | false | ignore case distinctions |
| -v, --invert-match | bool | false | select non-matching lines |
| -n, --line-number | bool | false | prefix each line with line number |
| -o, --only-matching | bool | false | show only matched parts |
| -q, --quiet | bool | false | suppress all normal output |

---

### file

**Category:** File Operations

**Usage:** `omni file [OPTION]... FILE... [flags]`

**Description:** Determine file type

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --brief | bool | false | do not prepend filenames |
| --json | bool | false | output as JSON |
| -i, --mime | bool | false | output MIME type |
| -L, --no-dereference | bool | false | don't follow symlinks |
| -F, --separator | string | : | use string as separator |

---

### find

**Category:** File Operations

**Usage:** `omni find [path...] [expression] [flags]`

**Description:** Search for files in a directory hierarchy

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --amin | string | - | access time [+-]N minutes |
| --atime | string | - | access time [+-]N days |
| --empty | bool | false | file is empty |
| --executable | bool | false | file is executable |
| --iname | string | - | case insensitive name match |
| --ipath | string | - | case insensitive path match |
| --iregex | string | - | case insensitive regex |
| --json | bool | false | output in JSON format |
| --maxdepth | int | 0 | maximum depth (0=unlimited) |
| --mindepth | int | 0 | minimum depth |
| --mmin | string | - | modification time [+-]N minutes |
| --mtime | string | - | modification time [+-]N days |
| --name | string | - | file name matches pattern |
| --not | bool | false | negate next test |
| --path | string | - | path matches pattern |
| -0, --print0 | bool | false | print with null terminator |
| --readable | bool | false | file is readable |
| --regex | string | - | path matches regex |
| --size | string | - | file size [+-]N[ckMG] |
| --type | string | - | file type (f=file, d=dir, l=link) |
| --writable | bool | false | file is writable |

---

### fold

**Category:** Text Processing

**Usage:** `omni fold [OPTION]... [FILE]... [flags]`

**Description:** Wrap each input line to fit in specified width

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --bytes | bool | false | count bytes rather than columns |
| -s, --spaces | bool | false | break at spaces |
| -w, --width | int | 80 | use WIDTH columns instead of 80 |

---

### for

**Category:** Other

**Usage:** `omni for`

**Description:** Loop and execute commands

**Subcommands:** `each`, `glob`, `lines`, `range`, `split`

---

### for each

**Category:** Other

**Usage:** `omni for each ITEM... -- COMMAND [flags]`

**Description:** Loop over a list of items

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | print commands without executing |
| --var | string | - | variable name to use |

---

### for glob

**Category:** Other

**Usage:** `omni for glob PATTERN -- COMMAND [flags]`

**Description:** Loop over files matching a pattern

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | print commands without executing |
| --var | string | - | variable name to use |

---

### for lines

**Category:** Other

**Usage:** `omni for lines [FILE] -- COMMAND [flags]`

**Description:** Loop over lines from stdin or file

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | print commands without executing |
| --var | string | - | variable name to use |

---

### for range

**Category:** Other

**Usage:** `omni for range START END [STEP] -- COMMAND [flags]`

**Description:** Loop over a numeric range

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | print commands without executing |
| --var | string | - | variable name to use |

---

### for split

**Category:** Text Processing

**Usage:** `omni for split DELIMITER INPUT -- COMMAND [flags]`

**Description:** Loop over items split by delimiter

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | print commands without executing |
| --var | string | - | variable name to use |

---

### free

**Category:** System Info

**Usage:** `omni free [OPTION]... [flags]`

**Description:** Display amount of free and used memory in the system

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --bytes | bool | false | show output in bytes |
| -g, --gibibytes | bool | false | show output in gibibytes |
| -H, --human | bool | false | show human-readable output |
| --json | bool | false | output as JSON |
| -k, --kibibytes | bool | false | show output in kibibytes |
| -m, --mebibytes | bool | false | show output in mebibytes |
| -t, --total | bool | false | show total for RAM + swap |
| -w, --wide | bool | false | wide output |

---

### gbc

**Category:** Other

**Usage:** `omni gbc [flags]`

**Description:** Git branch clean (alias)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | Show branches that would be deleted |

---

### generate

**Category:** Code Generation

**Usage:** `omni generate`

**Description:** Code generation utilities

**Subcommands:** `cobra`, `handler`, `repository`, `test`

---

### generate cobra

**Category:** Other

**Usage:** `omni generate cobra`

**Description:** Cobra CLI application generator

**Subcommands:** `add`, `config`, `init`

---

### generate cobra add

**Category:** Other

**Usage:** `omni generate cobra add <command-name> [flags]`

**Description:** Add a new command to an existing Cobra application

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --description | string | - | command description |
| --dir | string | - | project directory (defaults to current directory) |
| -p, --parent | string | root | parent command |

---

### generate cobra config

**Category:** Other

**Usage:** `omni generate cobra config [flags]`

**Description:** Manage cobra generator configuration

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --author | string | - | author name for config |
| --full | bool | false | set full in config |
| --init | bool | false | create a new configuration file |
| -l, --license | string | - | license type for config |
| --service | bool | false | set useService in config |
| --show | bool | false | show current configuration |
| --viper | bool | false | set useViper in config |

---

### generate cobra init

**Category:** Other

**Usage:** `omni generate cobra init <directory> [flags]`

**Description:** Initialize a new Cobra CLI application

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --author | string | - | author name |
| -d, --description | string | - | application description |
| --full | bool | false | full project with goreleaser, workflows, etc. |
| -l, --license | string | - | license type (MIT, Apache-2.0, BSD-3) |
| -m, --module | string | - | Go module path (required) |
| -n, --name | string | - | application name (defaults to directory name) |
| --service | bool | false | include service pattern with inovacc/config |
| --viper | bool | false | include viper for configuration |

---

### generate handler

**Category:** Other

**Usage:** `omni generate handler <name> [flags]`

**Description:** Generate HTTP handler

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --dir | string | internal/handler | output directory |
| -f, --framework | string | stdlib | framework: stdlib, chi, gin, echo |
| -m, --method | string | GET,POST,PUT,DELETE | HTTP methods (comma-separated) |
| --middleware | bool | false | include middleware support |
| -p, --package | string | handler | package name |
| --path | string | - | URL path pattern |

---

### generate repository

**Category:** Other

**Usage:** `omni generate repository <name> [flags]`

**Description:** Generate database repository

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --db | string | postgres | database type: postgres, mysql, sqlite |
| -d, --dir | string | internal/repository | output directory |
| -e, --entity | string | - | entity struct name |
| --interface | bool | true | generate interface |
| -p, --package | string | repository | package name |
| -t, --table | string | - | database table name |

---

### generate test

**Category:** Utilities

**Usage:** `omni generate test <file.go> [flags]`

**Description:** Generate tests for a Go source file

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --benchmark | bool | false | include benchmark tests |
| --fuzz | bool | false | include fuzz tests |
| --mock | bool | false | generate mock setup |
| --parallel | bool | false | add t.Parallel() calls |
| --table | bool | true | generate table-driven tests |

---

### git

**Category:** Other

**Usage:** `omni git`

**Description:** Git shortcuts and hacks

**Subcommands:** `amend`, `blame-line`, `branch-clean`, `diff-words`, `fetch-all`, `log-graph`, `pull-rebase`, `push`, `quick-commit`, `stash-staged`, `status`, `undo`

---

### git amend

**Category:** Other

**Usage:** `omni git amend`

**Description:** Amend last commit without editing

---

### git blame-line

**Category:** Other

**Usage:** `omni git blame-line <file> [flags]`

**Description:** Blame specific line range

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --end | int | 0 | End line number |
| --start | int | 0 | Start line number |

---

### git branch-clean

**Category:** Other

**Usage:** `omni git branch-clean [flags]`

**Description:** Delete merged branches

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | Show branches that would be deleted |

---

### git diff-words

**Category:** Other

**Usage:** `omni git diff-words`

**Description:** Word-level diff

---

### git fetch-all

**Category:** Other

**Usage:** `omni git fetch-all`

**Description:** Fetch all remotes with prune

---

### git log-graph

**Category:** Other

**Usage:** `omni git log-graph [flags]`

**Description:** Pretty log with graph

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 0 | Number of commits to show |

---

### git pull-rebase

**Category:** Other

**Usage:** `omni git pull-rebase`

**Description:** Pull with rebase

---

### git push

**Category:** Other

**Usage:** `omni git push [flags]`

**Description:** Push to remote

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --force | bool | false | Force push (with lease) |

---

### git quick-commit

**Category:** Other

**Usage:** `omni git quick-commit [flags]`

**Description:** Stage all and commit

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | true | Stage all changes before commit |
| -m, --message | string | - | Commit message (required) |

---

### git stash-staged

**Category:** Other

**Usage:** `omni git stash-staged [flags]`

**Description:** Stash only staged changes

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -m, --message | string | - | Stash message |

---

### git status

**Category:** Other

**Usage:** `omni git status`

**Description:** Short status

---

### git undo

**Category:** Other

**Usage:** `omni git undo`

**Description:** Undo last commit (soft reset)

---

### gops

**Category:** Other

**Usage:** `omni gops [PID] [flags]`

**Description:** Display Go process information

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -j, --json | bool | false | output as JSON |

---

### gqc

**Category:** Other

**Usage:** `omni gqc [flags]`

**Description:** Git quick commit (alias)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | true | Stage all changes before commit |
| -m, --message | string | - | Commit message (required) |

---

### grep

**Category:** Text Processing

**Usage:** `omni grep [options] PATTERN [FILE...] [flags]`

**Description:** Print lines that match patterns

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -A, --after-context | int | 0 | print NUM lines of trailing context |
| -B, --before-context | int | 0 | print NUM lines of leading context |
| -C, --context | int | 0 | print NUM lines of output context |
| -c, --count | bool | false | only print a count of matching lines per FILE |
| -E, --extended-regexp | bool | false | interpret PATTERN as an extended regular expression |
| -l, --files-with-matches | bool | false | only print FILE names containing matches |
| -L, --files-without-match | bool | false | only print FILE names not containing matches |
| -F, --fixed-strings | bool | false | interpret PATTERN as fixed strings |
| -i, --ignore-case | bool | false | ignore case distinctions in patterns and data |
| -v, --invert-match | bool | false | select non-matching lines |
| --json | bool | false | output as JSON |
| -n, --line-number | bool | false | prefix each line of output with line number |
| -x, --line-regexp | bool | false | match only whole lines |
| -m, --max-count | int | 0 | stop after NUM matches |
| --no-filename | bool | false | suppress the file name prefix on output |
| -o, --only-matching | bool | false | show only nonempty parts of lines that match |
| -q, --quiet | bool | false | suppress all normal output |
| -r, --recursive | bool | false | search directories recursively |
| -H, --with-filename | bool | false | print file name with output lines |
| -w, --word-regexp | bool | false | match only whole words |

---

### gunzip

**Category:** Archive

**Usage:** `omni gunzip [OPTION]... [FILE]... [flags]`

**Description:** Decompress gzip files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -f, --force | bool | false | force overwrite |
| -k, --keep | bool | false | keep original files |
| -c, --stdout | bool | false | write to stdout |
| -v, --verbose | bool | false | verbose mode |

---

### gzip

**Category:** Archive

**Usage:** `omni gzip [OPTION]... [FILE]... [flags]`

**Description:** Compress or decompress files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -9, --best | int | 0 | compress better |
| -d, --decompress | bool | false | decompress |
| -1, --fast | int | 0 | compress faster |
| -f, --force | bool | false | force overwrite |
| -k, --keep | bool | false | keep original files |
| -c, --stdout | bool | false | write to stdout |
| -v, --verbose | bool | false | verbose mode |

---

### hash

**Category:** Hash & Encoding

**Usage:** `omni hash [OPTION]... [FILE]... [flags]`

**Description:** Compute and check file hashes

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --algorithm | string | sha256 | hash algorithm (md5, sha1, sha256, sha512) |
| -b, --binary | bool | false | read in binary mode |
| -c, --check | bool | false | read checksums from FILE and check them |
| --json | bool | false | output as JSON |
| --quiet | bool | false | don't print OK for verified files |
| -r, --recursive | bool | false | hash files recursively |
| --status | bool | false | don't output anything, use status code |
| -w, --warn | bool | false | warn about improperly formatted lines |

---

### head

**Category:** Text Processing

**Usage:** `omni head [option]... [file]... [flags]`

**Description:** Output the first part of files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --bytes | int | 0 | print the first NUM bytes of each file |
| --json | bool | false | output as JSON |
| -n, --lines | int | 10 | print the first NUM lines instead of the first 10 |
| -q, --quiet | bool | false | never print headers giving file names |
| -v, --verbose | bool | false | always print headers giving file names |

---

### hex

**Category:** Other

**Usage:** `omni hex`

**Description:** Hexadecimal encoding and decoding utilities

**Subcommands:** `decode`, `encode`

---

### hex decode

**Category:** Other

**Usage:** `omni hex decode [HEX] [flags]`

**Description:** Decode hexadecimal to text

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### hex encode

**Category:** Other

**Usage:** `omni hex encode [TEXT] [flags]`

**Description:** Encode text to hexadecimal

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |
| -u, --upper | bool | false | use uppercase hex letters |

---

### html

**Category:** Other

**Usage:** `omni html`

**Description:** HTML utilities (format, encode, decode)

**Subcommands:** `decode`, `encode`, `fmt`, `minify`, `validate`

---

### html decode

**Category:** Other

**Usage:** `omni html decode [TEXT] [flags]`

**Description:** HTML decode text

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### html encode

**Category:** Other

**Usage:** `omni html encode [TEXT] [flags]`

**Description:** HTML encode text

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### html fmt

**Category:** Other

**Usage:** `omni html fmt [FILE] [flags]`

**Description:** Format/beautify HTML

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| --sort-attrs | bool | false | sort attributes alphabetically |

---

### html minify

**Category:** Other

**Usage:** `omni html minify [FILE]`

**Description:** Minify HTML

---

### html validate

**Category:** Other

**Usage:** `omni html validate [FILE] [flags]`

**Description:** Validate HTML syntax

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### id

**Category:** System Info

**Usage:** `omni id [OPTION]... [USER] [flags]`

**Description:** Print user and group information

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -g, --group | bool | false | print only the effective group ID |
| -G, --groups | bool | false | print all group IDs |
| --json | bool | false | output as JSON |
| -n, --name | bool | false | print a name instead of a number |
| -r, --real | bool | false | print the real ID instead of the effective ID |
| -u, --user | bool | false | print only the effective user ID |

---

### join

**Category:** Text Processing

**Usage:** `omni join [OPTION]... FILE1 FILE2 [flags]`

**Description:** Join lines of two files on a common field

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -1, --1 | int | 1 | join on this FIELD of file 1 |
| -2, --2 | int | 1 | join on this FIELD of file 2 |
| -a, --a | int | 0 | also print unpairable lines from file FILENUM (1 or 2) |
| -e, --e | string | - | replace missing fields with EMPTY |
| -i, --i | bool | false | ignore case when comparing fields |
| -t, --t | string | - | use CHAR as input and output field separator |
| -v, --v | int | 0 | print only unpairable lines from file FILENUM (1 or 2) |

---

### jq

**Category:** Data Processing

**Usage:** `omni jq [OPTION]... FILTER [FILE]... [flags]`

**Description:** Command-line JSON processor

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --compact-output | bool | false | compact output |
| -n, --null-input | bool | false | don't read any input |
| -r, --raw-output | bool | false | output raw strings |
| -s, --slurp | bool | false | read all inputs into array |
| -S, --sort-keys | bool | false | sort object keys |
| --tab | bool | false | use tabs for indentation |

---

### json

**Category:** Data Processing

**Usage:** `omni json`

**Description:** JSON utilities (format, minify, validate)

**Subcommands:** `fmt`, `fromcsv`, `fromtoml`, `fromxml`, `fromyaml`, `keys`, `minify`, `stats`, `tocsv`, `tostruct`, `toxml`, `toyaml`, `validate`

---

### json fmt

**Category:** Other

**Usage:** `omni json fmt [FILE]... [flags]`

**Description:** Beautify/format JSON with indentation

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -e, --escape-html | bool | false | escape HTML characters |
| -i, --indent | string |    | indentation string |
| -s, --sort-keys | bool | false | sort object keys |
| -t, --tab | bool | false | use tabs for indentation |

---

### json fromcsv

**Category:** Other

**Usage:** `omni json fromcsv [FILE] [flags]`

**Description:** Convert CSV to JSON array

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --array | bool | false | always output as array |
| -d, --delimiter | string | , | field delimiter |
| --no-header | bool | false | first row is data, not headers |

---

### json fromtoml

**Category:** Other

**Usage:** `omni json fromtoml [FILE] [flags]`

**Description:** Convert TOML to JSON

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -m, --minify | bool | false | output minified JSON |

---

### json fromxml

**Category:** Other

**Usage:** `omni json fromxml [FILE] [flags]`

**Description:** Convert XML to JSON

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --attr-prefix | string | - | prefix for attributes in JSON |
| --text-key | string | #text | key for text content |

---

### json fromyaml

**Category:** Other

**Usage:** `omni json fromyaml [FILE] [flags]`

**Description:** Convert YAML to JSON

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -m, --minify | bool | false | output minified JSON |

---

### json keys

**Category:** Other

**Usage:** `omni json keys [FILE] [flags]`

**Description:** List all keys in JSON object

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### json minify

**Category:** Other

**Usage:** `omni json minify [FILE]... [flags]`

**Description:** Compact JSON by removing whitespace

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -s, --sort-keys | bool | false | sort object keys |

---

### json stats

**Category:** Other

**Usage:** `omni json stats [FILE] [flags]`

**Description:** Show statistics about JSON data

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### json tocsv

**Category:** Other

**Usage:** `omni json tocsv [FILE] [flags]`

**Description:** Convert JSON array to CSV

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --delimiter | string | , | field delimiter |
| --no-header | bool | false | don't include header row |
| --no-quotes | bool | false | don't quote fields |

---

### json tostruct

**Category:** Other

**Usage:** `omni json tostruct [FILE] [flags]`

**Description:** Convert JSON to Go struct definition

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --inline | bool | false | inline nested structs |
| -n, --name | string | Root | struct name |
| --omitempty | bool | false | add omitempty to all fields |
| -p, --package | string | main | package name |

---

### json toxml

**Category:** Other

**Usage:** `omni json toxml [FILE] [flags]`

**Description:** Convert JSON to XML

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --attr-prefix | string | - | prefix for attributes |
| -i, --indent | string |    | indentation string |
| --item-tag | string | item | tag for array items |
| -r, --root | string | root | root element name |

---

### json toyaml

**Category:** Other

**Usage:** `omni json toyaml [FILE]`

**Description:** Convert JSON to YAML

---

### json validate

**Category:** Other

**Usage:** `omni json validate [FILE]... [flags]`

**Description:** Check if input is valid JSON

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output result as JSON |

---

### jwt

**Category:** Other

**Usage:** `omni jwt`

**Description:** JWT (JSON Web Token) utilities

**Subcommands:** `decode`

---

### jwt decode

**Category:** Other

**Usage:** `omni jwt decode [TOKEN] [flags]`

**Description:** Decode and inspect a JWT token

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -H, --header | bool | false | show only header |
| --json | bool | false | output as JSON |
| -p, --payload | bool | false | show only payload |
| --raw | bool | false | output raw JSON without formatting |

---

### kconfig

**Category:** Other

**Usage:** `omni kconfig`

**Description:** Show kubeconfig info

---

### kcs

**Category:** Other

**Usage:** `omni kcs [context]`

**Description:** Switch kubectl context

---

### kdebug

**Category:** Other

**Usage:** `omni kdebug <pod> [flags]`

**Description:** Debug pod with ephemeral container

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --image | string | - | Debug container image (default: busybox) |
| -n, --namespace | string | - | Namespace |

---

### kdp

**Category:** Other

**Usage:** `omni kdp <selector> [flags]`

**Description:** Delete pods by selector

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --force | bool | false | Force delete |
| -n, --namespace | string | - | Namespace |

---

### kdrain

**Category:** Other

**Usage:** `omni kdrain <node> [flags]`

**Description:** Drain node for maintenance

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --delete-emptydir | bool | false | Delete emptydir data |
| --ignore-daemonsets | bool | false | Ignore daemonsets |

---

### keb

**Category:** Other

**Usage:** `omni keb <pod> [flags]`

**Description:** Exec into pod with bash

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --container | string | - | Container name |
| -n, --namespace | string | - | Namespace |

---

### kga

**Category:** Other

**Usage:** `omni kga [flags]`

**Description:** Get all resources in namespace

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -A, --all-namespaces | bool | false | All namespaces |
| -n, --namespace | string | - | Namespace |

---

### kge

**Category:** Other

**Usage:** `omni kge [flags]`

**Description:** Get events sorted by time

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -A, --all-namespaces | bool | false | All namespaces |
| -n, --namespace | string | - | Namespace |

---

### kill

**Category:** System Info

**Usage:** `omni kill [OPTION]... PID... [flags]`

**Description:** Send a signal to a process

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -j, --json | bool | false | output as JSON |
| -l, --list | bool | false | list signal names |
| -s, --signal | string | - | specify the signal to be sent |
| -v, --verbose | bool | false | report successful signals |

---

### klf

**Category:** Other

**Usage:** `omni klf <pod> [flags]`

**Description:** Follow pod logs with timestamps

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --container | string | - | Container name |
| -n, --namespace | string | - | Namespace |
| --tail | int | 0 | Lines to show from end of logs |

---

### kns

**Category:** Other

**Usage:** `omni kns [namespace]`

**Description:** Switch default namespace

---

### kpf

**Category:** Other

**Usage:** `omni kpf <pod|svc/name> <local:remote> [flags]`

**Description:** Quick port forward

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --namespace | string | - | Namespace |

---

### krr

**Category:** Other

**Usage:** `omni krr <deployment> [flags]`

**Description:** Rollout restart deployment

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --namespace | string | - | Namespace |

---

### krun

**Category:** Other

**Usage:** `omni krun <name> --image=<image> [-- command] [flags]`

**Description:** Run a one-off pod

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --image | string | - | Container image (required) |
| -n, --namespace | string | - | Namespace |

---

### kscale

**Category:** Other

**Usage:** `omni kscale <deployment> <replicas> [flags]`

**Description:** Scale deployment

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --namespace | string | - | Namespace |

---

### ksuid

**Category:** Other

**Usage:** `omni ksuid [OPTION]... [flags]`

**Description:** Generate K-Sortable Unique IDentifiers

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | generate N KSUIDs |
| --json | bool | false | output as JSON |

---

### ktn

**Category:** Other

**Usage:** `omni ktn [flags]`

**Description:** Top nodes by resource usage

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --sort-by | string | - | Sort by (cpu or memory) |

---

### ktp

**Category:** Other

**Usage:** `omni ktp [flags]`

**Description:** Top pods by resource usage

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -A, --all-namespaces | bool | false | All namespaces |
| -n, --namespace | string | - | Namespace |
| --sort-by | string | - | Sort by (cpu or memory) |

---

### kubectl

**Category:** Other

**Usage:** `omni kubectl`

**Description:** Kubernetes CLI

---

### kwp

**Category:** Other

**Usage:** `omni kwp [flags]`

**Description:** Watch pods continuously

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -A, --all-namespaces | bool | false | All namespaces |
| -n, --namespace | string | - | Namespace |
| -l, --selector | string | - | Label selector |

---

### less

**Category:** Other

**Usage:** `omni less [OPTION]... [FILE] [flags]`

**Description:** View file contents with scrolling

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -N, --LINE-NUMBERS | bool | false | show line numbers |
| -S, --chop-long-lines | bool | false | truncate long lines |
| -i, --ignore-case | bool | false | case-insensitive search |
| -X, --no-init | bool | false | don't clear screen on start |
| -F, --quit-if-one-screen | bool | false | quit if content fits on one screen |
| -R, --raw-control-chars | bool | false | show raw control characters |

---

### lint

**Category:** Tooling

**Usage:** `omni lint [OPTION]... [FILE|DIR]... [flags]`

**Description:** Check Taskfiles for portability issues

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --fix | bool | false | auto-fix issues where possible |
| -f, --format | string | text | output format (text, json) |
| -q, --quiet | bool | false | only show errors, not warnings |
| --strict | bool | false | enable strict mode (more warnings become errors) |

---

### ln

**Category:** File Operations

**Usage:** `omni ln [OPTION]... TARGET LINK_NAME [flags]`

**Description:** Make links between files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --backup | bool | false | make a backup of each existing destination file |
| -f, --force | bool | false | remove existing destination files |
| -n, --no-dereference | bool | false | treat LINK_NAME as a normal file if it is a symlink |
| -r, --relative | bool | false | create symbolic links relative to link location |
| -s, --symbolic | bool | false | make symbolic links instead of hard links |
| -v, --verbose | bool | false | print name of each linked file |

---

### loc

**Category:** Other

**Usage:** `omni loc [PATH]... [flags]`

**Description:** Count lines of code by language

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -e, --exclude | stringSlice | [] | directories to exclude |
| --hidden | bool | false | include hidden files |
| --json | bool | false | output as JSON |

---

### logger

**Category:** Tooling

**Usage:** `omni logger [flags]`

**Description:** Configure omni command logging

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --disable | bool | false | Disable logging (unset environment variables) |
| -p, --path | string | - | Path to the log file |
| -s, --status | bool | false | Show current logging status |
| -v, --viewer | bool | false | View all log files sorted by time |

---

### ls

**Category:** Core

**Usage:** `omni ls [file...] [flags]`

**Description:** List directory contents

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | false | do not ignore entries starting with . |
| -A, --almost-all | bool | false | do not list implied . and .. |
| -F, --classify | bool | false | append indicator (*/=>@|) to entries |
| -d, --directory | bool | false | list directories themselves, not their contents |
| -H, --human-readable | bool | false | with -l, print sizes in human readable format |
| -i, --inode | bool | false | print the index number of each file |
| --json | bool | false | output in JSON format |
| -l, --long | bool | false | use a long listing format |
| -U, --no-sort | bool | false | do not sort; list entries in directory order |
| -1, --one | bool | false | list one file per line |
| -R, --recursive | bool | false | list subdirectories recursively |
| -r, --reverse | bool | false | reverse order while sorting |
| -S, --size | bool | false | sort by file size, largest first |
| -t, --time | bool | false | sort by modification time, newest first |

---

### md5sum

**Category:** Hash & Encoding

**Usage:** `omni md5sum [OPTION]... [FILE]... [flags]`

**Description:** Compute and check MD5 message digest

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --binary | bool | false | read in binary mode |
| -c, --check | bool | false | read checksums from FILE and check them |
| --quiet | bool | false | don't print OK for verified files |
| --status | bool | false | don't output anything, use status code |
| -w, --warn | bool | false | warn about improperly formatted lines |

---

### mkdir

**Category:** File Operations

**Usage:** `omni mkdir [directory...] [flags]`

**Description:** Create directories

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -p, --parents | bool | false | no error if existing, make parent directories as needed |

---

### more

**Category:** Other

**Usage:** `omni more [OPTION]... [FILE] [flags]`

**Description:** View file contents page by page

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --line-numbers | bool | false | show line numbers |

---

### move

**Category:** Other

**Usage:** `omni move`

**Description:** Alias for mv

---

### mv

**Category:** File Operations

**Usage:** `omni mv [source...] [destination]`

**Description:** Move (rename) files

---

### nanoid

**Category:** Other

**Usage:** `omni nanoid [OPTION]... [flags]`

**Description:** Generate compact, URL-safe unique IDs

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --alphabet | string | - | custom alphabet |
| -n, --count | int | 1 | generate N NanoIDs |
| --json | bool | false | output as JSON |
| -l, --length | int | 21 | length of NanoID |

---

### nl

**Category:** Text Processing

**Usage:** `omni nl [OPTION]... [FILE]... [flags]`

**Description:** Number lines of files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --body-numbering | string | t | use STYLE for numbering body lines |
| --json | bool | false | output as JSON |
| -i, --line-increment | int | 1 | line number increment |
| -n, --number-format | string | rn | insert line numbers according to FORMAT |
| -s, --number-separator | string | 	 | add STRING after line number |
| -w, --number-width | int | 6 | use N columns for line numbers |
| -v, --starting-line-number | int | 1 | first line number |

---

### paste

**Category:** Text Processing

**Usage:** `omni paste [OPTION]... [FILE]... [flags]`

**Description:** Merge lines of files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --delimiters | string | - | reuse characters from LIST instead of TABs |
| -s, --serial | bool | false | paste one file at a time instead of in parallel |
| -z, --zero-terminated | bool | false | line delimiter is NUL, not newline |

---

### pipe

**Category:** Other

**Usage:** `omni pipe {CMD}, {CMD}, ... | CMD | CMD [flags]`

**Description:** Chain omni commands without shell pipes

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output result as JSON with metadata |
| -s, --sep | string | | | command separator |
| --var | string | OUT | variable name for output substitution (default: OUT) |
| -v, --verbose | bool | false | show intermediate results |

---

### pipeline

**Category:** Other

**Usage:** `omni pipeline`

**Description:** Internal streaming pipeline engine

---

### printf

**Category:** Other

**Usage:** `omni printf FORMAT [ARG...] [flags]`

**Description:** Format and print data

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --no-newline | bool | false | do not append a trailing newline |

---

### ps

**Category:** System Info

**Usage:** `omni ps [OPTION]... [flags]`

**Description:** Report a snapshot of current processes

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | false | show processes for all users |
| -f, --full | bool | false | full-format listing |
| --go | bool | false | show only Go processes |
| -j, --json | bool | false | output as JSON |
| -l, --long | bool | false | long format |
| --no-headers | bool | false | don't print header line |
| -p, --pid | int | 0 | show process with specified PID |
| --sort | string | - | sort by column (pid, cpu, mem, time) |
| -u, --user | string | - | show processes for specified user |

---

### pwd

**Category:** Core

**Usage:** `omni pwd [flags]`

**Description:** Print working directory

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### random

**Category:** Security

**Usage:** `omni random [OPTION]... [flags]`

**Description:** Generate random values

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --charset | string | - | custom character set |
| -n, --count | int | 1 | number of values to generate |
| --json | bool | false | output as JSON |
| -l, --length | int | 16 | length of random strings |
| --max | int64 | 100 | maximum value for integers |
| --min | int64 | 0 | minimum value for integers |
| -s, --separator | string | 
 | separator between values |
| -t, --type | string | string | type: int, float, string, alpha, hex, password, bytes, custom |

---

### readlink

**Category:** Core

**Usage:** `omni readlink [OPTION]... FILE... [flags]`

**Description:** Print resolved symbolic links or canonical file names

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -f, --canonicalize | bool | false | canonicalize by following every symlink |
| -e, --canonicalize-existing | bool | false | canonicalize, all components must exist |
| -m, --canonicalize-missing | bool | false | canonicalize without requirements on existence |
| --json | bool | false | output as JSON |
| -n, --no-newline | bool | false | do not output the trailing delimiter |
| -q, --quiet | bool | false | suppress most error messages |
| -s, --silent | bool | false | suppress most error messages |
| -v, --verbose | bool | false | report error messages |
| -z, --zero | bool | false | end each output line with NUL, not newline |

---

### realpath

**Category:** Core

**Usage:** `omni realpath [path...] [flags]`

**Description:** Print the resolved path

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### remove

**Category:** Other

**Usage:** `omni remove`

**Description:** Alias for rm

---

### rev

**Category:** Text Processing

**Usage:** `omni rev [FILE]... [flags]`

**Description:** Reverse lines characterwise

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### rg

**Category:** Other

**Usage:** `omni rg [OPTIONS] PATTERN [PATH...] [flags]`

**Description:** Recursively search for a pattern (ripgrep-style)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -A, --after-context | int | 0 | show N lines after match |
| -B, --before-context | int | 0 | show N lines before match |
| -b, --byte-offset | bool | false | show byte offset of each line (not yet implemented) |
| --color | string | auto | when to use colors: auto, always, never |
| --colors | stringSlice | [] | custom color specification (e.g., 'path:fg:magenta') |
| --column | bool | false | show column numbers |
| -C, --context | int | 0 | show N lines before and after match |
| -c, --count | bool | false | only show count of matches per file |
| -l, --files-with-matches | bool | false | only show file names with matches |
| -F, --fixed-strings | bool | false | treat pattern as literal string |
| -L, --follow | bool | false | follow symbolic links |
| -g, --glob | stringSlice | [] | include/exclude files matching GLOB (prefix with ! to exclude) |
| --hidden | bool | false | search hidden files and directories |
| -i, --ignore-case | bool | false | case insensitive search |
| -v, --invert-match | bool | false | show non-matching lines |
| --json | bool | false | output results as JSON |
| --json-stream | bool | false | output results as streaming NDJSON (one JSON object per line) |
| -n, --line-number | bool | false | show line numbers |
| -m, --max-count | int | 0 | limit matches per file |
| --max-depth | int | 0 | limit directory traversal depth |
| -U, --multiline | bool | false | enable multiline matching |
| -H, --no-heading | bool | false | don't group matches by file name |
| --no-ignore | bool | false | don't respect gitignore files |
| -o, --only-matching | bool | false | show only matching part of line |
| --passthru | bool | false | show all lines, highlighting matches |
| -q, --quiet | bool | false | quiet mode, exit on first match |
| -r, --replace | string | - | replace matches with STRING |
| -S, --smart-case | bool | false | smart case (insensitive if pattern is all lowercase) |
| --stats | bool | false | show search statistics |
| -j, --threads | int | 0 | number of worker threads (default: CPU count) |
| --trim | bool | false | trim leading/trailing whitespace from each line |
| -t, --type | stringSlice | [] | only search files of TYPE (go, js, py, etc.) |
| -T, --type-not | stringSlice | [] | exclude files of TYPE |
| -w, --word-regexp | bool | false | only match whole words |

---

### rm

**Category:** File Operations

**Usage:** `omni rm [file...] [flags]`

**Description:** Remove files or directories

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -f, --force | bool | false | ignore nonexistent files and arguments, never prompt |
| --no-preserve-root | bool | false | do not treat protected paths specially (dangerous) |
| -r, --recursive | bool | false | remove directories and their contents recursively |

---

### rmdir

**Category:** File Operations

**Usage:** `omni rmdir [directory...] [flags]`

**Description:** Remove empty directories

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --no-preserve-root | bool | false | do not treat protected paths specially (dangerous) |

---

### sed

**Category:** Text Processing

**Usage:** `omni sed [OPTION]... {script} [FILE]... [flags]`

**Description:** Stream editor for filtering and transforming text

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -e, --expression | stringSlice | [] | add the script to the commands to be executed |
| -i, --in-place | bool | false | edit files in place |
| --in-place-suffix | string | - | backup suffix for in-place edit |
| -n, --quiet | bool | false | suppress automatic printing of pattern space |
| -r, --r | bool | false | use extended regular expressions (alias) |
| -E, --regexp-extended | bool | false | use extended regular expressions |

---

### seq

**Category:** Utilities

**Usage:** `omni seq [OPTION]... LAST or seq [OPTION]... FIRST LAST or seq [OPTION]... FIRST INCREMENT LAST [flags]`

**Description:** Print a sequence of numbers

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -w, --equal-width | bool | false | equalize width with leading zeros |
| -f, --format | string | - | use printf style FORMAT |
| --json | bool | false | output as JSON |
| -s, --separator | string | - | use STRING to separate numbers |

---

### sha256sum

**Category:** Hash & Encoding

**Usage:** `omni sha256sum [OPTION]... [FILE]... [flags]`

**Description:** Compute and check SHA256 message digest

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --binary | bool | false | read in binary mode |
| -c, --check | bool | false | read checksums from FILE and check them |
| --quiet | bool | false | don't print OK for verified files |
| --status | bool | false | don't output anything, use status code |
| -w, --warn | bool | false | warn about improperly formatted lines |

---

### sha512sum

**Category:** Hash & Encoding

**Usage:** `omni sha512sum [OPTION]... [FILE]... [flags]`

**Description:** Compute and check SHA512 message digest

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --binary | bool | false | read in binary mode |
| -c, --check | bool | false | read checksums from FILE and check them |
| --quiet | bool | false | don't print OK for verified files |
| --status | bool | false | don't output anything, use status code |
| -w, --warn | bool | false | warn about improperly formatted lines |

---

### shuf

**Category:** Text Processing

**Usage:** `omni shuf [OPTION]... [FILE] [flags]`

**Description:** Generate random permutations

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -e, --echo | bool | false | treat each ARG as an input line |
| -n, --head-count | int | 0 | output at most COUNT lines |
| -i, --input-range | string | - | treat each number LO through HI as an input line |
| --json | bool | false | output as JSON |
| -r, --repeat | bool | false | output lines can be repeated |
| -z, --zero-terminated | bool | false | line delimiter is NUL |

---

### sleep

**Category:** Utilities

**Usage:** `omni sleep NUMBER[SUFFIX]...`

**Description:** Delay for a specified amount of time

---

### snowflake

**Category:** Other

**Usage:** `omni snowflake [OPTION]... [flags]`

**Description:** Generate Twitter Snowflake-style IDs

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | generate N Snowflake IDs |
| --json | bool | false | output as JSON |
| -w, --worker | int64 | 0 | worker ID (0-1023) |

---

### sort

**Category:** Text Processing

**Usage:** `omni sort [option]... [file]... [flags]`

**Description:** Sort lines of text files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --check | bool | false | check for sorted input; do not sort |
| -d, --dictionary-order | bool | false | consider only blanks and alphanumeric characters |
| -t, --field-separator | string | - | use SEP instead of non-blank to blank transition |
| -f, --ignore-case | bool | false | fold lower case to upper case characters |
| -b, --ignore-leading-blanks | bool | false | ignore leading blanks |
| --json | bool | false | output as JSON |
| -k, --key | string | - | sort via a key |
| -n, --numeric-sort | bool | false | compare according to string numerical value |
| -o, --output | string | - | write result to FILE instead of standard output |
| -r, --reverse | bool | false | reverse the result of comparisons |
| -s, --stable | bool | false | stabilize sort by disabling last-resort comparison |
| -u, --unique | bool | false | with -c, check for strict ordering; without -c, output only the first of an equal run |

---

### split

**Category:** Text Processing

**Usage:** `omni split [OPTION]... [FILE [PREFIX]] [flags]`

**Description:** Split a file into pieces

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --bytes | string | - | put SIZE bytes per output file |
| -l, --lines | int | 0 | put NUMBER lines per output file |
| -d, --numeric-suffixes | bool | false | use numeric suffixes |
| -a, --suffix-length | int | 2 | generate suffixes of length N |
| --verbose | bool | false | print diagnostic for each output file |

---

### sql

**Category:** Other

**Usage:** `omni sql [FILE] [flags]`

**Description:** SQL utilities (format, minify, validate)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| -u, --uppercase | bool | true | uppercase keywords |

**Subcommands:** `fmt`, `minify`, `validate`

---

### sql fmt

**Category:** Other

**Usage:** `omni sql fmt [FILE] [flags]`

**Description:** Format/beautify SQL

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --dialect | string | generic | SQL dialect (mysql, postgres, sqlite, generic) |
| -i, --indent | string |    | indentation string |
| -u, --uppercase | bool | true | uppercase keywords |

---

### sql minify

**Category:** Other

**Usage:** `omni sql minify [FILE]`

**Description:** Minify SQL

---

### sql validate

**Category:** Other

**Usage:** `omni sql validate [FILE] [flags]`

**Description:** Validate SQL syntax

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --dialect | string | generic | SQL dialect |
| --json | bool | false | output as JSON |

---

### sqlite

**Category:** Database

**Usage:** `omni sqlite`

**Description:** SQLite database management

**Subcommands:** `check`, `columns`, `dump`, `import`, `indexes`, `query`, `schema`, `stats`, `tables`, `vacuum`

---

### sqlite check

**Category:** Other

**Usage:** `omni sqlite check <database>`

**Description:** Verify database integrity

---

### sqlite columns

**Category:** Other

**Usage:** `omni sqlite columns <database> <table>`

**Description:** Show table columns

---

### sqlite dump

**Category:** Other

**Usage:** `omni sqlite dump <database> [table]`

**Description:** Export database as SQL

---

### sqlite import

**Category:** Other

**Usage:** `omni sqlite import <database> <sql-file>`

**Description:** Import SQL file into database

---

### sqlite indexes

**Category:** Other

**Usage:** `omni sqlite indexes <database>`

**Description:** List all indexes

---

### sqlite query

**Category:** Other

**Usage:** `omni sqlite query <database> <sql> [flags]`

**Description:** Execute SQL query

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -H, --header | bool | false | show column headers |
| --log-data | bool | false | include result data in logs (use with caution for large results) |
| -s, --separator | string | | | column separator |

---

### sqlite schema

**Category:** Other

**Usage:** `omni sqlite schema <database> [table]`

**Description:** Show table schema

---

### sqlite stats

**Category:** Other

**Usage:** `omni sqlite stats <database>`

**Description:** Display database statistics

---

### sqlite tables

**Category:** Other

**Usage:** `omni sqlite tables <database>`

**Description:** List all tables in the database

---

### sqlite vacuum

**Category:** Other

**Usage:** `omni sqlite vacuum <database>`

**Description:** Optimize database

---

### stat

**Category:** File Operations

**Usage:** `omni stat [file...] [flags]`

**Description:** Display file or file system status

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | Output in JSON format |

---

### strings

**Category:** Text Processing

**Usage:** `omni strings [OPTION]... [FILE]... [flags]`

**Description:** Print the printable strings in files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --bytes | int | 4 | minimum string length |
| --json | bool | false | output as JSON |
| -t, --radix | string | - | print offset (d/o/x) |

---

### tac

**Category:** Text Processing

**Usage:** `omni tac [OPTION]... [FILE]... [flags]`

**Description:** Concatenate and print files in reverse

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --before | bool | false | attach the separator before instead of after |
| --json | bool | false | output as JSON |
| -r, --regex | bool | false | interpret the separator as a regular expression |
| -s, --separator | string | - | use STRING as the separator instead of newline |

---

### tagfixer

**Category:** Other

**Usage:** `omni tagfixer [PATH] [flags]`

**Description:** Fix and standardize Go struct tags

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --analyze | bool | false | analyze mode - generate report only |
| -c, --case | string | camel | target case type (camel, pascal, snake, kebab) |
| -d, --dry-run | bool | false | preview changes without writing |
| --json | bool | false | output as JSON |
| -r, --recursive | bool | true | process directories recursively |
| -t, --tags | string | json | comma-separated tag types to fix |
| -v, --verbose | bool | false | verbose output |

**Subcommands:** `analyze`

---

### tagfixer analyze

**Category:** Other

**Usage:** `omni tagfixer analyze [PATH] [flags]`

**Description:** Analyze struct tag usage patterns

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |
| -r, --recursive | bool | true | process directories recursively |
| -t, --tags | string | json,yaml,xml | comma-separated tag types to analyze |
| -v, --verbose | bool | false | verbose output (show all files) |

---

### tail

**Category:** Text Processing

**Usage:** `omni tail [option]... [file]... [flags]`

**Description:** Output the last part of files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --bytes | int | 0 | output the last NUM bytes |
| -f, --follow | bool | false | output appended data as the file grows |
| --json | bool | false | output as JSON |
| -n, --lines | int | 10 | output the last NUM lines, instead of the last 10 |
| -q, --quiet | bool | false | never output headers giving file names |
| --sleep-interval | duration | 1s | with -f, sleep for approximately N seconds between iterations |
| -v, --verbose | bool | false | always output headers giving file names |

---

### tar

**Category:** Archive

**Usage:** `omni tar [OPTION]... [FILE]... [flags]`

**Description:** Create, extract, or list archive files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --create | bool | false | create a new archive |
| -C, --directory | string | - | change to directory DIR |
| -x, --extract | bool | false | extract files from an archive |
| -f, --file | string | - | use archive file ARCHIVE |
| -z, --gzip | bool | false | filter through gzip |
| --json | bool | false | output as JSON (for list mode) |
| -t, --list | bool | false | list the contents of an archive |
| --strip-components | int | 0 | strip N leading path components |
| -v, --verbose | bool | false | verbosely list files processed |

---

### task

**Category:** Other

**Usage:** `omni task [TASK...] [flags]`

**Description:** Run tasks defined in Taskfile.yml

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --allow-external | bool | false | allow external (non-omni) commands |
| -d, --dir | string | - | working directory |
| --dry-run | bool | false | print commands without executing |
| -f, --force | bool | false | force run even if up-to-date |
| -l, --list | bool | false | list available tasks |
| -s, --silent | bool | false | suppress output |
| --summary | bool | false | show task summary |
| -t, --taskfile | string | - | path to Taskfile.yml |
| -v, --verbose | bool | false | verbose output |

---

### terraform

**Category:** Other

**Usage:** `omni terraform`

**Description:** Terraform CLI

**Subcommands:** `apply`, `console`, `destroy`, `fmt`, `get`, `graph`, `import`, `init`, `output`, `plan`, `providers`, `refresh`, `show`, `state`, `taint`, `test`, `untaint`, `validate`, `version`, `workspace`

---

### terraform apply

**Category:** Other

**Usage:** `omni terraform apply [plan-file] [flags]`

**Description:** Apply changes to infrastructure

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --auto-approve | bool | false | Skip interactive approval |
| --var | stringToString | [] | Set variables |
| --var-file | stringSlice | [] | Variable files |

---

### terraform console

**Category:** Other

**Usage:** `omni terraform console`

**Description:** Interactive console

---

### terraform destroy

**Category:** Other

**Usage:** `omni terraform destroy [flags]`

**Description:** Destroy managed infrastructure

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --auto-approve | bool | false | Skip interactive approval |
| --var | stringToString | [] | Set variables |
| --var-file | stringSlice | [] | Variable files |

---

### terraform fmt

**Category:** Other

**Usage:** `omni terraform fmt [flags]`

**Description:** Format configuration files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --check | bool | false | Check if formatted |
| --diff | bool | false | Display diffs |
| --recursive | bool | false | Process subdirectories |

---

### terraform get

**Category:** Other

**Usage:** `omni terraform get [flags]`

**Description:** Download modules

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --update | bool | false | Update modules |

---

### terraform graph

**Category:** Other

**Usage:** `omni terraform graph [flags]`

**Description:** Generate dependency graph

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --draw-cycles | bool | false | Draw cycles |
| --plan | string | - | Plan file |

---

### terraform import

**Category:** Other

**Usage:** `omni terraform import <address> <id> [flags]`

**Description:** Import existing infrastructure

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --var | stringToString | [] | Set variables |
| --var-file | stringSlice | [] | Variable files |

---

### terraform init

**Category:** Other

**Usage:** `omni terraform init [flags]`

**Description:** Initialize Terraform working directory

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --reconfigure | bool | false | Reconfigure backend |
| --upgrade | bool | false | Upgrade modules and plugins |

---

### terraform output

**Category:** Other

**Usage:** `omni terraform output [name] [flags]`

**Description:** Show output values

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | JSON output |

---

### terraform plan

**Category:** Other

**Usage:** `omni terraform plan [flags]`

**Description:** Create execution plan

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --destroy | bool | false | Create destroy plan |
| -o, --out | string | - | Write plan to file |
| --var | stringToString | [] | Set variables |
| --var-file | stringSlice | [] | Variable files |

---

### terraform providers

**Category:** Other

**Usage:** `omni terraform providers`

**Description:** Show provider information

---

### terraform refresh

**Category:** Other

**Usage:** `omni terraform refresh [flags]`

**Description:** Refresh state

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --var | stringToString | [] | Set variables |
| --var-file | stringSlice | [] | Variable files |

---

### terraform show

**Category:** Other

**Usage:** `omni terraform show [plan-file] [flags]`

**Description:** Show plan or state

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | JSON output |

---

### terraform state

**Category:** Other

**Usage:** `omni terraform state`

**Description:** State management commands

**Subcommands:** `list`, `mv`, `rm`, `show`

---

### terraform state list

**Category:** Other

**Usage:** `omni terraform state list [addresses...]`

**Description:** List resources in state

---

### terraform state mv

**Category:** File Operations

**Usage:** `omni terraform state mv <source> <destination> [flags]`

**Description:** Move resource in state

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | Dry run |

---

### terraform state rm

**Category:** File Operations

**Usage:** `omni terraform state rm <addresses...> [flags]`

**Description:** Remove resources from state

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | Dry run |

---

### terraform state show

**Category:** Other

**Usage:** `omni terraform state show <address>`

**Description:** Show a resource in state

---

### terraform taint

**Category:** Other

**Usage:** `omni terraform taint <address>`

**Description:** Mark resource for recreation

---

### terraform test

**Category:** Utilities

**Usage:** `omni terraform test`

**Description:** Run tests

---

### terraform untaint

**Category:** Other

**Usage:** `omni terraform untaint <address>`

**Description:** Remove taint from resource

---

### terraform validate

**Category:** Other

**Usage:** `omni terraform validate`

**Description:** Validate configuration

---

### terraform version

**Category:** Tooling

**Usage:** `omni terraform version`

**Description:** Show Terraform version

---

### terraform workspace

**Category:** Other

**Usage:** `omni terraform workspace`

**Description:** Workspace management commands

**Subcommands:** `delete`, `list`, `new`, `select`, `show`

---

### terraform workspace delete

**Category:** Other

**Usage:** `omni terraform workspace delete <name> [flags]`

**Description:** Delete workspace

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --force | bool | false | Force delete |

---

### terraform workspace list

**Category:** Other

**Usage:** `omni terraform workspace list`

**Description:** List workspaces

---

### terraform workspace new

**Category:** Other

**Usage:** `omni terraform workspace new <name>`

**Description:** Create new workspace

---

### terraform workspace select

**Category:** Other

**Usage:** `omni terraform workspace select <name>`

**Description:** Select workspace

---

### terraform workspace show

**Category:** Other

**Usage:** `omni terraform workspace show`

**Description:** Show current workspace

---

### testcheck

**Category:** Other

**Usage:** `omni testcheck [directory] [flags]`

**Description:** Check test coverage for Go packages

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | false | show all packages (default shows only missing) |
| -j, --json | bool | false | output as JSON |
| -s, --summary | bool | false | show only summary |
| -v, --verbose | bool | false | show test file names |

---

### time

**Category:** Utilities

**Usage:** `omni time`

**Description:** Time a simple command or give resource usage

---

### toml

**Category:** Other

**Usage:** `omni toml`

**Description:** TOML utilities

**Subcommands:** `fmt`, `validate`

---

### toml fmt

**Category:** Other

**Usage:** `omni toml fmt [FILE] [flags]`

**Description:** Format TOML

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | int | 2 | indentation width |

---

### toml validate

**Category:** Other

**Usage:** `omni toml validate [FILE...] [flags]`

**Description:** Validate TOML syntax

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### top

**Category:** Other

**Usage:** `omni top [flags]`

**Description:** Display system processes sorted by resource usage

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --go | bool | false | show only Go processes |
| -j, --json | bool | false | output as JSON |
| -n, --num | int | 10 | number of processes to show |
| --sort | string | cpu | sort by column: cpu, mem, pid |

---

### touch

**Category:** File Operations

**Usage:** `omni touch [file...]`

**Description:** Update the access and modification times of each FILE to the current time

---

### tr

**Category:** Text Processing

**Usage:** `omni tr [OPTION]... SET1 [SET2] [flags]`

**Description:** Translate or delete characters

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --complement | bool | false | use the complement of SET1 |
| -d, --delete | bool | false | delete characters in SET1, do not translate |
| -s, --squeeze-repeats | bool | false | replace repeated characters with single occurrence |
| -t, --truncate-set1 | bool | false | first truncate SET1 to length of SET2 |

---

### tree

**Category:** Core

**Usage:** `omni tree [path] [flags]`

**Description:** Display directory tree structure

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | false | show hidden files |
| --date | bool | false | show modification dates |
| -d, --depth | int | -1 | maximum depth to scan (-1 for unlimited) |
| --dirs-only | bool | false | show only directories |
| --hash | bool | false | show SHA256 hash for files |
| -i, --ignore | string | - | patterns to ignore (comma-separated) |
| -j, --json | bool | false | output as JSON format |
| --no-color | bool | false | disable colored output |
| --no-dir-slash | bool | false | don't add trailing slash to directory names |
| --size | bool | false | show file sizes |
| -s, --stats | bool | false | show statistics |

---

### ulid

**Category:** Other

**Usage:** `omni ulid [OPTION]... [flags]`

**Description:** Generate Universally Unique Lexicographically Sortable Identifiers

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | generate N ULIDs |
| --json | bool | false | output as JSON |
| -l, --lower | bool | false | output in lowercase |

---

### uname

**Category:** System Info

**Usage:** `omni uname [flags]`

**Description:** Print system information

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | false | print all information |
| -i, --hardware-platform | bool | false | print the hardware platform |
| --json | bool | false | output as JSON |
| -s, --kernel-name | bool | false | print the kernel name |
| -r, --kernel-release | bool | false | print the kernel release |
| -v, --kernel-version | bool | false | print the kernel version |
| -m, --machine | bool | false | print the machine hardware name |
| -n, --nodename | bool | false | print the network node hostname |
| -o, --operating-system | bool | false | print the operating system |
| -p, --processor | bool | false | print the processor type |

---

### uniq

**Category:** Text Processing

**Usage:** `omni uniq [option]... [input [output]] [flags]`

**Description:** Report or omit repeated lines

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -D, --all-repeated | bool | false | print all duplicate lines |
| -w, --check-chars | int | 0 | compare no more than N characters in lines |
| -c, --count | bool | false | prefix lines by the number of occurrences |
| -i, --ignore-case | bool | false | ignore differences in case when comparing |
| --json | bool | false | output as JSON |
| -d, --repeated | bool | false | only print duplicate lines, one for each group |
| -s, --skip-chars | int | 0 | avoid comparing the first N characters |
| -f, --skip-fields | int | 0 | avoid comparing the first N fields |
| -u, --unique | bool | false | only print unique lines |
| -z, --zero-terminated | bool | false | line delimiter is NUL, not newline |

---

### unxz

**Category:** Archive

**Usage:** `omni unxz [OPTION]... [FILE]... [flags]`

**Description:** Decompress xz files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -f, --force | bool | false | force overwrite |
| -k, --keep | bool | false | keep original files |
| -c, --stdout | bool | false | write to stdout |
| -v, --verbose | bool | false | verbose mode |

---

### unzip

**Category:** Archive

**Usage:** `omni unzip [OPTION]... ZIPFILE [flags]`

**Description:** Extract files from a zip archive

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --directory | string | - | extract files into directory |
| --json | bool | false | output as JSON (for list mode) |
| -l, --list | bool | false | list contents without extracting |
| --strip-components | int | 0 | strip N leading path components |
| -v, --verbose | bool | false | verbose output |

---

### uptime

**Category:** System Info

**Usage:** `omni uptime [OPTION]... [flags]`

**Description:** Tell how long the system has been running

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |
| -p, --pretty | bool | false | show uptime in pretty format |
| -s, --since | bool | false | system up since |

---

### url

**Category:** Other

**Usage:** `omni url`

**Description:** URL encoding and decoding utilities

**Subcommands:** `decode`, `encode`

---

### url decode

**Category:** Other

**Usage:** `omni url decode [TEXT] [flags]`

**Description:** URL decode text

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --component | bool | false | use query component decoding |
| --json | bool | false | output as JSON |

---

### url encode

**Category:** Other

**Usage:** `omni url encode [TEXT] [flags]`

**Description:** URL encode text

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --component | bool | false | use query component encoding (more aggressive) |
| --json | bool | false | output as JSON |

---

### uuid

**Category:** Security

**Usage:** `omni uuid [OPTION]... [flags]`

**Description:** Generate random UUIDs

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | generate N UUIDs |
| --json | bool | false | output as JSON |
| -x, --no-dashes | bool | false | output without dashes |
| -u, --upper | bool | false | output in uppercase |
| -v, --version | int | 4 | UUID version (4 or 7) |

---

### watch

**Category:** Utilities

**Usage:** `omni watch [OPTION]... COMMAND [flags]`

**Description:** Execute a program periodically, showing output fullscreen

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --beep | bool | false | beep if command has a non-zero exit |
| -g, --chgexit | bool | false | exit when output changes |
| -c, --color | bool | false | interpret ANSI color sequences |
| -d, --differences | bool | false | highlight differences between updates |
| -e, --errexit | bool | false | exit if command has a non-zero exit |
| -n, --interval | float64 | 2 | seconds to wait between updates |
| -t, --no-title | bool | false | turn off header |
| --only-changes | bool | false | only display output when it changes |
| -p, --precise | bool | false | attempt run command in precise intervals |

---

### wc

**Category:** Text Processing

**Usage:** `omni wc [option]... [file]... [flags]`

**Description:** Print newline, word, and byte counts for each file

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --bytes | bool | false | print the byte counts |
| -m, --chars | bool | false | print the character counts |
| --json | bool | false | output in JSON format |
| -l, --lines | bool | false | print the newline counts |
| -L, --max-line-length | bool | false | print the maximum display width |
| -w, --words | bool | false | print the word counts |

---

### which

**Category:** System Info

**Usage:** `omni which [OPTION]... COMMAND... [flags]`

**Description:** Locate a command

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | false | print all matches |
| --json | bool | false | output as JSON |

---

### whoami

**Category:** System Info

**Usage:** `omni whoami [flags]`

**Description:** Print effective username

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### xargs

**Category:** Utilities

**Usage:** `omni xargs [OPTION]... [COMMAND [INITIAL-ARGS]] [flags]`

**Description:** Build and execute command lines from standard input

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -I, --I | string | - | replace occurrences of REPLACE-STR |
| -d, --delimiter | string | - | input items are separated by DELIM |
| -n, --max-args | int | 0 | use at most MAX arguments per command line |
| -P, --max-procs | int | 1 | run at most MAX processes at a time |
| -r, --no-run-if-empty | bool | false | if there are no arguments, do not run |
| -0, --null | bool | false | input items are separated by a null character |
| -t, --verbose | bool | false | print commands before executing them |

---

### xml

**Category:** Other

**Usage:** `omni xml [FILE] [flags]`

**Description:** XML utilities (format, validate, convert)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| -m, --minify | bool | false | minify XML (remove whitespace) |

**Subcommands:** `fmt`, `fromjson`, `tojson`, `validate`

---

### xml fmt

**Category:** Other

**Usage:** `omni xml fmt [FILE] [flags]`

**Description:** Format XML

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| -m, --minify | bool | false | minify XML (remove whitespace) |

---

### xml fromjson

**Category:** Other

**Usage:** `omni xml fromjson [FILE] [flags]`

**Description:** Convert JSON to XML

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --attr-prefix | string | - | prefix for attributes |
| -i, --indent | string |    | indentation string |
| --item-tag | string | item | tag for array items |
| -r, --root | string | root | root element name |

---

### xml tojson

**Category:** Other

**Usage:** `omni xml tojson [FILE] [flags]`

**Description:** Convert XML to JSON

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --attr-prefix | string | - | prefix for attributes in JSON |
| --text-key | string | #text | key for text content |

---

### xml validate

**Category:** Other

**Usage:** `omni xml validate [FILE...] [flags]`

**Description:** Validate XML syntax

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### xxd

**Category:** Other

**Usage:** `omni xxd [OPTIONS] [FILE] [flags]`

**Description:** Make a hex dump or reverse it

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --bits | bool | false | binary digit dump (bits instead of hex) |
| -c, --cols | int | 16 | format <cols> octets per line (default 16) |
| -g, --groupsize | int | 2 | separate output with <bytes> spaces (default 2) |
| -i, --include | bool | false | output in C include file style |
| -l, --len | int | 0 | stop after <len> octets |
| -p, --plain | bool | false | output plain hex dump (no addresses or ASCII) |
| -r, --reverse | bool | false | reverse operation: convert hex dump to binary |
| -s, --seek | int | 0 | start at <seek> bytes offset |
| -u, --uppercase | bool | false | use uppercase hex letters |

---

### xz

**Category:** Archive

**Usage:** `omni xz [OPTION]... [FILE]... [flags]`

**Description:** Compress or decompress xz files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --decompress | bool | false | decompress |
| -f, --force | bool | false | force overwrite |
| -k, --keep | bool | false | keep original files |
| -l, --list | bool | false | list compressed file info |
| -c, --stdout | bool | false | write to stdout |
| -v, --verbose | bool | false | verbose mode |

---

### xzcat

**Category:** Archive

**Usage:** `omni xzcat [FILE]...`

**Description:** Decompress and print xz files

---

### yaml

**Category:** Other

**Usage:** `omni yaml`

**Description:** YAML utilities

**Subcommands:** `fmt`, `tostruct`, `validate`

---

### yaml fmt

**Category:** Other

**Usage:** `omni yaml fmt [FILE] [flags]`

**Description:** Format YAML

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | int | 2 | indentation width |
| --json | bool | false | output as JSON instead of YAML |

---

### yaml tostruct

**Category:** Other

**Usage:** `omni yaml tostruct [FILE] [flags]`

**Description:** Convert YAML to Go struct definition

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --inline | bool | false | inline nested structs |
| -n, --name | string | Root | struct name |
| --omitempty | bool | false | add omitempty to all fields |
| -p, --package | string | main | package name |

---

### yaml validate

**Category:** Other

**Usage:** `omni yaml validate [FILE...] [flags]`

**Description:** Validate YAML syntax

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |
| --strict | bool | false | fail on unknown fields |

---

### yes

**Category:** Core

**Usage:** `omni yes [STRING]...`

**Description:** Output a string repeatedly until killed

---

### yq

**Category:** Data Processing

**Usage:** `omni yq [OPTION]... FILTER [FILE]... [flags]`

**Description:** Command-line YAML processor

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --compact-output | bool | false | compact output |
| -I, --indent | int | 2 | indentation level |
| -n, --null-input | bool | false | don't read any input |
| -o, --output-format | string | yaml | output format (yaml or json) |
| -r, --raw-output | bool | false | output raw strings |

---

### zcat

**Category:** Archive

**Usage:** `omni zcat [FILE]...`

**Description:** Decompress and print gzip files

---

### zip

**Category:** Archive

**Usage:** `omni zip [OPTION]... ZIPFILE FILE... [flags]`

**Description:** Package and compress files into a zip archive

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -C, --directory | string | - | change to directory before adding |
| -r, --recursive | bool | false | recurse into directories |
| -v, --verbose | bool | false | verbose output |

---

## Library API

Command implementations live under internal/cli/ and cannot be imported by external projects (Go's internal package restriction). omni is designed as a CLI tool, not a library. To reuse the logic, fork the repository or use omni as a subprocess.

**Import pattern:** `internal/cli/<command>/<command>.go`

## Architecture

Hexagonal architecture with clear separation between CLI and library layers

```
omni/
  cmd/  # Cobra CLI command definitions (thin wrappers)
  internal/cli/  # Library implementations with Options structs and Run* functions
  internal/flags/  # Feature flags and environment configuration
  internal/logger/  # Structured logging with slog
  main.go  # Entry point calling cmd.Execute()
```
