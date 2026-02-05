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

Commands: `aws`, `brdoc`, `buf`, `case`, `cloud`, `copy`, `cron`, `css`, `csv`, `curl`, `for`, `gbc`, `git`, `gops`, `gqc`, `hex`, `html`, `jwt`, `kconfig`, `kcs`, `kdebug`, `kdp`, `kdrain`, `keb`, `kga`, `kge`, `klf`, `kns`, `kpf`, `krr`, `krun`, `kscale`, `ksuid`, `ktn`, `ktp`, `kubectl`, `kwp`, `less`, `loc`, `more`, `move`, `nanoid`, `pipe`, `pipeline`, `printf`, `remove`, `rg`, `snowflake`, `sql`, `tagfixer`, `task`, `terraform`, `testcheck`, `toml`, `top`, `ulid`, `url`, `vault`, `xml`, `xxd`, `yaml`

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

**Details:**

Generate comprehensive, AI-optimized documentation about the omni application.

Unlike cmdtree which shows a visual tree, aicontext produces a detailed context
document designed for AI consumption, including:

  - Application overview and design principles
  - Command categories with descriptions
  - Complete command reference with all flags
  - Library API usage examples
  - Architecture documentation

Examples:

  # Generate markdown documentation (default)
  omni aicontext

  # Generate JSON for programmatic use
  omni aicontext --json

  # Generate compact output (no examples or long descriptions)
  omni aicontext --compact

  # Filter to a specific category
  omni aicontext --category "Text Processing"

  # Write to a file
  omni aicontext --output context.md

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --category | string | - | filter to specific category |
| --compact | bool | false | omit examples and long descriptions |
| --json | bool | false | output as structured JSON |
| -o, --output | string | - | write to file instead of stdout |

**Examples:**

```bash
omni aicontext
omni aicontext --json
omni aicontext --compact
omni aicontext --category "Text Processing"
omni aicontext --output context.md
```

---

### arch

**Category:** System Info

**Usage:** `omni arch [flags]`

**Description:** Print machine architecture

**Details:**

Print the machine hardware name (similar to uname -m).

Examples:
  omni arch    # x86_64, aarch64, etc.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni arch    # x86_64, aarch64, etc.
```

---

### awk

**Category:** Text Processing

**Usage:** `omni awk [OPTION]... 'program' [FILE]... [flags]`

**Description:** Pattern scanning and processing language

**Details:**

Awk scans each input file for lines that match any of a set of patterns.

This is a simplified subset of AWK supporting:
  - Field access: $0 (whole line), $1, $2, etc.
  - Pattern blocks: BEGIN{}, END{}, /regex/{}
  - Print statements: print, print $1, print $1,$2
  - Built-in variable: NF (number of fields)

  -F fs          use fs for the input field separator
  -v var=value   assign value to variable var

Examples:
  omni awk '{print $1}' file.txt          # print first field
  omni awk -F: '{print $1}' /etc/passwd   # use : as separator
  omni awk '/pattern/{print}' file.txt    # print matching lines
  omni awk 'BEGIN{print "start"} {print} END{print "end"}' file
  omni awk '{print $1, $NF}' file.txt     # print first and last field

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -v, --assign | stringSlice | [] | assign value to variable (var=value) |
| -F, --field-separator | string | - | use FS for the input field separator |

**Examples:**

```bash
omni awk '{print $1}' file.txt          # print first field
omni awk -F: '{print $1}' /etc/passwd   # use : as separator
omni awk '/pattern/{print}' file.txt    # print matching lines
omni awk 'BEGIN{print "start"} {print} END{print "end"}' file
omni awk '{print $1, $NF}' file.txt     # print first and last field
```

---

### aws

**Category:** Other

**Usage:** `omni aws`

**Description:** AWS CLI operations

**Details:**

AWS CLI operations for core services.

Supported services:
  s3    - S3 bucket and object operations
  ec2   - EC2 instance operations
  iam   - IAM user, role, and policy operations
  sts   - STS identity and credential operations
  ssm   - SSM Parameter Store operations

Configuration:
  AWS credentials are loaded from standard AWS SDK sources:
  - Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
  - Shared credentials file (~/.aws/credentials)
  - IAM role (when running on EC2)

Examples:
  # Get caller identity
  omni aws sts get-caller-identity

  # List S3 buckets
  omni aws s3 ls

  # List S3 objects
  omni aws s3 ls s3://my-bucket/prefix/

  # Describe EC2 instances
  omni aws ec2 describe-instances

  # Get SSM parameter
  omni aws ssm get-parameter --name /app/config

**Examples:**

```bash
omni aws sts get-caller-identity
omni aws s3 ls
omni aws s3 ls s3://my-bucket/prefix/
omni aws ec2 describe-instances
omni aws ssm get-parameter --name /app/config
```

**Subcommands:** `ec2`, `iam`, `s3`, `ssm`, `sts`

---

### aws ec2

**Category:** Other

**Usage:** `omni aws ec2`

**Description:** AWS EC2 operations

**Details:**

AWS EC2 instance and resource operations.

**Subcommands:** `describe-instances`, `describe-security-groups`, `describe-vpcs`, `start-instances`, `stop-instances`

---

### aws ec2 describe-instances

**Category:** Other

**Usage:** `omni aws ec2 describe-instances [flags]`

**Description:** Describe EC2 instances

**Details:**

Describes one or more EC2 instances.

Examples:
  omni aws ec2 describe-instances
  omni aws ec2 describe-instances --instance-ids i-1234567890abcdef0
  omni aws ec2 describe-instances --filters "Name=tag:Name,Values=prod-*"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --filters | stringSlice | [] | filters in format 'Name=value,Values=v1,v2' |
| --instance-ids | stringSlice | [] | instance IDs |
| --max-results | int32 | 0 | maximum number of results |

**Examples:**

```bash
omni aws ec2 describe-instances
omni aws ec2 describe-instances --instance-ids i-1234567890abcdef0
omni aws ec2 describe-instances --filters "Name=tag:Name,Values=prod-*"
```

---

### aws ec2 describe-security-groups

**Category:** Other

**Usage:** `omni aws ec2 describe-security-groups [flags]`

**Description:** Describe security groups

**Details:**

Describes one or more security groups.

Examples:
  omni aws ec2 describe-security-groups
  omni aws ec2 describe-security-groups --group-ids sg-1234567890abcdef0

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --filters | stringSlice | [] | filters |
| --group-ids | stringSlice | [] | security group IDs |

**Examples:**

```bash
omni aws ec2 describe-security-groups
omni aws ec2 describe-security-groups --group-ids sg-1234567890abcdef0
```

---

### aws ec2 describe-vpcs

**Category:** Other

**Usage:** `omni aws ec2 describe-vpcs [flags]`

**Description:** Describe VPCs

**Details:**

Describes one or more VPCs.

Examples:
  omni aws ec2 describe-vpcs
  omni aws ec2 describe-vpcs --vpc-ids vpc-1234567890abcdef0

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --filters | stringSlice | [] | filters |
| --vpc-ids | stringSlice | [] | VPC IDs |

**Examples:**

```bash
omni aws ec2 describe-vpcs
omni aws ec2 describe-vpcs --vpc-ids vpc-1234567890abcdef0
```

---

### aws ec2 start-instances

**Category:** Other

**Usage:** `omni aws ec2 start-instances [flags]`

**Description:** Start EC2 instances

**Details:**

Starts one or more stopped instances.

Examples:
  omni aws ec2 start-instances --instance-ids i-1234567890abcdef0
  omni aws ec2 start-instances --instance-ids i-1234,i-5678

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --instance-ids | stringSlice | [] | instance IDs (required) |

**Examples:**

```bash
omni aws ec2 start-instances --instance-ids i-1234567890abcdef0
omni aws ec2 start-instances --instance-ids i-1234,i-5678
```

---

### aws ec2 stop-instances

**Category:** Other

**Usage:** `omni aws ec2 stop-instances [flags]`

**Description:** Stop EC2 instances

**Details:**

Stops one or more running instances.

Examples:
  omni aws ec2 stop-instances --instance-ids i-1234567890abcdef0
  omni aws ec2 stop-instances --instance-ids i-1234 --force

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --force | bool | false | force stop without graceful shutdown |
| --instance-ids | stringSlice | [] | instance IDs (required) |

**Examples:**

```bash
omni aws ec2 stop-instances --instance-ids i-1234567890abcdef0
omni aws ec2 stop-instances --instance-ids i-1234 --force
```

---

### aws iam

**Category:** Other

**Usage:** `omni aws iam`

**Description:** AWS IAM operations

**Details:**

AWS Identity and Access Management (IAM) operations.

**Subcommands:** `get-policy`, `get-role`, `get-user`, `list-policies`, `list-roles`

---

### aws iam get-policy

**Category:** Other

**Usage:** `omni aws iam get-policy [flags]`

**Description:** Get IAM policy information

**Details:**

Retrieves information about the specified managed policy.

Examples:
  omni aws iam get-policy --policy-arn arn:aws:iam::123456789012:policy/MyPolicy

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --policy-arn | string | - | policy ARN (required) |

**Examples:**

```bash
omni aws iam get-policy --policy-arn arn:aws:iam::123456789012:policy/MyPolicy
```

---

### aws iam get-role

**Category:** Other

**Usage:** `omni aws iam get-role [flags]`

**Description:** Get IAM role information

**Details:**

Retrieves information about the specified role, including the role's path, GUID,
ARN, and the role's trust policy.

Examples:
  omni aws iam get-role --role-name MyRole

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --role-name | string | - | role name (required) |

**Examples:**

```bash
omni aws iam get-role --role-name MyRole
```

---

### aws iam get-user

**Category:** Other

**Usage:** `omni aws iam get-user [flags]`

**Description:** Get IAM user information

**Details:**

Retrieves information about the specified IAM user, including the user's creation date,
path, unique ID, and ARN. If no user name is specified, returns information about
the IAM user whose credentials are used to call the operation.

Examples:
  omni aws iam get-user
  omni aws iam get-user --user-name myuser

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --user-name | string | - | user name (optional, defaults to current user) |

**Examples:**

```bash
omni aws iam get-user
omni aws iam get-user --user-name myuser
```

---

### aws iam list-policies

**Category:** Other

**Usage:** `omni aws iam list-policies [flags]`

**Description:** List IAM policies

**Details:**

Lists all the managed policies that are available in your AWS account.

Examples:
  omni aws iam list-policies
  omni aws iam list-policies --scope Local
  omni aws iam list-policies --only-attached

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --max-items | int32 | 0 | maximum number of items |
| --only-attached | bool | false | only show attached policies |
| --path-prefix | string | - | path prefix filter |
| --scope | string | All | scope: All, AWS, Local |

**Examples:**

```bash
omni aws iam list-policies
omni aws iam list-policies --scope Local
omni aws iam list-policies --only-attached
```

---

### aws iam list-roles

**Category:** Other

**Usage:** `omni aws iam list-roles [flags]`

**Description:** List IAM roles

**Details:**

Lists the IAM roles that have the specified path prefix.

Examples:
  omni aws iam list-roles
  omni aws iam list-roles --path-prefix /service-role/

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --max-items | int32 | 0 | maximum number of items |
| --path-prefix | string | - | path prefix filter |

**Examples:**

```bash
omni aws iam list-roles
omni aws iam list-roles --path-prefix /service-role/
```

---

### aws s3

**Category:** Other

**Usage:** `omni aws s3`

**Description:** AWS S3 operations

**Details:**

AWS S3 bucket and object operations.

Examples:
  # List buckets
  omni aws s3 ls

  # List objects in a bucket
  omni aws s3 ls s3://my-bucket/

  # Copy file to S3
  omni aws s3 cp file.txt s3://my-bucket/file.txt

  # Download file from S3
  omni aws s3 cp s3://my-bucket/file.txt ./file.txt

  # Remove object
  omni aws s3 rm s3://my-bucket/file.txt

  # Create bucket
  omni aws s3 mb s3://my-new-bucket

  # Generate presigned URL
  omni aws s3 presign s3://my-bucket/file.txt

**Examples:**

```bash
omni aws s3 ls
omni aws s3 ls s3://my-bucket/
omni aws s3 cp file.txt s3://my-bucket/file.txt
omni aws s3 cp s3://my-bucket/file.txt ./file.txt
omni aws s3 rm s3://my-bucket/file.txt
omni aws s3 mb s3://my-new-bucket
omni aws s3 presign s3://my-bucket/file.txt
```

**Subcommands:** `cp`, `ls`, `mb`, `presign`, `rb`, `rm`

---

### aws s3 cp

**Category:** File Operations

**Usage:** `omni aws s3 cp <SOURCE> <DESTINATION> [flags]`

**Description:** Copy files to/from S3

**Details:**

Copies files between local filesystem and S3, or between S3 locations.

Examples:
  # Upload to S3
  omni aws s3 cp file.txt s3://my-bucket/file.txt

  # Download from S3
  omni aws s3 cp s3://my-bucket/file.txt ./file.txt

  # Copy between S3 locations
  omni aws s3 cp s3://bucket1/file.txt s3://bucket2/file.txt

  # Dry run
  omni aws s3 cp file.txt s3://my-bucket/file.txt --dryrun

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dryrun | bool | false | display operations without executing |
| --quiet | bool | false | suppress output |
| --recursive | bool | false | copy recursively |

**Examples:**

```bash
omni aws s3 cp file.txt s3://my-bucket/file.txt
omni aws s3 cp s3://my-bucket/file.txt ./file.txt
omni aws s3 cp s3://bucket1/file.txt s3://bucket2/file.txt
omni aws s3 cp file.txt s3://my-bucket/file.txt --dryrun
```

---

### aws s3 ls

**Category:** Core

**Usage:** `omni aws s3 ls [S3_URI] [flags]`

**Description:** List S3 objects or buckets

**Details:**

Lists S3 objects in a bucket or all buckets.

Examples:
  omni aws s3 ls
  omni aws s3 ls s3://my-bucket/
  omni aws s3 ls s3://my-bucket/prefix/ --recursive

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --human-readable | bool | false | display file sizes in human-readable format |
| --recursive | bool | false | list recursively |
| --summarize | bool | false | display summary information |

**Examples:**

```bash
omni aws s3 ls
omni aws s3 ls s3://my-bucket/
omni aws s3 ls s3://my-bucket/prefix/ --recursive
```

---

### aws s3 mb

**Category:** Other

**Usage:** `omni aws s3 mb <S3_URI>`

**Description:** Create an S3 bucket

**Details:**

Creates an S3 bucket.

Examples:
  omni aws s3 mb s3://my-new-bucket
  omni aws s3 mb s3://my-new-bucket --region us-west-2

**Examples:**

```bash
omni aws s3 mb s3://my-new-bucket
omni aws s3 mb s3://my-new-bucket --region us-west-2
```

---

### aws s3 presign

**Category:** Other

**Usage:** `omni aws s3 presign <S3_URI> [flags]`

**Description:** Generate a presigned URL

**Details:**

Generates a presigned URL for an S3 object.

Examples:
  omni aws s3 presign s3://my-bucket/file.txt
  omni aws s3 presign s3://my-bucket/file.txt --expires-in 3600

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --expires-in | int | 900 | URL expiration time in seconds (default 15 minutes) |

**Examples:**

```bash
omni aws s3 presign s3://my-bucket/file.txt
omni aws s3 presign s3://my-bucket/file.txt --expires-in 3600
```

---

### aws s3 rb

**Category:** Other

**Usage:** `omni aws s3 rb <S3_URI> [flags]`

**Description:** Remove an S3 bucket

**Details:**

Deletes an S3 bucket. The bucket must be empty unless --force is specified.

Examples:
  omni aws s3 rb s3://my-bucket
  omni aws s3 rb s3://my-bucket --force

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --force | bool | false | delete all objects before removing bucket |

**Examples:**

```bash
omni aws s3 rb s3://my-bucket
omni aws s3 rb s3://my-bucket --force
```

---

### aws s3 rm

**Category:** File Operations

**Usage:** `omni aws s3 rm <S3_URI> [flags]`

**Description:** Remove S3 objects

**Details:**

Deletes objects from S3.

Examples:
  omni aws s3 rm s3://my-bucket/file.txt
  omni aws s3 rm s3://my-bucket/prefix/ --recursive
  omni aws s3 rm s3://my-bucket/prefix/ --recursive --dryrun

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dryrun | bool | false | display operations without executing |
| --quiet | bool | false | suppress output |
| --recursive | bool | false | delete recursively |

**Examples:**

```bash
omni aws s3 rm s3://my-bucket/file.txt
omni aws s3 rm s3://my-bucket/prefix/ --recursive
omni aws s3 rm s3://my-bucket/prefix/ --recursive --dryrun
```

---

### aws ssm

**Category:** Other

**Usage:** `omni aws ssm`

**Description:** AWS SSM Parameter Store operations

**Details:**

AWS Systems Manager Parameter Store operations.

**Subcommands:** `delete-parameter`, `get-parameter`, `get-parameters`, `get-parameters-by-path`, `put-parameter`

---

### aws ssm delete-parameter

**Category:** Other

**Usage:** `omni aws ssm delete-parameter [flags]`

**Description:** Delete a parameter

**Details:**

Deletes a parameter.

Examples:
  omni aws ssm delete-parameter --name /app/config

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --name | string | - | parameter name (required) |

**Examples:**

```bash
omni aws ssm delete-parameter --name /app/config
```

---

### aws ssm get-parameter

**Category:** Other

**Usage:** `omni aws ssm get-parameter [flags]`

**Description:** Get a parameter value

**Details:**

Retrieves information about a parameter.

Examples:
  omni aws ssm get-parameter --name /app/config
  omni aws ssm get-parameter --name /app/secret --with-decryption

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --name | string | - | parameter name (required) |
| --with-decryption | bool | false | decrypt SecureString values |

**Examples:**

```bash
omni aws ssm get-parameter --name /app/config
omni aws ssm get-parameter --name /app/secret --with-decryption
```

---

### aws ssm get-parameters

**Category:** Other

**Usage:** `omni aws ssm get-parameters [flags]`

**Description:** Get multiple parameters

**Details:**

Retrieves information about multiple parameters.

Examples:
  omni aws ssm get-parameters --names /app/config,/app/secret
  omni aws ssm get-parameters --names /app/config --names /app/secret --with-decryption

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --names | stringSlice | [] | parameter names (required) |
| --with-decryption | bool | false | decrypt SecureString values |

**Examples:**

```bash
omni aws ssm get-parameters --names /app/config,/app/secret
omni aws ssm get-parameters --names /app/config --names /app/secret --with-decryption
```

---

### aws ssm get-parameters-by-path

**Category:** Other

**Usage:** `omni aws ssm get-parameters-by-path [flags]`

**Description:** Get parameters by path

**Details:**

Retrieves all parameters within a hierarchy.

Examples:
  omni aws ssm get-parameters-by-path --path /app/
  omni aws ssm get-parameters-by-path --path /app/ --recursive --with-decryption

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --max-results | int32 | 10 | maximum results per page |
| --path | string | - | parameter path (required) |
| --recursive | bool | false | include nested parameters |
| --with-decryption | bool | false | decrypt SecureString values |

**Examples:**

```bash
omni aws ssm get-parameters-by-path --path /app/
omni aws ssm get-parameters-by-path --path /app/ --recursive --with-decryption
```

---

### aws ssm put-parameter

**Category:** Other

**Usage:** `omni aws ssm put-parameter [flags]`

**Description:** Create or update a parameter

**Details:**

Creates or updates a parameter.

Examples:
  omni aws ssm put-parameter --name /app/config --value "config-value" --type String
  omni aws ssm put-parameter --name /app/secret --value "secret" --type SecureString
  omni aws ssm put-parameter --name /app/config --value "new-value" --overwrite

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --description | string | - | parameter description |
| --key-id | string | - | KMS key for SecureString |
| --name | string | - | parameter name (required) |
| --overwrite | bool | false | overwrite existing parameter |
| --type | string | String | parameter type: String, StringList, SecureString |
| --value | string | - | parameter value (required) |

**Examples:**

```bash
omni aws ssm put-parameter --name /app/config --value "config-value" --type String
omni aws ssm put-parameter --name /app/secret --value "secret" --type SecureString
omni aws ssm put-parameter --name /app/config --value "new-value" --overwrite
```

---

### aws sts

**Category:** Other

**Usage:** `omni aws sts`

**Description:** AWS STS operations

**Details:**

AWS Security Token Service (STS) operations.

**Subcommands:** `assume-role`, `get-caller-identity`

---

### aws sts assume-role

**Category:** Other

**Usage:** `omni aws sts assume-role [flags]`

**Description:** Assume an IAM role

**Details:**

Returns a set of temporary security credentials that you can use to access AWS resources.

Examples:
  omni aws sts assume-role --role-arn arn:aws:iam::123456789012:role/MyRole --role-session-name MySession

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --duration-seconds | int32 | 3600 | duration of the session in seconds |
| --external-id | string | - | external ID for cross-account access |
| --role-arn | string | - | ARN of the role to assume (required) |
| --role-session-name | string | - | session name (required) |

**Examples:**

```bash
omni aws sts assume-role --role-arn arn:aws:iam::123456789012:role/MyRole --role-session-name MySession
```

---

### aws sts get-caller-identity

**Category:** Other

**Usage:** `omni aws sts get-caller-identity`

**Description:** Get details about the IAM identity calling the API

**Details:**

Returns details about the IAM user or role whose credentials are used to call the operation.

Examples:
  omni aws sts get-caller-identity

**Examples:**

```bash
omni aws sts get-caller-identity
```

---

### base32

**Category:** Hash & Encoding

**Usage:** `omni base32 [OPTION]... [FILE] [flags]`

**Description:** Base32 encode or decode data

**Details:**

Base32 encode or decode FILE, or standard input, to standard output.

With no FILE, or when FILE is -, read standard input.

  -d, --decode    decode data
  -w, --wrap=N    wrap encoded lines after N characters (default 76, 0 = no wrap)

Examples:
  echo "hello" | omni base32           # encode
  echo "NBSWY3DP" | omni base32 -d     # decode
  omni base32 file.bin                 # encode file

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --decode | bool | false | decode data |
| -w, --wrap | int | 76 | wrap encoded lines after N characters (0 = no wrap) |

**Examples:**

```bash
omni base32 file.bin                 # encode file
```

---

### base58

**Category:** Hash & Encoding

**Usage:** `omni base58 [OPTION]... [FILE] [flags]`

**Description:** Base58 encode or decode data (Bitcoin alphabet)

**Details:**

Base58 encode or decode FILE, or standard input, to standard output.

Uses Bitcoin/IPFS alphabet: 123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz

With no FILE, or when FILE is -, read standard input.

  -d, --decode    decode data

Examples:
  echo "hello" | omni base58           # encode
  omni base58 -d encoded.txt           # decode

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --decode | bool | false | decode data |

**Examples:**

```bash
omni base58 -d encoded.txt           # decode
```

---

### base64

**Category:** Hash & Encoding

**Usage:** `omni base64 [OPTION]... [FILE] [flags]`

**Description:** Base64 encode or decode data

**Details:**

Base64 encode or decode FILE, or standard input, to standard output.

With no FILE, or when FILE is -, read standard input.

  -d, --decode    decode data
  -w, --wrap=N    wrap encoded lines after N characters (default 76, 0 = no wrap)
  -i, --ignore-garbage  ignore non-alphabet characters when decoding

Examples:
  echo "hello" | omni base64           # encode
  echo "aGVsbG8K" | omni base64 -d     # decode
  omni base64 file.bin                 # encode file
  omni base64 -d encoded.txt           # decode file

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --decode | bool | false | decode data |
| -i, --ignore-garbage | bool | false | ignore non-alphabet characters when decoding |
| -w, --wrap | int | 76 | wrap encoded lines after N characters (0 = no wrap) |

**Examples:**

```bash
omni base64 file.bin                 # encode file
omni base64 -d encoded.txt           # decode file
```

---

### basename

**Category:** Core

**Usage:** `omni basename NAME [SUFFIX] [flags]`

**Description:** Strip directory and suffix from file names

**Details:**

Print NAME with any leading directory components removed.
If specified, also remove a trailing SUFFIX.

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

**Details:**

bbolt provides commands for working with BoltDB databases.

This is a CLI wrapper around etcd-io/bbolt for database inspection,
manipulation, and maintenance operations.

Subcommands:
  info       Display database page size
  stats      Show database statistics
  buckets    List all buckets
  keys       List keys in a bucket
  get        Get value for a key
  put        Store a key-value pair
  delete     Delete a key
  dump       Dump bucket contents
  compact    Compact database to new file
  check      Verify database integrity
  pages      List database pages
  page       Hex dump of a page
  create-bucket  Create a new bucket
  delete-bucket  Delete a bucket

Examples:
  omni bbolt stats mydb.bolt
  omni bbolt buckets mydb.bolt
  omni bbolt keys mydb.bolt users
  omni bbolt get mydb.bolt users user1
  omni bbolt put mydb.bolt config version 1.0.0
  omni bbolt compact mydb.bolt mydb-compact.bolt

**Examples:**

```bash
omni bbolt stats mydb.bolt
omni bbolt buckets mydb.bolt
omni bbolt keys mydb.bolt users
omni bbolt get mydb.bolt users user1
omni bbolt put mydb.bolt config version 1.0.0
omni bbolt compact mydb.bolt mydb-compact.bolt
```

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

**Details:**

Brazilian document validation, generation, and formatting.

Subcommands:
  cpf     CPF (Cadastro de Pessoas Físicas) operations
  cnpj    CNPJ (Cadastro Nacional de Pessoa Jurídica) operations

Examples:
  omni brdoc cpf --generate           # generate a valid CPF
  omni brdoc cpf --validate 123.456.789-09
  omni brdoc cnpj --generate          # generate alphanumeric CNPJ
  omni brdoc cnpj --generate --legacy # generate numeric-only CNPJ

**Examples:**

```bash
omni brdoc cpf --generate           # generate a valid CPF
omni brdoc cpf --validate 123.456.789-09
omni brdoc cnpj --generate          # generate alphanumeric CNPJ
omni brdoc cnpj --generate --legacy # generate numeric-only CNPJ
```

**Subcommands:** `cnpj`, `cpf`

---

### brdoc cnpj

**Category:** Other

**Usage:** `omni brdoc cnpj [CNPJ...] [flags]`

**Description:** CNPJ operations (generate, validate, format)

**Details:**

CNPJ (Cadastro Nacional de Pessoa Jurídica) operations.

Supports both numeric and alphanumeric CNPJ formats per SERPRO specification.

Flags:
  -g, --generate    Generate valid CNPJ(s)
  -v, --validate    Validate CNPJ(s)
  -f, --format      Format CNPJ(s) as XX.XXX.XXX/XXXX-XX
  -n, --count       Number of CNPJs to generate (default 1)
  -l, --legacy      Generate numeric-only CNPJ (14 digits)
  --json            Output as JSON

Examples:
  omni brdoc cnpj --generate              # generate alphanumeric CNPJ
  omni brdoc cnpj --generate --legacy     # generate numeric-only CNPJ
  omni brdoc cnpj --generate -n 5         # generate 5 CNPJs
  omni brdoc cnpj --validate 12.ABC.345/01DE-35
  omni brdoc cnpj --validate 11222333000181
  omni brdoc cnpj --format 11222333000181
  omni brdoc cnpj --generate --json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | number of CNPJs to generate |
| -f, --format | bool | false | format CNPJ(s) |
| -g, --generate | bool | false | generate valid CNPJ(s) |
| --json | bool | false | output as JSON |
| -l, --legacy | bool | false | generate numeric-only CNPJ |
| -v, --validate | bool | false | validate CNPJ(s) |

**Examples:**

```bash
omni brdoc cnpj --generate              # generate alphanumeric CNPJ
omni brdoc cnpj --generate --legacy     # generate numeric-only CNPJ
omni brdoc cnpj --generate -n 5         # generate 5 CNPJs
omni brdoc cnpj --validate 12.ABC.345/01DE-35
omni brdoc cnpj --validate 11222333000181
omni brdoc cnpj --format 11222333000181
omni brdoc cnpj --generate --json
```

---

### brdoc cpf

**Category:** Other

**Usage:** `omni brdoc cpf [CPF...] [flags]`

**Description:** CPF operations (generate, validate, format)

**Details:**

CPF (Cadastro de Pessoas Físicas) operations.

Flags:
  -g, --generate    Generate valid CPF(s)
  -v, --validate    Validate CPF(s)
  -f, --format      Format CPF(s) as XXX.XXX.XXX-XX
  -n, --count       Number of CPFs to generate (default 1)
  --json            Output as JSON

Examples:
  omni brdoc cpf --generate              # generate one CPF
  omni brdoc cpf --generate -n 5         # generate 5 CPFs
  omni brdoc cpf --validate 12345678909
  omni brdoc cpf --validate 123.456.789-09
  omni brdoc cpf --format 12345678909
  omni brdoc cpf --generate --json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | number of CPFs to generate |
| -f, --format | bool | false | format CPF(s) |
| -g, --generate | bool | false | generate valid CPF(s) |
| --json | bool | false | output as JSON |
| -v, --validate | bool | false | validate CPF(s) |

**Examples:**

```bash
omni brdoc cpf --generate              # generate one CPF
omni brdoc cpf --generate -n 5         # generate 5 CPFs
omni brdoc cpf --validate 12345678909
omni brdoc cpf --validate 123.456.789-09
omni brdoc cpf --format 12345678909
omni brdoc cpf --generate --json
```

---

### buf

**Category:** Other

**Usage:** `omni buf`

**Description:** Protocol buffer utilities (lint, format, compile, generate)

**Details:**

Protocol buffer utilities inspired by buf.build.

Subcommands:
  lint       Lint proto files
  format     Format proto files
  compile    Compile proto files
  breaking   Check for breaking changes
  generate   Generate code from proto files
  mod        Module management (init)
  ls-files   List proto files

Examples:
  omni buf lint
  omni buf format --write
  omni buf compile -o image.bin
  omni buf breaking --against ../v1
  omni buf generate
  omni buf mod init buf.build/org/repo

**Examples:**

```bash
omni buf lint
omni buf format --write
omni buf compile -o image.bin
omni buf breaking --against ../v1
omni buf generate
omni buf mod init buf.build/org/repo
```

**Subcommands:** `breaking`, `compile`, `format`, `generate`, `lint`, `ls-files`, `mod`

---

### buf breaking

**Category:** Other

**Usage:** `omni buf breaking [DIR] [flags]`

**Description:** Check for breaking changes

**Details:**

Check for breaking changes against a previous version.

Flags:
  --against=PATH         Source to compare against (required)
  --exclude-path=PATH    Paths to exclude
  --exclude-imports      Don't check imported files
  --error-format=FORMAT  Output format: text, json, github-actions

Breaking change rules:
  FILE_NO_DELETE      Files cannot be deleted
  PACKAGE_NO_DELETE   Packages cannot be changed
  MESSAGE_NO_DELETE   Messages cannot be deleted
  FIELD_NO_DELETE     Fields cannot be deleted
  FIELD_SAME_TYPE     Field types cannot change
  ENUM_NO_DELETE      Enums cannot be deleted
  SERVICE_NO_DELETE   Services cannot be deleted
  RPC_NO_DELETE       RPCs cannot be deleted

Examples:
  omni buf breaking --against ../v1
  omni buf breaking --against ./baseline --error-format=json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --against | string | - | source to compare against (required) |
| --error-format | string | text | output format (text, json, github-actions) |
| --exclude-imports | bool | false | don't check imported files |
| --exclude-path | stringSlice | [] | paths to exclude |

**Examples:**

```bash
omni buf breaking --against ../v1
omni buf breaking --against ./baseline --error-format=json
```

---

### buf compile

**Category:** Other

**Usage:** `omni buf compile [DIR] [flags]`

**Description:** Compile proto files

**Details:**

Compile proto files and output an image.

Flags:
  -o, --output=FILE      Output file (.bin or .json)
  --exclude-path=PATH    Paths to exclude
  --error-format=FORMAT  Output format: text, json, github-actions

Examples:
  omni buf compile
  omni buf compile -o image.bin
  omni buf compile -o image.json
  omni buf compile --exclude-path=vendor

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --error-format | string | text | output format (text, json, github-actions) |
| --exclude-path | stringSlice | [] | paths to exclude |
| -o, --output | string | - | output file path |

**Examples:**

```bash
omni buf compile
omni buf compile -o image.bin
omni buf compile -o image.json
omni buf compile --exclude-path=vendor
```

---

### buf format

**Category:** Other

**Usage:** `omni buf format [DIR] [flags]`

**Description:** Format proto files

**Details:**

Format proto files with consistent style.

Flags:
  -w, --write       Rewrite files in place
  -d, --diff        Display diff instead of formatted output
  --exit-code       Exit with non-zero if files are not formatted

Examples:
  omni buf format
  omni buf format --write
  omni buf format --diff
  omni buf format --exit-code  # for CI

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --diff | bool | false | display diff |
| --exit-code | bool | false | exit with non-zero if files unformatted |
| -w, --write | bool | false | rewrite files in place |

**Examples:**

```bash
omni buf format
omni buf format --write
omni buf format --diff
omni buf format --exit-code  # for CI
```

---

### buf generate

**Category:** Code Generation

**Usage:** `omni buf generate [DIR] [flags]`

**Description:** Generate code from proto files

**Details:**

Generate code using plugins defined in buf.gen.yaml.

Flags:
  --template=FILE        Alternate buf.gen.yaml location
  -o, --output=DIR       Base output directory
  --include-imports      Include imported files in generation

buf.gen.yaml example:
  version: v1
  plugins:
    - local: protoc-gen-go
      out: gen/go
      opt:
        - paths=source_relative

Examples:
  omni buf generate
  omni buf generate --template=custom.gen.yaml
  omni buf generate -o ./generated

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --include-imports | bool | false | include imported files |
| -o, --output | string | - | base output directory |
| --template | string | - | alternate buf.gen.yaml location |

**Examples:**

```bash
omni buf generate
omni buf generate --template=custom.gen.yaml
omni buf generate -o ./generated
```

---

### buf lint

**Category:** Tooling

**Usage:** `omni buf lint [DIR] [flags]`

**Description:** Lint proto files

**Details:**

Lint proto files for style and structure issues.

Uses buf.yaml configuration if present. Default rules: STANDARD

Flags:
  --error-format=FORMAT  Output format: text, json, github-actions (default: text)
  --exclude-path=PATH    Paths to exclude (can be repeated)
  --config=FILE          Custom config file path

Categories:
  MINIMAL    Minimal set of rules
  BASIC      Basic rules (includes MINIMAL)
  STANDARD   Standard rules (includes BASIC)
  COMMENTS   Comment-related rules

Examples:
  omni buf lint
  omni buf lint ./proto
  omni buf lint --error-format=json
  omni buf lint --exclude-path=vendor

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --config | string | - | custom config file path |
| --error-format | string | text | output format (text, json, github-actions) |
| --exclude-path | stringSlice | [] | paths to exclude |

**Examples:**

```bash
omni buf lint
omni buf lint ./proto
omni buf lint --error-format=json
omni buf lint --exclude-path=vendor
```

---

### buf ls-files

**Category:** Other

**Usage:** `omni buf ls-files [DIR]`

**Description:** List proto files in the module

**Details:**

List all proto files in the module.

Examples:
  omni buf ls-files
  omni buf ls-files ./proto

**Examples:**

```bash
omni buf ls-files
omni buf ls-files ./proto
```

---

### buf mod

**Category:** Other

**Usage:** `omni buf mod`

**Description:** Module management commands

**Details:**

Module management commands.

Subcommands:
  init    Initialize a new buf module
  update  Update dependencies

**Subcommands:** `init`, `update`

---

### buf mod init

**Category:** Other

**Usage:** `omni buf mod init [NAME] [flags]`

**Description:** Initialize a new buf module

**Details:**

Initialize a new buf.yaml configuration file.

Examples:
  omni buf mod init
  omni buf mod init buf.build/org/repo

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dir | string | . | directory to initialize |

**Examples:**

```bash
omni buf mod init
omni buf mod init buf.build/org/repo
```

---

### buf mod update

**Category:** Other

**Usage:** `omni buf mod update`

**Description:** Update dependencies

**Details:**

Update dependencies listed in buf.yaml.

Note: Full dependency resolution requires network access to BSR.

Examples:
  omni buf mod update

**Examples:**

```bash
omni buf mod update
```

---

### bunzip2

**Category:** Archive

**Usage:** `omni bunzip2 [OPTION]... [FILE]... [flags]`

**Description:** Decompress bzip2 files

**Details:**

Decompress FILEs in bzip2 format.

Equivalent to bzip2 -d.

Examples:
  omni bunzip2 file.txt.bz2    # decompress
  omni bunzip2 -k file.txt.bz2 # keep original

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -f, --force | bool | false | force overwrite |
| -k, --keep | bool | false | keep original files |
| -c, --stdout | bool | false | write to stdout |
| -v, --verbose | bool | false | verbose mode |

**Examples:**

```bash
omni bunzip2 file.txt.bz2    # decompress
omni bunzip2 -k file.txt.bz2 # keep original
```

---

### bzcat

**Category:** Archive

**Usage:** `omni bzcat [FILE]...`

**Description:** Decompress and print bzip2 files

**Details:**

Decompress and print FILEs to stdout.

Equivalent to bzip2 -dc.

Examples:
  omni bzcat file.txt.bz2      # print decompressed content

**Examples:**

```bash
omni bzcat file.txt.bz2      # print decompressed content
```

---

### bzip2

**Category:** Archive

**Usage:** `omni bzip2 [OPTION]... [FILE]... [flags]`

**Description:** Decompress bzip2 files

**Details:**

Decompress FILEs using bzip2 format.

Note: Only decompression is supported (Go stdlib limitation).

  -d, --decompress   decompress (required)
  -k, --keep         keep original files
  -f, --force        force overwrite
  -c, --stdout       write to stdout
  -v, --verbose      verbose mode

Examples:
  omni bzip2 -d file.txt.bz2   # decompress
  omni bzip2 -dk file.txt.bz2  # decompress, keep original

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --decompress | bool | false | decompress |
| -f, --force | bool | false | force overwrite |
| -k, --keep | bool | false | keep original files |
| -c, --stdout | bool | false | write to stdout |
| -v, --verbose | bool | false | verbose mode |

**Examples:**

```bash
omni bzip2 -d file.txt.bz2   # decompress
omni bzip2 -dk file.txt.bz2  # decompress, keep original
```

---

### case

**Category:** Other

**Usage:** `omni case`

**Description:** Text case conversion utilities

**Details:**

Convert text between different case conventions.

Subcommands:
  upper     UPPERCASE
  lower     lowercase
  title     Title Case
  sentence  Sentence case
  camel     camelCase
  pascal    PascalCase
  snake     snake_case
  kebab     kebab-case
  constant  CONSTANT_CASE
  dot       dot.case
  path      path/case
  swap      sWAP cASE
  toggle    Toggle first char
  detect    Detect case type
  all       Show all conversions

Examples:
  omni case upper "hello world"       # HELLO WORLD
  omni case camel "hello world"       # helloWorld
  omni case snake "helloWorld"        # hello_world
  echo "hello" | omni case upper      # HELLO

**Examples:**

```bash
omni case upper "hello world"       # HELLO WORLD
omni case camel "hello world"       # helloWorld
omni case snake "helloWorld"        # hello_world
```

**Subcommands:** `all`, `camel`, `constant`, `detect`, `dot`, `kebab`, `lower`, `pascal`, `path`, `sentence`, `snake`, `swap`, `title`, `toggle`, `upper`

---

### case all

**Category:** Other

**Usage:** `omni case all [TEXT...] [flags]`

**Description:** Show all case conversions

**Details:**

Convert text to all supported case types and display results.

Examples:
  omni case all "hello world"
  omni case all "helloWorld"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case all "hello world"
omni case all "helloWorld"
```

---

### case camel

**Category:** Other

**Usage:** `omni case camel [TEXT...] [flags]`

**Description:** Convert to camelCase

**Details:**

Convert text to camelCase.

Examples:
  omni case camel "hello world"       # helloWorld
  omni case camel "Hello_World"       # helloWorld
  omni case camel "hello-world"       # helloWorld

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case camel "hello world"       # helloWorld
omni case camel "Hello_World"       # helloWorld
omni case camel "hello-world"       # helloWorld
```

---

### case constant

**Category:** Other

**Usage:** `omni case constant [TEXT...] [flags]`

**Description:** Convert to CONSTANT_CASE

**Details:**

Convert text to CONSTANT_CASE (SCREAMING_SNAKE_CASE).

Examples:
  omni case constant "hello world"    # HELLO_WORLD
  omni case constant "helloWorld"     # HELLO_WORLD

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case constant "hello world"    # HELLO_WORLD
omni case constant "helloWorld"     # HELLO_WORLD
```

---

### case detect

**Category:** Other

**Usage:** `omni case detect [TEXT...] [flags]`

**Description:** Detect the case type of text

**Details:**

Detect the case type of the input text.

Examples:
  omni case detect "helloWorld"       # camel
  omni case detect "hello_world"      # snake
  omni case detect "HELLO_WORLD"      # constant

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case detect "helloWorld"       # camel
omni case detect "hello_world"      # snake
omni case detect "HELLO_WORLD"      # constant
```

---

### case dot

**Category:** Other

**Usage:** `omni case dot [TEXT...] [flags]`

**Description:** Convert to dot.case

**Details:**

Convert text to dot.case.

Examples:
  omni case dot "hello world"         # hello.world
  omni case dot "helloWorld"          # hello.world

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case dot "hello world"         # hello.world
omni case dot "helloWorld"          # hello.world
```

---

### case kebab

**Category:** Other

**Usage:** `omni case kebab [TEXT...] [flags]`

**Description:** Convert to kebab-case

**Details:**

Convert text to kebab-case.

Examples:
  omni case kebab "hello world"       # hello-world
  omni case kebab "helloWorld"        # hello-world
  omni case kebab "HelloWorld"        # hello-world

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case kebab "hello world"       # hello-world
omni case kebab "helloWorld"        # hello-world
omni case kebab "HelloWorld"        # hello-world
```

---

### case lower

**Category:** Other

**Usage:** `omni case lower [TEXT...] [flags]`

**Description:** Convert to lowercase

**Details:**

Convert text to lowercase.

Examples:
  omni case lower "HELLO WORLD"       # hello world
  echo "HELLO" | omni case lower      # hello

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case lower "HELLO WORLD"       # hello world
```

---

### case pascal

**Category:** Other

**Usage:** `omni case pascal [TEXT...] [flags]`

**Description:** Convert to PascalCase

**Details:**

Convert text to PascalCase.

Examples:
  omni case pascal "hello world"      # HelloWorld
  omni case pascal "hello_world"      # HelloWorld
  omni case pascal "hello-world"      # HelloWorld

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case pascal "hello world"      # HelloWorld
omni case pascal "hello_world"      # HelloWorld
omni case pascal "hello-world"      # HelloWorld
```

---

### case path

**Category:** Other

**Usage:** `omni case path [TEXT...] [flags]`

**Description:** Convert to path/case

**Details:**

Convert text to path/case.

Examples:
  omni case path "hello world"        # hello/world
  omni case path "helloWorld"         # hello/world

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case path "hello world"        # hello/world
omni case path "helloWorld"         # hello/world
```

---

### case sentence

**Category:** Other

**Usage:** `omni case sentence [TEXT...] [flags]`

**Description:** Convert to Sentence case

**Details:**

Convert text to Sentence case (capitalize first letter only).

Examples:
  omni case sentence "hello world"    # Hello world
  echo "HELLO WORLD" | omni case sentence

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case sentence "hello world"    # Hello world
```

---

### case snake

**Category:** Other

**Usage:** `omni case snake [TEXT...] [flags]`

**Description:** Convert to snake_case

**Details:**

Convert text to snake_case.

Examples:
  omni case snake "hello world"       # hello_world
  omni case snake "helloWorld"        # hello_world
  omni case snake "HelloWorld"        # hello_world

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case snake "hello world"       # hello_world
omni case snake "helloWorld"        # hello_world
omni case snake "HelloWorld"        # hello_world
```

---

### case swap

**Category:** Other

**Usage:** `omni case swap [TEXT...] [flags]`

**Description:** Swap case of each character

**Details:**

Swap the case of each character (upper becomes lower, lower becomes upper).

Examples:
  omni case swap "Hello World"        # hELLO wORLD
  omni case swap "helloWorld"         # HELLOwORLD

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case swap "Hello World"        # hELLO wORLD
omni case swap "helloWorld"         # HELLOwORLD
```

---

### case title

**Category:** Other

**Usage:** `omni case title [TEXT...] [flags]`

**Description:** Convert to Title Case

**Details:**

Convert text to Title Case (capitalize first letter of each word).

Examples:
  omni case title "hello world"       # Hello World
  echo "hello world" | omni case title

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case title "hello world"       # Hello World
```

---

### case toggle

**Category:** Other

**Usage:** `omni case toggle [TEXT...] [flags]`

**Description:** Toggle first character's case

**Details:**

Toggle the case of the first character.

Examples:
  omni case toggle "hello"            # Hello
  omni case toggle "Hello"            # hello

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case toggle "hello"            # Hello
omni case toggle "Hello"            # hello
```

---

### case upper

**Category:** Other

**Usage:** `omni case upper [TEXT...] [flags]`

**Description:** Convert to UPPERCASE

**Details:**

Convert text to UPPERCASE.

Examples:
  omni case upper "hello world"       # HELLO WORLD
  omni case upper hello world         # HELLO WORLD
  echo "hello" | omni case upper      # HELLO

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni case upper "hello world"       # HELLO WORLD
omni case upper hello world         # HELLO WORLD
```

---

### cat

**Category:** Core

**Usage:** `omni cat [file...] [flags]`

**Description:** Concatenate files and print on the standard output

**Details:**

Concatenate FILE(s) to standard output.
With no FILE, or when FILE is -, read standard input.

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

**Details:**

Change the mode of each FILE to MODE.

MODE can be:
  - Octal number (e.g., 755, 644)
  - Symbolic mode (e.g., u+x, go-w, a=rw)

Symbolic mode format: [ugoa][+-=][rwx]
  u = user, g = group, o = others, a = all
  + = add, - = remove, = = set exactly
  r = read, w = write, x = execute

Options:
  -R, --recursive  change files and directories recursively
  -v, --verbose    output a diagnostic for every file processed
  -c, --changes    like verbose but report only when a change is made
  -f, --silent     suppress most error messages
      --reference  use RFILE's mode instead of MODE values

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

**Details:**

Change the owner and/or group of each FILE to OWNER and/or GROUP.

OWNER can be specified as:
  - User name (e.g., root)
  - Numeric user ID (e.g., 0)
  - OWNER:GROUP to change both
  - OWNER: to change owner and set group to owner's login group
  - :GROUP to change only the group

Options:
  -R, --recursive   operate on files and directories recursively
  -v, --verbose     output a diagnostic for every file processed
  -c, --changes     like verbose but report only when a change is made
  -f, --silent      suppress most error messages
  -h, --no-dereference  affect symbolic links instead of referenced file
      --reference   use RFILE's owner and group
      --preserve-root  fail to operate recursively on '/'

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

**Details:**

Cloud profile management for AWS, Azure, and GCP.

Securely store and manage credentials for multiple cloud providers.
Credentials are encrypted using AES-256-GCM with per-profile key isolation.

Supported providers:
  aws    - Amazon Web Services
  azure  - Microsoft Azure
  gcp    - Google Cloud Platform

Examples:
  # Add an AWS profile
  omni cloud profile add myaws --provider aws --region us-east-1

  # List all profiles
  omni cloud profile list

  # Use a profile with AWS commands
  export OMNI_CLOUD_PROFILE=myaws
  omni aws s3 ls

  # Or use the omni: prefix
  omni aws s3 ls --profile omni:myaws

**Examples:**

```bash
omni cloud profile add myaws --provider aws --region us-east-1
omni cloud profile list
omni aws s3 ls
omni aws s3 ls --profile omni:myaws
```

**Subcommands:** `profile`

---

### cloud profile

**Category:** Other

**Usage:** `omni cloud profile`

**Description:** Manage cloud profiles

**Details:**

Manage cloud profiles for AWS, Azure, and GCP.

**Subcommands:** `add`, `delete`, `import`, `list`, `show`, `use`

---

### cloud profile add

**Category:** Other

**Usage:** `omni cloud profile add <name> [flags]`

**Description:** Add a new cloud profile

**Details:**

Add a new cloud profile with encrypted credentials.

For AWS:
  Prompts for Access Key ID and Secret Access Key, or use flags.

For Azure:
  Prompts for Tenant ID, Client ID, Client Secret, and Subscription ID.

For GCP:
  Requires --key-file pointing to a service account JSON file.

Examples:
  # Add AWS profile interactively
  omni cloud profile add myaws --provider aws --region us-east-1

  # Add AWS profile with flags
  omni cloud profile add myaws --provider aws \
    --access-key-id AKIAXXXXXXXX \
    --secret-access-key XXXXXXXX

  # Add Azure profile
  omni cloud profile add myazure --provider azure

  # Add GCP profile
  omni cloud profile add mygcp --provider gcp --key-file /path/to/sa.json

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

**Examples:**

```bash
omni cloud profile add myaws --provider aws --region us-east-1
omni cloud profile add myaws --provider aws \
omni cloud profile add myazure --provider azure
omni cloud profile add mygcp --provider gcp --key-file /path/to/sa.json
```

---

### cloud profile delete

**Category:** Other

**Usage:** `omni cloud profile delete <name> [flags]`

**Description:** Delete a cloud profile

**Details:**

Delete a cloud profile and its encrypted credentials.

Examples:
  omni cloud profile delete myaws --provider aws
  omni cloud profile delete myaws --provider aws --force

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -f, --force | bool | false | Skip confirmation |
| -p, --provider | string | - | Provider (required) |

**Examples:**

```bash
omni cloud profile delete myaws --provider aws
omni cloud profile delete myaws --provider aws --force
```

---

### cloud profile import

**Category:** Other

**Usage:** `omni cloud profile import [name] [flags]`

**Description:** Import credentials from existing cloud CLI

**Details:**

Import existing credentials from AWS, Azure, or GCP CLI configurations.

AWS:
  Reads from ~/.aws/credentials and ~/.aws/config
  Use --source to specify which AWS profile to import (default: "default")

Azure:
  Requires a service principal JSON file (Azure CLI tokens cannot be migrated)
  Use --source to specify the service principal file path
  Create one with: az ad sp create-for-rbac --name omni-sp --sdk-auth > ~/.azure/sp.json

GCP:
  Imports service account credentials from GOOGLE_APPLICATION_CREDENTIALS
  or ~/.config/gcloud/ directory
  Note: Application Default Credentials (authorized_user) cannot be migrated

Examples:
  # Import default AWS profile
  omni cloud profile import --provider aws

  # Import specific AWS profile with custom name
  omni cloud profile import myaws --provider aws --source prod

  # List available AWS profiles
  omni cloud profile import --provider aws --list

  # Import Azure service principal
  omni cloud profile import --provider azure --source ~/.azure/sp.json

  # Import GCP service account
  omni cloud profile import --provider gcp --source /path/to/sa.json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --default | bool | false | Set as default profile after import |
| --list | bool | false | List available profiles/credentials to import |
| -p, --provider | string | - | Cloud provider (aws, azure, gcp) (required) |
| -s, --source | string | - | Source profile/file to import from |

**Examples:**

```bash
omni cloud profile import --provider aws
omni cloud profile import myaws --provider aws --source prod
omni cloud profile import --provider aws --list
omni cloud profile import --provider azure --source ~/.azure/sp.json
omni cloud profile import --provider gcp --source /path/to/sa.json
```

---

### cloud profile list

**Category:** Other

**Usage:** `omni cloud profile list [flags]`

**Description:** List cloud profiles

**Details:**

List all cloud profiles, optionally filtered by provider.

Examples:
  omni cloud profile list
  omni cloud profile list --provider aws

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -p, --provider | string | - | Filter by provider (aws, azure, gcp) |

**Examples:**

```bash
omni cloud profile list
omni cloud profile list --provider aws
```

---

### cloud profile show

**Category:** Other

**Usage:** `omni cloud profile show <name> [flags]`

**Description:** Show profile details

**Details:**

Show details of a cloud profile (without credentials).

Examples:
  omni cloud profile show myaws
  omni cloud profile show myaws --provider aws

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -p, --provider | string | - | Provider (defaults to aws) |

**Examples:**

```bash
omni cloud profile show myaws
omni cloud profile show myaws --provider aws
```

---

### cloud profile use

**Category:** Other

**Usage:** `omni cloud profile use <name> [flags]`

**Description:** Set a profile as default

**Details:**

Set a profile as the default for its provider.

Examples:
  omni cloud profile use myaws
  omni cloud profile use myaws --provider aws

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -p, --provider | string | - | Provider (defaults to aws) |

**Examples:**

```bash
omni cloud profile use myaws
omni cloud profile use myaws --provider aws
```

---

### cmdtree

**Category:** Tooling

**Usage:** `omni cmdtree [flags]`

**Description:** Display command tree visualization

**Details:**

Display a tree visualization of all available commands with descriptions.

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

**Details:**

Compare two files byte by byte.

  -s, --silent   suppress all normal output
  -l, --verbose  output byte numbers and differing byte values
  -b, --print-bytes  print differing bytes
  -i, --ignore-initial=SKIP  skip first SKIP bytes
  -n, --bytes=LIMIT  compare at most LIMIT bytes

Exit status:
  0  files are identical
  1  files differ
  2  trouble

Examples:
  omni cmp file1.bin file2.bin     # compare files
  omni cmp -s file1 file2          # silent, check exit status
  omni cmp -l file1 file2          # show all differences

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --bytes | int64 | 0 | compare at most LIMIT bytes |
| -i, --ignore-initial | int64 | 0 | skip first SKIP bytes |
| --json | bool | false | output as JSON |
| -b, --print-bytes | bool | false | print differing bytes |
| -s, --silent | bool | false | suppress all output |
| -l, --verbose | bool | false | output byte numbers and values |

**Examples:**

```bash
omni cmp file1.bin file2.bin     # compare files
omni cmp -s file1 file2          # silent, check exit status
omni cmp -l file1 file2          # show all differences
```

---

### column

**Category:** Text Processing

**Usage:** `omni column [OPTION]... [FILE]... [flags]`

**Description:** Columnate lists

**Details:**

Format input into multiple columns.

With no FILE, or when FILE is -, read standard input.

  -t, --table            determine column count based on input
  -s, --separator=STRING delimiter characters for -t option
  -o, --output-separator=STRING  output separator for table mode
  -c, --columns=N        output width in characters (default 80)
  -x, --fillrows         fill rows before columns
  -R, --right            right-align columns

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

**Details:**

Compare sorted files FILE1 and FILE2 line by line.

With no options, produce three-column output:
  Column 1: lines unique to FILE1
  Column 2: lines unique to FILE2
  Column 3: lines common to both files

  -1                 suppress column 1 (lines unique to FILE1)
  -2                 suppress column 2 (lines unique to FILE2)
  -3                 suppress column 3 (lines common to both)
  --check-order      check that input is correctly sorted
  --nocheck-order    do not check input order
  --output-delimiter use STR as output delimiter
  -z, --zero-terminated  line delimiter is NUL

Examples:
  omni comm file1.txt file2.txt        # show all columns
  omni comm -12 file1.txt file2.txt    # show only common lines
  omni comm -3 file1.txt file2.txt     # show only unique lines

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

**Examples:**

```bash
omni comm file1.txt file2.txt        # show all columns
omni comm -12 file1.txt file2.txt    # show only common lines
omni comm -3 file1.txt file2.txt     # show only unique lines
```

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

**Details:**

Copy SOURCE to DEST, or multiple SOURCE(s) to DIRECTORY.

---

### cron

**Category:** Other

**Usage:** `omni cron EXPRESSION [flags]`

**Description:** Parse and explain cron expressions

**Details:**

Parse cron expressions and display human-readable explanations.

Cron expression format: MINUTE HOUR DAY MONTH WEEKDAY

Field values:
  MINUTE   0-59
  HOUR     0-23
  DAY      1-31
  MONTH    1-12 (or names: jan, feb, etc.)
  WEEKDAY  0-7 (0 and 7 = Sunday, or names: sun, mon, etc.)

Special characters:
  *    Any value
  ,    List separator (e.g., 1,3,5)
  -    Range (e.g., 1-5)
  /    Step (e.g., */5)

Aliases:
  @yearly    Once a year (0 0 1 1 *)
  @monthly   Once a month (0 0 1 * *)
  @weekly    Once a week (0 0 * * 0)
  @daily     Once a day (0 0 * * *)
  @hourly    Once an hour (0 * * * *)

Examples:
  omni cron "*/15 * * * *"              # Every 15 minutes
  omni cron "0 9 * * 1-5"               # 9 AM on weekdays
  omni cron "0 0 1 * *"                 # First day of month at midnight
  omni cron "30 4 1,15 * *"             # 4:30 AM on 1st and 15th
  omni cron "@daily"                    # Every day at midnight
  omni cron --next 5 "0 */2 * * *"      # Show next 5 runs

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |
| --next | int | 0 | show next N scheduled runs |
| --validate | bool | false | only validate the expression |

**Examples:**

```bash
omni cron "*/15 * * * *"              # Every 15 minutes
omni cron "0 9 * * 1-5"               # 9 AM on weekdays
omni cron "0 0 1 * *"                 # First day of month at midnight
omni cron "30 4 1,15 * *"             # 4:30 AM on 1st and 15th
omni cron "@daily"                    # Every day at midnight
omni cron --next 5 "0 */2 * * *"      # Show next 5 runs
```

---

### css

**Category:** Other

**Usage:** `omni css [FILE] [flags]`

**Description:** CSS utilities (format, minify, validate)

**Details:**

CSS utilities for formatting, minifying, and validating CSS.

When called directly, formats CSS (same as 'css fmt').

Subcommands:
  fmt         Format/beautify CSS
  minify      Minify CSS
  validate    Validate CSS syntax

Examples:
  omni css file.css
  omni css fmt file.css
  omni css minify file.css
  omni css validate file.css
  echo 'body{margin:0}' | omni css
  omni css "body{margin:0;padding:0}"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| --sort-props | bool | false | sort properties alphabetically |
| --sort-rules | bool | false | sort selectors alphabetically |

**Examples:**

```bash
omni css file.css
omni css fmt file.css
omni css minify file.css
omni css validate file.css
omni css "body{margin:0;padding:0}"
```

**Subcommands:** `fmt`, `minify`, `validate`

---

### css fmt

**Category:** Other

**Usage:** `omni css fmt [FILE] [flags]`

**Description:** Format/beautify CSS

**Details:**

Format CSS with proper indentation.

  -i, --indent=STR     indentation string (default "  ")
  --sort-props         sort properties alphabetically
  --sort-rules         sort selectors alphabetically

Examples:
  omni css fmt file.css
  omni css fmt "body{margin:0;padding:0}"
  cat file.css | omni css fmt
  omni css fmt --sort-props file.css

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| --sort-props | bool | false | sort properties alphabetically |
| --sort-rules | bool | false | sort selectors alphabetically |

**Examples:**

```bash
omni css fmt file.css
omni css fmt "body{margin:0;padding:0}"
omni css fmt --sort-props file.css
```

---

### css minify

**Category:** Other

**Usage:** `omni css minify [FILE]`

**Description:** Minify CSS

**Details:**

Minify CSS by removing unnecessary whitespace and comments.

Examples:
  omni css minify file.css
  cat file.css | omni css minify

**Examples:**

```bash
omni css minify file.css
```

---

### css validate

**Category:** Other

**Usage:** `omni css validate [FILE] [flags]`

**Description:** Validate CSS syntax

**Details:**

Validate CSS syntax.

Exit codes:
  0  Valid CSS
  1  Invalid CSS or error

  --json    output result as JSON

Examples:
  omni css validate file.css
  omni css validate "body { margin: 0; }"
  omni css validate --json file.css

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni css validate file.css
omni css validate "body { margin: 0; }"
omni css validate --json file.css
```

---

### csv

**Category:** Other

**Usage:** `omni csv`

**Description:** CSV utilities (convert to/from JSON)

**Details:**

CSV utilities for converting between CSV and JSON formats.

Subcommands:
  tojson    Convert CSV to JSON array
  fromjson  Convert JSON array to CSV

Examples:
  omni csv tojson file.csv             # convert CSV to JSON
  omni csv fromjson file.json          # convert JSON to CSV
  cat data.csv | omni csv tojson       # from stdin
  omni csv tojson -d ";" file.csv      # custom delimiter

**Examples:**

```bash
omni csv tojson file.csv             # convert CSV to JSON
omni csv fromjson file.json          # convert JSON to CSV
omni csv tojson -d ";" file.csv      # custom delimiter
```

**Subcommands:** `fromjson`, `tojson`

---

### csv fromjson

**Category:** Other

**Usage:** `omni csv fromjson [FILE] [flags]`

**Description:** Convert JSON array to CSV

**Details:**

Convert JSON array of objects to CSV format.

Nested objects are flattened with dot notation (e.g., address.city).

  --no-header          don't include header row
  -d, --delimiter=STR  field delimiter (default ",")
  --no-quotes          don't quote fields

Examples:
  omni csv fromjson file.json
  echo '[{"name":"John","age":30}]' | omni csv fromjson
  omni csv fromjson -d ";" file.json   # semicolon delimiter
  omni csv fromjson --no-header file.json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --delimiter | string | , | field delimiter |
| --no-header | bool | false | don't include header row |
| --no-quotes | bool | false | don't quote fields |

**Examples:**

```bash
omni csv fromjson file.json
omni csv fromjson -d ";" file.json   # semicolon delimiter
omni csv fromjson --no-header file.json
```

---

### csv tojson

**Category:** Other

**Usage:** `omni csv tojson [FILE] [flags]`

**Description:** Convert CSV to JSON array

**Details:**

Convert CSV data to JSON array of objects.

  --no-header          first row is data, not headers
  -d, --delimiter=STR  field delimiter (default ",")
  -a, --array          always output as array (even for single row)

Examples:
  omni csv tojson file.csv
  cat file.csv | omni csv tojson
  omni csv tojson -d ";" file.csv      # semicolon delimiter
  omni csv tojson --no-header file.csv

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --array | bool | false | always output as array |
| -d, --delimiter | string | , | field delimiter |
| --no-header | bool | false | first row is data, not headers |

**Examples:**

```bash
omni csv tojson file.csv
omni csv tojson -d ";" file.csv      # semicolon delimiter
omni csv tojson --no-header file.csv
```

---

### curl

**Category:** Other

**Usage:** `omni curl [METHOD] URL [ITEM...] [flags]`

**Description:** HTTP client with httpie-like syntax

**Details:**

HTTP client inspired by curlie/httpie.

Supports httpie-like syntax for headers and data:
  key:value     HTTP header
  key=value     JSON data field
  key==value    URL query parameter
  @file         Request body from file

Methods: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS

Examples:
  omni curl https://api.example.com/users
  omni curl POST https://api.example.com/users name=John email=john@example.com
  omni curl https://api.example.com/users Authorization:"Bearer token"
  omni curl https://api.example.com/search q==hello
  omni curl POST https://api.example.com/upload @data.json
  omni curl -v https://api.example.com/users
  omni curl --json https://api.example.com/users

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

**Examples:**

```bash
omni curl https://api.example.com/users
omni curl POST https://api.example.com/users name=John email=john@example.com
omni curl https://api.example.com/users Authorization:"Bearer token"
omni curl https://api.example.com/search q==hello
omni curl POST https://api.example.com/upload @data.json
omni curl -v https://api.example.com/users
omni curl --json https://api.example.com/users
```

---

### cut

**Category:** Text Processing

**Usage:** `omni cut [OPTION]... [FILE]... [flags]`

**Description:** Remove sections from each line of files

**Details:**

Print selected parts of lines from each FILE to standard output.

With no FILE, or when FILE is -, read standard input.

Mandatory arguments to long options are mandatory for short options too.
  -b, --bytes=LIST        select only these bytes
  -c, --characters=LIST   select only these characters
  -d, --delimiter=DELIM   use DELIM instead of TAB for field delimiter
  -f, --fields=LIST       select only these fields
  -s, --only-delimited    do not print lines not containing delimiters
      --complement        complement the set of selected bytes, characters or fields
      --output-delimiter=STRING  use STRING as the output delimiter

Use one, and only one of -b, -c or -f.  Each LIST is made up of one
range, or many ranges separated by commas.  Each range is one of:
  N     N'th byte, character or field, counted from 1
  N-    from N'th byte, character or field, to end of line
  N-M   from N'th to M'th (included) byte, character or field
  -M    from first to M'th (included) byte, character or field

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

**Details:**

Display the current time in the given FORMAT, or set the system date.

FORMAT controls the output. Interpreted sequences are:
  %Y   year
  %m   month (01..12)
  %d   day of month (01..31)
  %H   hour (00..23)
  %M   minute (00..59)
  %S   second (00..60)

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

**Details:**

Copy a file, converting and formatting according to the operands.

  if=FILE     read from FILE instead of stdin
  of=FILE     write to FILE instead of stdout
  bs=BYTES    read and write up to BYTES bytes at a time
  ibs=BYTES   read up to BYTES bytes at a time (default: 512)
  obs=BYTES   write BYTES bytes at a time (default: 512)
  count=N     copy only N input blocks
  skip=N      skip N ibs-sized blocks at start of input
  seek=N      skip N obs-sized blocks at start of output
  conv=CONVS  convert the file as per the comma separated symbol list
  status=LEVEL  LEVEL of information to print to stderr:
              'none' suppresses everything but error messages,
              'noxfer' suppresses the final transfer statistics,
              'progress' shows periodic transfer statistics

CONV symbols:
  lcase       change upper case to lower case
  ucase       change lower case to upper case
  swab        swap every pair of input bytes
  notrunc     do not truncate the output file
  fsync       physically write output file data before finishing

BYTES may be followed by multiplicative suffixes:
  K=1024, M=1024*1024, G=1024*1024*1024

Examples:
  omni dd if=input.txt of=output.txt               # copy file
  omni dd if=/dev/zero of=file.bin bs=1M count=10  # create 10MB file
  omni dd if=file.txt conv=ucase                   # convert to uppercase
  omni dd if=disk.img of=backup.img bs=4K          # disk image backup

**Examples:**

```bash
omni dd if=input.txt of=output.txt               # copy file
omni dd if=/dev/zero of=file.bin bs=1M count=10  # create 10MB file
omni dd if=file.txt conv=ucase                   # convert to uppercase
omni dd if=disk.img of=backup.img bs=4K          # disk image backup
```

---

### decrypt

**Category:** Security

**Usage:** `omni decrypt [OPTION]... [FILE] [flags]`

**Description:** Decrypt data using AES-256-GCM

**Details:**

Decrypt FILE or standard input using AES-256-GCM.

  -p, --password STRING   password for decryption
  -P, --password-file FILE  read password from file
  -k, --key-file FILE     use key file for decryption
  -o, --output FILE       write output to file
  -a, --armor             input is ASCII armored (base64)
  -i, --iterations N      PBKDF2 iterations (default 100000)

Password can also be set via omni_PASSWORD environment variable.

Examples:
  omni decrypt -p mypassword secret.enc
  omni decrypt -p mypassword -a < secret.b64
  omni decrypt -P ~/.password -o file.txt secret.enc
  cat secret.enc | omni_PASSWORD=pass omni decrypt

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

**Examples:**

```bash
omni decrypt -p mypassword secret.enc
omni decrypt -p mypassword -a < secret.b64
omni decrypt -P ~/.password -o file.txt secret.enc
```

---

### df

**Category:** System Info

**Usage:** `omni df [OPTION]... [FILE]... [flags]`

**Description:** Report file system disk space usage

**Details:**

Show information about the file system on which each FILE resides,
or all file systems by default.

  -h, --human-readable  print sizes in human readable format (e.g., 1K 234M 2G)
  -i, --inodes          list inode information instead of block usage
  -B, --block-size=SIZE scale sizes by SIZE before printing them
      --total           produce a grand total
  -t, --type=TYPE       limit listing to file systems of type TYPE
  -x, --exclude-type=TYPE  exclude file systems of type TYPE
  -l, --local           limit listing to local file systems
  -P, --portability     use the POSIX output format

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

**Details:**

Compare files line by line.

Output format:
  - Unified diff (default): shows context with +/- markers
  - Side-by-side (-y): shows files in parallel columns
  - Brief (-q): only report if files differ

Special modes:
  --json    Compare JSON files structurally
  -r        Recursively compare directories

Examples:
  omni diff file1.txt file2.txt
  omni diff -u 5 old.txt new.txt         # 5 lines of context
  omni diff -y file1.txt file2.txt       # side-by-side
  omni diff -q dir1/ dir2/               # brief comparison
  omni diff --json config1.json config2.json
  omni diff -r dir1/ dir2/               # recursive

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

**Examples:**

```bash
omni diff file1.txt file2.txt
omni diff -u 5 old.txt new.txt         # 5 lines of context
omni diff -y file1.txt file2.txt       # side-by-side
omni diff -q dir1/ dir2/               # brief comparison
omni diff --json config1.json config2.json
omni diff -r dir1/ dir2/               # recursive
```

---

### dirname

**Category:** Core

**Usage:** `omni dirname [path...] [flags]`

**Description:** Strip last component from file name

**Details:**

Output each NAME with its last non-slash component and trailing slashes removed.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### dotenv

**Category:** Data Processing

**Usage:** `omni dotenv [OPTION]... [FILE]... [flags]`

**Description:** Load environment variables from .env files

**Details:**

Parse and display environment variables from .env files.

With no FILE, reads from .env in the current directory.

  -e, --export    output as export statements (for shell sourcing)
  -q, --quiet     suppress warnings
  -x, --expand    expand variables in values

The .env file format:
  # Comments start with #
  KEY=value
  KEY="quoted value"
  KEY='single quoted'
  export KEY=value    # export prefix is optional

Examples:
  omni dotenv                    # display vars from .env
  omni dotenv .env.local         # display vars from specific file
  omni dotenv -e                 # output as export statements
  eval $(omni dotenv -e)         # load vars into shell
  omni dotenv -x                 # expand ${VAR} references

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -x, --expand | bool | false | expand variables in values |
| -e, --export | bool | false | output as export statements |
| -q, --quiet | bool | false | suppress warnings |

**Examples:**

```bash
omni dotenv                    # display vars from .env
omni dotenv .env.local         # display vars from specific file
omni dotenv -e                 # output as export statements
omni dotenv -x                 # expand ${VAR} references
```

---

### du

**Category:** System Info

**Usage:** `omni du [OPTION]... [FILE]... [flags]`

**Description:** Estimate file space usage

**Details:**

Summarize disk usage of each FILE, recursively for directories.

  -a, --all             write counts for all files, not just directories
  -b, --bytes           equivalent to --apparent-size --block-size=1
  -c, --total           produce a grand total
  -h, --human-readable  print sizes in human readable format (e.g., 1K 234M 2G)
  -s, --summarize       display only a total for each argument
  -d, --max-depth=N     print the total for a directory only if it is N or fewer
                        levels below the command line argument
  -x, --one-file-system skip directories on different file systems
      --apparent-size   print apparent sizes, rather than disk usage
  -0, --null            end each output line with NUL, not newline
  -B, --block-size=SIZE scale sizes by SIZE before printing them

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

**Details:**

Echo the STRING(s) to standard output.

Examples:
  omni echo Hello World     # outputs 'Hello World'
  omni echo -n "no newline" # outputs without trailing newline
  omni echo -e "tab\there"  # outputs with tab character

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -e, --escape | bool | false | enable interpretation of backslash escapes |
| --json | bool | false | output as JSON |
| -E, --no-escape | bool | false | disable interpretation of backslash escapes (default) |
| -n, --no-newline | bool | false | do not output the trailing newline |

**Examples:**

```bash
omni echo Hello World     # outputs 'Hello World'
omni echo -n "no newline" # outputs without trailing newline
omni echo -e "tab\there"  # outputs with tab character
```

---

### egrep

**Category:** Text Processing

**Usage:** `omni egrep [options] PATTERN [FILE...] [flags]`

**Description:** Print lines that match patterns (extended regexp)

**Details:**

Search for PATTERN in each FILE using extended regular expressions.
This is equivalent to 'grep -E'.

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

**Details:**

Encrypt FILE or standard input using AES-256-GCM.

Uses PBKDF2 for key derivation with SHA-256.

  -p, --password STRING   password for encryption
  -P, --password-file FILE  read password from file
  -k, --key-file FILE     use key file for encryption
  -o, --output FILE       write output to file
  -a, --armor             ASCII armor (base64) output
  -i, --iterations N      PBKDF2 iterations (default 100000)

Password can also be set via omni_PASSWORD environment variable.

Examples:
  echo "secret" | omni encrypt -p mypassword
  omni encrypt -p mypassword -o secret.enc file.txt
  omni encrypt -P ~/.password -a file.txt
  omni_PASSWORD=pass omni encrypt file.txt

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

**Examples:**

```bash
omni encrypt -p mypassword -o secret.enc file.txt
omni encrypt -P ~/.password -a file.txt
```

---

### env

**Category:** System Info

**Usage:** `omni env [NAME...] [flags]`

**Description:** Print environment variables

**Details:**

Print the values of the specified environment variables.
If no NAME is specified, print all environment variables.

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

**Details:**

Search for PATTERN in each FILE using fixed strings (no regex).
This is equivalent to 'grep -F'.

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

**Details:**

Determine the type of each FILE.

  -b, --brief           do not prepend filenames to output
  -i, --mime            output MIME type strings
  -h, --no-dereference  don't follow symlinks
  -F, --separator       use string as separator instead of ':'

Examples:
  omni file image.png          # PNG image data
  omni file -i document.pdf    # application/pdf
  omni file -b script.sh       # output type only
  omni file *                  # check all files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --brief | bool | false | do not prepend filenames |
| --json | bool | false | output as JSON |
| -i, --mime | bool | false | output MIME type |
| -L, --no-dereference | bool | false | don't follow symlinks |
| -F, --separator | string | : | use string as separator |

**Examples:**

```bash
omni file image.png          # PNG image data
omni file -i document.pdf    # application/pdf
omni file -b script.sh       # output type only
omni file *                  # check all files
```

---

### find

**Category:** File Operations

**Usage:** `omni find [path...] [expression] [flags]`

**Description:** Search for files in a directory hierarchy

**Details:**

Search for files in a directory hierarchy.

Tests:
  -name PATTERN      file name matches shell PATTERN
  -iname PATTERN     like -name, but case insensitive
  -path PATTERN      path matches shell PATTERN
  -ipath PATTERN     like -path, but case insensitive
  -regex PATTERN     path matches regular expression PATTERN
  -iregex PATTERN    like -regex, but case insensitive
  -type TYPE         file type: f(ile), d(ir), l(ink), p(ipe), s(ocket)
  -size N[cwbkMGTP]  file size is N units (c=bytes, k=KB, M=MB, G=GB)
                     prefix + for greater than, - for less than
  -mindepth N        do not apply tests at levels less than N
  -maxdepth N        descend at most N levels
  -mtime N           modified N*24 hours ago (+N=more than, -N=less than)
  -mmin N            modified N minutes ago
  -atime N           accessed N*24 hours ago
  -amin N            accessed N minutes ago
  -empty             file is empty or directory has no entries
  -executable        matches files which are executable
  -readable          matches files which are readable
  -writable          matches files which are writable

Actions:
  -print0            print full path with null terminator

Operators:
  -not               negate the next test

Examples:
  omni find . -name "*.go"                    # find Go files
  omni find . -type f -size +1M               # find files larger than 1MB
  omni find /tmp -type f -mtime +7            # files modified more than 7 days ago
  omni find . -name "*.log" -empty            # find empty log files
  omni find . -type d -name "node_modules"    # find node_modules directories
  omni find . -maxdepth 2 -type f             # files at most 2 levels deep
  omni find . -name "*.txt" -print0           # null-separated output

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

**Examples:**

```bash
omni find . -name "*.go"                    # find Go files
omni find . -type f -size +1M               # find files larger than 1MB
omni find /tmp -type f -mtime +7            # files modified more than 7 days ago
omni find . -name "*.log" -empty            # find empty log files
omni find . -type d -name "node_modules"    # find node_modules directories
omni find . -maxdepth 2 -type f             # files at most 2 levels deep
omni find . -name "*.txt" -print0           # null-separated output
```

---

### fold

**Category:** Text Processing

**Usage:** `omni fold [OPTION]... [FILE]... [flags]`

**Description:** Wrap each input line to fit in specified width

**Details:**

Wrap input lines in each FILE, writing to standard output.

With no FILE, or when FILE is -, read standard input.

  -w, --width=WIDTH  use WIDTH columns instead of 80
  -b, --bytes        count bytes rather than columns
  -s, --spaces       break at spaces

Examples:
  omni fold -w 40 file.txt     # wrap lines at 40 columns
  omni fold -s -w 72 README    # wrap at spaces, 72 columns

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --bytes | bool | false | count bytes rather than columns |
| -s, --spaces | bool | false | break at spaces |
| -w, --width | int | 80 | use WIDTH columns instead of 80 |

**Examples:**

```bash
omni fold -w 40 file.txt     # wrap lines at 40 columns
omni fold -s -w 72 README    # wrap at spaces, 72 columns
```

---

### for

**Category:** Other

**Usage:** `omni for`

**Description:** Loop and execute commands

**Details:**

Loop over items and execute commands for each.

Subcommands:
  range   Loop over a numeric range
  each    Loop over a list of items
  lines   Loop over lines from stdin or file
  split   Loop over items split by delimiter
  glob    Loop over files matching a pattern

Variable substitution:
  $item or ${item}   Current item value
  $i or ${i}         Current index (0-based)
  $n or ${n}         Current line number (1-based)
  $file or ${file}   Current file path

Examples:
  omni for range 1 5 -- echo $i
  omni for each a b c -- echo "Item: $item"
  omni for lines file.txt -- echo "Line: $line"
  omni for split "," "a,b,c" -- echo $item
  omni for glob "*.txt" -- cat $file

**Examples:**

```bash
omni for range 1 5 -- echo $i
omni for each a b c -- echo "Item: $item"
omni for lines file.txt -- echo "Line: $line"
omni for split "," "a,b,c" -- echo $item
omni for glob "*.txt" -- cat $file
```

**Subcommands:** `each`, `glob`, `lines`, `range`, `split`

---

### for each

**Category:** Other

**Usage:** `omni for each ITEM... -- COMMAND [flags]`

**Description:** Loop over a list of items

**Details:**

Loop over each item in the provided list.

Variable: $item or ${item}

Examples:
  omni for each apple banana cherry -- echo "Fruit: $item"
  omni for each *.go -- echo "File: $item"
  omni for each a b c --var=x -- echo "Value: $x"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | print commands without executing |
| --var | string | - | variable name to use |

**Examples:**

```bash
omni for each apple banana cherry -- echo "Fruit: $item"
omni for each *.go -- echo "File: $item"
omni for each a b c --var=x -- echo "Value: $x"
```

---

### for glob

**Category:** Other

**Usage:** `omni for glob PATTERN -- COMMAND [flags]`

**Description:** Loop over files matching a pattern

**Details:**

Loop over files matching a glob pattern.

Variable: $file or ${file}

Patterns:
  *.txt       All .txt files in current directory
  **/*.go     All .go files recursively
  src/*.js    All .js files in src/

Examples:
  omni for glob "*.txt" -- cat $file
  omni for glob "**/*.go" -- wc -l $file
  omni for glob "src/*.js" --dry-run -- echo $file

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | print commands without executing |
| --var | string | - | variable name to use |

**Examples:**

```bash
omni for glob "*.txt" -- cat $file
omni for glob "**/*.go" -- wc -l $file
omni for glob "src/*.js" --dry-run -- echo $file
```

---

### for lines

**Category:** Other

**Usage:** `omni for lines [FILE] -- COMMAND [flags]`

**Description:** Loop over lines from stdin or file

**Details:**

Loop over each line from stdin or a file.

Variables:
  $line or ${line}   Current line content
  $n or ${n}         Current line number (1-based)

Examples:
  cat file.txt | omni for lines -- echo "Line $n: $line"
  omni for lines input.txt -- echo "$n: $line"
  omni for lines --var=x -- echo "Got: $x"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | print commands without executing |
| --var | string | - | variable name to use |

**Examples:**

```bash
omni for lines input.txt -- echo "$n: $line"
omni for lines --var=x -- echo "Got: $x"
```

---

### for range

**Category:** Other

**Usage:** `omni for range START END [STEP] -- COMMAND [flags]`

**Description:** Loop over a numeric range

**Details:**

Loop from START to END (inclusive) with optional STEP.

Variable: $i or ${i}

Examples:
  omni for range 1 5 -- echo $i
  omni for range 10 0 -2 -- echo $i
  omni for range 1 100 -- echo "Number: $i"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | print commands without executing |
| --var | string | - | variable name to use |

**Examples:**

```bash
omni for range 1 5 -- echo $i
omni for range 10 0 -2 -- echo $i
omni for range 1 100 -- echo "Number: $i"
```

---

### for split

**Category:** Text Processing

**Usage:** `omni for split DELIMITER INPUT -- COMMAND [flags]`

**Description:** Loop over items split by delimiter

**Details:**

Split input by delimiter and loop over each item.

Variables:
  $item or ${item}   Current item
  $i or ${i}         Current index (0-based)

Examples:
  omni for split "," "a,b,c" -- echo "Item: $item"
  omni for split ":" "$PATH" -- echo "Dir: $item"
  omni for split "\\n" "$(cat file.txt)" -- process $item

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | print commands without executing |
| --var | string | - | variable name to use |

**Examples:**

```bash
omni for split "," "a,b,c" -- echo "Item: $item"
omni for split ":" "$PATH" -- echo "Dir: $item"
omni for split "\\n" "$(cat file.txt)" -- process $item
```

---

### free

**Category:** System Info

**Usage:** `omni free [OPTION]... [flags]`

**Description:** Display amount of free and used memory in the system

**Details:**

Display the total amount of free and used physical and swap memory
in the system, as well as the buffers and caches used by the kernel.

  -b, --bytes         show output in bytes
  -k, --kibibytes     show output in kibibytes (default)
  -m, --mebibytes     show output in mebibytes
  -g, --gibibytes     show output in gibibytes
  -h, --human         show human-readable output
  -w, --wide          wide output
  -t, --total         show total for RAM + swap

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

**Details:**

Alias for 'git branch-clean'.
Delete merged branches.

Examples:
  omni gbc --dry-run

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | Show branches that would be deleted |

**Examples:**

```bash
omni gbc --dry-run
```

---

### generate

**Category:** Code Generation

**Usage:** `omni generate`

**Description:** Code generation utilities

**Details:**

generate provides code generation utilities for scaffolding applications.

Subcommands:
  cobra init    Initialize a new Cobra CLI application
  cobra add     Add a new command to an existing Cobra application
  cobra config  Manage cobra generator configuration
  handler       Generate HTTP handler
  repository    Generate database repository
  test          Generate tests for a Go source file

Configuration:
  Default values can be set in ~/.cobra.yaml (compatible with cobra-cli).
  Command-line flags override config file values.

Examples:
  omni generate cobra init myapp --module github.com/user/myapp
  omni generate cobra add serve --parent root
  omni generate cobra config --show
  omni generate handler user --method GET,POST --framework chi
  omni generate repository user --entity User --table users
  omni generate test internal/cli/foo/foo.go

**Examples:**

```bash
omni generate cobra init myapp --module github.com/user/myapp
omni generate cobra add serve --parent root
omni generate cobra config --show
omni generate handler user --method GET,POST --framework chi
omni generate repository user --entity User --table users
omni generate test internal/cli/foo/foo.go
```

**Subcommands:** `cobra`, `handler`, `repository`, `test`

---

### generate cobra

**Category:** Other

**Usage:** `omni generate cobra`

**Description:** Cobra CLI application generator

**Details:**

Generate Cobra CLI applications and commands.

Configuration file (~/.cobra.yaml):
  author: Your Name <email@example.com>
  license: MIT
  useViper: true
  useService: false
  full: false

Subcommands:
  init    Initialize a new Cobra CLI application
  add     Add a new command to an existing application
  config  Manage generator configuration

**Subcommands:** `add`, `config`, `init`

---

### generate cobra add

**Category:** Other

**Usage:** `omni generate cobra add <command-name> [flags]`

**Description:** Add a new command to an existing Cobra application

**Details:**

Add a new command to an existing Cobra CLI application.

Creates a new command file in the cmd/ directory with the proper structure.

Examples:
  omni generate cobra add serve
  omni generate cobra add serve --parent root
  omni generate cobra add list --parent user --description "List all users"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --description | string | - | command description |
| --dir | string | - | project directory (defaults to current directory) |
| -p, --parent | string | root | parent command |

**Examples:**

```bash
omni generate cobra add serve
omni generate cobra add serve --parent root
omni generate cobra add list --parent user --description "List all users"
```

---

### generate cobra config

**Category:** Other

**Usage:** `omni generate cobra config [flags]`

**Description:** Manage cobra generator configuration

**Details:**

Manage the cobra generator configuration file.

The configuration file is stored at ~/.cobra.yaml and is compatible
with cobra-cli's configuration format.

Available options in config file:
  author: Your Name <email@example.com>
  license: MIT | Apache-2.0 | BSD-3
  useViper: true | false
  useService: true | false
  full: true | false

Examples:
  omni generate cobra config --show
  omni generate cobra config --init
  omni generate cobra config --init --author "John Doe" --license MIT

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

**Examples:**

```bash
omni generate cobra config --show
omni generate cobra config --init
omni generate cobra config --init --author "John Doe" --license MIT
```

---

### generate cobra init

**Category:** Other

**Usage:** `omni generate cobra init <directory> [flags]`

**Description:** Initialize a new Cobra CLI application

**Details:**

Initialize a new Cobra CLI application with all necessary scaffolding.

Configuration:
  Reads defaults from ~/.cobra.yaml if present.
  Command-line flags override config file values.

Creates (basic mode):
  - main.go          Entry point
  - cmd/root.go      Root command
  - cmd/version.go   Version command
  - go.mod           Go module
  - README.md        Documentation
  - Taskfile.yml     Task runner
  - .gitignore       Git ignore file
  - .editorconfig    Editor configuration
  - LICENSE          License file (optional)

With --viper:
  - internal/config/config.go  Viper configuration

With --service:
  - internal/parameters/config.go  Service parameters
  - internal/service/service.go    Service handler (uses inovacc/config)

With --full (includes all above plus):
  - .goreleaser.yaml              GoReleaser configuration
  - .golangci.yml                 GolangCI-Lint configuration (v2)
  - tools.go                      Build tool dependencies
  - .github/workflows/build.yml   GitHub Actions build workflow
  - .github/workflows/test.yml    GitHub Actions test workflow
  - .github/workflows/release.yaml GitHub Actions release workflow

Examples:
  omni generate cobra init myapp --module github.com/user/myapp
  omni generate cobra init ./apps/cli --module github.com/user/cli --viper
  omni generate cobra init myapp --module github.com/user/myapp --license MIT --author "John Doe"
  omni generate cobra init myapp --module github.com/user/myapp --full --service

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

**Examples:**

```bash
omni generate cobra init myapp --module github.com/user/myapp
omni generate cobra init ./apps/cli --module github.com/user/cli --viper
omni generate cobra init myapp --module github.com/user/myapp --license MIT --author "John Doe"
omni generate cobra init myapp --module github.com/user/myapp --full --service
```

---

### generate handler

**Category:** Other

**Usage:** `omni generate handler <name> [flags]`

**Description:** Generate HTTP handler

**Details:**

Generate an HTTP handler with the specified name.

Supports multiple frameworks: stdlib, chi, gin, echo

  -p, --package      Package name (default: "handler")
  -d, --dir          Output directory (default: "internal/handler")
  -m, --method       HTTP methods: GET,POST,PUT,DELETE,PATCH (default: GET,POST,PUT,DELETE)
  --path             URL path pattern
  --middleware       Include middleware support
  -f, --framework    Framework: stdlib, chi, gin, echo (default: stdlib)

Examples:
  omni generate handler user
  omni generate handler user --method GET,POST --framework chi
  omni generate handler user --dir handlers --package handlers
  omni generate handler product --middleware --framework gin

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --dir | string | internal/handler | output directory |
| -f, --framework | string | stdlib | framework: stdlib, chi, gin, echo |
| -m, --method | string | GET,POST,PUT,DELETE | HTTP methods (comma-separated) |
| --middleware | bool | false | include middleware support |
| -p, --package | string | handler | package name |
| --path | string | - | URL path pattern |

**Examples:**

```bash
omni generate handler user
omni generate handler user --method GET,POST --framework chi
omni generate handler user --dir handlers --package handlers
omni generate handler product --middleware --framework gin
```

---

### generate repository

**Category:** Other

**Usage:** `omni generate repository <name> [flags]`

**Description:** Generate database repository

**Details:**

Generate a database repository with the specified name.

  -p, --package      Package name (default: "repository")
  -d, --dir          Output directory (default: "internal/repository")
  -e, --entity       Entity struct name (default: capitalized name)
  -t, --table        Database table name (default: lowercase name + "s")
  --db               Database type: postgres, mysql, sqlite (default: postgres)
  --interface        Generate interface (default: true)

Examples:
  omni generate repository user
  omni generate repository user --entity User --table users
  omni generate repository product --db mysql
  omni generate repository order --interface=false

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --db | string | postgres | database type: postgres, mysql, sqlite |
| -d, --dir | string | internal/repository | output directory |
| -e, --entity | string | - | entity struct name |
| --interface | bool | true | generate interface |
| -p, --package | string | repository | package name |
| -t, --table | string | - | database table name |

**Examples:**

```bash
omni generate repository user
omni generate repository user --entity User --table users
omni generate repository product --db mysql
omni generate repository order --interface=false
```

---

### generate test

**Category:** Utilities

**Usage:** `omni generate test <file.go> [flags]`

**Description:** Generate tests for a Go source file

**Details:**

Generate test stubs for exported functions in a Go source file.

Parses the input file and generates test functions for all exported
functions and methods.

  --table           Generate table-driven tests (default: true)
  --parallel        Add t.Parallel() calls
  --mock            Generate mock setup
  --benchmark       Include benchmark tests
  --fuzz            Include fuzz tests (Go 1.18+)

Examples:
  omni generate test internal/cli/foo/foo.go
  omni generate test pkg/service/user.go --parallel
  omni generate test handler.go --table=false
  omni generate test service.go --benchmark --mock

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --benchmark | bool | false | include benchmark tests |
| --fuzz | bool | false | include fuzz tests |
| --mock | bool | false | generate mock setup |
| --parallel | bool | false | add t.Parallel() calls |
| --table | bool | true | generate table-driven tests |

**Examples:**

```bash
omni generate test internal/cli/foo/foo.go
omni generate test pkg/service/user.go --parallel
omni generate test handler.go --table=false
omni generate test service.go --benchmark --mock
```

---

### git

**Category:** Other

**Usage:** `omni git`

**Description:** Git shortcuts and hacks

**Details:**

Git shortcut commands for common operations.

**Subcommands:** `amend`, `blame-line`, `branch-clean`, `diff-words`, `fetch-all`, `log-graph`, `pull-rebase`, `push`, `quick-commit`, `stash-staged`, `status`, `undo`

---

### git amend

**Category:** Other

**Usage:** `omni git amend`

**Description:** Amend last commit without editing

**Details:**

Amend the last commit without editing the message.
Equivalent to: git commit --amend --no-edit

Examples:
  omni git amend

**Examples:**

```bash
omni git amend
```

---

### git blame-line

**Category:** Other

**Usage:** `omni git blame-line <file> [flags]`

**Description:** Blame specific line range

**Details:**

Show blame for a specific line range in a file.

Examples:
  omni git blame-line main.go --start 10 --end 20

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --end | int | 0 | End line number |
| --start | int | 0 | Start line number |

**Examples:**

```bash
omni git blame-line main.go --start 10 --end 20
```

---

### git branch-clean

**Category:** Other

**Usage:** `omni git branch-clean [flags]`

**Description:** Delete merged branches

**Details:**

Delete local branches that have been merged into the current branch.
Skips main, master, and develop branches.

Examples:
  omni git branch-clean
  omni git bc --dry-run

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | Show branches that would be deleted |

**Examples:**

```bash
omni git branch-clean
omni git bc --dry-run
```

---

### git diff-words

**Category:** Other

**Usage:** `omni git diff-words`

**Description:** Word-level diff

**Details:**

Show word-level diff instead of line-level.
Equivalent to: git diff --word-diff

Examples:
  omni git diff-words
  omni git diff-words HEAD~1

**Examples:**

```bash
omni git diff-words
omni git diff-words HEAD~1
```

---

### git fetch-all

**Category:** Other

**Usage:** `omni git fetch-all`

**Description:** Fetch all remotes with prune

**Details:**

Fetch all remotes with prune.
Equivalent to: git fetch --all --prune

Examples:
  omni git fetch-all
  omni git fa

**Examples:**

```bash
omni git fetch-all
omni git fa
```

---

### git log-graph

**Category:** Other

**Usage:** `omni git log-graph [flags]`

**Description:** Pretty log with graph

**Details:**

Show a pretty git log with graph visualization.
Equivalent to: git log --oneline --graph --decorate --all

Examples:
  omni git log-graph
  omni git lg -n 20

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 0 | Number of commits to show |

**Examples:**

```bash
omni git log-graph
omni git lg -n 20
```

---

### git pull-rebase

**Category:** Other

**Usage:** `omni git pull-rebase`

**Description:** Pull with rebase

**Details:**

Pull from remote with rebase.
Equivalent to: git pull --rebase

Examples:
  omni git pull-rebase
  omni git pr

**Examples:**

```bash
omni git pull-rebase
omni git pr
```

---

### git push

**Category:** Other

**Usage:** `omni git push [flags]`

**Description:** Push to remote

**Details:**

Push to the remote repository.

Examples:
  omni git push
  omni git push --force

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --force | bool | false | Force push (with lease) |

**Examples:**

```bash
omni git push
omni git push --force
```

---

### git quick-commit

**Category:** Other

**Usage:** `omni git quick-commit [flags]`

**Description:** Stage all and commit

**Details:**

Stage all changes and commit with a message.
Equivalent to: git add -A && git commit -m "message"

Examples:
  omni git quick-commit -m "fix bug"
  omni git qc -m "add feature"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | true | Stage all changes before commit |
| -m, --message | string | - | Commit message (required) |

**Examples:**

```bash
omni git quick-commit -m "fix bug"
omni git qc -m "add feature"
```

---

### git stash-staged

**Category:** Other

**Usage:** `omni git stash-staged [flags]`

**Description:** Stash only staged changes

**Details:**

Stash only staged changes, leaving unstaged changes in the working directory.

Examples:
  omni git stash-staged
  omni git stash-staged -m "WIP: feature"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -m, --message | string | - | Stash message |

**Examples:**

```bash
omni git stash-staged
omni git stash-staged -m "WIP: feature"
```

---

### git status

**Category:** Other

**Usage:** `omni git status`

**Description:** Short status

**Details:**

Show short git status.
Equivalent to: git status -sb

Examples:
  omni git status
  omni git st

**Examples:**

```bash
omni git status
omni git st
```

---

### git undo

**Category:** Other

**Usage:** `omni git undo`

**Description:** Undo last commit (soft reset)

**Details:**

Undo the last commit, keeping changes staged.
Equivalent to: git reset --soft HEAD~1

Examples:
  omni git undo

**Examples:**

```bash
omni git undo
```

---

### gops

**Category:** Other

**Usage:** `omni gops [PID] [flags]`

**Description:** Display Go process information

**Details:**

Display information about running Go processes.

Uses google/gops to detect Go processes and show their version
and build information.

Examples:
  omni gops           # list all Go processes
  omni gops -j        # output as JSON
  omni gops 1234      # show info for specific PID

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -j, --json | bool | false | output as JSON |

**Examples:**

```bash
omni gops           # list all Go processes
omni gops -j        # output as JSON
omni gops 1234      # show info for specific PID
```

---

### gqc

**Category:** Other

**Usage:** `omni gqc [flags]`

**Description:** Git quick commit (alias)

**Details:**

Alias for 'git quick-commit'.
Stage all changes and commit with a message.

Examples:
  omni gqc -m "fix bug"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | true | Stage all changes before commit |
| -m, --message | string | - | Commit message (required) |

**Examples:**

```bash
omni gqc -m "fix bug"
```

---

### grep

**Category:** Text Processing

**Usage:** `omni grep [options] PATTERN [FILE...] [flags]`

**Description:** Print lines that match patterns

**Details:**

Search for PATTERN in each FILE.
When FILE is '-', read standard input.
With no FILE, read '.' if recursive; otherwise, read standard input.

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

**Details:**

Decompress FILEs in gzip format.

Equivalent to gzip -d.

Examples:
  omni gunzip file.txt.gz      # decompress
  omni gunzip -k file.txt.gz   # keep original

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -f, --force | bool | false | force overwrite |
| -k, --keep | bool | false | keep original files |
| -c, --stdout | bool | false | write to stdout |
| -v, --verbose | bool | false | verbose mode |

**Examples:**

```bash
omni gunzip file.txt.gz      # decompress
omni gunzip -k file.txt.gz   # keep original
```

---

### gzip

**Category:** Archive

**Usage:** `omni gzip [OPTION]... [FILE]... [flags]`

**Description:** Compress or decompress files

**Details:**

Compress or decompress FILEs using gzip format.

  -d, --decompress   decompress
  -k, --keep         keep original files
  -f, --force        force overwrite
  -c, --stdout       write to stdout
  -v, --verbose      verbose mode
  -1 to -9           compression level (default 6)

Examples:
  omni gzip file.txt           # compress to file.txt.gz
  omni gzip -d file.txt.gz     # decompress
  omni gzip -k file.txt        # keep original
  omni gzip -c file.txt > out.gz  # write to stdout

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

**Examples:**

```bash
omni gzip file.txt           # compress to file.txt.gz
omni gzip -d file.txt.gz     # decompress
omni gzip -k file.txt        # keep original
omni gzip -c file.txt > out.gz  # write to stdout
```

---

### hash

**Category:** Hash & Encoding

**Usage:** `omni hash [OPTION]... [FILE]... [flags]`

**Description:** Compute and check file hashes

**Details:**

Print or check cryptographic hashes (checksums).

With no FILE, or when FILE is -, read standard input.

  -a, --algorithm ALG  hash algorithm: md5, sha1, sha256 (default), sha512
  -c, --check          read checksums from FILE and check them
  -b, --binary         read in binary mode
  -r, --recursive      hash files recursively in directories
      --quiet          don't print OK for each verified file
      --status         don't output anything, status code shows success
  -w, --warn           warn about improperly formatted checksum lines

Examples:
  omni hash file.txt                    # SHA256 hash
  omni hash -a md5 file.txt             # MD5 hash
  omni hash -r ./dir                    # hash all files in directory
  omni hash -c checksums.txt            # verify checksums
  omni hash file1 file2 > checksums.txt # create checksum file

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

**Examples:**

```bash
omni hash file.txt                    # SHA256 hash
omni hash -a md5 file.txt             # MD5 hash
omni hash -r ./dir                    # hash all files in directory
omni hash -c checksums.txt            # verify checksums
omni hash file1 file2 > checksums.txt # create checksum file
```

---

### head

**Category:** Text Processing

**Usage:** `omni head [option]... [file]... [flags]`

**Description:** Output the first part of files

**Details:**

Print the first 10 lines of each FILE to standard output.
With more than one FILE, precede each with a header giving the file name.
With no FILE, or when FILE is -, read standard input.

Numeric shortcuts are supported: -80 is equivalent to -n 80.

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

**Details:**

Hexadecimal encoding and decoding utilities.

Subcommands:
  encode    Encode text to hexadecimal
  decode    Decode hexadecimal to text

Examples:
  omni hex encode "hello"
  omni hex decode "68656c6c6f"
  echo "test" | omni hex encode

**Examples:**

```bash
omni hex encode "hello"
omni hex decode "68656c6c6f"
```

**Subcommands:** `decode`, `encode`

---

### hex decode

**Category:** Other

**Usage:** `omni hex decode [HEX] [flags]`

**Description:** Decode hexadecimal to text

**Details:**

Decode hexadecimal string back to text.

Accepts hex strings with or without separators (spaces, colons, dashes).

Examples:
  omni hex decode "68656c6c6f"         # Output: hello
  omni hex decode "68:65:6c:6c:6f"     # With colons
  omni hex decode "68 65 6c 6c 6f"     # With spaces
  echo "74657374" | omni hex decode    # Read from stdin

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni hex decode "68656c6c6f"         # Output: hello
omni hex decode "68:65:6c:6c:6f"     # With colons
omni hex decode "68 65 6c 6c 6f"     # With spaces
```

---

### hex encode

**Category:** Other

**Usage:** `omni hex encode [TEXT] [flags]`

**Description:** Encode text to hexadecimal

**Details:**

Encode text to hexadecimal representation.

Each byte is converted to its two-character hex representation.

Examples:
  omni hex encode "hello"              # Output: 68656c6c6f
  omni hex encode --upper "hello"      # Output: 68656C6C6F
  echo "test" | omni hex encode        # Read from stdin
  omni hex encode file.txt             # Read from file

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |
| -u, --upper | bool | false | use uppercase hex letters |

**Examples:**

```bash
omni hex encode "hello"              # Output: 68656c6c6f
omni hex encode --upper "hello"      # Output: 68656C6C6F
omni hex encode file.txt             # Read from file
```

---

### html

**Category:** Other

**Usage:** `omni html`

**Description:** HTML utilities (format, encode, decode)

**Details:**

HTML utilities for formatting, encoding, and decoding.

Subcommands:
  fmt       Format/beautify HTML
  minify    Minify HTML
  validate  Validate HTML syntax
  encode    HTML encode text (escape special characters)
  decode    HTML decode text (unescape entities)

Examples:
  omni html fmt file.html
  omni html minify file.html
  omni html validate file.html
  omni html encode "<script>alert('xss')</script>"
  omni html decode "&lt;div&gt;content&lt;/div&gt;"

**Examples:**

```bash
omni html fmt file.html
omni html minify file.html
omni html validate file.html
omni html encode "<script>alert('xss')</script>"
omni html decode "&lt;div&gt;content&lt;/div&gt;"
```

**Subcommands:** `decode`, `encode`, `fmt`, `minify`, `validate`

---

### html decode

**Category:** Other

**Usage:** `omni html decode [TEXT] [flags]`

**Description:** HTML decode text

**Details:**

HTML decode text by unescaping HTML entities.

Converts HTML entities like &lt;, &gt;, &amp;, &quot; back to their original characters.

Examples:
  omni html decode "&lt;script&gt;"
  omni html decode "Tom &amp; Jerry"
  echo "&lt;div&gt;" | omni html decode

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni html decode "&lt;script&gt;"
omni html decode "Tom &amp; Jerry"
```

---

### html encode

**Category:** Other

**Usage:** `omni html encode [TEXT] [flags]`

**Description:** HTML encode text

**Details:**

HTML encode text by escaping special characters.

Converts characters like <, >, &, ", and ' to their HTML entity equivalents.

Examples:
  omni html encode "<script>alert('xss')</script>"
  omni html encode "Tom & Jerry"
  echo "<div>" | omni html encode

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni html encode "<script>alert('xss')</script>"
omni html encode "Tom & Jerry"
```

---

### html fmt

**Category:** Other

**Usage:** `omni html fmt [FILE] [flags]`

**Description:** Format/beautify HTML

**Details:**

Format HTML with proper indentation.

  -i, --indent=STR     indentation string (default "  ")
  --sort-attrs         sort attributes alphabetically

Examples:
  omni html fmt file.html
  omni html fmt "<div><p>text</p></div>"
  cat file.html | omni html fmt
  omni html fmt --sort-attrs file.html

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| --sort-attrs | bool | false | sort attributes alphabetically |

**Examples:**

```bash
omni html fmt file.html
omni html fmt "<div><p>text</p></div>"
omni html fmt --sort-attrs file.html
```

---

### html minify

**Category:** Other

**Usage:** `omni html minify [FILE]`

**Description:** Minify HTML

**Details:**

Minify HTML by removing unnecessary whitespace and comments.

Examples:
  omni html minify file.html
  cat file.html | omni html minify

**Examples:**

```bash
omni html minify file.html
```

---

### html validate

**Category:** Other

**Usage:** `omni html validate [FILE] [flags]`

**Description:** Validate HTML syntax

**Details:**

Validate HTML syntax.

Exit codes:
  0  Valid HTML
  1  Invalid HTML or error

  --json    output result as JSON

Examples:
  omni html validate file.html
  omni html validate "<div><p>text</p></div>"
  omni html validate --json file.html

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni html validate file.html
omni html validate "<div><p>text</p></div>"
omni html validate --json file.html
```

---

### id

**Category:** System Info

**Usage:** `omni id [OPTION]... [USER] [flags]`

**Description:** Print user and group information

**Details:**

Print user and group information for the specified USER,
or (when USER omitted) for the current user.

  -g, --group   print only the effective group ID
  -G, --groups  print all group IDs
  -n, --name    print a name instead of a number, for -ugG
  -r, --real    print the real ID instead of the effective ID, with -ugG
  -u, --user    print only the effective user ID

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

**Details:**

For each pair of input lines with identical join fields, write a line to
standard output. The default join field is the first, delimited by blanks.

  -1 FIELD       join on this FIELD of file 1
  -2 FIELD       join on this FIELD of file 2
  -t CHAR        use CHAR as input and output field separator
  -i             ignore differences in case when comparing fields
  -a FILENUM     also print unpairable lines from file FILENUM
  -v FILENUM     print only unpairable lines from file FILENUM
  -e EMPTY       replace missing fields with EMPTY

When FILE1 or FILE2 is -, read standard input.

Examples:
  omni join file1.txt file2.txt           # join on first field
  omni join -1 2 -2 1 file1.txt file2.txt # join field 2 of file1 with field 1 of file2
  omni join -t ',' data1.csv data2.csv    # join CSV files

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

**Examples:**

```bash
omni join file1.txt file2.txt           # join on first field
omni join -1 2 -2 1 file1.txt file2.txt # join field 2 of file1 with field 1 of file2
omni join -t ',' data1.csv data2.csv    # join CSV files
```

---

### jq

**Category:** Data Processing

**Usage:** `omni jq [OPTION]... FILTER [FILE]... [flags]`

**Description:** Command-line JSON processor

**Details:**

jq is a lightweight command-line JSON processor.

This is a simplified implementation supporting common operations:
  .           identity (output input unchanged)
  .field      access object field
  .field.sub  nested field access
  .[n]        array index
  .[]         iterate array elements
  keys        get object/array keys
  length      get length
  type        get type name

  -r          output raw strings (no quotes)
  -c          compact output
  -s          slurp: read all inputs into array
  -n          null input
  --tab       use tabs for indentation

Examples:
  echo '{"name":"John"}' | omni jq '.name'
  echo '[1,2,3]' | omni jq '.[]'
  echo '{"a":{"b":1}}' | omni jq '.a.b'
  omni jq -r '.name' data.json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --compact-output | bool | false | compact output |
| -n, --null-input | bool | false | don't read any input |
| -r, --raw-output | bool | false | output raw strings |
| -s, --slurp | bool | false | read all inputs into array |
| -S, --sort-keys | bool | false | sort object keys |
| --tab | bool | false | use tabs for indentation |

**Examples:**

```bash
omni jq -r '.name' data.json
```

---

### json

**Category:** Data Processing

**Usage:** `omni json`

**Description:** JSON utilities (format, minify, validate)

**Details:**

JSON utilities for formatting, minifying, and validating JSON data.

Subcommands:
  fmt       Beautify/format JSON with indentation
  minify    Compact JSON by removing whitespace
  validate  Check if input is valid JSON
  stats     Show statistics about JSON data
  keys      List all keys in JSON object
  toyaml    Convert JSON to YAML
  fromyaml  Convert YAML to JSON
  fromtoml  Convert TOML to JSON
  tostruct  Convert JSON to Go struct definition
  tocsv     Convert JSON array to CSV
  fromcsv   Convert CSV to JSON array
  toxml     Convert JSON to XML
  fromxml   Convert XML to JSON

Examples:
  omni json fmt file.json              # beautify JSON
  omni json minify file.json           # compact JSON
  omni json validate file.json         # check if valid
  echo '{"a":1}' | omni json fmt       # from stdin
  omni json stats file.json            # show statistics
  omni json toyaml file.json           # convert to YAML
  omni json fromyaml file.yaml         # convert from YAML
  omni json fromtoml file.toml         # convert from TOML
  omni json tostruct file.json         # convert to Go struct
  omni json tocsv file.json            # convert to CSV
  omni json fromcsv file.csv           # convert from CSV
  omni json toxml file.json            # convert to XML
  omni json fromxml file.xml           # convert from XML

**Examples:**

```bash
omni json fmt file.json              # beautify JSON
omni json minify file.json           # compact JSON
omni json validate file.json         # check if valid
omni json stats file.json            # show statistics
omni json toyaml file.json           # convert to YAML
omni json fromyaml file.yaml         # convert from YAML
omni json fromtoml file.toml         # convert from TOML
omni json tostruct file.json         # convert to Go struct
omni json tocsv file.json            # convert to CSV
omni json fromcsv file.csv           # convert from CSV
omni json toxml file.json            # convert to XML
omni json fromxml file.xml           # convert from XML
```

**Subcommands:** `fmt`, `fromcsv`, `fromtoml`, `fromxml`, `fromyaml`, `keys`, `minify`, `stats`, `tocsv`, `tostruct`, `toxml`, `toyaml`, `validate`

---

### json fmt

**Category:** Other

**Usage:** `omni json fmt [FILE]... [flags]`

**Description:** Beautify/format JSON with indentation

**Details:**

Format JSON with proper indentation and line breaks.

  -i, --indent=STR   indentation string (default "  ")
  -t, --tab          use tabs for indentation
  -s, --sort-keys    sort object keys alphabetically
  -e, --escape-html  escape HTML characters (<, >, &)

Examples:
  omni json fmt file.json              # beautify with 2-space indent
  omni json fmt -t file.json           # use tabs
  omni json fmt -s file.json           # sort keys
  echo '{"b":2,"a":1}' | omni json fmt -s

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -e, --escape-html | bool | false | escape HTML characters |
| -i, --indent | string |    | indentation string |
| -s, --sort-keys | bool | false | sort object keys |
| -t, --tab | bool | false | use tabs for indentation |

**Examples:**

```bash
omni json fmt file.json              # beautify with 2-space indent
omni json fmt -t file.json           # use tabs
omni json fmt -s file.json           # sort keys
```

---

### json fromcsv

**Category:** Other

**Usage:** `omni json fromcsv [FILE] [flags]`

**Description:** Convert CSV to JSON array

**Details:**

Convert CSV data to JSON array of objects.

  --no-header          first row is data, not headers
  -d, --delimiter=STR  field delimiter (default ",")
  -a, --array          always output as array (even for single row)

Examples:
  omni json fromcsv file.csv
  cat file.csv | omni json fromcsv
  omni json fromcsv -d ";" file.csv    # semicolon delimiter
  omni json fromcsv --no-header file.csv

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --array | bool | false | always output as array |
| -d, --delimiter | string | , | field delimiter |
| --no-header | bool | false | first row is data, not headers |

**Examples:**

```bash
omni json fromcsv file.csv
omni json fromcsv -d ";" file.csv    # semicolon delimiter
omni json fromcsv --no-header file.csv
```

---

### json fromtoml

**Category:** Other

**Usage:** `omni json fromtoml [FILE] [flags]`

**Description:** Convert TOML to JSON

**Details:**

Convert TOML data to JSON format.

  -m, --minify    output minified JSON (no indentation)

Examples:
  omni json fromtoml file.toml
  cat file.toml | omni json fromtoml
  omni json fromtoml -m file.toml     # minified output
  omni json fromtoml file.toml > output.json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -m, --minify | bool | false | output minified JSON |

**Examples:**

```bash
omni json fromtoml file.toml
omni json fromtoml -m file.toml     # minified output
omni json fromtoml file.toml > output.json
```

---

### json fromxml

**Category:** Other

**Usage:** `omni json fromxml [FILE] [flags]`

**Description:** Convert XML to JSON

**Details:**

Convert XML data to JSON format.

  --attr-prefix=STR    prefix for attributes in JSON (default "-")
  --text-key=STR       key for text content (default "#text")

Examples:
  omni json fromxml file.xml
  cat file.xml | omni json fromxml
  omni json fromxml --attr-prefix=@ file.xml

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --attr-prefix | string | - | prefix for attributes in JSON |
| --text-key | string | #text | key for text content |

**Examples:**

```bash
omni json fromxml file.xml
omni json fromxml --attr-prefix=@ file.xml
```

---

### json fromyaml

**Category:** Other

**Usage:** `omni json fromyaml [FILE] [flags]`

**Description:** Convert YAML to JSON

**Details:**

Convert YAML data to JSON format.

  -m, --minify    output minified JSON (no indentation)

Examples:
  omni json fromyaml file.yaml
  cat file.yaml | omni json fromyaml
  omni json fromyaml -m file.yaml     # minified output
  omni json fromyaml file.yaml > output.json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -m, --minify | bool | false | output minified JSON |

**Examples:**

```bash
omni json fromyaml file.yaml
omni json fromyaml -m file.yaml     # minified output
omni json fromyaml file.yaml > output.json
```

---

### json keys

**Category:** Other

**Usage:** `omni json keys [FILE] [flags]`

**Description:** List all keys in JSON object

**Details:**

List all keys (paths) in a JSON object recursively.

Examples:
  omni json keys file.json
  echo '{"a":{"b":1}}' | omni json keys

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni json keys file.json
```

---

### json minify

**Category:** Other

**Usage:** `omni json minify [FILE]... [flags]`

**Description:** Compact JSON by removing whitespace

**Details:**

Remove all unnecessary whitespace from JSON.

Examples:
  omni json minify file.json
  cat file.json | omni json minify
  omni json minify -s file.json        # also sort keys

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -s, --sort-keys | bool | false | sort object keys |

**Examples:**

```bash
omni json minify file.json
omni json minify -s file.json        # also sort keys
```

---

### json stats

**Category:** Other

**Usage:** `omni json stats [FILE] [flags]`

**Description:** Show statistics about JSON data

**Details:**

Display statistics about JSON data including type, depth, size, etc.

Examples:
  omni json stats file.json
  echo '[1,2,3]' | omni json stats

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni json stats file.json
```

---

### json tocsv

**Category:** Other

**Usage:** `omni json tocsv [FILE] [flags]`

**Description:** Convert JSON array to CSV

**Details:**

Convert JSON array of objects to CSV format.

Nested objects are flattened with dot notation (e.g., address.city).

  --no-header          don't include header row
  -d, --delimiter=STR  field delimiter (default ",")
  --no-quotes          don't quote fields

Examples:
  omni json tocsv file.json
  echo '[{"name":"John","age":30}]' | omni json tocsv
  omni json tocsv -d ";" file.json     # semicolon delimiter
  omni json tocsv --no-header file.json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --delimiter | string | , | field delimiter |
| --no-header | bool | false | don't include header row |
| --no-quotes | bool | false | don't quote fields |

**Examples:**

```bash
omni json tocsv file.json
omni json tocsv -d ";" file.json     # semicolon delimiter
omni json tocsv --no-header file.json
```

---

### json tostruct

**Category:** Other

**Usage:** `omni json tostruct [FILE] [flags]`

**Description:** Convert JSON to Go struct definition

**Details:**

Convert JSON data to a Go struct definition.

  -n, --name=NAME      struct name (default "Root")
  -p, --package=PKG    package name (default "main")
  --inline             inline nested structs
  --omitempty          add omitempty to all fields

Examples:
  omni json tostruct file.json
  echo '{"name":"test","count":1}' | omni json tostruct
  omni json tostruct -n User -p models file.json
  omni json tostruct --omitempty file.json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --inline | bool | false | inline nested structs |
| -n, --name | string | Root | struct name |
| --omitempty | bool | false | add omitempty to all fields |
| -p, --package | string | main | package name |

**Examples:**

```bash
omni json tostruct file.json
omni json tostruct -n User -p models file.json
omni json tostruct --omitempty file.json
```

---

### json toxml

**Category:** Other

**Usage:** `omni json toxml [FILE] [flags]`

**Description:** Convert JSON to XML

**Details:**

Convert JSON data to XML format.

  -r, --root=NAME      root element name (default "root")
  -i, --indent=STR     indentation string (default "  ")
  --item-tag=NAME      tag for array items (default "item")
  --attr-prefix=STR    prefix for attributes (default "-")

Examples:
  omni json toxml file.json
  echo '{"name":"John"}' | omni json toxml
  omni json toxml -r person file.json   # custom root
  omni json toxml --item-tag=entry file.json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --attr-prefix | string | - | prefix for attributes |
| -i, --indent | string |    | indentation string |
| --item-tag | string | item | tag for array items |
| -r, --root | string | root | root element name |

**Examples:**

```bash
omni json toxml file.json
omni json toxml -r person file.json   # custom root
omni json toxml --item-tag=entry file.json
```

---

### json toyaml

**Category:** Other

**Usage:** `omni json toyaml [FILE]`

**Description:** Convert JSON to YAML

**Details:**

Convert JSON data to YAML format.

Examples:
  omni json toyaml file.json
  echo '{"name":"test"}' | omni json toyaml
  omni json toyaml file.json > output.yaml

**Examples:**

```bash
omni json toyaml file.json
omni json toyaml file.json > output.yaml
```

---

### json validate

**Category:** Other

**Usage:** `omni json validate [FILE]... [flags]`

**Description:** Check if input is valid JSON

**Details:**

Validate JSON syntax without outputting the data.

Exit codes:
  0  Valid JSON
  1  Invalid JSON or error

  --json    output result as JSON

Examples:
  omni json validate file.json
  omni json validate --json file.json
  echo '{"valid": true}' | omni json validate

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output result as JSON |

**Examples:**

```bash
omni json validate file.json
omni json validate --json file.json
```

---

### jwt

**Category:** Other

**Usage:** `omni jwt`

**Description:** JWT (JSON Web Token) utilities

**Details:**

JWT (JSON Web Token) utilities.

Subcommands:
  decode    Decode and inspect a JWT token

Examples:
  omni jwt decode "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  echo $TOKEN | omni jwt decode
  omni jwt decode --header token.txt

**Examples:**

```bash
omni jwt decode "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
omni jwt decode --header token.txt
```

**Subcommands:** `decode`

---

### jwt decode

**Category:** Other

**Usage:** `omni jwt decode [TOKEN] [flags]`

**Description:** Decode and inspect a JWT token

**Details:**

Decode and inspect a JWT token.

Displays the header and payload of a JWT token. Does NOT verify the signature
(use a proper JWT library for that). Useful for debugging and inspection.

Examples:
  omni jwt decode "eyJhbGciOiJIUzI1NiIs..."
  omni jwt decode --header "eyJhbGciOiJIUzI1NiIs..."
  omni jwt decode --payload "eyJhbGciOiJIUzI1NiIs..."
  omni jwt decode --json "eyJhbGciOiJIUzI1NiIs..."
  echo $TOKEN | omni jwt decode

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -H, --header | bool | false | show only header |
| --json | bool | false | output as JSON |
| -p, --payload | bool | false | show only payload |
| --raw | bool | false | output raw JSON without formatting |

**Examples:**

```bash
omni jwt decode "eyJhbGciOiJIUzI1NiIs..."
omni jwt decode --header "eyJhbGciOiJIUzI1NiIs..."
omni jwt decode --payload "eyJhbGciOiJIUzI1NiIs..."
omni jwt decode --json "eyJhbGciOiJIUzI1NiIs..."
```

---

### kconfig

**Category:** Other

**Usage:** `omni kconfig`

**Description:** Show kubeconfig info

**Details:**

Show current kubeconfig context and cluster info.

Examples:
  omni kconfig

**Examples:**

```bash
omni kconfig
```

---

### kcs

**Category:** Other

**Usage:** `omni kcs [context]`

**Description:** Switch kubectl context

**Details:**

Switch to a different kubectl context.
Without arguments, lists available contexts.

Examples:
  omni kcs              # list contexts
  omni kcs production   # switch to production

**Examples:**

```bash
omni kcs              # list contexts
omni kcs production   # switch to production
```

---

### kdebug

**Category:** Other

**Usage:** `omni kdebug <pod> [flags]`

**Description:** Debug pod with ephemeral container

**Details:**

Run an ephemeral debug container in a pod.
Equivalent to: kubectl debug -it <pod> --image=<image>

Examples:
  omni kdebug mypod
  omni kdebug mypod --image=nicolaka/netshoot

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --image | string | - | Debug container image (default: busybox) |
| -n, --namespace | string | - | Namespace |

**Examples:**

```bash
omni kdebug mypod
omni kdebug mypod --image=nicolaka/netshoot
```

---

### kdp

**Category:** Other

**Usage:** `omni kdp <selector> [flags]`

**Description:** Delete pods by selector

**Details:**

Delete pods by label selector.
Equivalent to: kubectl delete pods -l <selector>

Examples:
  omni kdp app=nginx
  omni kdp app=nginx -n default --force

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --force | bool | false | Force delete |
| -n, --namespace | string | - | Namespace |

**Examples:**

```bash
omni kdp app=nginx
omni kdp app=nginx -n default --force
```

---

### kdrain

**Category:** Other

**Usage:** `omni kdrain <node> [flags]`

**Description:** Drain node for maintenance

**Details:**

Drain a node for maintenance.
Equivalent to: kubectl drain <node>

Examples:
  omni kdrain mynode
  omni kdrain mynode --ignore-daemonsets

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --delete-emptydir | bool | false | Delete emptydir data |
| --ignore-daemonsets | bool | false | Ignore daemonsets |

**Examples:**

```bash
omni kdrain mynode
omni kdrain mynode --ignore-daemonsets
```

---

### keb

**Category:** Other

**Usage:** `omni keb <pod> [flags]`

**Description:** Exec into pod with bash

**Details:**

Exec into a pod with bash (falls back to sh).
Equivalent to: kubectl exec -it <pod> -- /bin/bash

Examples:
  omni keb mypod
  omni keb mypod -n default
  omni keb mypod -c mycontainer

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --container | string | - | Container name |
| -n, --namespace | string | - | Namespace |

**Examples:**

```bash
omni keb mypod
omni keb mypod -n default
omni keb mypod -c mycontainer
```

---

### kga

**Category:** Other

**Usage:** `omni kga [flags]`

**Description:** Get all resources in namespace

**Details:**

Get all common resources in a namespace.
Equivalent to: kubectl get pods,svc,deploy,... -o wide

Examples:
  omni kga
  omni kga -n kube-system
  omni kga -A

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -A, --all-namespaces | bool | false | All namespaces |
| -n, --namespace | string | - | Namespace |

**Examples:**

```bash
omni kga
omni kga -n kube-system
omni kga -A
```

---

### kge

**Category:** Other

**Usage:** `omni kge [flags]`

**Description:** Get events sorted by time

**Details:**

Get events sorted by last timestamp.
Equivalent to: kubectl get events --sort-by='.lastTimestamp'

Examples:
  omni kge
  omni kge -n kube-system
  omni kge -A

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -A, --all-namespaces | bool | false | All namespaces |
| -n, --namespace | string | - | Namespace |

**Examples:**

```bash
omni kge
omni kge -n kube-system
omni kge -A
```

---

### kill

**Category:** System Info

**Usage:** `omni kill [OPTION]... PID... [flags]`

**Description:** Send a signal to a process

**Details:**

Send the specified signal to the specified processes.

  -s, --signal=SIGNAL  specify the signal to be sent
  -l, --list           list signal names
  -v, --verbose        report successful signals
  -j, --json           output as JSON

Signal can be specified by name (e.g., HUP, KILL, TERM) or number.
Common signals:
   1) SIGHUP       2) SIGINT       3) SIGQUIT
   9) SIGKILL     15) SIGTERM (default)

Examples:
  omni kill 1234           # send SIGTERM to process 1234
  omni kill -9 1234        # send SIGKILL to process 1234
  omni kill -s HUP 1234    # send SIGHUP to process 1234
  omni kill -l             # list all signal names
  omni kill -l -j          # list signals as JSON
  omni kill -j 1234        # kill with JSON output

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -j, --json | bool | false | output as JSON |
| -l, --list | bool | false | list signal names |
| -s, --signal | string | - | specify the signal to be sent |
| -v, --verbose | bool | false | report successful signals |

**Examples:**

```bash
omni kill 1234           # send SIGTERM to process 1234
omni kill -9 1234        # send SIGKILL to process 1234
omni kill -s HUP 1234    # send SIGHUP to process 1234
omni kill -l             # list all signal names
omni kill -l -j          # list signals as JSON
omni kill -j 1234        # kill with JSON output
```

---

### klf

**Category:** Other

**Usage:** `omni klf <pod> [flags]`

**Description:** Follow pod logs with timestamps

**Details:**

Follow logs for a pod with timestamps.
Equivalent to: kubectl logs -f --timestamps <pod>

Examples:
  omni klf mypod
  omni klf mypod -n default -c mycontainer
  omni klf mypod --tail 100

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --container | string | - | Container name |
| -n, --namespace | string | - | Namespace |
| --tail | int | 0 | Lines to show from end of logs |

**Examples:**

```bash
omni klf mypod
omni klf mypod -n default -c mycontainer
omni klf mypod --tail 100
```

---

### kns

**Category:** Other

**Usage:** `omni kns [namespace]`

**Description:** Switch default namespace

**Details:**

Switch the default namespace for the current context.
Without arguments, lists all namespaces.

Examples:
  omni kns           # list namespaces
  omni kns default   # switch to default namespace

**Examples:**

```bash
omni kns           # list namespaces
omni kns default   # switch to default namespace
```

---

### kpf

**Category:** Other

**Usage:** `omni kpf <pod|svc/name> <local:remote> [flags]`

**Description:** Quick port forward

**Details:**

Quick port forward to a pod or service.
Equivalent to: kubectl port-forward <target> <local>:<remote>

Examples:
  omni kpf mypod 8080:80
  omni kpf svc/myservice 3000:80 -n default

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --namespace | string | - | Namespace |

**Examples:**

```bash
omni kpf mypod 8080:80
omni kpf svc/myservice 3000:80 -n default
```

---

### krr

**Category:** Other

**Usage:** `omni krr <deployment> [flags]`

**Description:** Rollout restart deployment

**Details:**

Restart a deployment using rollout restart.
Equivalent to: kubectl rollout restart deployment/<name>

Examples:
  omni krr mydeployment
  omni krr mydeployment -n default

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --namespace | string | - | Namespace |

**Examples:**

```bash
omni krr mydeployment
omni krr mydeployment -n default
```

---

### krun

**Category:** Other

**Usage:** `omni krun <name> --image=<image> [-- command] [flags]`

**Description:** Run a one-off pod

**Details:**

Run a one-off pod that auto-deletes after completion.
Equivalent to: kubectl run <name> --image=<image> --rm -it --restart=Never

Examples:
  omni krun test --image=busybox -- sh
  omni krun curl --image=curlimages/curl -- curl google.com

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --image | string | - | Container image (required) |
| -n, --namespace | string | - | Namespace |

**Examples:**

```bash
omni krun test --image=busybox -- sh
omni krun curl --image=curlimages/curl -- curl google.com
```

---

### kscale

**Category:** Other

**Usage:** `omni kscale <deployment> <replicas> [flags]`

**Description:** Scale deployment

**Details:**

Scale a deployment to the specified number of replicas.

Examples:
  omni kscale mydeployment 3
  omni kscale mydeployment 0 -n default

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --namespace | string | - | Namespace |

**Examples:**

```bash
omni kscale mydeployment 3
omni kscale mydeployment 0 -n default
```

---

### ksuid

**Category:** Other

**Usage:** `omni ksuid [OPTION]... [flags]`

**Description:** Generate K-Sortable Unique IDentifiers

**Details:**

Generate KSUIDs (K-Sortable Unique IDentifiers).

KSUIDs are 27-character, base62-encoded identifiers that are:
- Globally unique
- Naturally sortable by generation time
- URL-safe and case-sensitive

Structure: 4-byte timestamp + 16-byte random payload

  -n, --count=N   generate N KSUIDs (default 1)
  --json          output as JSON

Examples:
  omni ksuid                  # generate one KSUID
  omni ksuid -n 5             # generate 5 KSUIDs
  omni ksuid --json           # JSON output

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | generate N KSUIDs |
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni ksuid                  # generate one KSUID
omni ksuid -n 5             # generate 5 KSUIDs
omni ksuid --json           # JSON output
```

---

### ktn

**Category:** Other

**Usage:** `omni ktn [flags]`

**Description:** Top nodes by resource usage

**Details:**

Show top nodes by resource usage.
Equivalent to: kubectl top nodes

Examples:
  omni ktn
  omni ktn --sort-by cpu

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --sort-by | string | - | Sort by (cpu or memory) |

**Examples:**

```bash
omni ktn
omni ktn --sort-by cpu
```

---

### ktp

**Category:** Other

**Usage:** `omni ktp [flags]`

**Description:** Top pods by resource usage

**Details:**

Show top pods by resource usage.
Equivalent to: kubectl top pods

Examples:
  omni ktp
  omni ktp -n default
  omni ktp --sort-by cpu

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -A, --all-namespaces | bool | false | All namespaces |
| -n, --namespace | string | - | Namespace |
| --sort-by | string | - | Sort by (cpu or memory) |

**Examples:**

```bash
omni ktp
omni ktp -n default
omni ktp --sort-by cpu
```

---

### kubectl

**Category:** Other

**Usage:** `omni kubectl`

**Description:** Kubernetes CLI

**Details:**

Kubernetes command-line tool integrated into omni.

This is a full integration of kubectl, supporting all kubectl commands and flags.
You can use 'omni kubectl' or the shorter alias 'omni k'.

Examples:
  omni kubectl get pods
  omni k get pods -A
  omni k describe node mynode
  omni k logs -f mypod
  omni k exec -it mypod -- /bin/sh
  omni k apply -f manifest.yaml

**Examples:**

```bash
omni kubectl get pods
omni k get pods -A
omni k describe node mynode
omni k logs -f mypod
omni k exec -it mypod -- /bin/sh
omni k apply -f manifest.yaml
```

---

### kwp

**Category:** Other

**Usage:** `omni kwp [flags]`

**Description:** Watch pods continuously

**Details:**

Watch pods continuously.
Equivalent to: kubectl get pods -w

Examples:
  omni kwp
  omni kwp -n default
  omni kwp -l app=nginx

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -A, --all-namespaces | bool | false | All namespaces |
| -n, --namespace | string | - | Namespace |
| -l, --selector | string | - | Label selector |

**Examples:**

```bash
omni kwp
omni kwp -n default
omni kwp -l app=nginx
```

---

### less

**Category:** Other

**Usage:** `omni less [OPTION]... [FILE] [flags]`

**Description:** View file contents with scrolling

**Details:**

Display file contents one screen at a time with scrolling support.

Navigation:
  j, Down, Enter  Scroll down one line
  k, Up           Scroll up one line
  Space, PgDn     Scroll down one page
  PgUp            Scroll up one page
  g, Home         Go to beginning
  G, End          Go to end
  /               Search forward
  n               Next search match
  N               Previous search match
  h               Show help
  q               Quit

Examples:
  omni less file.txt
  omni less -N file.txt     # with line numbers
  omni less -S file.txt     # chop long lines
  cat file.txt | omni less  # from stdin

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -N, --LINE-NUMBERS | bool | false | show line numbers |
| -S, --chop-long-lines | bool | false | truncate long lines |
| -i, --ignore-case | bool | false | case-insensitive search |
| -X, --no-init | bool | false | don't clear screen on start |
| -F, --quit-if-one-screen | bool | false | quit if content fits on one screen |
| -R, --raw-control-chars | bool | false | show raw control characters |

**Examples:**

```bash
omni less file.txt
omni less -N file.txt     # with line numbers
omni less -S file.txt     # chop long lines
```

---

### lint

**Category:** Tooling

**Usage:** `omni lint [OPTION]... [FILE|DIR]... [flags]`

**Description:** Check Taskfiles for portability issues

**Details:**

Lint Taskfiles for cross-platform portability.

Checks for:
  - Shell commands that should use omni equivalents
  - Non-portable commands (package managers, OS-specific tools)
  - Bash-specific syntax ([[ ]], <<<, etc.)
  - Hardcoded Unix paths
  - Pipe chains that may fail silently

Severity levels:
  error   - Will likely fail on some platforms
  warning - May cause issues, should be reviewed
  info    - Suggestions for improvement

Examples:
  omni lint                            # lint Taskfile.yml in current dir
  omni lint Taskfile.yml               # lint specific file
  omni lint ./tasks/                   # lint all Taskfiles in directory
  omni lint --strict Taskfile.yml      # enable strict mode
  omni lint -q Taskfile.yml            # only show errors

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --fix | bool | false | auto-fix issues where possible |
| -f, --format | string | text | output format (text, json) |
| -q, --quiet | bool | false | only show errors, not warnings |
| --strict | bool | false | enable strict mode (more warnings become errors) |

**Examples:**

```bash
omni lint                            # lint Taskfile.yml in current dir
omni lint Taskfile.yml               # lint specific file
omni lint ./tasks/                   # lint all Taskfiles in directory
omni lint --strict Taskfile.yml      # enable strict mode
omni lint -q Taskfile.yml            # only show errors
```

---

### ln

**Category:** File Operations

**Usage:** `omni ln [OPTION]... TARGET LINK_NAME [flags]`

**Description:** Make links between files

**Details:**

Create a link to TARGET with the name LINK_NAME.
Create hard links by default, symbolic links with --symbolic.

  -s, --symbolic     make symbolic links instead of hard links
  -f, --force        remove existing destination files
  -n, --no-dereference  treat LINK_NAME as a normal file if it is a symlink
  -v, --verbose      print name of each linked file
  -b, --backup       make a backup of each existing destination file
  -r, --relative     create symbolic links relative to link location

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

**Details:**

Count lines of code, comments, and blanks by programming language.

Similar to tokei, cloc, or sloccount. Automatically detects language by
file extension and counts code, comments, and blank lines.

Supported languages: Go, Rust, JavaScript, TypeScript, Python, Java, C, C++,
C#, Ruby, PHP, Swift, Kotlin, Scala, Shell, Lua, SQL, HTML, CSS, and more.

Default excludes: .git, node_modules, vendor, __pycache__, .idea, .vscode,
target, build, dist

Examples:
  omni loc                         # count in current directory
  omni loc ./src                   # count in specific directory
  omni loc --json .                # output as JSON
  omni loc --exclude test .        # exclude "test" directory
  omni loc --hidden .              # include hidden files

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -e, --exclude | stringSlice | [] | directories to exclude |
| --hidden | bool | false | include hidden files |
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni loc                         # count in current directory
omni loc ./src                   # count in specific directory
omni loc --json .                # output as JSON
omni loc --exclude test .        # exclude "test" directory
omni loc --hidden .              # include hidden files
```

---

### logger

**Category:** Tooling

**Usage:** `omni logger [flags]`

**Description:** Configure omni command logging

**Details:**

Configure omni command logging by outputting shell export statements.

Usage with eval to set environment variables:
  eval "$(omni logger --path /path/to/omni.log)"

To disable logging:
  eval "$(omni logger --disable)"

To view all log files:
  omni logger --viewer

Environment variables set:
  OMNI_LOG_ENABLED - Set to "true" to enable logging
  OMNI_LOG_PATH    - Path to the log file

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --disable | bool | false | Disable logging (unset environment variables) |
| -p, --path | string | - | Path to the log file |
| -s, --status | bool | false | Show current logging status |
| -v, --viewer | bool | false | View all log files sorted by time |

**Examples:**

```bash
omni logger --viewer
```

---

### ls

**Category:** Core

**Usage:** `omni ls [file...] [flags]`

**Description:** List directory contents

**Details:**

List information about the FILEs (the current directory by default).
Sort entries alphabetically if none of -tSU is specified.

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

**Details:**

Print or check MD5 (128-bit) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Note: MD5 is cryptographically broken and should not be used for security.
Use SHA256 for secure hashing.

Examples:
  omni md5sum file.txt           # compute hash
  omni md5sum -c checksums.txt   # verify checksums

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --binary | bool | false | read in binary mode |
| -c, --check | bool | false | read checksums from FILE and check them |
| --quiet | bool | false | don't print OK for verified files |
| --status | bool | false | don't output anything, use status code |
| -w, --warn | bool | false | warn about improperly formatted lines |

**Examples:**

```bash
omni md5sum file.txt           # compute hash
omni md5sum -c checksums.txt   # verify checksums
```

---

### mkdir

**Category:** File Operations

**Usage:** `omni mkdir [directory...] [flags]`

**Description:** Create directories

**Details:**

Create the DIRECTORY(ies), if they do not already exist.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -p, --parents | bool | false | no error if existing, make parent directories as needed |

---

### more

**Category:** Other

**Usage:** `omni more [OPTION]... [FILE] [flags]`

**Description:** View file contents page by page

**Details:**

Display file contents one screen at a time.

more is a simpler pager than less - it's designed to show content
and quit when reaching the end.

Navigation:
  Space, Enter    Scroll down one page
  q               Quit

Examples:
  omni more file.txt
  omni more -n file.txt     # with line numbers
  cat file.txt | omni more  # from stdin

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --line-numbers | bool | false | show line numbers |

**Examples:**

```bash
omni more file.txt
omni more -n file.txt     # with line numbers
```

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

**Details:**

Rename SOURCE to DEST, or move SOURCE(s) to DIRECTORY.

---

### nanoid

**Category:** Other

**Usage:** `omni nanoid [OPTION]... [flags]`

**Description:** Generate compact, URL-safe unique IDs

**Details:**

Generate NanoIDs - compact, URL-safe, unique string identifiers.

NanoIDs are:
- Shorter than UUID (21 chars vs 36)
- URL-safe (using A-Za-z0-9_-)
- Cryptographically secure
- Customizable length and alphabet

Default: 21 characters from URL-safe alphabet (64 chars)

  -n, --count=N     generate N NanoIDs (default 1)
  -l, --length=N    length of NanoID (default 21)
  -a, --alphabet=S  custom alphabet
  --json            output as JSON

Examples:
  omni nanoid                    # generate one NanoID (21 chars)
  omni nanoid -n 5               # generate 5 NanoIDs
  omni nanoid -l 10              # shorter 10-char NanoID
  omni nanoid -a "0123456789"    # numeric only
  omni nanoid --json             # JSON output

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --alphabet | string | - | custom alphabet |
| -n, --count | int | 1 | generate N NanoIDs |
| --json | bool | false | output as JSON |
| -l, --length | int | 21 | length of NanoID |

**Examples:**

```bash
omni nanoid                    # generate one NanoID (21 chars)
omni nanoid -n 5               # generate 5 NanoIDs
omni nanoid -l 10              # shorter 10-char NanoID
omni nanoid -a "0123456789"    # numeric only
omni nanoid --json             # JSON output
```

---

### nl

**Category:** Text Processing

**Usage:** `omni nl [OPTION]... [FILE]... [flags]`

**Description:** Number lines of files

**Details:**

Write each FILE to standard output, with line numbers added.

With no FILE, or when FILE is -, read standard input.

  -b, --body-numbering=STYLE      use STYLE for numbering body lines
  -n, --number-format=FORMAT      insert line numbers according to FORMAT
  -s, --number-separator=STRING   add STRING after line number
  -v, --starting-line-number=N    first line number on each logical page
  -i, --line-increment=N          line number increment at each line
  -w, --number-width=N            use N columns for line numbers

STYLE is one of:
  a      number all lines
  t      number only nonempty lines (default for body)
  n      number no lines

FORMAT is one of:
  ln     left justified, no leading zeros
  rn     right justified, no leading zeros (default)
  rz     right justified, leading zeros

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

**Details:**

Write lines consisting of the sequentially corresponding lines from
each FILE, separated by TABs, to standard output.

With no FILE, or when FILE is -, read standard input.

  -d, --delimiters=LIST   reuse characters from LIST instead of TABs
  -s, --serial            paste one file at a time instead of in parallel
  -z, --zero-terminated   line delimiter is NUL, not newline

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

**Details:**

Chain multiple omni commands together, passing output from one to the next.

This allows creating pipelines of omni commands without using shell pipes,
making scripts more portable and avoiding shell-specific behavior.

Commands can be separated by:
  - Curly braces with commas: {cmd1}, {cmd2}, {cmd3} (recommended)
  - The | character: cmd1 | cmd2 | cmd3
  - A custom separator with --sep
  - As separate quoted arguments

Variable Substitution:
  Use $OUT (or custom var name with --var) to substitute previous output:
  - $OUT or ${OUT}     - single value substitution (uses last line)
  - [$OUT...]          - iterate over each line of output

Examples:
  # Using braces (recommended - clearest syntax)
  omni pipe '{ls -la}', '{grep .go}', '{wc -l}'
  omni pipe '{cat file.txt}', '{sort}', '{uniq}'
  omni pipe '{cat data.json}', '{jq .users[]}'

  # Using | separator (quote the whole thing)
  omni pipe "cat file.txt | grep pattern | sort | uniq"

  # Using separate arguments with | between
  omni pipe cat file.txt \| grep error \| sort \| uniq -c

  # Using custom separator
  omni pipe --sep "->" "cat file.txt -> grep error -> sort"

  # Multiple quoted commands
  omni pipe "cat file.txt" "grep pattern" "sort" "uniq"

  # With stdin
  echo "hello world" | omni pipe '{grep hello}', '{wc -l}'

  # Verbose mode to see intermediate results
  omni pipe -v '{cat file.txt}', '{head -10}', '{sort}'

  # JSON output with pipeline metadata
  omni pipe --json '{cat file.txt}', '{wc -l}'

  # Variable substitution - create folder with UUID
  omni pipe '{uuid -v 7}', '{mkdir $OUT}'

  # Custom variable name
  omni pipe --var UUID '{uuid -v 7}', '{mkdir $UUID}'

  # Iteration - create folder for each UUID
  omni pipe '{uuid -v 7 -n 10}', '{mkdir [$OUT...]}'

Supported commands include all omni commands:
  cat, grep, head, tail, sort, uniq, wc, cut, tr, sed, awk,
  base64, hex, json, jq, yq, curl, and many more.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output result as JSON with metadata |
| -s, --sep | string | | | command separator |
| --var | string | OUT | variable name for output substitution (default: OUT) |
| -v, --verbose | bool | false | show intermediate results |

**Examples:**

```bash
omni pipe '{ls -la}', '{grep .go}', '{wc -l}'
omni pipe '{cat file.txt}', '{sort}', '{uniq}'
omni pipe '{cat data.json}', '{jq .users[]}'
omni pipe "cat file.txt | grep pattern | sort | uniq"
omni pipe cat file.txt \| grep error \| sort \| uniq -c
omni pipe --sep "->" "cat file.txt -> grep error -> sort"
omni pipe "cat file.txt" "grep pattern" "sort" "uniq"
omni pipe -v '{cat file.txt}', '{head -10}', '{sort}'
omni pipe --json '{cat file.txt}', '{wc -l}'
omni pipe '{uuid -v 7}', '{mkdir $OUT}'
omni pipe --var UUID '{uuid -v 7}', '{mkdir $UUID}'
omni pipe '{uuid -v 7 -n 10}', '{mkdir [$OUT...]}'
```

---

### pipeline

**Category:** Other

**Usage:** `omni pipeline`

**Description:** Internal streaming pipeline engine

**Details:**

Internal streaming pipeline engine for chaining commands.

---

### printf

**Category:** Other

**Usage:** `omni printf FORMAT [ARG...] [flags]`

**Description:** Format and print data

**Details:**

Format and print data using printf-style format specifiers.

Format specifiers:
  %s    String
  %d    Decimal integer
  %i    Integer (same as %d)
  %o    Octal
  %x    Lowercase hexadecimal
  %X    Uppercase hexadecimal
  %b    Binary
  %f    Floating point
  %e    Scientific notation
  %g    Compact floating point
  %c    Character
  %q    Quoted string
  %%    Literal percent sign

Escape sequences:
  \n    Newline
  \t    Tab
  \r    Carriage return
  \\    Backslash
  \xHH  Hex character
  \NNN  Octal character

Width and precision:
  %10s   Right-aligned, width 10
  %-10s  Left-aligned, width 10
  %.5s   Max 5 characters
  %10.5s Width 10, max 5 characters
  %08d   Zero-padded, width 8

Examples:
  omni printf "Hello, %s!" World
  omni printf "Number: %d, Hex: %x" 255 255
  omni printf "Pi: %.2f" 3.14159
  omni printf "Name: %-10s Age: %3d" Alice 25
  omni printf "Tab:\tNewline:\n"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --no-newline | bool | false | do not append a trailing newline |

**Examples:**

```bash
omni printf "Hello, %s!" World
omni printf "Number: %d, Hex: %x" 255 255
omni printf "Pi: %.2f" 3.14159
omni printf "Name: %-10s Age: %3d" Alice 25
omni printf "Tab:\tNewline:\n"
```

---

### ps

**Category:** System Info

**Usage:** `omni ps [OPTION]... [flags]`

**Description:** Report a snapshot of current processes

**Details:**

Display information about active processes.

  -a             show processes for all users
  -f             full-format listing
  -l             long format
  -u USER        show processes for specified user
  -p PID         show process with specified PID
  --no-headers   don't print header line
  --sort COL     sort by column (pid, cpu, mem, time)
  -j, --json     output as JSON
  --go           show only Go processes (detected via gops)

Go processes are automatically detected and marked. Use --go to filter
only Go processes, or -j/--json to see the is_go field in output.

Note: On Linux, reads /proc filesystem. On Windows, uses Win32 API.

Examples:
  omni ps                 # show current user's processes
  omni ps -a              # show all processes
  omni ps -f              # full format listing
  omni ps -p 1234         # show specific process
  omni ps --sort cpu      # sort by CPU usage
  omni ps -j              # output as JSON
  omni ps --go            # show only Go processes
  omni ps --go -j         # Go processes as JSON

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

**Examples:**

```bash
omni ps                 # show current user's processes
omni ps -a              # show all processes
omni ps -f              # full format listing
omni ps -p 1234         # show specific process
omni ps --sort cpu      # sort by CPU usage
omni ps -j              # output as JSON
omni ps --go            # show only Go processes
omni ps --go -j         # Go processes as JSON
```

---

### pwd

**Category:** Core

**Usage:** `omni pwd [flags]`

**Description:** Print working directory

**Details:**

Print the full filename of the current working directory.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### random

**Category:** Security

**Usage:** `omni random [OPTION]... [flags]`

**Description:** Generate random values

**Details:**

Generate random numbers, strings, or bytes using crypto/rand.

Types:
  int, integer    random integer (use --min, --max)
  float, decimal  random float between 0 and 1
  string, str     random alphanumeric string (default)
  alpha           random letters only
  alnum           random alphanumeric
  hex             random hexadecimal string
  password        random password (letters, digits, symbols)
  bytes           random bytes as hex
  custom          use custom charset (-c)

  -n, --count N     number of values to generate
  -l, --length N    length of strings (default 16)
  -t, --type TYPE   value type (default: string)
  --min N           minimum for integers
  --max N           maximum for integers
  -c, --charset STR custom character set
  -s, --separator   separator between values (default: newline)

Examples:
  omni random                         # random 16-char string
  omni random -t int --max 100        # random int 0-99
  omni random -t hex -l 32            # random 32-char hex
  omni random -t password -l 20       # random password
  omni random -n 5 -t int --max 10    # 5 random ints 0-9
  omni random -t custom -c "abc123"   # from custom charset

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

**Examples:**

```bash
omni random                         # random 16-char string
omni random -t int --max 100        # random int 0-99
omni random -t hex -l 32            # random 32-char hex
omni random -t password -l 20       # random password
omni random -n 5 -t int --max 10    # 5 random ints 0-9
omni random -t custom -c "abc123"   # from custom charset
```

---

### readlink

**Category:** Core

**Usage:** `omni readlink [OPTION]... FILE... [flags]`

**Description:** Print resolved symbolic links or canonical file names

**Details:**

Print value of a symbolic link or canonical file name.

  -f, --canonicalize            canonicalize by following every symlink
  -e, --canonicalize-existing   canonicalize, all components must exist
  -m, --canonicalize-missing    canonicalize without requirements on existence
  -n, --no-newline              do not output the trailing delimiter
  -q, --quiet                   suppress most error messages
  -s, --silent                  suppress most error messages
  -v, --verbose                 report error messages
  -z, --zero                    end each output line with NUL, not newline

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

**Details:**

Print the resolved absolute file name; all components must exist.

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

**Details:**

Reverse the characters in each line of FILE(s) or standard input.

Examples:
  echo "hello" | omni rev     # olleh
  omni rev file.txt           # reverse each line in file

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni rev file.txt           # reverse each line in file
```

---

### rg

**Category:** Other

**Usage:** `omni rg [OPTIONS] PATTERN [PATH...] [flags]`

**Description:** Recursively search for a pattern (ripgrep-style)

**Details:**

Recursively search current directory for a regex pattern.

rg is a line-oriented search tool that recursively searches the current
directory for a regex pattern. By default, rg respects gitignore rules
and automatically skips hidden files/directories and binary files.

This is inspired by ripgrep (https://github.com/BurntSushi/ripgrep).

Examples:
  # Search for pattern in current directory
  omni rg "pattern"

  # Search in specific directory
  omni rg "pattern" ./src

  # Case insensitive search
  omni rg -i "pattern"

  # Search only Go files
  omni rg -t go "func main"

  # Search with context (3 lines before and after)
  omni rg -C 3 "error"

  # Show only filenames with matches
  omni rg -l "TODO"

  # Count matches per file
  omni rg -c "pattern"

  # Include hidden files
  omni rg --hidden "pattern"

  # Don't respect gitignore
  omni rg --no-ignore "pattern"

  # Search for literal string (no regex)
  omni rg -F "func()"

  # JSON output
  omni rg --json "pattern"

  # Streaming JSON output (NDJSON)
  omni rg --json-stream "pattern"

  # Glob patterns
  omni rg -g "*.go" -g "!*_test.go" "pattern"

  # Control parallelism
  omni rg --threads 4 "pattern"

File Types:
  go, js, ts, py, rust, c, cpp, java, rb, php, sh, json, yaml, toml,
  xml, html, css, md, sql, proto, dockerfile, make, txt

Gitignore Support:
  rg respects multiple ignore sources (in order of precedence):
  - ~/.config/git/ignore (global gitignore)
  - .git/info/exclude (per-repo excludes)
  - .gitignore files (walked up from target directory)
  - .ignore files (ripgrep-specific, same hierarchy)

  Supports negation patterns (!pattern) to re-include files.
  Supports directory-only patterns (pattern/).

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

**Examples:**

```bash
omni rg "pattern"
omni rg "pattern" ./src
omni rg -i "pattern"
omni rg -t go "func main"
omni rg -C 3 "error"
omni rg -l "TODO"
omni rg -c "pattern"
omni rg --hidden "pattern"
omni rg --no-ignore "pattern"
omni rg -F "func()"
omni rg --json "pattern"
omni rg --json-stream "pattern"
omni rg -g "*.go" -g "!*_test.go" "pattern"
omni rg --threads 4 "pattern"
```

---

### rm

**Category:** File Operations

**Usage:** `omni rm [file...] [flags]`

**Description:** Remove files or directories

**Details:**

Remove the FILE(s).

Protected paths (system directories, SSH keys, credentials, etc.) cannot be
deleted without explicit override flags. Use --force for non-critical
protected paths, or --no-preserve-root for critical system paths.

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

**Details:**

Remove the DIRECTORY(ies), if they are empty.

Protected paths (system directories, credentials, etc.) cannot be
deleted without the --no-preserve-root flag.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --no-preserve-root | bool | false | do not treat protected paths specially (dangerous) |

---

### sed

**Category:** Text Processing

**Usage:** `omni sed [OPTION]... {script} [FILE]... [flags]`

**Description:** Stream editor for filtering and transforming text

**Details:**

Sed is a stream editor. A stream editor is used to perform basic text
transformations on an input stream (a file or input from a pipeline).

  -e script      add the script to the commands to be executed
  -i[SUFFIX]     edit files in place (makes backup if SUFFIX supplied)
  -n             suppress automatic printing of pattern space
  -E, -r         use extended regular expressions

Supported commands:
  s/regexp/replacement/flags  substitute
  d                           delete pattern space
  p                           print pattern space
  q                           quit

Examples:
  omni sed 's/old/new/' file.txt        # replace first occurrence
  omni sed 's/old/new/g' file.txt       # replace all occurrences
  omni sed -i.bak 's/foo/bar/g' file    # in-place edit with backup
  omni sed '/pattern/d' file.txt        # delete matching lines
  omni sed -n '/pattern/p' file.txt     # print only matching lines

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -e, --expression | stringSlice | [] | add the script to the commands to be executed |
| -i, --in-place | bool | false | edit files in place |
| --in-place-suffix | string | - | backup suffix for in-place edit |
| -n, --quiet | bool | false | suppress automatic printing of pattern space |
| -r, --r | bool | false | use extended regular expressions (alias) |
| -E, --regexp-extended | bool | false | use extended regular expressions |

**Examples:**

```bash
omni sed 's/old/new/' file.txt        # replace first occurrence
omni sed 's/old/new/g' file.txt       # replace all occurrences
omni sed -i.bak 's/foo/bar/g' file    # in-place edit with backup
omni sed '/pattern/d' file.txt        # delete matching lines
omni sed -n '/pattern/p' file.txt     # print only matching lines
```

---

### seq

**Category:** Utilities

**Usage:** `omni seq [OPTION]... LAST or seq [OPTION]... FIRST LAST or seq [OPTION]... FIRST INCREMENT LAST [flags]`

**Description:** Print a sequence of numbers

**Details:**

Print numbers from FIRST to LAST, in steps of INCREMENT.

  -s, --separator=STRING  use STRING to separate numbers (default: \n)
  -f, --format=FORMAT     use printf style floating-point FORMAT
  -w, --equal-width       equalize width by padding with leading zeros

Examples:
  omni seq 5               # print 1 2 3 4 5
  omni seq 2 5             # print 2 3 4 5
  omni seq 1 2 10          # print 1 3 5 7 9
  omni seq -w 1 10         # print 01 02 ... 10
  omni seq -s ', ' 1 5     # print 1, 2, 3, 4, 5
  omni seq 0.5 0.1 1.0     # print 0.5 0.6 0.7 0.8 0.9 1.0

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -w, --equal-width | bool | false | equalize width with leading zeros |
| -f, --format | string | - | use printf style FORMAT |
| --json | bool | false | output as JSON |
| -s, --separator | string | - | use STRING to separate numbers |

**Examples:**

```bash
omni seq 5               # print 1 2 3 4 5
omni seq 2 5             # print 2 3 4 5
omni seq 1 2 10          # print 1 3 5 7 9
omni seq -w 1 10         # print 01 02 ... 10
omni seq -s ', ' 1 5     # print 1, 2, 3, 4, 5
omni seq 0.5 0.1 1.0     # print 0.5 0.6 0.7 0.8 0.9 1.0
```

---

### sha256sum

**Category:** Hash & Encoding

**Usage:** `omni sha256sum [OPTION]... [FILE]... [flags]`

**Description:** Compute and check SHA256 message digest

**Details:**

Print or check SHA256 (256-bit) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Examples:
  omni sha256sum file.txt           # compute hash
  omni sha256sum -c checksums.txt   # verify checksums

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --binary | bool | false | read in binary mode |
| -c, --check | bool | false | read checksums from FILE and check them |
| --quiet | bool | false | don't print OK for verified files |
| --status | bool | false | don't output anything, use status code |
| -w, --warn | bool | false | warn about improperly formatted lines |

**Examples:**

```bash
omni sha256sum file.txt           # compute hash
omni sha256sum -c checksums.txt   # verify checksums
```

---

### sha512sum

**Category:** Hash & Encoding

**Usage:** `omni sha512sum [OPTION]... [FILE]... [flags]`

**Description:** Compute and check SHA512 message digest

**Details:**

Print or check SHA512 (512-bit) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Examples:
  omni sha512sum file.txt           # compute hash
  omni sha512sum -c checksums.txt   # verify checksums

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --binary | bool | false | read in binary mode |
| -c, --check | bool | false | read checksums from FILE and check them |
| --quiet | bool | false | don't print OK for verified files |
| --status | bool | false | don't output anything, use status code |
| -w, --warn | bool | false | warn about improperly formatted lines |

**Examples:**

```bash
omni sha512sum file.txt           # compute hash
omni sha512sum -c checksums.txt   # verify checksums
```

---

### shuf

**Category:** Text Processing

**Usage:** `omni shuf [OPTION]... [FILE] [flags]`

**Description:** Generate random permutations

**Details:**

Write a random permutation of the input lines to standard output.

  -e, --echo          treat each ARG as an input line
  -i, --input-range   treat each number LO through HI as an input line
  -n, --head-count    output at most COUNT lines
  -r, --repeat        output lines can be repeated (with -n)
  -z, --zero-terminated  line delimiter is NUL, not newline

Examples:
  omni shuf file.txt              # shuffle lines of file
  omni shuf -e a b c d            # shuffle arguments
  omni shuf -i 1-10               # shuffle numbers 1-10
  omni shuf -n 5 file.txt         # output 5 random lines
  omni shuf -rn 10 -e yes no      # 10 random picks with repetition

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -e, --echo | bool | false | treat each ARG as an input line |
| -n, --head-count | int | 0 | output at most COUNT lines |
| -i, --input-range | string | - | treat each number LO through HI as an input line |
| --json | bool | false | output as JSON |
| -r, --repeat | bool | false | output lines can be repeated |
| -z, --zero-terminated | bool | false | line delimiter is NUL |

**Examples:**

```bash
omni shuf file.txt              # shuffle lines of file
omni shuf -e a b c d            # shuffle arguments
omni shuf -i 1-10               # shuffle numbers 1-10
omni shuf -n 5 file.txt         # output 5 random lines
omni shuf -rn 10 -e yes no      # 10 random picks with repetition
```

---

### sleep

**Category:** Utilities

**Usage:** `omni sleep NUMBER[SUFFIX]...`

**Description:** Delay for a specified amount of time

**Details:**

Pause for NUMBER seconds. SUFFIX may be:
  s   seconds (default)
  m   minutes
  h   hours
  d   days

NUMBER may be a decimal.

Examples:
  omni sleep 5           # sleep 5 seconds
  omni sleep 0.5         # sleep 0.5 seconds
  omni sleep 1m          # sleep 1 minute
  omni sleep 1h 30m      # sleep 1.5 hours

**Examples:**

```bash
omni sleep 5           # sleep 5 seconds
omni sleep 0.5         # sleep 0.5 seconds
omni sleep 1m          # sleep 1 minute
omni sleep 1h 30m      # sleep 1.5 hours
```

---

### snowflake

**Category:** Other

**Usage:** `omni snowflake [OPTION]... [flags]`

**Description:** Generate Twitter Snowflake-style IDs

**Details:**

Generate Snowflake IDs - distributed, time-sortable unique identifiers.

Snowflake IDs are 64-bit integers with embedded timestamp:
- 1 bit: unused (sign)
- 41 bits: timestamp (milliseconds, ~69 years)
- 10 bits: worker ID (0-1023)
- 12 bits: sequence (0-4095 per ms per worker)

Features:
- Roughly time-ordered
- Distributed generation (with worker IDs)
- ~4 million IDs per second per worker

  -n, --count=N     generate N Snowflake IDs (default 1)
  -w, --worker=N    worker ID (0-1023, default 0)
  --json            output as JSON

Examples:
  omni snowflake                 # generate one Snowflake ID
  omni snowflake -n 5            # generate 5 IDs
  omni snowflake -w 42           # use worker ID 42
  omni snowflake --json          # JSON output

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | generate N Snowflake IDs |
| --json | bool | false | output as JSON |
| -w, --worker | int64 | 0 | worker ID (0-1023) |

**Examples:**

```bash
omni snowflake                 # generate one Snowflake ID
omni snowflake -n 5            # generate 5 IDs
omni snowflake -w 42           # use worker ID 42
omni snowflake --json          # JSON output
```

---

### sort

**Category:** Text Processing

**Usage:** `omni sort [option]... [file]... [flags]`

**Description:** Sort lines of text files

**Details:**

Write sorted concatenation of all FILE(s) to standard output.

With no FILE, or when FILE is -, read standard input.

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

**Details:**

Output pieces of FILE to PREFIXaa, PREFIXab, ...;
default PREFIX is 'x'.

  -l, --lines=NUMBER   put NUMBER lines per output file
  -b, --bytes=SIZE     put SIZE bytes per output file
  -a, --suffix-length  generate suffixes of length N (default 2)
  -d, --numeric-suffixes  use numeric suffixes instead of alphabetic
      --verbose        print a diagnostic just before each output file is opened

SIZE may have a suffix: K=1024, M=1024*1024, G=1024*1024*1024

Examples:
  omni split file.txt              # split into 1000-line files
  omni split -l 100 file.txt       # split into 100-line files
  omni split -b 1M file.bin        # split into 1MB files
  omni split -d file.txt chunk_    # use numeric suffixes

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -b, --bytes | string | - | put SIZE bytes per output file |
| -l, --lines | int | 0 | put NUMBER lines per output file |
| -d, --numeric-suffixes | bool | false | use numeric suffixes |
| -a, --suffix-length | int | 2 | generate suffixes of length N |
| --verbose | bool | false | print diagnostic for each output file |

**Examples:**

```bash
omni split file.txt              # split into 1000-line files
omni split -l 100 file.txt       # split into 100-line files
omni split -b 1M file.bin        # split into 1MB files
omni split -d file.txt chunk_    # use numeric suffixes
```

---

### sql

**Category:** Other

**Usage:** `omni sql [FILE] [flags]`

**Description:** SQL utilities (format, minify, validate)

**Details:**

SQL utilities for formatting, minifying, and validating SQL.

When called directly, formats SQL (same as 'sql fmt').

Subcommands:
  fmt         Format/beautify SQL
  minify      Compact SQL
  validate    Validate SQL syntax

Examples:
  omni sql file.sql
  omni sql fmt file.sql
  omni sql minify file.sql
  omni sql validate file.sql
  echo 'select * from users' | omni sql
  omni sql "SELECT * FROM users WHERE id=1"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| -u, --uppercase | bool | true | uppercase keywords |

**Examples:**

```bash
omni sql file.sql
omni sql fmt file.sql
omni sql minify file.sql
omni sql validate file.sql
omni sql "SELECT * FROM users WHERE id=1"
```

**Subcommands:** `fmt`, `minify`, `validate`

---

### sql fmt

**Category:** Other

**Usage:** `omni sql fmt [FILE] [flags]`

**Description:** Format/beautify SQL

**Details:**

Format SQL with proper indentation and keyword capitalization.

  -i, --indent=STR     indentation string (default "  ")
  -u, --uppercase      uppercase keywords (default: true)
  -d, --dialect=NAME   SQL dialect: mysql, postgres, sqlite (default: generic)

Examples:
  omni sql fmt file.sql
  omni sql fmt "select * from users where id = 1"
  cat file.sql | omni sql fmt
  omni sql fmt --indent "    " file.sql

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --dialect | string | generic | SQL dialect (mysql, postgres, sqlite, generic) |
| -i, --indent | string |    | indentation string |
| -u, --uppercase | bool | true | uppercase keywords |

**Examples:**

```bash
omni sql fmt file.sql
omni sql fmt "select * from users where id = 1"
omni sql fmt --indent "    " file.sql
```

---

### sql minify

**Category:** Other

**Usage:** `omni sql minify [FILE]`

**Description:** Minify SQL

**Details:**

Minify SQL by removing unnecessary whitespace.

Examples:
  omni sql minify file.sql
  cat file.sql | omni sql minify

**Examples:**

```bash
omni sql minify file.sql
```

---

### sql validate

**Category:** Other

**Usage:** `omni sql validate [FILE] [flags]`

**Description:** Validate SQL syntax

**Details:**

Validate SQL syntax without outputting the query.

Exit codes:
  0  Valid SQL
  1  Invalid SQL or error

  --json    output result as JSON

Examples:
  omni sql validate file.sql
  omni sql validate "SELECT * FROM users"
  omni sql validate --json file.sql

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --dialect | string | generic | SQL dialect |
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni sql validate file.sql
omni sql validate "SELECT * FROM users"
omni sql validate --json file.sql
```

---

### sqlite

**Category:** Database

**Usage:** `omni sqlite`

**Description:** SQLite database management

**Details:**

sqlite provides commands for working with SQLite databases.

This is a CLI wrapper around modernc.org/sqlite (pure Go, no CGO)
for database inspection, querying, and maintenance operations.

Subcommands:
  stats      Show database statistics
  tables     List all tables
  schema     Show table schema
  columns    Show table columns
  indexes    List all indexes
  query      Execute SQL query
  vacuum     Optimize database
  check      Verify database integrity
  dump       Export database as SQL
  import     Import SQL file

Examples:
  omni sqlite stats mydb.sqlite
  omni sqlite tables mydb.sqlite
  omni sqlite schema mydb.sqlite users
  omni sqlite query mydb.sqlite "SELECT * FROM users"
  omni sqlite dump mydb.sqlite > backup.sql
  omni sqlite import mydb.sqlite backup.sql

**Examples:**

```bash
omni sqlite stats mydb.sqlite
omni sqlite tables mydb.sqlite
omni sqlite schema mydb.sqlite users
omni sqlite query mydb.sqlite "SELECT * FROM users"
omni sqlite dump mydb.sqlite > backup.sql
omni sqlite import mydb.sqlite backup.sql
```

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

**Details:**

Execute SQL query against a SQLite database.

Query logging can be enabled with the omni logger command:
  eval "$(omni logger --path /path/to/logs)"

With logging enabled, queries and results are recorded to log files.
Use --log-data to include result data in logs (use with caution for large results).

Examples:
  omni sqlite query mydb.sqlite "SELECT * FROM users"
  omni sqlite query mydb.sqlite "SELECT * FROM users" --header
  omni sqlite query mydb.sqlite "SELECT * FROM users" --json
  omni sqlite query mydb.sqlite "SELECT * FROM users" --log-data

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -H, --header | bool | false | show column headers |
| --log-data | bool | false | include result data in logs (use with caution for large results) |
| -s, --separator | string | | | column separator |

**Examples:**

```bash
omni sqlite query mydb.sqlite "SELECT * FROM users"
omni sqlite query mydb.sqlite "SELECT * FROM users" --header
omni sqlite query mydb.sqlite "SELECT * FROM users" --json
omni sqlite query mydb.sqlite "SELECT * FROM users" --log-data
```

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

**Details:**

Display file or file system status.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | Output in JSON format |

---

### strings

**Category:** Text Processing

**Usage:** `omni strings [OPTION]... [FILE]... [flags]`

**Description:** Print the printable strings in files

**Details:**

Print the sequences of printable characters in files.

  -n, --bytes=MIN   print sequences of at least MIN characters (default 4)
  -t, --radix=TYPE  print offset in TYPE: d=decimal, o=octal, x=hex

Examples:
  omni strings binary.exe         # extract strings from binary
  omni strings -n 8 file.bin      # strings of at least 8 chars
  omni strings -t x program       # show hex offsets

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --bytes | int | 4 | minimum string length |
| --json | bool | false | output as JSON |
| -t, --radix | string | - | print offset (d/o/x) |

**Examples:**

```bash
omni strings binary.exe         # extract strings from binary
omni strings -n 8 file.bin      # strings of at least 8 chars
omni strings -t x program       # show hex offsets
```

---

### tac

**Category:** Text Processing

**Usage:** `omni tac [OPTION]... [FILE]... [flags]`

**Description:** Concatenate and print files in reverse

**Details:**

Write each FILE to standard output, last line first.

With no FILE, or when FILE is -, read standard input.

  -b, --before             attach the separator before instead of after
  -r, --regex              interpret the separator as a regular expression
  -s, --separator=STRING   use STRING as the separator instead of newline

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

**Details:**

Fix and standardize struct tags in Go files.

Supports multiple casing conventions: camelCase, PascalCase, kebab-case, snake_case.

Case Types:
  camel   - camelCase (default)
  pascal  - PascalCase
  snake   - snake_case
  kebab   - kebab-case

Flags:
  -c, --case        Target case type (default: camel)
  -t, --tags        Tag types to fix (default: json)
  -r, --recursive   Process directories recursively
  -d, --dry-run     Preview changes without writing
  -a, --analyze     Analyze mode - generate report only
  -v, --verbose     Verbose output
  --json            Output as JSON

Examples:
  omni tagfixer                           # Fix json tags in current dir (camelCase)
  omni tagfixer ./pkg                     # Fix in specific directory
  omni tagfixer -c snake                  # Use snake_case
  omni tagfixer -c kebab                  # Use kebab-case
  omni tagfixer -t json,yaml,xml          # Fix multiple tag types
  omni tagfixer -d                        # Dry-run (preview)
  omni tagfixer -a                        # Analyze only
  omni tagfixer -a -v                     # Detailed analysis
  omni tagfixer --json                    # JSON output

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

**Examples:**

```bash
omni tagfixer                           # Fix json tags in current dir (camelCase)
omni tagfixer ./pkg                     # Fix in specific directory
omni tagfixer -c snake                  # Use snake_case
omni tagfixer -c kebab                  # Use kebab-case
omni tagfixer -t json,yaml,xml          # Fix multiple tag types
omni tagfixer -d                        # Dry-run (preview)
omni tagfixer -a                        # Analyze only
omni tagfixer -a -v                     # Detailed analysis
omni tagfixer --json                    # JSON output
```

**Subcommands:** `analyze`

---

### tagfixer analyze

**Category:** Other

**Usage:** `omni tagfixer analyze [PATH] [flags]`

**Description:** Analyze struct tag usage patterns

**Details:**

Analyze Go files to understand current struct tag patterns.

Generates a report showing:
  - Total files, structs, and fields analyzed
  - Tag type statistics (json, yaml, xml, etc.)
  - Case type distribution
  - Consistency score (0-100%)
  - Recommended case type based on existing patterns

Examples:
  omni tagfixer analyze                   # Analyze current directory
  omni tagfixer analyze ./pkg             # Analyze specific directory
  omni tagfixer analyze -r                # Recursive analysis
  omni tagfixer analyze -v                # Verbose (show all files)
  omni tagfixer analyze --json            # JSON output

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |
| -r, --recursive | bool | true | process directories recursively |
| -t, --tags | string | json,yaml,xml | comma-separated tag types to analyze |
| -v, --verbose | bool | false | verbose output (show all files) |

**Examples:**

```bash
omni tagfixer analyze                   # Analyze current directory
omni tagfixer analyze ./pkg             # Analyze specific directory
omni tagfixer analyze -r                # Recursive analysis
omni tagfixer analyze -v                # Verbose (show all files)
omni tagfixer analyze --json            # JSON output
```

---

### tail

**Category:** Text Processing

**Usage:** `omni tail [option]... [file]... [flags]`

**Description:** Output the last part of files

**Details:**

Print the last 10 lines of each FILE to standard output.
With more than one FILE, precede each with a header giving the file name.
With no FILE, or when FILE is -, read standard input.

Numeric shortcuts are supported: -80 is equivalent to -n 80.

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

**Details:**

Manipulate tape archive files.

  -c, --create           create a new archive
  -x, --extract          extract files from an archive
  -t, --list             list the contents of an archive
  -f, --file=ARCHIVE     use archive file ARCHIVE
  -v, --verbose          verbosely list files processed
  -z, --gzip             filter through gzip
  -C, --directory=DIR    change to directory DIR
      --strip-components=N  strip N leading path components

Examples:
  omni tar -cvf archive.tar dir/        # create tar archive
  omni tar -czvf archive.tar.gz dir/    # create gzipped tar
  omni tar -xvf archive.tar             # extract tar archive
  omni tar -xzvf archive.tar.gz         # extract gzipped tar
  omni tar -tvf archive.tar             # list contents
  omni tar -xvf archive.tar -C /dest    # extract to directory

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

**Examples:**

```bash
omni tar -cvf archive.tar dir/        # create tar archive
omni tar -czvf archive.tar.gz dir/    # create gzipped tar
omni tar -xvf archive.tar             # extract tar archive
omni tar -xzvf archive.tar.gz         # extract gzipped tar
omni tar -tvf archive.tar             # list contents
omni tar -xvf archive.tar -C /dest    # extract to directory
```

---

### task

**Category:** Other

**Usage:** `omni task [TASK...] [flags]`

**Description:** Run tasks defined in Taskfile.yml

**Details:**

A task runner that executes tasks defined in Taskfile.yml.

By default, only omni internal commands are supported. Use --allow-external
to enable execution of external shell commands (golangci-lint, go, npm, etc).

Examples:
  # List available tasks
  omni task --list

  # Run the default task
  omni task

  # Run a specific task
  omni task build

  # Run multiple tasks
  omni task build test

  # Run with external commands enabled
  omni task --allow-external lint

  # Dry run (show commands without executing)
  omni task --dry-run build

  # Force run even if up-to-date
  omni task --force build

  # Show task summary
  omni task --summary build

Taskfile Format:
  version: '3'

  vars:
    BUILD_DIR: ./build

  tasks:
    default:
      deps: [build]

    build:
      desc: Build the project
      cmds:
        - omni mkdir -p {{.BUILD_DIR}}
        - omni cp -r src/* {{.BUILD_DIR}}/

    lint:
      desc: Run linter (requires --allow-external)
      cmds:
        - golangci-lint run --fix ./...

    clean:
      desc: Clean build artifacts
      cmds:
        - omni rm -rf {{.BUILD_DIR}}

Supported Features:
  - Task dependencies (deps)
  - Variable expansion ({{.VAR}})
  - Task includes (includes)
  - Status checks for up-to-date detection
  - Deferred commands
  - Task aliases
  - External commands (with --allow-external)

Limitations:
  - Dynamic variables (sh:) are not supported

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

**Examples:**

```bash
omni task --list
omni task
omni task build
omni task build test
omni task --allow-external lint
omni task --dry-run build
omni task --force build
omni task --summary build
```

---

### terraform

**Category:** Other

**Usage:** `omni terraform`

**Description:** Terraform CLI

**Details:**

Terraform CLI wrapper integrated into omni.

Provides access to all Terraform commands. You can use 'omni terraform'
or the shorter alias 'omni tf'.

Examples:
  omni terraform init
  omni tf plan
  omni tf apply -auto-approve
  omni tf destroy

**Examples:**

```bash
omni terraform init
omni tf plan
omni tf apply -auto-approve
omni tf destroy
```

**Subcommands:** `apply`, `console`, `destroy`, `fmt`, `get`, `graph`, `import`, `init`, `output`, `plan`, `providers`, `refresh`, `show`, `state`, `taint`, `test`, `untaint`, `validate`, `version`, `workspace`

---

### terraform apply

**Category:** Other

**Usage:** `omni terraform apply [plan-file] [flags]`

**Description:** Apply changes to infrastructure

**Details:**

Apply the changes required to reach the desired state.

Examples:
  omni tf apply
  omni tf apply plan.tfplan
  omni tf apply -auto-approve

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --auto-approve | bool | false | Skip interactive approval |
| --var | stringToString | [] | Set variables |
| --var-file | stringSlice | [] | Variable files |

**Examples:**

```bash
omni tf apply
omni tf apply plan.tfplan
omni tf apply -auto-approve
```

---

### terraform console

**Category:** Other

**Usage:** `omni terraform console`

**Description:** Interactive console

**Details:**

Launch an interactive console for evaluating expressions.

Examples:
  omni tf console

**Examples:**

```bash
omni tf console
```

---

### terraform destroy

**Category:** Other

**Usage:** `omni terraform destroy [flags]`

**Description:** Destroy managed infrastructure

**Details:**

Destroy all remote objects managed by Terraform.

Examples:
  omni tf destroy
  omni tf destroy -auto-approve

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --auto-approve | bool | false | Skip interactive approval |
| --var | stringToString | [] | Set variables |
| --var-file | stringSlice | [] | Variable files |

**Examples:**

```bash
omni tf destroy
omni tf destroy -auto-approve
```

---

### terraform fmt

**Category:** Other

**Usage:** `omni terraform fmt [flags]`

**Description:** Format configuration files

**Details:**

Reformat configuration files to a canonical format.

Examples:
  omni tf fmt
  omni tf fmt -check
  omni tf fmt -recursive

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --check | bool | false | Check if formatted |
| --diff | bool | false | Display diffs |
| --recursive | bool | false | Process subdirectories |

**Examples:**

```bash
omni tf fmt
omni tf fmt -check
omni tf fmt -recursive
```

---

### terraform get

**Category:** Other

**Usage:** `omni terraform get [flags]`

**Description:** Download modules

**Details:**

Download and install modules for the configuration.

Examples:
  omni tf get
  omni tf get -update

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --update | bool | false | Update modules |

**Examples:**

```bash
omni tf get
omni tf get -update
```

---

### terraform graph

**Category:** Other

**Usage:** `omni terraform graph [flags]`

**Description:** Generate dependency graph

**Details:**

Generate a visual representation of dependencies.

Examples:
  omni tf graph
  omni tf graph | dot -Tpng > graph.png

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --draw-cycles | bool | false | Draw cycles |
| --plan | string | - | Plan file |

**Examples:**

```bash
omni tf graph
omni tf graph | dot -Tpng > graph.png
```

---

### terraform import

**Category:** Other

**Usage:** `omni terraform import <address> <id> [flags]`

**Description:** Import existing infrastructure

**Details:**

Import existing infrastructure into Terraform state.

Examples:
  omni tf import aws_instance.example i-1234567890

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --var | stringToString | [] | Set variables |
| --var-file | stringSlice | [] | Variable files |

**Examples:**

```bash
omni tf import aws_instance.example i-1234567890
```

---

### terraform init

**Category:** Other

**Usage:** `omni terraform init [flags]`

**Description:** Initialize Terraform working directory

**Details:**

Initialize a new or existing Terraform working directory.

Examples:
  omni tf init
  omni tf init -upgrade
  omni tf init -reconfigure

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --reconfigure | bool | false | Reconfigure backend |
| --upgrade | bool | false | Upgrade modules and plugins |

**Examples:**

```bash
omni tf init
omni tf init -upgrade
omni tf init -reconfigure
```

---

### terraform output

**Category:** Other

**Usage:** `omni terraform output [name] [flags]`

**Description:** Show output values

**Details:**

Read an output variable from the state file.

Examples:
  omni tf output
  omni tf output instance_ip
  omni tf output -json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | JSON output |

**Examples:**

```bash
omni tf output
omni tf output instance_ip
omni tf output -json
```

---

### terraform plan

**Category:** Other

**Usage:** `omni terraform plan [flags]`

**Description:** Create execution plan

**Details:**

Create an execution plan showing what Terraform will do.

Examples:
  omni tf plan
  omni tf plan -out=plan.tfplan
  omni tf plan -var "region=us-east-1"

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --destroy | bool | false | Create destroy plan |
| -o, --out | string | - | Write plan to file |
| --var | stringToString | [] | Set variables |
| --var-file | stringSlice | [] | Variable files |

**Examples:**

```bash
omni tf plan
omni tf plan -out=plan.tfplan
omni tf plan -var "region=us-east-1"
```

---

### terraform providers

**Category:** Other

**Usage:** `omni terraform providers`

**Description:** Show provider information

**Details:**

Show the providers required for this configuration.

Examples:
  omni tf providers

**Examples:**

```bash
omni tf providers
```

---

### terraform refresh

**Category:** Other

**Usage:** `omni terraform refresh [flags]`

**Description:** Refresh state

**Details:**

Update the state file with real-world infrastructure.

Examples:
  omni tf refresh

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --var | stringToString | [] | Set variables |
| --var-file | stringSlice | [] | Variable files |

**Examples:**

```bash
omni tf refresh
```

---

### terraform show

**Category:** Other

**Usage:** `omni terraform show [plan-file] [flags]`

**Description:** Show plan or state

**Details:**

Show a human-readable output from a plan file or state.

Examples:
  omni tf show
  omni tf show plan.tfplan
  omni tf show -json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | JSON output |

**Examples:**

```bash
omni tf show
omni tf show plan.tfplan
omni tf show -json
```

---

### terraform state

**Category:** Other

**Usage:** `omni terraform state`

**Description:** State management commands

**Details:**

Commands for managing Terraform state.

**Subcommands:** `list`, `mv`, `rm`, `show`

---

### terraform state list

**Category:** Other

**Usage:** `omni terraform state list [addresses...]`

**Description:** List resources in state

**Details:**

List resources in the Terraform state.

Examples:
  omni tf state list
  omni tf state list aws_instance.example

**Examples:**

```bash
omni tf state list
omni tf state list aws_instance.example
```

---

### terraform state mv

**Category:** File Operations

**Usage:** `omni terraform state mv <source> <destination> [flags]`

**Description:** Move resource in state

**Details:**

Move a resource from one address to another.

Examples:
  omni tf state mv aws_instance.old aws_instance.new

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | Dry run |

**Examples:**

```bash
omni tf state mv aws_instance.old aws_instance.new
```

---

### terraform state rm

**Category:** File Operations

**Usage:** `omni terraform state rm <addresses...> [flags]`

**Description:** Remove resources from state

**Details:**

Remove resources from the Terraform state.

Examples:
  omni tf state rm aws_instance.example

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --dry-run | bool | false | Dry run |

**Examples:**

```bash
omni tf state rm aws_instance.example
```

---

### terraform state show

**Category:** Other

**Usage:** `omni terraform state show <address>`

**Description:** Show a resource in state

**Details:**

Show the attributes of a single resource in the state.

Examples:
  omni tf state show aws_instance.example

**Examples:**

```bash
omni tf state show aws_instance.example
```

---

### terraform taint

**Category:** Other

**Usage:** `omni terraform taint <address>`

**Description:** Mark resource for recreation

**Details:**

Mark a resource instance as not fully functional.

Examples:
  omni tf taint aws_instance.example

**Examples:**

```bash
omni tf taint aws_instance.example
```

---

### terraform test

**Category:** Utilities

**Usage:** `omni terraform test`

**Description:** Run tests

**Details:**

Execute Terraform test files.

Examples:
  omni tf test

**Examples:**

```bash
omni tf test
```

---

### terraform untaint

**Category:** Other

**Usage:** `omni terraform untaint <address>`

**Description:** Remove taint from resource

**Details:**

Remove the taint state from a resource instance.

Examples:
  omni tf untaint aws_instance.example

**Examples:**

```bash
omni tf untaint aws_instance.example
```

---

### terraform validate

**Category:** Other

**Usage:** `omni terraform validate`

**Description:** Validate configuration

**Details:**

Validate the configuration files.

Examples:
  omni tf validate

**Examples:**

```bash
omni tf validate
```

---

### terraform version

**Category:** Tooling

**Usage:** `omni terraform version`

**Description:** Show Terraform version

**Details:**

Show the current Terraform version.

Examples:
  omni tf version

**Examples:**

```bash
omni tf version
```

---

### terraform workspace

**Category:** Other

**Usage:** `omni terraform workspace`

**Description:** Workspace management commands

**Details:**

Commands for managing Terraform workspaces.

**Subcommands:** `delete`, `list`, `new`, `select`, `show`

---

### terraform workspace delete

**Category:** Other

**Usage:** `omni terraform workspace delete <name> [flags]`

**Description:** Delete workspace

**Details:**

Delete a workspace.

Examples:
  omni tf workspace delete old-workspace

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --force | bool | false | Force delete |

**Examples:**

```bash
omni tf workspace delete old-workspace
```

---

### terraform workspace list

**Category:** Other

**Usage:** `omni terraform workspace list`

**Description:** List workspaces

**Details:**

List all available workspaces.

Examples:
  omni tf workspace list

**Examples:**

```bash
omni tf workspace list
```

---

### terraform workspace new

**Category:** Other

**Usage:** `omni terraform workspace new <name>`

**Description:** Create new workspace

**Details:**

Create a new workspace.

Examples:
  omni tf workspace new development

**Examples:**

```bash
omni tf workspace new development
```

---

### terraform workspace select

**Category:** Other

**Usage:** `omni terraform workspace select <name>`

**Description:** Select workspace

**Details:**

Select a workspace to use.

Examples:
  omni tf workspace select production

**Examples:**

```bash
omni tf workspace select production
```

---

### terraform workspace show

**Category:** Other

**Usage:** `omni terraform workspace show`

**Description:** Show current workspace

**Details:**

Show the name of the current workspace.

Examples:
  omni tf workspace show

**Examples:**

```bash
omni tf workspace show
```

---

### testcheck

**Category:** Other

**Usage:** `omni testcheck [directory] [flags]`

**Description:** Check test coverage for Go packages

**Details:**

Scan a directory for Go packages and report which have tests.

By default, only shows packages WITHOUT tests. Use --all to show all packages.

Examples:
  omni testcheck ./pkg/cli/           # Check packages in pkg/cli
  omni testcheck .                    # Check current directory
  omni testcheck --all ./pkg/         # Show all packages
  omni testcheck --summary ./pkg/     # Show only summary
  omni testcheck --json ./pkg/        # Output as JSON

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | false | show all packages (default shows only missing) |
| -j, --json | bool | false | output as JSON |
| -s, --summary | bool | false | show only summary |
| -v, --verbose | bool | false | show test file names |

**Examples:**

```bash
omni testcheck ./pkg/cli/           # Check packages in pkg/cli
omni testcheck .                    # Check current directory
omni testcheck --all ./pkg/         # Show all packages
omni testcheck --summary ./pkg/     # Show only summary
omni testcheck --json ./pkg/        # Output as JSON
```

---

### time

**Category:** Utilities

**Usage:** `omni time`

**Description:** Time a simple command or give resource usage

**Details:**

The time utility executes and times the specified utility. After the
utility finishes, time writes to the standard error stream, the total
time elapsed.

Note: Since omni doesn't execute external commands, this command
provides timing utilities and can measure internal operations.

Examples:
  omni time sleep 2    # Time a sleep operation
  omni time            # Just show current time info

**Examples:**

```bash
omni time sleep 2    # Time a sleep operation
omni time            # Just show current time info
```

---

### toml

**Category:** Other

**Usage:** `omni toml`

**Description:** TOML utilities

**Details:**

TOML utilities for validation and formatting.

Subcommands:
  validate    Validate TOML syntax
  fmt         Format/beautify TOML

Examples:
  omni toml validate config.toml
  omni toml fmt config.toml

**Examples:**

```bash
omni toml validate config.toml
omni toml fmt config.toml
```

**Subcommands:** `fmt`, `validate`

---

### toml fmt

**Category:** Other

**Usage:** `omni toml fmt [FILE] [flags]`

**Description:** Format TOML

**Details:**

Format and beautify TOML.

Parses TOML and outputs it with consistent formatting.

Examples:
  omni toml fmt config.toml
  cat config.toml | omni toml fmt
  omni toml fmt --indent 4 config.toml

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | int | 2 | indentation width |

**Examples:**

```bash
omni toml fmt config.toml
omni toml fmt --indent 4 config.toml
```

---

### toml validate

**Category:** Other

**Usage:** `omni toml validate [FILE...] [flags]`

**Description:** Validate TOML syntax

**Details:**

Validate TOML syntax for one or more files.

Checks that the input is valid TOML.

Examples:
  omni toml validate config.toml
  omni toml validate *.toml
  cat config.toml | omni toml validate
  omni toml validate --json config.toml

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni toml validate config.toml
omni toml validate *.toml
omni toml validate --json config.toml
```

---

### top

**Category:** Other

**Usage:** `omni top [flags]`

**Description:** Display system processes sorted by resource usage

**Details:**

Display system processes sorted by CPU or memory usage.

  -n NUM         number of processes to show (default 10)
  --sort COL     sort by column: cpu (default), mem, pid
  -j, --json     output as JSON
  --go           show only Go processes

Note: This is a snapshot view. For real-time monitoring, use system top.

Examples:
  omni top                 # show top 10 by CPU
  omni top -n 20           # show top 20 processes
  omni top --sort mem      # sort by memory
  omni top --go            # show top Go processes
  omni top -j              # output as JSON

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --go | bool | false | show only Go processes |
| -j, --json | bool | false | output as JSON |
| -n, --num | int | 10 | number of processes to show |
| --sort | string | cpu | sort by column: cpu, mem, pid |

**Examples:**

```bash
omni top                 # show top 10 by CPU
omni top -n 20           # show top 20 processes
omni top --sort mem      # sort by memory
omni top --go            # show top Go processes
omni top -j              # output as JSON
```

---

### touch

**Category:** File Operations

**Usage:** `omni touch [file...]`

**Description:** Update the access and modification times of each FILE to the current time

**Details:**

Update the access and modification times of each FILE to the current time. A FILE argument that does not exist is created empty.

---

### tr

**Category:** Text Processing

**Usage:** `omni tr [OPTION]... SET1 [SET2] [flags]`

**Description:** Translate or delete characters

**Details:**

Translate, squeeze, and/or delete characters from standard input,
writing to standard output.

  -c, --complement    use the complement of SET1
  -d, --delete        delete characters in SET1, do not translate
  -s, --squeeze-repeats  replace each sequence of a repeated character
                         that is listed in the last SET, with a single
                         occurrence of that character
  -t, --truncate-set1 first truncate SET1 to length of SET2

SETs are specified as strings of characters.  Most represent themselves.
Interpreted sequences are:

  \n     new line
  \r     return
  \t     horizontal tab
  \\     backslash

  CHAR1-CHAR2  all characters from CHAR1 to CHAR2 in ascending order

  [:alnum:]    all letters and digits
  [:alpha:]    all letters
  [:digit:]    all digits
  [:lower:]    all lower case letters
  [:upper:]    all upper case letters
  [:space:]    all horizontal or vertical whitespace
  [:blank:]    all horizontal whitespace
  [:punct:]    all punctuation characters
  [:graph:]    all printable characters, not including space
  [:print:]    all printable characters, including space
  [:xdigit:]   all hexadecimal digits

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

**Details:**

Display a tree visualization of directory contents.

Examples:
  omni tree                          # current directory
  omni tree /path/to/dir             # specific directory
  omni tree -a                       # show hidden files
  omni tree -d 3                     # limit depth to 3
  omni tree -i "node_modules,.git"   # ignore patterns
  omni tree --dirs-only              # show only directories
  omni tree -s                       # show statistics
  omni tree --json                   # output as JSON

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

**Examples:**

```bash
omni tree                          # current directory
omni tree /path/to/dir             # specific directory
omni tree -a                       # show hidden files
omni tree -d 3                     # limit depth to 3
omni tree -i "node_modules,.git"   # ignore patterns
omni tree --dirs-only              # show only directories
omni tree -s                       # show statistics
omni tree --json                   # output as JSON
```

---

### ulid

**Category:** Other

**Usage:** `omni ulid [OPTION]... [flags]`

**Description:** Generate Universally Unique Lexicographically Sortable Identifiers

**Details:**

Generate ULIDs (Universally Unique Lexicographically Sortable Identifiers).

ULIDs are 26-character, Crockford's base32-encoded identifiers that are:
- 128-bit compatible with UUID
- Lexicographically sortable
- Case insensitive
- URL-safe (no special characters)

Structure: 48-bit timestamp (ms) + 80-bit randomness

  -n, --count=N   generate N ULIDs (default 1)
  -l, --lower     output in lowercase
  --json          output as JSON

Examples:
  omni ulid                   # generate one ULID
  omni ulid -n 5              # generate 5 ULIDs
  omni ulid -l                # lowercase output
  omni ulid --json            # JSON output

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | generate N ULIDs |
| --json | bool | false | output as JSON |
| -l, --lower | bool | false | output in lowercase |

**Examples:**

```bash
omni ulid                   # generate one ULID
omni ulid -n 5              # generate 5 ULIDs
omni ulid -l                # lowercase output
omni ulid --json            # JSON output
```

---

### uname

**Category:** System Info

**Usage:** `omni uname [flags]`

**Description:** Print system information

**Details:**

Print certain system information. With no OPTION, same as -s.

  -a, --all                print all information
  -s, --kernel-name        print the kernel name
  -n, --nodename           print the network node hostname
  -r, --kernel-release     print the kernel release
  -v, --kernel-version     print the kernel version
  -m, --machine            print the machine hardware name
  -p, --processor          print the processor type
  -i, --hardware-platform  print the hardware platform
  -o, --operating-system   print the operating system

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

**Details:**

Filter adjacent matching lines from INPUT (or standard input),
writing to OUTPUT (or standard output).

With no options, matching lines are merged to the first occurrence.

Note: 'uniq' does not detect repeated lines unless they are adjacent.
You may want to sort the input first, or use 'sort -u' without 'uniq'.

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

**Details:**

Decompress FILEs in xz format.

Equivalent to xz -d.

Note: Full decompression requires external library.

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

**Details:**

Extract files from a zip archive.

  -l, --list        list contents without extracting
  -v, --verbose     verbose output
  -d, --directory   extract files into directory
      --strip-components=N  strip N leading path components

Examples:
  omni unzip archive.zip              # extract to current directory
  omni unzip -d /dest archive.zip     # extract to specific directory
  omni unzip -l archive.zip           # list contents
  omni unzip -v archive.zip           # verbose extraction

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --directory | string | - | extract files into directory |
| --json | bool | false | output as JSON (for list mode) |
| -l, --list | bool | false | list contents without extracting |
| --strip-components | int | 0 | strip N leading path components |
| -v, --verbose | bool | false | verbose output |

**Examples:**

```bash
omni unzip archive.zip              # extract to current directory
omni unzip -d /dest archive.zip     # extract to specific directory
omni unzip -l archive.zip           # list contents
omni unzip -v archive.zip           # verbose extraction
```

---

### uptime

**Category:** System Info

**Usage:** `omni uptime [OPTION]... [flags]`

**Description:** Tell how long the system has been running

**Details:**

Print the current time, how long the system has been running,
how many users are currently logged on, and the system load averages
for the past 1, 5, and 15 minutes.

  -p, --pretty   show uptime in pretty format
  -s, --since    system up since, in yyyy-mm-dd HH:MM:SS format

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

**Details:**

URL encoding and decoding utilities.

Subcommands:
  encode    URL encode text
  decode    URL decode text

Examples:
  omni url encode "hello world"
  omni url decode "hello%20world"
  echo "test string" | omni url encode
  omni url encode --component "a=b&c=d"

**Examples:**

```bash
omni url encode "hello world"
omni url decode "hello%20world"
omni url encode --component "a=b&c=d"
```

**Subcommands:** `decode`, `encode`

---

### url decode

**Category:** Other

**Usage:** `omni url decode [TEXT] [flags]`

**Description:** URL decode text

**Details:**

URL decode percent-encoded text.

By default uses path decoding. Use --component for query string decoding.

Examples:
  omni url decode "hello%20world"         # Output: hello world
  omni url decode --component "a%3Db"     # Output: a=b
  echo "test%20string" | omni url decode  # Read from stdin

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --component | bool | false | use query component decoding |
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni url decode "hello%20world"         # Output: hello world
omni url decode --component "a%3Db"     # Output: a=b
```

---

### url encode

**Category:** Other

**Usage:** `omni url encode [TEXT] [flags]`

**Description:** URL encode text

**Details:**

URL encode text for safe use in URLs.

By default uses path encoding. Use --component for query string encoding
which is more aggressive (encodes more characters).

Examples:
  omni url encode "hello world"           # Output: hello%20world
  omni url encode --component "a=b&c=d"   # Output: a%3Db%26c%3Dd
  echo "test" | omni url encode           # Read from stdin
  omni url encode file.txt                # Read from file

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --component | bool | false | use query component encoding (more aggressive) |
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni url encode "hello world"           # Output: hello%20world
omni url encode --component "a=b&c=d"   # Output: a%3Db%26c%3Dd
omni url encode file.txt                # Read from file
```

---

### uuid

**Category:** Security

**Usage:** `omni uuid [OPTION]... [flags]`

**Description:** Generate random UUIDs

**Details:**

Generate random UUIDs (Universally Unique Identifiers).

Versions:
  4  Random UUID (default) - fully random, no ordering
  7  Time-ordered UUID - timestamp + random, sortable

  -v, --version=N UUID version (4 or 7, default 4)
  -n, --count=N   generate N UUIDs (default 1)
  -u, --upper     output in uppercase
  -x, --no-dashes output without dashes
  --json          output as JSON

Examples:
  omni uuid                  # generate one UUID v4
  omni uuid -v 7             # generate time-ordered UUID v7
  omni uuid -n 5             # generate 5 UUIDs
  omni uuid -u               # uppercase output
  omni uuid -x               # no dashes (32 hex chars)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -n, --count | int | 1 | generate N UUIDs |
| --json | bool | false | output as JSON |
| -x, --no-dashes | bool | false | output without dashes |
| -u, --upper | bool | false | output in uppercase |
| -v, --version | int | 4 | UUID version (4 or 7) |

**Examples:**

```bash
omni uuid                  # generate one UUID v4
omni uuid -v 7             # generate time-ordered UUID v7
omni uuid -n 5             # generate 5 UUIDs
omni uuid -u               # uppercase output
omni uuid -x               # no dashes (32 hex chars)
```

---

### vault

**Category:** Other

**Usage:** `omni vault`

**Description:** HashiCorp Vault CLI

**Details:**

HashiCorp Vault CLI for secrets management.

Environment variables:
  VAULT_ADDR       Vault server address (default: https://127.0.0.1:8200)
  VAULT_TOKEN      Authentication token
  VAULT_NAMESPACE  Vault namespace

Examples:
  omni vault status
  omni vault login -method=token
  omni vault kv get secret/myapp
  omni vault kv put secret/myapp key=value

**Examples:**

```bash
omni vault status
omni vault login -method=token
omni vault kv get secret/myapp
omni vault kv put secret/myapp key=value
```

**Subcommands:** `delete`, `kv`, `list`, `login`, `read`, `status`, `token`, `write`

---

### vault delete

**Category:** Other

**Usage:** `omni vault delete <path>`

**Description:** Delete secrets

**Details:**

Delete a secret at the given path.

Examples:
  omni vault delete secret/data/myapp

**Examples:**

```bash
omni vault delete secret/data/myapp
```

---

### vault kv

**Category:** Other

**Usage:** `omni vault kv`

**Description:** KV secrets engine operations

**Details:**

Interact with Vault's KV secrets engine (v2).

**Subcommands:** `delete`, `destroy`, `get`, `list`, `metadata`, `put`, `undelete`

---

### vault kv delete

**Category:** Other

**Usage:** `omni vault kv delete <path> [flags]`

**Description:** Delete a secret from KV

**Details:**

Soft delete a secret from the KV secrets engine.

Examples:
  omni vault kv delete secret/myapp
  omni vault kv delete -versions=1,2,3 secret/myapp

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --versions | string | - | Comma-separated versions to delete |

**Examples:**

```bash
omni vault kv delete secret/myapp
omni vault kv delete -versions=1,2,3 secret/myapp
```

---

### vault kv destroy

**Category:** Other

**Usage:** `omni vault kv destroy <path> [flags]`

**Description:** Permanently destroy secret versions

**Details:**

Permanently destroy versions of a secret in the KV secrets engine.

Examples:
  omni vault kv destroy -versions=1,2,3 secret/myapp

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --versions | string | - | Comma-separated versions to destroy (required) |

**Examples:**

```bash
omni vault kv destroy -versions=1,2,3 secret/myapp
```

---

### vault kv get

**Category:** Other

**Usage:** `omni vault kv get <path> [flags]`

**Description:** Get a secret from KV

**Details:**

Retrieve a secret from the KV secrets engine.

Examples:
  omni vault kv get secret/myapp
  omni vault kv get -mount=kv myapp
  omni vault kv get -version=2 secret/myapp
  omni vault kv get -field=password secret/myapp

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --field | string | - | Print only this field |
| --version | int | 0 | Version to retrieve (0 = latest) |

**Examples:**

```bash
omni vault kv get secret/myapp
omni vault kv get -mount=kv myapp
omni vault kv get -version=2 secret/myapp
omni vault kv get -field=password secret/myapp
```

---

### vault kv list

**Category:** Other

**Usage:** `omni vault kv list [path]`

**Description:** List secrets in KV

**Details:**

List secrets at a path in the KV secrets engine.

Examples:
  omni vault kv list secret/
  omni vault kv list -mount=kv myapp/

**Examples:**

```bash
omni vault kv list secret/
omni vault kv list -mount=kv myapp/
```

---

### vault kv metadata

**Category:** Other

**Usage:** `omni vault kv metadata`

**Description:** KV metadata operations

**Details:**

Manage metadata for secrets in the KV secrets engine.

**Subcommands:** `delete`, `get`

---

### vault kv metadata delete

**Category:** Other

**Usage:** `omni vault kv metadata delete <path>`

**Description:** Delete all versions and metadata

**Details:**

Permanently delete all versions and metadata for a secret.

Examples:
  omni vault kv metadata delete secret/myapp

**Examples:**

```bash
omni vault kv metadata delete secret/myapp
```

---

### vault kv metadata get

**Category:** Other

**Usage:** `omni vault kv metadata get <path>`

**Description:** Get secret metadata

**Details:**

Retrieve metadata for a secret.

Examples:
  omni vault kv metadata get secret/myapp

**Examples:**

```bash
omni vault kv metadata get secret/myapp
```

---

### vault kv put

**Category:** Other

**Usage:** `omni vault kv put <path> [key=value...]`

**Description:** Put a secret into KV

**Details:**

Write a secret to the KV secrets engine.

Examples:
  omni vault kv put secret/myapp key=value
  omni vault kv put -mount=kv myapp username=admin password=secret

**Examples:**

```bash
omni vault kv put secret/myapp key=value
omni vault kv put -mount=kv myapp username=admin password=secret
```

---

### vault kv undelete

**Category:** Other

**Usage:** `omni vault kv undelete <path> [flags]`

**Description:** Restore deleted secret versions

**Details:**

Restore deleted versions of a secret in the KV secrets engine.

Examples:
  omni vault kv undelete -versions=1,2,3 secret/myapp

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --versions | string | - | Comma-separated versions to restore (required) |

**Examples:**

```bash
omni vault kv undelete -versions=1,2,3 secret/myapp
```

---

### vault list

**Category:** Other

**Usage:** `omni vault list <path>`

**Description:** List secrets

**Details:**

List secrets at the given path.

Examples:
  omni vault list secret/metadata/

**Examples:**

```bash
omni vault list secret/metadata/
```

---

### vault login

**Category:** Other

**Usage:** `omni vault login [token] [flags]`

**Description:** Authenticate to Vault

**Details:**

Authenticate to Vault using various methods.

Methods:
  token     - Direct token authentication (default)
  userpass  - Username/password authentication
  approle   - AppRole authentication

Examples:
  omni vault login                           # Prompts for token
  omni vault login s.xxxxx                   # Direct token
  omni vault login -method=userpass -username=admin
  omni vault login -method=approle -role-id=xxx -secret-id=xxx

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --method | string | token | Auth method (token, userpass, approle) |
| --no-store | bool | false | Don't save token to file |
| --password | string | - | Password for userpass auth |
| --path | string | - | Mount path for auth method |
| --role-id | string | - | Role ID for approle auth |
| --secret-id | string | - | Secret ID for approle auth |
| --username | string | - | Username for userpass auth |

**Examples:**

```bash
omni vault login                           # Prompts for token
omni vault login s.xxxxx                   # Direct token
omni vault login -method=userpass -username=admin
omni vault login -method=approle -role-id=xxx -secret-id=xxx
```

---

### vault read

**Category:** Other

**Usage:** `omni vault read <path> [flags]`

**Description:** Read secrets

**Details:**

Read a secret from Vault at the given path.

Examples:
  omni vault read secret/data/myapp
  omni vault read -field=password secret/data/myapp

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --field | string | - | Print only this field |

**Examples:**

```bash
omni vault read secret/data/myapp
omni vault read -field=password secret/data/myapp
```

---

### vault status

**Category:** Other

**Usage:** `omni vault status`

**Description:** Show Vault server status

**Details:**

Show the status of the Vault server.

Examples:
  omni vault status

**Examples:**

```bash
omni vault status
```

---

### vault token

**Category:** Other

**Usage:** `omni vault token`

**Description:** Token operations

**Details:**

Token management operations.

**Subcommands:** `lookup`, `renew`, `revoke`

---

### vault token lookup

**Category:** Other

**Usage:** `omni vault token lookup`

**Description:** Lookup current token

**Details:**

Display information about the current token.

Examples:
  omni vault token lookup

**Examples:**

```bash
omni vault token lookup
```

---

### vault token renew

**Category:** Other

**Usage:** `omni vault token renew [flags]`

**Description:** Renew current token

**Details:**

Renew the current token's lease.

Examples:
  omni vault token renew
  omni vault token renew -increment=3600

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --increment | int | 0 | Lease increment in seconds |

**Examples:**

```bash
omni vault token renew
omni vault token renew -increment=3600
```

---

### vault token revoke

**Category:** Other

**Usage:** `omni vault token revoke`

**Description:** Revoke current token

**Details:**

Revoke the current token.

Examples:
  omni vault token revoke

**Examples:**

```bash
omni vault token revoke
```

---

### vault write

**Category:** Other

**Usage:** `omni vault write <path> [key=value...]`

**Description:** Write secrets

**Details:**

Write a secret to Vault at the given path.

Examples:
  omni vault write secret/data/myapp key=value
  omni vault write secret/data/myapp username=admin password=secret

**Examples:**

```bash
omni vault write secret/data/myapp key=value
omni vault write secret/data/myapp username=admin password=secret
```

---

### watch

**Category:** Utilities

**Usage:** `omni watch [OPTION]... COMMAND [flags]`

**Description:** Execute a program periodically, showing output fullscreen

**Details:**

Execute a command repeatedly, displaying its output.

Note: Since omni doesn't execute external commands, this version
monitors files or directories for changes.

  -n, --interval=SECS   seconds to wait between updates (default 2)
  -d, --differences     highlight differences between successive updates
  -t, --no-title        turn off header showing the command and time
  -b, --beep            beep if command has a non-zero exit
  -e, --errexit         exit if command has a non-zero exit
  -p, --precise         attempt run command in precise intervals
  -c, --color           interpret ANSI color and style sequences
  -g, --chgexit         exit when command output changes
  --only-changes        only display output when it changes

Examples:
  omni watch -n 1 file myfile.txt    # Watch a file for changes
  omni watch -n 5 dir ./logs         # Watch a directory

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

**Examples:**

```bash
omni watch -n 1 file myfile.txt    # Watch a file for changes
omni watch -n 5 dir ./logs         # Watch a directory
```

---

### wc

**Category:** Text Processing

**Usage:** `omni wc [option]... [file]... [flags]`

**Description:** Print newline, word, and byte counts for each file

**Details:**

Print newline, word, and byte counts for each FILE, and a total line if
more than one FILE is specified. A word is a non-zero-length sequence of
characters delimited by white space.

With no FILE, or when FILE is -, read standard input.

The options below may be used to select which counts are printed, always in
the following order: newline, word, character, byte, maximum line length.

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

**Details:**

Write the full path of COMMAND(s) to standard output.

  -a, --all   print all matching executables in PATH, not just the first

Examples:
  omni which go              # /usr/local/go/bin/go
  omni which python python3  # locate multiple commands
  omni which -a python       # show all python executables

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -a, --all | bool | false | print all matches |
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni which go              # /usr/local/go/bin/go
omni which python python3  # locate multiple commands
omni which -a python       # show all python executables
```

---

### whoami

**Category:** System Info

**Usage:** `omni whoami [flags]`

**Description:** Print effective username

**Details:**

Print the user name associated with the current effective user ID.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

---

### xargs

**Category:** Utilities

**Usage:** `omni xargs [OPTION]... [COMMAND [INITIAL-ARGS]] [flags]`

**Description:** Build and execute command lines from standard input

**Details:**

Read items from standard input, delimited by blanks or newlines, and
execute a command for each item.

Note: Since omni doesn't execute external commands, this version
reads and prints arguments from stdin. It can be used to transform
input for piping to other tools.

  -0, --null            input items are separated by a null character
  -d, --delimiter=DELIM  input items are separated by DELIM
  -n, --max-args=MAX    use at most MAX arguments per command line
  -P, --max-procs=MAX   run at most MAX processes at a time
  -r, --no-run-if-empty if there are no arguments, do not run COMMAND
  -t, --verbose         print commands before executing them
  -I REPLACE-STR        replace occurrences of REPLACE-STR in initial args

Examples:
  echo "a b c" | omni xargs        # prints: a b c
  echo -e "a\nb\nc" | omni xargs   # prints: a b c
  echo -e "a\nb\nc" | omni xargs -n 1  # prints each on separate line

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

**Details:**

XML utilities for formatting, validation, and conversion.

When called directly, formats XML (same as 'xml fmt').

Subcommands:
  fmt         Format/beautify XML
  validate    Validate XML syntax
  tojson      Convert XML to JSON
  fromjson    Convert JSON to XML

Examples:
  omni xml file.xml
  omni xml fmt file.xml
  omni xml validate file.xml
  omni xml "<root><item>value</item></root>"
  cat file.xml | omni xml
  omni xml --minify file.xml
  omni xml tojson file.xml
  omni xml fromjson file.json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| -m, --minify | bool | false | minify XML (remove whitespace) |

**Examples:**

```bash
omni xml file.xml
omni xml fmt file.xml
omni xml validate file.xml
omni xml "<root><item>value</item></root>"
omni xml --minify file.xml
omni xml tojson file.xml
omni xml fromjson file.json
```

**Subcommands:** `fmt`, `fromjson`, `tojson`, `validate`

---

### xml fmt

**Category:** Other

**Usage:** `omni xml fmt [FILE] [flags]`

**Description:** Format XML

**Details:**

Format and beautify XML.

Reads XML from a file, argument, or stdin and outputs formatted XML
with proper indentation. Use --minify to remove whitespace.

Examples:
  omni xml fmt file.xml
  omni xml fmt "<root><item>value</item></root>"
  cat file.xml | omni xml fmt
  omni xml fmt --minify file.xml
  omni xml fmt --indent "    " file.xml

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | string |    | indentation string |
| -m, --minify | bool | false | minify XML (remove whitespace) |

**Examples:**

```bash
omni xml fmt file.xml
omni xml fmt "<root><item>value</item></root>"
omni xml fmt --minify file.xml
omni xml fmt --indent "    " file.xml
```

---

### xml fromjson

**Category:** Other

**Usage:** `omni xml fromjson [FILE] [flags]`

**Description:** Convert JSON to XML

**Details:**

Convert JSON data to XML format.

  -r, --root=NAME      root element name (default "root")
  -i, --indent=STR     indentation string (default "  ")
  --item-tag=NAME      tag for array items (default "item")
  --attr-prefix=STR    prefix for attributes (default "-")

Examples:
  omni xml fromjson file.json
  echo '{"name":"John"}' | omni xml fromjson
  omni xml fromjson -r person file.json

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --attr-prefix | string | - | prefix for attributes |
| -i, --indent | string |    | indentation string |
| --item-tag | string | item | tag for array items |
| -r, --root | string | root | root element name |

**Examples:**

```bash
omni xml fromjson file.json
omni xml fromjson -r person file.json
```

---

### xml tojson

**Category:** Other

**Usage:** `omni xml tojson [FILE] [flags]`

**Description:** Convert XML to JSON

**Details:**

Convert XML data to JSON format.

  --attr-prefix=STR    prefix for attributes in JSON (default "-")
  --text-key=STR       key for text content (default "#text")

Examples:
  omni xml tojson file.xml
  cat file.xml | omni xml tojson
  omni xml tojson --attr-prefix=@ file.xml

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --attr-prefix | string | - | prefix for attributes in JSON |
| --text-key | string | #text | key for text content |

**Examples:**

```bash
omni xml tojson file.xml
omni xml tojson --attr-prefix=@ file.xml
```

---

### xml validate

**Category:** Other

**Usage:** `omni xml validate [FILE...] [flags]`

**Description:** Validate XML syntax

**Details:**

Validate XML syntax for one or more files.

Checks that the input is well-formed XML.

Examples:
  omni xml validate file.xml
  omni xml validate *.xml
  cat file.xml | omni xml validate
  omni xml validate --json file.xml

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |

**Examples:**

```bash
omni xml validate file.xml
omni xml validate *.xml
omni xml validate --json file.xml
```

---

### xxd

**Category:** Other

**Usage:** `omni xxd [OPTIONS] [FILE] [flags]`

**Description:** Make a hex dump or reverse it

**Details:**

Make a hex dump of a file or standard input, or reverse it.

xxd creates a hex dump of a given file or standard input. It can also convert
a hex dump back to its original binary form.

Output Modes:
  (default)    Traditional hex dump with addresses and ASCII
  -p, --plain  Output only hex bytes, no addresses or ASCII
  -i, --include  Output as C include file (array definition)
  -b, --bits   Binary digit dump instead of hex

Reverse Mode:
  -r, --reverse  Convert hex dump back to binary

Examples:
  # Basic hex dump
  omni xxd file.bin
  omni xxd < file.bin
  echo "hello" | omni xxd

  # Plain hex output (like 'hex encode' but for binary files)
  omni xxd -p file.bin

  # C include file output
  omni xxd -i data.bin > data.h

  # Binary dump (bits instead of hex)
  omni xxd -b file.bin

  # Reverse hex dump back to binary
  omni xxd -r hexdump.txt > original.bin
  omni xxd -p file.bin | omni xxd -r -p > copy.bin

  # Limit output to first N bytes
  omni xxd -l 16 file.bin

  # Start at offset
  omni xxd -s 100 file.bin

  # Custom columns and grouping
  omni xxd -c 8 -g 1 file.bin

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

**Examples:**

```bash
omni xxd file.bin
omni xxd < file.bin
omni xxd -p file.bin
omni xxd -i data.bin > data.h
omni xxd -b file.bin
omni xxd -r hexdump.txt > original.bin
omni xxd -p file.bin | omni xxd -r -p > copy.bin
omni xxd -l 16 file.bin
omni xxd -s 100 file.bin
omni xxd -c 8 -g 1 file.bin
```

---

### xz

**Category:** Archive

**Usage:** `omni xz [OPTION]... [FILE]... [flags]`

**Description:** Compress or decompress xz files

**Details:**

Compress or decompress FILEs using xz format.

Note: Full xz support requires external library. Basic info/listing is supported.

  -d, --decompress   decompress
  -k, --keep         keep original files
  -f, --force        force overwrite
  -c, --stdout       write to stdout
  -v, --verbose      verbose mode
  -l, --list         list compressed file info

Examples:
  omni xz -l file.xz           # list info
  omni xz -d file.txt.xz       # decompress (requires external lib)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -d, --decompress | bool | false | decompress |
| -f, --force | bool | false | force overwrite |
| -k, --keep | bool | false | keep original files |
| -l, --list | bool | false | list compressed file info |
| -c, --stdout | bool | false | write to stdout |
| -v, --verbose | bool | false | verbose mode |

**Examples:**

```bash
omni xz -l file.xz           # list info
omni xz -d file.txt.xz       # decompress (requires external lib)
```

---

### xzcat

**Category:** Archive

**Usage:** `omni xzcat [FILE]...`

**Description:** Decompress and print xz files

**Details:**

Decompress and print FILEs to stdout.

Equivalent to xz -dc.

Note: Full decompression requires external library.

---

### yaml

**Category:** Other

**Usage:** `omni yaml`

**Description:** YAML utilities

**Details:**

YAML utilities for validation and formatting.

Subcommands:
  validate    Validate YAML syntax
  fmt         Format/beautify YAML
  tostruct    Convert YAML to Go struct definition

Examples:
  omni yaml validate config.yaml
  omni yaml fmt config.yaml
  omni yaml tostruct config.yaml

**Examples:**

```bash
omni yaml validate config.yaml
omni yaml fmt config.yaml
omni yaml tostruct config.yaml
```

**Subcommands:** `fmt`, `tostruct`, `validate`

---

### yaml fmt

**Category:** Other

**Usage:** `omni yaml fmt [FILE] [flags]`

**Description:** Format YAML

**Details:**

Format and beautify YAML.

Parses YAML and outputs it with consistent formatting.

Examples:
  omni yaml fmt config.yaml
  cat config.yaml | omni yaml fmt
  omni yaml fmt --indent 4 config.yaml

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -i, --indent | int | 2 | indentation width |
| --json | bool | false | output as JSON instead of YAML |

**Examples:**

```bash
omni yaml fmt config.yaml
omni yaml fmt --indent 4 config.yaml
```

---

### yaml tostruct

**Category:** Other

**Usage:** `omni yaml tostruct [FILE] [flags]`

**Description:** Convert YAML to Go struct definition

**Details:**

Convert YAML data to a Go struct definition.

  -n, --name=NAME      struct name (default "Root")
  -p, --package=PKG    package name (default "main")
  --inline             inline nested structs
  --omitempty          add omitempty to all fields

Examples:
  omni yaml tostruct config.yaml
  cat config.yaml | omni yaml tostruct
  omni yaml tostruct -n Config -p models config.yaml
  omni yaml tostruct --omitempty config.yaml

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --inline | bool | false | inline nested structs |
| -n, --name | string | Root | struct name |
| --omitempty | bool | false | add omitempty to all fields |
| -p, --package | string | main | package name |

**Examples:**

```bash
omni yaml tostruct config.yaml
omni yaml tostruct -n Config -p models config.yaml
omni yaml tostruct --omitempty config.yaml
```

---

### yaml validate

**Category:** Other

**Usage:** `omni yaml validate [FILE...] [flags]`

**Description:** Validate YAML syntax

**Details:**

Validate YAML syntax for one or more files.

Checks that the input is valid YAML. Supports multi-document YAML files.

Examples:
  omni yaml validate config.yaml
  omni yaml validate *.yaml
  omni yaml validate --strict config.yaml
  cat config.yaml | omni yaml validate
  omni yaml validate --json config.yaml

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --json | bool | false | output as JSON |
| --strict | bool | false | fail on unknown fields |

**Examples:**

```bash
omni yaml validate config.yaml
omni yaml validate *.yaml
omni yaml validate --strict config.yaml
omni yaml validate --json config.yaml
```

---

### yes

**Category:** Core

**Usage:** `omni yes [STRING]...`

**Description:** Output a string repeatedly until killed

**Details:**

Repeatedly output a line with all specified STRING(s), or 'y'.

Examples:
  omni yes              # outputs 'y' forever
  omni yes hello        # outputs 'hello' forever
  omni yes | head -5    # outputs 5 'y' lines

**Examples:**

```bash
omni yes              # outputs 'y' forever
omni yes hello        # outputs 'hello' forever
omni yes | head -5    # outputs 5 'y' lines
```

---

### yq

**Category:** Data Processing

**Usage:** `omni yq [OPTION]... FILTER [FILE]... [flags]`

**Description:** Command-line YAML processor

**Details:**

yq is a lightweight command-line YAML processor.

Uses the same filter syntax as jq:
  .           identity
  .field      access field
  .field.sub  nested access
  .[n]        array index
  .[]         iterate array

  -r          output raw strings
  -c          compact JSON output
  -o json     output as JSON
  -o yaml     output as YAML (default)
  -n          null input

Examples:
  omni yq '.name' config.yaml
  omni yq -o json '.' config.yaml    # convert YAML to JSON
  echo "name: John" | omni yq '.name'
  omni yq '.items[]' data.yaml

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -c, --compact-output | bool | false | compact output |
| -I, --indent | int | 2 | indentation level |
| -n, --null-input | bool | false | don't read any input |
| -o, --output-format | string | yaml | output format (yaml or json) |
| -r, --raw-output | bool | false | output raw strings |

**Examples:**

```bash
omni yq '.name' config.yaml
omni yq -o json '.' config.yaml    # convert YAML to JSON
omni yq '.items[]' data.yaml
```

---

### zcat

**Category:** Archive

**Usage:** `omni zcat [FILE]...`

**Description:** Decompress and print gzip files

**Details:**

Decompress and print FILEs to stdout.

Equivalent to gzip -dc.

Examples:
  omni zcat file.txt.gz        # print decompressed content
  omni zcat file.gz | grep x   # decompress and grep

**Examples:**

```bash
omni zcat file.txt.gz        # print decompressed content
omni zcat file.gz | grep x   # decompress and grep
```

---

### zip

**Category:** Archive

**Usage:** `omni zip [OPTION]... ZIPFILE FILE... [flags]`

**Description:** Package and compress files into a zip archive

**Details:**

Create a zip archive from files and directories.

  -v, --verbose     verbose output
  -r, --recursive   recurse into directories (default for directories)
  -C, --directory   change to directory before adding files

Examples:
  omni zip archive.zip file1.txt file2.txt   # create zip
  omni zip archive.zip dir/                   # zip directory
  omni zip -v archive.zip file.txt           # verbose output

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| -C, --directory | string | - | change to directory before adding |
| -r, --recursive | bool | false | recurse into directories |
| -v, --verbose | bool | false | verbose output |

**Examples:**

```bash
omni zip archive.zip file1.txt file2.txt   # create zip
omni zip archive.zip dir/                   # zip directory
omni zip -v archive.zip file.txt           # verbose output
```

---

## Library API

Command implementations live under internal/cli/ and cannot be imported by external projects (Go's internal package restriction). omni is designed as a CLI tool, not a library. To reuse the logic, fork the repository or use omni as a subprocess.

**Import pattern:** `internal/cli/<command>/<command>.go`

**Examples:**

```go
// Internal structure (not importable externally):
// Each command follows the pattern:
//   - Options struct for configuration
//   - Run<Command>(w io.Writer, args []string, opts Options) error
//   - Helper functions for implementation

// Example: internal/cli/cat/cat.go
type CatOptions struct {
    NumberAll bool
    JSON      bool
}
func RunCat(w io.Writer, args []string, opts CatOptions) error
```

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
