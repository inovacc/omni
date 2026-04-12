# Command Tree

```
omni
+-- aicontext
|   Usage: omni aicontext [flags]
|   Description: Generate concise context for AI coding agents.

Examples:
  omni aicontext              # Markdown output
  omni aicontext --json       # JSON output
  omni aicontext -c text      # Filter by category
  omni aicontext -o ctx.md    # Write to file
|   
|   Flags:
|     -c, --category string     filter category
|     -j, --json                JSON output
|         --no-structure        omit project structure
|     -o, --output string       write to file
|   
+-- arch
|   Usage: omni arch
|   Description: Print the machine hardware name (similar to uname -m).

Examples:
  omni arch    # x86_64, aarch64, etc.
|   
+-- awk
|   Usage: omni awk [OPTION]... 'program' [FILE]... [flags]
|   Description: Awk scans each input file for lines that match any of a set of patterns.

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
|   
|   Flags:
|     -v, --assign stringSlice  assign value to variable (var=value)
|     -F, --field-separator string  use FS for the input field separator
|   
+-- aws
|   Usage: omni aws
|   Description: AWS CLI operations for core services.

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
|   
|   +-- ec2
|   |   Usage: omni aws ec2
|   |   Description: AWS EC2 instance and resource operations.
|   |   
|   |   +-- describe-instances
|   |   |   Usage: omni aws ec2 describe-instances [flags]
|   |   |   Description: Describes one or more EC2 instances.

Examples:
  omni aws ec2 describe-instances
  omni aws ec2 describe-instances --instance-ids i-1234567890abcdef0
  omni aws ec2 describe-instances --filters "Name=tag:Name,Values=prod-*"
|   |   |   
|   |   |   Flags:
|   |   |         --filters stringSlice  filters in format 'Name=value,Values=v1,v2'
|   |   |         --instance-ids stringSlice  instance IDs
|   |   |         --max-results int32   maximum number of results
|   |   |   
|   |   +-- describe-security-groups
|   |   |   Usage: omni aws ec2 describe-security-groups [flags]
|   |   |   Description: Describes one or more security groups.

Examples:
  omni aws ec2 describe-security-groups
  omni aws ec2 describe-security-groups --group-ids sg-1234567890abcdef0
|   |   |   
|   |   |   Flags:
|   |   |         --filters stringSlice  filters
|   |   |         --group-ids stringSlice  security group IDs
|   |   |   
|   |   +-- describe-vpcs
|   |   |   Usage: omni aws ec2 describe-vpcs [flags]
|   |   |   Description: Describes one or more VPCs.

Examples:
  omni aws ec2 describe-vpcs
  omni aws ec2 describe-vpcs --vpc-ids vpc-1234567890abcdef0
|   |   |   
|   |   |   Flags:
|   |   |         --filters stringSlice  filters
|   |   |         --vpc-ids stringSlice  VPC IDs
|   |   |   
|   |   +-- start-instances
|   |   |   Usage: omni aws ec2 start-instances [flags]
|   |   |   Description: Starts one or more stopped instances.

Examples:
  omni aws ec2 start-instances --instance-ids i-1234567890abcdef0
  omni aws ec2 start-instances --instance-ids i-1234,i-5678
|   |   |   
|   |   |   Flags:
|   |   |         --instance-ids stringSlice  instance IDs (required)
|   |   |   
|   |   \-- stop-instances
|   |       Usage: omni aws ec2 stop-instances [flags]
|   |       Description: Stops one or more running instances.

Examples:
  omni aws ec2 stop-instances --instance-ids i-1234567890abcdef0
  omni aws ec2 stop-instances --instance-ids i-1234 --force
|   |       
|   |       Flags:
|   |             --force               force stop without graceful shutdown
|   |             --instance-ids stringSlice  instance IDs (required)
|   |       
|   +-- iam
|   |   Usage: omni aws iam
|   |   Description: AWS Identity and Access Management (IAM) operations.
|   |   
|   |   +-- get-policy
|   |   |   Usage: omni aws iam get-policy [flags]
|   |   |   Description: Retrieves information about the specified managed policy.

Examples:
  omni aws iam get-policy --policy-arn arn:aws:iam::123456789012:policy/MyPolicy
|   |   |   
|   |   |   Flags:
|   |   |         --policy-arn string   policy ARN (required)
|   |   |   
|   |   +-- get-role
|   |   |   Usage: omni aws iam get-role [flags]
|   |   |   Description: Retrieves information about the specified role, including the role's path, GUID,
ARN, and the role's trust policy.

Examples:
  omni aws iam get-role --role-name MyRole
|   |   |   
|   |   |   Flags:
|   |   |         --role-name string    role name (required)
|   |   |   
|   |   +-- get-user
|   |   |   Usage: omni aws iam get-user [flags]
|   |   |   Description: Retrieves information about the specified IAM user, including the user's creation date,
path, unique ID, and ARN. If no user name is specified, returns information about
the IAM user whose credentials are used to call the operation.

Examples:
  omni aws iam get-user
  omni aws iam get-user --user-name myuser
|   |   |   
|   |   |   Flags:
|   |   |         --user-name string    user name (optional, defaults to current user)
|   |   |   
|   |   +-- list-policies
|   |   |   Usage: omni aws iam list-policies [flags]
|   |   |   Description: Lists all the managed policies that are available in your AWS account.

Examples:
  omni aws iam list-policies
  omni aws iam list-policies --scope Local
  omni aws iam list-policies --only-attached
|   |   |   
|   |   |   Flags:
|   |   |         --max-items int32     maximum number of items
|   |   |         --only-attached       only show attached policies
|   |   |         --path-prefix string  path prefix filter
|   |   |         --scope string        scope: All, AWS, Local
|   |   |   
|   |   \-- list-roles
|   |       Usage: omni aws iam list-roles [flags]
|   |       Description: Lists the IAM roles that have the specified path prefix.

Examples:
  omni aws iam list-roles
  omni aws iam list-roles --path-prefix /service-role/
|   |       
|   |       Flags:
|   |             --max-items int32     maximum number of items
|   |             --path-prefix string  path prefix filter
|   |       
|   +-- s3
|   |   Usage: omni aws s3
|   |   Description: AWS S3 bucket and object operations.

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
|   |   
|   |   +-- cp
|   |   |   Usage: omni aws s3 cp <SOURCE> <DESTINATION> [flags]
|   |   |   Description: Copies files between local filesystem and S3, or between S3 locations.

Examples:
  # Upload to S3
  omni aws s3 cp file.txt s3://my-bucket/file.txt

  # Download from S3
  omni aws s3 cp s3://my-bucket/file.txt ./file.txt

  # Copy between S3 locations
  omni aws s3 cp s3://bucket1/file.txt s3://bucket2/file.txt

  # Dry run
  omni aws s3 cp file.txt s3://my-bucket/file.txt --dryrun
|   |   |   
|   |   |   Flags:
|   |   |         --dryrun              display operations without executing
|   |   |         --quiet               suppress output
|   |   |         --recursive           copy recursively
|   |   |   
|   |   +-- ls
|   |   |   Usage: omni aws s3 ls [S3_URI] [flags]
|   |   |   Description: Lists S3 objects in a bucket or all buckets.

Examples:
  omni aws s3 ls
  omni aws s3 ls s3://my-bucket/
  omni aws s3 ls s3://my-bucket/prefix/ --recursive
|   |   |   
|   |   |   Flags:
|   |   |         --human-readable      display file sizes in human-readable format
|   |   |         --recursive           list recursively
|   |   |         --summarize           display summary information
|   |   |   
|   |   +-- mb
|   |   |   Usage: omni aws s3 mb <S3_URI>
|   |   |   Description: Creates an S3 bucket.

Examples:
  omni aws s3 mb s3://my-new-bucket
  omni aws s3 mb s3://my-new-bucket --region us-west-2
|   |   |   
|   |   +-- presign
|   |   |   Usage: omni aws s3 presign <S3_URI> [flags]
|   |   |   Description: Generates a presigned URL for an S3 object.

Examples:
  omni aws s3 presign s3://my-bucket/file.txt
  omni aws s3 presign s3://my-bucket/file.txt --expires-in 3600
|   |   |   
|   |   |   Flags:
|   |   |         --expires-in int      URL expiration time in seconds (default 15 minutes)
|   |   |   
|   |   +-- rb
|   |   |   Usage: omni aws s3 rb <S3_URI> [flags]
|   |   |   Description: Deletes an S3 bucket. The bucket must be empty unless --force is specified.

Examples:
  omni aws s3 rb s3://my-bucket
  omni aws s3 rb s3://my-bucket --force
|   |   |   
|   |   |   Flags:
|   |   |         --force               delete all objects before removing bucket
|   |   |   
|   |   \-- rm
|   |       Usage: omni aws s3 rm <S3_URI> [flags]
|   |       Description: Deletes objects from S3.

Examples:
  omni aws s3 rm s3://my-bucket/file.txt
  omni aws s3 rm s3://my-bucket/prefix/ --recursive
  omni aws s3 rm s3://my-bucket/prefix/ --recursive --dryrun
|   |       
|   |       Flags:
|   |             --dryrun              display operations without executing
|   |             --quiet               suppress output
|   |             --recursive           delete recursively
|   |       
|   +-- ssm
|   |   Usage: omni aws ssm
|   |   Description: AWS Systems Manager Parameter Store operations.
|   |   
|   |   +-- delete-parameter
|   |   |   Usage: omni aws ssm delete-parameter [flags]
|   |   |   Description: Deletes a parameter.

Examples:
  omni aws ssm delete-parameter --name /app/config
|   |   |   
|   |   |   Flags:
|   |   |         --name string         parameter name (required)
|   |   |   
|   |   +-- get-parameter
|   |   |   Usage: omni aws ssm get-parameter [flags]
|   |   |   Description: Retrieves information about a parameter.

Examples:
  omni aws ssm get-parameter --name /app/config
  omni aws ssm get-parameter --name /app/secret --with-decryption
|   |   |   
|   |   |   Flags:
|   |   |         --name string         parameter name (required)
|   |   |         --with-decryption     decrypt SecureString values
|   |   |   
|   |   +-- get-parameters
|   |   |   Usage: omni aws ssm get-parameters [flags]
|   |   |   Description: Retrieves information about multiple parameters.

Examples:
  omni aws ssm get-parameters --names /app/config,/app/secret
  omni aws ssm get-parameters --names /app/config --names /app/secret --with-decryption
|   |   |   
|   |   |   Flags:
|   |   |         --names stringSlice   parameter names (required)
|   |   |         --with-decryption     decrypt SecureString values
|   |   |   
|   |   +-- get-parameters-by-path
|   |   |   Usage: omni aws ssm get-parameters-by-path [flags]
|   |   |   Description: Retrieves all parameters within a hierarchy.

Examples:
  omni aws ssm get-parameters-by-path --path /app/
  omni aws ssm get-parameters-by-path --path /app/ --recursive --with-decryption
|   |   |   
|   |   |   Flags:
|   |   |         --max-results int32   maximum results per page
|   |   |         --path string         parameter path (required)
|   |   |         --recursive           include nested parameters
|   |   |         --with-decryption     decrypt SecureString values
|   |   |   
|   |   \-- put-parameter
|   |       Usage: omni aws ssm put-parameter [flags]
|   |       Description: Creates or updates a parameter.

Examples:
  omni aws ssm put-parameter --name /app/config --value "config-value" --type String
  omni aws ssm put-parameter --name /app/secret --value "secret" --type SecureString
  omni aws ssm put-parameter --name /app/config --value "new-value" --overwrite
|   |       
|   |       Flags:
|   |             --description string  parameter description
|   |             --key-id string       KMS key for SecureString
|   |             --name string         parameter name (required)
|   |             --overwrite           overwrite existing parameter
|   |             --type string         parameter type: String, StringList, SecureString
|   |             --value string        parameter value (required)
|   |       
|   \-- sts
|       Usage: omni aws sts
|       Description: AWS Security Token Service (STS) operations.
|       
|       +-- assume-role
|       |   Usage: omni aws sts assume-role [flags]
|       |   Description: Returns a set of temporary security credentials that you can use to access AWS resources.

Examples:
  omni aws sts assume-role --role-arn arn:aws:iam::123456789012:role/MyRole --role-session-name MySession
|       |   
|       |   Flags:
|       |         --duration-seconds int32  duration of the session in seconds
|       |         --external-id string  external ID for cross-account access
|       |         --role-arn string     ARN of the role to assume (required)
|       |         --role-session-name string  session name (required)
|       |   
|       \-- get-caller-identity
|           Usage: omni aws sts get-caller-identity
|           Description: Returns details about the IAM user or role whose credentials are used to call the operation.

Examples:
  omni aws sts get-caller-identity
|           
+-- banner
|   Usage: omni banner [TEXT] [flags]
|   Description: Generate FIGlet-style ASCII art text banners.

Supports multiple fonts and reads text from arguments or stdin.

  -f, --font=NAME   font name (default "standard")
  -w, --width=N     max output width (0 = unlimited)
  -l, --list        list available fonts

Examples:
  omni banner "Hello World"
  omni banner -f slant "omni"
  omni banner -f small "test"
  omni banner --list
  echo "piped" | omni banner
|   
|   Flags:
|     -f, --font string         font name
|     -l, --list                list available fonts
|     -w, --width int           max output width (0 = unlimited)
|   
+-- base32
|   Usage: omni base32 [OPTION]... [FILE] [flags]
|   Description: Base32 encode or decode FILE, or standard input, to standard output.

With no FILE, or when FILE is -, read standard input.

  -d, --decode    decode data
  -w, --wrap=N    wrap encoded lines after N characters (default 76, 0 = no wrap)

Examples:
  echo "hello" | omni base32           # encode
  echo "NBSWY3DP" | omni base32 -d     # decode
  omni base32 file.bin                 # encode file
|   
|   Flags:
|     -d, --decode              decode data
|     -w, --wrap int            wrap encoded lines after N characters (0 = no wrap)
|   
+-- base58
|   Usage: omni base58 [OPTION]... [FILE] [flags]
|   Description: Base58 encode or decode FILE, or standard input, to standard output.

Uses Bitcoin/IPFS alphabet: 123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz

With no FILE, or when FILE is -, read standard input.

  -d, --decode    decode data

Examples:
  echo "hello" | omni base58           # encode
  omni base58 -d encoded.txt           # decode
|   
|   Flags:
|     -d, --decode              decode data
|   
+-- base64
|   Usage: omni base64 [OPTION]... [FILE] [flags]
|   Description: Base64 encode or decode FILE, or standard input, to standard output.

With no FILE, or when FILE is -, read standard input.

  -d, --decode    decode data
  -w, --wrap=N    wrap encoded lines after N characters (default 76, 0 = no wrap)
  -i, --ignore-garbage  ignore non-alphabet characters when decoding

Examples:
  echo "hello" | omni base64           # encode
  echo "aGVsbG8K" | omni base64 -d     # decode
  omni base64 file.bin                 # encode file
  omni base64 -d encoded.txt           # decode file
|   
|   Flags:
|     -d, --decode              decode data
|     -i, --ignore-garbage      ignore non-alphabet characters when decoding
|     -w, --wrap int            wrap encoded lines after N characters (0 = no wrap)
|   
+-- basename
|   Usage: omni basename NAME [SUFFIX] [flags]
|   Description: Print NAME with any leading directory components removed.
If specified, also remove a trailing SUFFIX.
|   
|   Flags:
|     -s, --suffix string       remove a trailing SUFFIX
|   
+-- bbolt
|   Usage: omni bbolt
|   Description: bbolt provides commands for working with BoltDB databases.

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
|   
|   +-- buckets
|   |   Usage: omni bbolt buckets <database>
|   |   Description: List all buckets in the database
|   |   
|   +-- check
|   |   Usage: omni bbolt check <database>
|   |   Description: Verify database integrity
|   |   
|   +-- compact
|   |   Usage: omni bbolt compact <source> <destination>
|   |   Description: Compact database to a new file
|   |   
|   +-- create-bucket
|   |   Usage: omni bbolt create-bucket <database> <bucket>
|   |   Description: Create a new bucket
|   |   
|   +-- delete
|   |   Usage: omni bbolt delete <database> <bucket> <key>
|   |   Description: Delete a key from a bucket
|   |   
|   +-- delete-bucket
|   |   Usage: omni bbolt delete-bucket <database> <bucket>
|   |   Description: Delete a bucket
|   |   
|   +-- dump
|   |   Usage: omni bbolt dump <database> <bucket> [flags]
|   |   Description: Dump all keys and values in a bucket
|   |   
|   |   Flags:
|   |         --hex                 display values in hexadecimal
|   |         --prefix string       filter keys by prefix
|   |   
|   +-- get
|   |   Usage: omni bbolt get <database> <bucket> <key> [flags]
|   |   Description: Get value for a key
|   |   
|   |   Flags:
|   |         --hex                 display value in hexadecimal
|   |   
|   +-- info
|   |   Usage: omni bbolt info <database>
|   |   Description: Display database information
|   |   
|   +-- keys
|   |   Usage: omni bbolt keys <database> <bucket> [flags]
|   |   Description: List keys in a bucket
|   |   
|   |   Flags:
|   |         --prefix string       filter keys by prefix
|   |   
|   +-- page
|   |   Usage: omni bbolt page <database> <page-id>
|   |   Description: Hex dump of a specific page
|   |   
|   +-- pages
|   |   Usage: omni bbolt pages <database>
|   |   Description: List database pages
|   |   
|   +-- put
|   |   Usage: omni bbolt put <database> <bucket> <key> <value>
|   |   Description: Store a key-value pair
|   |   
|   \-- stats
|       Usage: omni bbolt stats <database>
|       Description: Display database statistics
|       
+-- brdoc
|   Usage: omni brdoc
|   Description: Brazilian document validation, generation, and formatting.

Subcommands:
  cpf     CPF (Cadastro de Pessoas Físicas) operations
  cnpj    CNPJ (Cadastro Nacional de Pessoa Jurídica) operations

Examples:
  omni brdoc cpf --generate           # generate a valid CPF
  omni brdoc cpf --validate 123.456.789-09
  omni brdoc cnpj --generate          # generate alphanumeric CNPJ
  omni brdoc cnpj --generate --legacy # generate numeric-only CNPJ
|   
|   +-- cnpj
|   |   Usage: omni brdoc cnpj [CNPJ...] [flags]
|   |   Description: CNPJ (Cadastro Nacional de Pessoa Jurídica) operations.

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
|   |   
|   |   Flags:
|   |     -n, --count int           number of CNPJs to generate
|   |     -f, --format              format CNPJ(s)
|   |     -g, --generate            generate valid CNPJ(s)
|   |         --json                output as JSON
|   |     -l, --legacy              generate numeric-only CNPJ
|   |     -v, --validate            validate CNPJ(s)
|   |   
|   \-- cpf
|       Usage: omni brdoc cpf [CPF...] [flags]
|       Description: CPF (Cadastro de Pessoas Físicas) operations.

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
|       
|       Flags:
|         -n, --count int           number of CPFs to generate
|         -f, --format              format CPF(s)
|         -g, --generate            generate valid CPF(s)
|             --json                output as JSON
|         -v, --validate            validate CPF(s)
|       
+-- buf
|   Usage: omni buf
|   Description: Protocol buffer utilities inspired by buf.build.

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
|   
|   +-- breaking
|   |   Usage: omni buf breaking [DIR] [flags]
|   |   Description: Check for breaking changes against a previous version.

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
|   |   
|   |   Flags:
|   |         --against string      source to compare against (required)
|   |         --error-format string  output format (text, json, github-actions)
|   |         --exclude-imports     don't check imported files
|   |         --exclude-path stringSlice  paths to exclude
|   |   
|   +-- compile
|   |   Usage: omni buf compile [DIR] [flags]
|   |   Description: Compile proto files and output an image.

Flags:
  -o, --output=FILE      Output file (.bin or .json)
  --exclude-path=PATH    Paths to exclude
  --error-format=FORMAT  Output format: text, json, github-actions

Examples:
  omni buf compile
  omni buf compile -o image.bin
  omni buf compile -o image.json
  omni buf compile --exclude-path=vendor
|   |   
|   |   Flags:
|   |         --error-format string  output format (text, json, github-actions)
|   |         --exclude-path stringSlice  paths to exclude
|   |     -o, --output string       output file path
|   |   
|   +-- format
|   |   Usage: omni buf format [DIR] [flags]
|   |   Description: Format proto files with consistent style.

Flags:
  -w, --write       Rewrite files in place
  -d, --diff        Display diff instead of formatted output
  --exit-code       Exit with non-zero if files are not formatted

Examples:
  omni buf format
  omni buf format --write
  omni buf format --diff
  omni buf format --exit-code  # for CI
|   |   
|   |   Flags:
|   |     -d, --diff                display diff
|   |         --exit-code           exit with non-zero if files unformatted
|   |     -w, --write               rewrite files in place
|   |   
|   +-- generate
|   |   Usage: omni buf generate [DIR] [flags]
|   |   Description: Generate code using plugins defined in buf.gen.yaml.

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
|   |   
|   |   Flags:
|   |         --include-imports     include imported files
|   |     -o, --output string       base output directory
|   |         --template string     alternate buf.gen.yaml location
|   |   
|   +-- lint
|   |   Usage: omni buf lint [DIR] [flags]
|   |   Description: Lint proto files for style and structure issues.

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
|   |   
|   |   Flags:
|   |         --config string       custom config file path
|   |         --error-format string  output format (text, json, github-actions)
|   |         --exclude-path stringSlice  paths to exclude
|   |   
|   +-- ls-files
|   |   Usage: omni buf ls-files [DIR]
|   |   Description: List all proto files in the module.

Examples:
  omni buf ls-files
  omni buf ls-files ./proto
|   |   
|   \-- mod
|       Usage: omni buf mod
|       Description: Module management commands.

Subcommands:
  init    Initialize a new buf module
  update  Update dependencies
|       
|       +-- init
|       |   Usage: omni buf mod init [NAME] [flags]
|       |   Description: Initialize a new buf.yaml configuration file.

Examples:
  omni buf mod init
  omni buf mod init buf.build/org/repo
|       |   
|       |   Flags:
|       |         --dir string          directory to initialize
|       |   
|       \-- update
|           Usage: omni buf mod update
|           Description: Update dependencies listed in buf.yaml.

Note: Full dependency resolution requires network access to BSR.

Examples:
  omni buf mod update
|           
+-- bunzip2
|   Usage: omni bunzip2 [OPTION]... [FILE]... [flags]
|   Description: Decompress FILEs in bzip2 format.

Equivalent to bzip2 -d.

Examples:
  omni bunzip2 file.txt.bz2    # decompress
  omni bunzip2 -k file.txt.bz2 # keep original
|   
|   Flags:
|     -f, --force               force overwrite
|     -k, --keep                keep original files
|     -c, --stdout              write to stdout
|     -v, --verbose             verbose mode
|   
+-- bzcat
|   Usage: omni bzcat [FILE]...
|   Description: Decompress and print FILEs to stdout.

Equivalent to bzip2 -dc.

Examples:
  omni bzcat file.txt.bz2      # print decompressed content
|   
+-- bzip2
|   Usage: omni bzip2 [OPTION]... [FILE]... [flags]
|   Description: Decompress FILEs using bzip2 format.

Note: Only decompression is supported (Go stdlib limitation).

  -d, --decompress   decompress (required)
  -k, --keep         keep original files
  -f, --force        force overwrite
  -c, --stdout       write to stdout
  -v, --verbose      verbose mode

Examples:
  omni bzip2 -d file.txt.bz2   # decompress
  omni bzip2 -dk file.txt.bz2  # decompress, keep original
|   
|   Flags:
|     -d, --decompress          decompress
|     -f, --force               force overwrite
|     -k, --keep                keep original files
|     -c, --stdout              write to stdout
|     -v, --verbose             verbose mode
|   
+-- case
|   Usage: omni case
|   Description: Convert text between different case conventions.

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
|   
|   +-- all
|   |   Usage: omni case all [TEXT...]
|   |   Description: Convert text to all supported case types and display results.

Examples:
  omni case all "hello world"
  omni case all "helloWorld"
|   |   
|   +-- camel
|   |   Usage: omni case camel [TEXT...]
|   |   Description: Convert text to camelCase.

Examples:
  omni case camel "hello world"       # helloWorld
  omni case camel "Hello_World"       # helloWorld
  omni case camel "hello-world"       # helloWorld
|   |   
|   +-- constant
|   |   Usage: omni case constant [TEXT...]
|   |   Description: Convert text to CONSTANT_CASE (SCREAMING_SNAKE_CASE).

Examples:
  omni case constant "hello world"    # HELLO_WORLD
  omni case constant "helloWorld"     # HELLO_WORLD
|   |   
|   +-- detect
|   |   Usage: omni case detect [TEXT...]
|   |   Description: Detect the case type of the input text.

Examples:
  omni case detect "helloWorld"       # camel
  omni case detect "hello_world"      # snake
  omni case detect "HELLO_WORLD"      # constant
|   |   
|   +-- dot
|   |   Usage: omni case dot [TEXT...]
|   |   Description: Convert text to dot.case.

Examples:
  omni case dot "hello world"         # hello.world
  omni case dot "helloWorld"          # hello.world
|   |   
|   +-- kebab
|   |   Usage: omni case kebab [TEXT...]
|   |   Description: Convert text to kebab-case.

Examples:
  omni case kebab "hello world"       # hello-world
  omni case kebab "helloWorld"        # hello-world
  omni case kebab "HelloWorld"        # hello-world
|   |   
|   +-- lower
|   |   Usage: omni case lower [TEXT...]
|   |   Description: Convert text to lowercase.

Examples:
  omni case lower "HELLO WORLD"       # hello world
  echo "HELLO" | omni case lower      # hello
|   |   
|   +-- pascal
|   |   Usage: omni case pascal [TEXT...]
|   |   Description: Convert text to PascalCase.

Examples:
  omni case pascal "hello world"      # HelloWorld
  omni case pascal "hello_world"      # HelloWorld
  omni case pascal "hello-world"      # HelloWorld
|   |   
|   +-- path
|   |   Usage: omni case path [TEXT...]
|   |   Description: Convert text to path/case.

Examples:
  omni case path "hello world"        # hello/world
  omni case path "helloWorld"         # hello/world
|   |   
|   +-- sentence
|   |   Usage: omni case sentence [TEXT...]
|   |   Description: Convert text to Sentence case (capitalize first letter only).

Examples:
  omni case sentence "hello world"    # Hello world
  echo "HELLO WORLD" | omni case sentence
|   |   
|   +-- snake
|   |   Usage: omni case snake [TEXT...]
|   |   Description: Convert text to snake_case.

Examples:
  omni case snake "hello world"       # hello_world
  omni case snake "helloWorld"        # hello_world
  omni case snake "HelloWorld"        # hello_world
|   |   
|   +-- swap
|   |   Usage: omni case swap [TEXT...]
|   |   Description: Swap the case of each character (upper becomes lower, lower becomes upper).

Examples:
  omni case swap "Hello World"        # hELLO wORLD
  omni case swap "helloWorld"         # HELLOwORLD
|   |   
|   +-- title
|   |   Usage: omni case title [TEXT...]
|   |   Description: Convert text to Title Case (capitalize first letter of each word).

Examples:
  omni case title "hello world"       # Hello World
  echo "hello world" | omni case title
|   |   
|   +-- toggle
|   |   Usage: omni case toggle [TEXT...]
|   |   Description: Toggle the case of the first character.

Examples:
  omni case toggle "hello"            # Hello
  omni case toggle "Hello"            # hello
|   |   
|   \-- upper
|       Usage: omni case upper [TEXT...]
|       Description: Convert text to UPPERCASE.

Examples:
  omni case upper "hello world"       # HELLO WORLD
  omni case upper hello world         # HELLO WORLD
  echo "hello" | omni case upper      # HELLO
|       
+-- cat
|   Usage: omni cat [file...] [flags]
|   Description: Concatenate FILE(s) to standard output.
With no FILE, or when FILE is -, read standard input.
|   
|   Flags:
|     -e, --e                   equivalent to -vE
|         --json                output as JSON array of lines
|     -n, --number              number all output lines
|     -b, --number-nonblank     number nonempty output lines, overrides -n
|     -A, --show-all            equivalent to -vET
|     -E, --show-ends           display $ at end of each line
|     -v, --show-nonprinting    use ^ and M- notation, except for LFD and TAB
|     -T, --show-tabs           display TAB characters as ^I
|     -s, --squeeze-blank       suppress repeated empty output lines
|     -t, --t                   equivalent to -vT
|   
+-- chmod
|   Usage: omni chmod [OPTION]... MODE[,MODE]... FILE... [flags]
|   Description: Change the mode of each FILE to MODE.

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
|   
|   Flags:
|     -c, --changes             like verbose but report only when a change is made
|     -R, --recursive           change files and directories recursively
|         --reference string    use RFILE's mode instead of MODE values
|     -f, --silent              suppress most error messages
|     -v, --verbose             output a diagnostic for every file processed
|   
+-- chown
|   Usage: omni chown [OPTION]... OWNER[:GROUP] FILE... [flags]
|   Description: Change the owner and/or group of each FILE to OWNER and/or GROUP.

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
|   
|   Flags:
|     -c, --changes             like verbose but report only when a change is made
|     -h, --no-dereference      affect symbolic links instead of referenced file
|         --preserve-root       fail to operate recursively on '/'
|     -R, --recursive           operate on files and directories recursively
|         --reference string    use RFILE's owner and group
|     -f, --silent              suppress most error messages
|     -v, --verbose             output a diagnostic for every file processed
|   
+-- cloud
|   Usage: omni cloud
|   Description: Cloud profile management for AWS, Azure, and GCP.

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
|   
|   \-- profile
|       Usage: omni cloud profile
|       Description: Manage cloud profiles for AWS, Azure, and GCP.
|       
|       +-- add
|       |   Usage: omni cloud profile add <name> [flags]
|       |   Description: Add a new cloud profile with encrypted credentials.

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
|       |   
|       |   Flags:
|       |         --access-key-id string  AWS Access Key ID
|       |         --account-id string   Account/Subscription ID
|       |         --client-id string    Azure Client ID
|       |         --client-secret string  Azure Client Secret
|       |         --default             Set as default profile
|       |         --key-file string     Path to GCP service account JSON file
|       |     -p, --provider string     Cloud provider (aws, azure, gcp) (required)
|       |         --region string       Default region for the profile
|       |         --role-arn string     IAM Role ARN (AWS only)
|       |         --secret-access-key string  AWS Secret Access Key
|       |         --session-token string  AWS Session Token (optional)
|       |         --subscription-id string  Azure Subscription ID
|       |         --tenant-id string    Azure Tenant ID
|       |   
|       +-- delete
|       |   Usage: omni cloud profile delete <name> [flags]
|       |   Description: Delete a cloud profile and its encrypted credentials.

Examples:
  omni cloud profile delete myaws --provider aws
  omni cloud profile delete myaws --provider aws --force
|       |   
|       |   Flags:
|       |     -f, --force               Skip confirmation
|       |     -p, --provider string     Provider (required)
|       |   
|       +-- import
|       |   Usage: omni cloud profile import [name] [flags]
|       |   Description: Import existing credentials from AWS, Azure, or GCP CLI configurations.

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
|       |   
|       |   Flags:
|       |         --default             Set as default profile after import
|       |         --list                List available profiles/credentials to import
|       |     -p, --provider string     Cloud provider (aws, azure, gcp) (required)
|       |     -s, --source string       Source profile/file to import from
|       |   
|       +-- list
|       |   Usage: omni cloud profile list [flags]
|       |   Description: List all cloud profiles, optionally filtered by provider.

Examples:
  omni cloud profile list
  omni cloud profile list --provider aws
|       |   
|       |   Flags:
|       |     -p, --provider string     Filter by provider (aws, azure, gcp)
|       |   
|       +-- show
|       |   Usage: omni cloud profile show <name> [flags]
|       |   Description: Show details of a cloud profile (without credentials).

Examples:
  omni cloud profile show myaws
  omni cloud profile show myaws --provider aws
|       |   
|       |   Flags:
|       |     -p, --provider string     Provider (defaults to aws)
|       |   
|       \-- use
|           Usage: omni cloud profile use <name> [flags]
|           Description: Set a profile as the default for its provider.

Examples:
  omni cloud profile use myaws
  omni cloud profile use myaws --provider aws
|           
|           Flags:
|             -p, --provider string     Provider (defaults to aws)
|           
+-- cmdtree
|   Usage: omni cmdtree [flags]
|   Description: Display a tree visualization of all available commands with descriptions.
|   
|   Flags:
|     -b, --brief               Show compact tree with short descriptions only
|     -c, --command string      Show details for a specific command only
|         --json                output as JSON
|         --table               output as aligned table
|     -v, --verbose             Show full details for all commands (default)
|   
+-- cmp
|   Usage: omni cmp [OPTION]... FILE1 FILE2 [flags]
|   Description: Compare two files byte by byte.

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
|   
|   Flags:
|     -n, --bytes int64         compare at most LIMIT bytes
|     -i, --ignore-initial int64  skip first SKIP bytes
|     -b, --print-bytes         print differing bytes
|     -s, --silent              suppress all output
|     -l, --verbose             output byte numbers and values
|   
+-- column
|   Usage: omni column [OPTION]... [FILE]... [flags]
|   Description: Format input into multiple columns.

With no FILE, or when FILE is -, read standard input.

  -t, --table            determine column count based on input
  -s, --separator=STRING delimiter characters for -t option
  -o, --output-separator=STRING  output separator for table mode
  -c, --columns=N        output width in characters (default 80)
  -x, --fillrows         fill rows before columns
  -R, --right            right-align columns
|   
|   Flags:
|     -c, --columns int         output width in characters
|     -x, --fillrows            fill rows before columns
|     -H, --headers string      column headers (comma-separated)
|     -J, --json                output as JSON
|     -o, --output-separator string  output separator for table mode
|     -R, --right               right-align columns
|     -s, --separator string    delimiter characters for -t option
|     -t, --table               determine column count based on input
|   
+-- comm
|   Usage: omni comm [OPTION]... FILE1 FILE2 [flags]
|   Description: Compare sorted files FILE1 and FILE2 line by line.

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
|   
|   Flags:
|     -1, --1                   suppress column 1
|     -2, --2                   suppress column 2
|     -3, --3                   suppress column 3
|         --check-order         check input is sorted
|         --nocheck-order       do not check input order
|         --output-delimiter string  use STR as output delimiter
|     -z, --zero-terminated     line delimiter is NUL
|   
+-- copy
|   Usage: omni copy
|   Description: copy is an alias for the cp command.

Usage:
  omni copy SOURCE DEST

See 'omni cp --help' for full options.
|   
+-- cp
|   Usage: omni cp [source...] [destination]
|   Description: Copy SOURCE to DEST, or multiple SOURCE(s) to DIRECTORY.
|   
+-- crc32sum
|   Usage: omni crc32sum [OPTION]... [FILE]... [flags]
|   Description: Print or check CRC32 (IEEE polynomial) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Examples:
  omni crc32sum file.txt           # compute CRC32
  omni crc32sum -c checksums.txt   # verify checksums
  omni crc32sum file1 file2        # hash multiple files
|   
|   Flags:
|     -b, --binary              read in binary mode
|     -c, --check               read checksums from FILE and check them
|         --quiet               don't print OK for verified files
|         --status              don't output anything, use status code
|     -w, --warn                warn about improperly formatted lines
|   
+-- crc64sum
|   Usage: omni crc64sum [OPTION]... [FILE]... [flags]
|   Description: Print or check CRC64 (ECMA polynomial) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Examples:
  omni crc64sum file.txt           # compute CRC64
  omni crc64sum -c checksums.txt   # verify checksums
  omni crc64sum file1 file2        # hash multiple files
|   
|   Flags:
|     -b, --binary              read in binary mode
|     -c, --check               read checksums from FILE and check them
|         --quiet               don't print OK for verified files
|         --status              don't output anything, use status code
|     -w, --warn                warn about improperly formatted lines
|   
+-- cron
|   Usage: omni cron EXPRESSION [flags]
|   Description: Parse cron expressions and display human-readable explanations.

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
|   
|   Flags:
|         --next int            show next N scheduled runs
|         --validate            only validate the expression
|   
+-- css
|   Usage: omni css [FILE] [flags]
|   Description: CSS utilities for formatting, minifying, and validating CSS.

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
|   
|   Flags:
|     -i, --indent string       indentation string
|         --sort-props          sort properties alphabetically
|         --sort-rules          sort selectors alphabetically
|   
|   +-- fmt
|   |   Usage: omni css fmt [FILE] [flags]
|   |   Description: Format CSS with proper indentation.

  -i, --indent=STR     indentation string (default "  ")
  --sort-props         sort properties alphabetically
  --sort-rules         sort selectors alphabetically

Examples:
  omni css fmt file.css
  omni css fmt "body{margin:0;padding:0}"
  cat file.css | omni css fmt
  omni css fmt --sort-props file.css
|   |   
|   |   Flags:
|   |     -i, --indent string       indentation string
|   |         --sort-props          sort properties alphabetically
|   |         --sort-rules          sort selectors alphabetically
|   |   
|   +-- minify
|   |   Usage: omni css minify [FILE]
|   |   Description: Minify CSS by removing unnecessary whitespace and comments.

Examples:
  omni css minify file.css
  cat file.css | omni css minify
|   |   
|   \-- validate
|       Usage: omni css validate [FILE]
|       Description: Validate CSS syntax.

Exit codes:
  0  Valid CSS
  1  Invalid CSS or error

  --json    output result as JSON

Examples:
  omni css validate file.css
  omni css validate "body { margin: 0; }"
  omni css validate --json file.css
|       
+-- csv
|   Usage: omni csv
|   Description: CSV utilities for converting between CSV and JSON formats.

Subcommands:
  tojson    Convert CSV to JSON array
  fromjson  Convert JSON array to CSV

Examples:
  omni csv tojson file.csv             # convert CSV to JSON
  omni csv fromjson file.json          # convert JSON to CSV
  cat data.csv | omni csv tojson       # from stdin
  omni csv tojson -d ";" file.csv      # custom delimiter
|   
|   +-- fromjson
|   |   Usage: omni csv fromjson [FILE] [flags]
|   |   Description: Convert JSON array of objects to CSV format.

Nested objects are flattened with dot notation (e.g., address.city).

  --no-header          don't include header row
  -d, --delimiter=STR  field delimiter (default ",")
  --no-quotes          don't quote fields

Examples:
  omni csv fromjson file.json
  echo '[{"name":"John","age":30}]' | omni csv fromjson
  omni csv fromjson -d ";" file.json   # semicolon delimiter
  omni csv fromjson --no-header file.json
|   |   
|   |   Flags:
|   |     -d, --delimiter string    field delimiter
|   |         --no-header           don't include header row
|   |         --no-quotes           don't quote fields
|   |   
|   \-- tojson
|       Usage: omni csv tojson [FILE] [flags]
|       Description: Convert CSV data to JSON array of objects.

  --no-header          first row is data, not headers
  -d, --delimiter=STR  field delimiter (default ",")
  -a, --array          always output as array (even for single row)

Examples:
  omni csv tojson file.csv
  cat file.csv | omni csv tojson
  omni csv tojson -d ";" file.csv      # semicolon delimiter
  omni csv tojson --no-header file.csv
|       
|       Flags:
|         -a, --array               always output as array
|         -d, --delimiter string    field delimiter
|             --no-header           first row is data, not headers
|       
+-- curl
|   Usage: omni curl [METHOD] URL [ITEM...] [flags]
|   Description: HTTP client inspired by curlie/httpie.

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
|   
|   Flags:
|     -d, --data string         request body data
|     -f, --form                send as form data instead of JSON
|     -H, --header stringArray  custom header (can be used multiple times)
|     -k, --insecure            skip TLS verification
|         --json                output response as structured JSON
|     -L, --location            follow redirects
|     -t, --timeout int         request timeout in seconds
|     -v, --verbose             show request/response details
|   
+-- cut
|   Usage: omni cut [OPTION]... [FILE]... [flags]
|   Description: Print selected parts of lines from each FILE to standard output.

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
|   
|   Flags:
|     -b, --bytes string        select only these bytes
|     -c, --characters string   select only these characters
|         --complement          complement the set of selected bytes, characters or fields
|     -d, --delimiter string    use DELIM instead of TAB for field delimiter
|     -f, --fields string       select only these fields
|     -s, --only-delimited      do not print lines not containing delimiters
|         --output-delimiter string  use STRING as the output delimiter
|   
+-- date
|   Usage: omni date [+FORMAT] [flags]
|   Description: Display the current time in the given FORMAT, or set the system date.

FORMAT controls the output. Interpreted sequences are:
  %Y   year
  %m   month (01..12)
  %d   day of month (01..31)
  %H   hour (00..23)
  %M   minute (00..59)
  %S   second (00..60)
|   
|   Flags:
|         --iso-8601            output date/time in ISO 8601 format
|     -u, --utc                 print Coordinated Universal Time (UTC)
|   
+-- dd
|   Usage: omni dd [OPERAND]...
|   Description: Copy a file, converting and formatting according to the operands.

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
|   
+-- decrypt
|   Usage: omni decrypt [OPTION]... [FILE] [flags]
|   Description: Decrypt FILE or standard input using AES-256-GCM.

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
|   
|   Flags:
|     -a, --armor               input is ASCII armored (base64)
|     -b, --base64              input is base64 (same as -a)
|     -i, --iterations int      PBKDF2 iterations
|     -k, --key-file string     use key file for decryption
|     -o, --output string       write output to file
|     -p, --password string     password for decryption
|     -P, --password-file string  read password from file
|   
+-- df
|   Usage: omni df [OPTION]... [FILE]... [flags]
|   Description: Show information about the file system on which each FILE resides,
or all file systems by default.

  -h, --human-readable  print sizes in human readable format (e.g., 1K 234M 2G)
  -i, --inodes          list inode information instead of block usage
  -B, --block-size=SIZE scale sizes by SIZE before printing them
      --total           produce a grand total
  -t, --type=TYPE       limit listing to file systems of type TYPE
  -x, --exclude-type=TYPE  exclude file systems of type TYPE
  -l, --local           limit listing to local file systems
  -P, --portability     use the POSIX output format
|   
|   Flags:
|     -B, --block-size int64    scale sizes by SIZE before printing them
|     -x, --exclude-type string  exclude file systems of type TYPE
|     -H, --human-readable      print sizes in human readable format
|     -i, --inodes              list inode information instead of block usage
|     -l, --local               limit listing to local file systems
|     -P, --portability         use the POSIX output format
|         --total               produce a grand total
|     -t, --type string         limit listing to file systems of type TYPE
|   
+-- diff
|   Usage: omni diff [OPTION]... FILE1 FILE2 [flags]
|   Description: Compare files line by line.

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
|   
|   Flags:
|     -q, --brief               report only when files differ
|         --color               colorize the output
|     -B, --ignore-blank-lines  ignore changes where lines are all blank
|     -i, --ignore-case         ignore case differences
|     -b, --ignore-space-change  ignore changes in amount of white space
|         --json                compare as JSON files
|     -r, --recursive           recursively compare subdirectories
|     -y, --side-by-side        output in two columns
|         --suppress-common-lines  do not output common lines in side-by-side
|     -u, --unified int         output NUM lines of unified context
|     -W, --width int           output at most NUM columns
|   
+-- dirname
|   Usage: omni dirname [path...]
|   Description: Output each NAME with its last non-slash component and trailing slashes removed.
|   
+-- dotenv
|   Usage: omni dotenv [OPTION]... [FILE]... [flags]
|   Description: Parse and display environment variables from .env files.

With no FILE, reads from .env in the current directory.

  -e, --export      output as export statements (for shell sourcing)
  -s, --shell TYPE  target shell (auto, bash, zsh, fish, powershell, cmd, nushell)
  -q, --quiet       suppress warnings
  -x, --expand      expand variables in values

The .env file format:
  # Comments start with #
  KEY=value
  KEY="quoted value"
  KEY='single quoted'
  export KEY=value    # export prefix is optional

Shell export formats:
  bash/zsh:    export KEY="value"
  powershell:  $env:KEY = "value"
  cmd:         set KEY=value
  fish:        set -gx KEY "value"
  nushell:     $env.KEY = "value"

Examples:
  omni dotenv                    # display vars from .env
  omni dotenv .env.local         # display vars from specific file
  omni dotenv -e                 # output as export statements (auto-detect shell)
  omni dotenv -e -s powershell   # output for PowerShell
  omni dotenv -e -s fish         # output for Fish shell

Load into shell:
  Bash/Zsh:    eval $(omni dotenv -e)
  PowerShell:  omni dotenv -e -s powershell | Invoke-Expression
  Fish:        omni dotenv -e -s fish | source
  CMD:         for /f "tokens=*" %i in ('omni dotenv -e -s cmd') do %i
|   
|   Flags:
|     -x, --expand              expand variables in values
|     -e, --export              output as export statements
|     -q, --quiet               suppress warnings
|     -s, --shell string        target shell (auto, bash, zsh, fish, powershell, cmd, nushell)
|   
+-- du
|   Usage: omni du [OPTION]... [FILE]... [flags]
|   Description: Summarize disk usage of each FILE, recursively for directories.

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
|   
|   Flags:
|     -a, --all                 write counts for all files, not just directories
|         --apparent-size       print apparent sizes, rather than disk usage
|     -B, --block-size int64    scale sizes by SIZE before printing them
|     -b, --bytes               equivalent to --apparent-size --block-size=1
|     -H, --human-readable      print sizes in human readable format
|     -d, --max-depth int       print total for directory only if N or fewer levels deep
|     -0, --null                end each output line with NUL, not newline
|     -x, --one-file-system     skip directories on different file systems
|     -s, --summarize           display only a total for each argument
|     -c, --total               produce a grand total
|   
+-- echo
|   Usage: omni echo [STRING]... [flags]
|   Description: Echo the STRING(s) to standard output.

Examples:
  omni echo Hello World     # outputs 'Hello World'
  omni echo -n "no newline" # outputs without trailing newline
  omni echo -e "tab\there"  # outputs with tab character
|   
|   Flags:
|     -e, --escape              enable interpretation of backslash escapes
|     -E, --no-escape           disable interpretation of backslash escapes (default)
|     -n, --no-newline          do not output the trailing newline
|   
+-- egrep
|   Usage: omni egrep [options] PATTERN [FILE...] [flags]
|   Description: Search for PATTERN in each FILE using extended regular expressions.
This is equivalent to 'grep -E'.
|   
|   Flags:
|     -C, --context int         print NUM lines of context
|     -c, --count               only print a count of matching lines
|     -l, --files-with-matches  only print FILE names with matches
|     -i, --ignore-case         ignore case distinctions
|     -v, --invert-match        select non-matching lines
|     -n, --line-number         prefix each line with line number
|     -o, --only-matching       show only matched parts
|     -q, --quiet               suppress all normal output
|   
+-- encrypt
|   Usage: omni encrypt [OPTION]... [FILE] [flags]
|   Description: Encrypt FILE or standard input using AES-256-GCM.

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
|   
|   Flags:
|     -a, --armor               ASCII armor (base64) output
|     -b, --base64              base64 output (same as -a)
|     -i, --iterations int      PBKDF2 iterations
|     -k, --key-file string     use key file for encryption
|     -o, --output string       write output to file
|     -p, --password string     password for encryption
|     -P, --password-file string  read password from file
|   
+-- env
|   Usage: omni env [NAME...] [flags]
|   Description: Print the values of the specified environment variables.
If no NAME is specified, print all environment variables.
|   
|   Flags:
|     -i, --ignore-environment  start with an empty environment
|     -0, --null                end each output line with NUL, not newline
|     -u, --unset string        remove variable from the environment
|   
+-- exec
|   Usage: omni exec <command> [args...] [flags]
|   Description: Safe wrapper for external commands that can't be reimplemented in Go.

Before executing, omni inspects the command to detect missing credentials
(registry tokens, cloud provider keys, kubeconfig, etc.) and warns you.

Examples:
  omni exec pnpm install
  omni exec --force docker build .
  omni exec --dry-run aws s3 ls
  omni exec --strict kubectl get pods
|   
|   Flags:
|         --dry-run             Show credential checks without executing
|     -f, --force               Skip credential checks, execute immediately
|         --no-prompt           Don't prompt, just warn and proceed
|         --strict              Abort if any credentials are missing (CI mode)
|   
+-- exist
|   Usage: omni exist
|   Description: Check existence of various targets with proper exit codes for scripting.

Exit status:
  0  target exists
  1  target not found

Examples:
  omni exist file go.mod              # check if regular file exists
  omni exist dir cmd                  # check if directory exists
  omni exist path go.mod              # check if any path exists
  omni exist command go               # check if command is in PATH
  omni exist env PATH                 # check if env var is set
  omni exist process 1234             # check if PID is running
  omni exist port 8080                # check if TCP port is listening
  omni exist -q file go.mod && echo yes  # quiet mode for scripting
|   
|   +-- command
|   |   Usage: omni exist command <name>
|   |   Description: Check if a command exists in PATH
|   |   
|   +-- dir
|   |   Usage: omni exist dir <path>
|   |   Description: Check if a directory exists
|   |   
|   +-- env
|   |   Usage: omni exist env <name>
|   |   Description: Check if an environment variable is set
|   |   
|   +-- file
|   |   Usage: omni exist file <path>
|   |   Description: Check if a regular file exists
|   |   
|   +-- path
|   |   Usage: omni exist path <path>
|   |   Description: Check if any path exists (file, dir, symlink)
|   |   
|   +-- port
|   |   Usage: omni exist port <number>
|   |   Description: Check if a TCP port is listening
|   |   
|   \-- process
|       Usage: omni exist process <name|pid>
|       Description: Check if a process is running
|       
+-- fgrep
|   Usage: omni fgrep [options] PATTERN [FILE...] [flags]
|   Description: Search for PATTERN in each FILE using fixed strings (no regex).
This is equivalent to 'grep -F'.
|   
|   Flags:
|     -C, --context int         print NUM lines of context
|     -c, --count               only print a count of matching lines
|     -l, --files-with-matches  only print FILE names with matches
|     -i, --ignore-case         ignore case distinctions
|     -v, --invert-match        select non-matching lines
|     -n, --line-number         prefix each line with line number
|     -o, --only-matching       show only matched parts
|     -q, --quiet               suppress all normal output
|   
+-- file
|   Usage: omni file [OPTION]... FILE... [flags]
|   Description: Determine the type of each FILE.

  -b, --brief           do not prepend filenames to output
  -i, --mime            output MIME type strings
  -h, --no-dereference  don't follow symlinks
  -F, --separator       use string as separator instead of ':'

Examples:
  omni file image.png          # PNG image data
  omni file -i document.pdf    # application/pdf
  omni file -b script.sh       # output type only
  omni file *                  # check all files
|   
|   Flags:
|     -b, --brief               do not prepend filenames
|     -i, --mime                output MIME type
|     -L, --no-dereference      don't follow symlinks
|     -F, --separator string    use string as separator
|   
+-- find
|   Usage: omni find [path...] [expression] [flags]
|   Description: Search for files in a directory hierarchy.

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
|   
|   Flags:
|         --amin string         access time [+-]N minutes
|         --atime string        access time [+-]N days
|         --empty               file is empty
|         --executable          file is executable
|         --iname string        case insensitive name match
|         --ipath string        case insensitive path match
|         --iregex string       case insensitive regex
|         --maxdepth int        maximum depth (0=unlimited)
|         --mindepth int        minimum depth
|         --mmin string         modification time [+-]N minutes
|         --mtime string        modification time [+-]N days
|         --name string         file name matches pattern
|         --not                 negate next test
|         --path string         path matches pattern
|     -0, --print0              print with null terminator
|         --readable            file is readable
|         --regex string        path matches regex
|         --size string         file size [+-]N[ckMG]
|         --type string         file type (f=file, d=dir, l=link)
|         --writable            file is writable
|   
+-- fold
|   Usage: omni fold [OPTION]... [FILE]... [flags]
|   Description: Wrap input lines in each FILE, writing to standard output.

With no FILE, or when FILE is -, read standard input.

  -w, --width=WIDTH  use WIDTH columns instead of 80
  -b, --bytes        count bytes rather than columns
  -s, --spaces       break at spaces

Examples:
  omni fold -w 40 file.txt     # wrap lines at 40 columns
  omni fold -s -w 72 README    # wrap at spaces, 72 columns
|   
|   Flags:
|     -b, --bytes               count bytes rather than columns
|     -s, --spaces              break at spaces
|     -w, --width int           use WIDTH columns instead of 80
|   
+-- for
|   Usage: omni for
|   Description: Loop over items and execute commands for each.

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
|   
|   +-- each
|   |   Usage: omni for each ITEM... -- COMMAND [flags]
|   |   Description: Loop over each item in the provided list.

Variable: $item or ${item}

Examples:
  omni for each apple banana cherry -- echo "Fruit: $item"
  omni for each *.go -- echo "File: $item"
  omni for each a b c --var=x -- echo "Value: $x"
|   |   
|   |   Flags:
|   |         --dry-run             print commands without executing
|   |         --var string          variable name to use
|   |   
|   +-- glob
|   |   Usage: omni for glob PATTERN -- COMMAND [flags]
|   |   Description: Loop over files matching a glob pattern.

Variable: $file or ${file}

Patterns:
  *.txt       All .txt files in current directory
  **/*.go     All .go files recursively
  src/*.js    All .js files in src/

Examples:
  omni for glob "*.txt" -- cat $file
  omni for glob "**/*.go" -- wc -l $file
  omni for glob "src/*.js" --dry-run -- echo $file
|   |   
|   |   Flags:
|   |         --dry-run             print commands without executing
|   |         --var string          variable name to use
|   |   
|   +-- lines
|   |   Usage: omni for lines [FILE] -- COMMAND [flags]
|   |   Description: Loop over each line from stdin or a file.

Variables:
  $line or ${line}   Current line content
  $n or ${n}         Current line number (1-based)

Examples:
  cat file.txt | omni for lines -- echo "Line $n: $line"
  omni for lines input.txt -- echo "$n: $line"
  omni for lines --var=x -- echo "Got: $x"
|   |   
|   |   Flags:
|   |         --dry-run             print commands without executing
|   |         --var string          variable name to use
|   |   
|   +-- range
|   |   Usage: omni for range START END [STEP] -- COMMAND [flags]
|   |   Description: Loop from START to END (inclusive) with optional STEP.

Variable: $i or ${i}

Examples:
  omni for range 1 5 -- echo $i
  omni for range 10 0 -2 -- echo $i
  omni for range 1 100 -- echo "Number: $i"
|   |   
|   |   Flags:
|   |         --dry-run             print commands without executing
|   |         --var string          variable name to use
|   |   
|   \-- split
|       Usage: omni for split DELIMITER INPUT -- COMMAND [flags]
|       Description: Split input by delimiter and loop over each item.

Variables:
  $item or ${item}   Current item
  $i or ${i}         Current index (0-based)

Examples:
  omni for split "," "a,b,c" -- echo "Item: $item"
  omni for split ":" "$PATH" -- echo "Dir: $item"
  omni for split "\\n" "$(cat file.txt)" -- process $item
|       
|       Flags:
|             --dry-run             print commands without executing
|             --var string          variable name to use
|       
+-- free
|   Usage: omni free [OPTION]... [flags]
|   Description: Display the total amount of free and used physical and swap memory
in the system, as well as the buffers and caches used by the kernel.

  -b, --bytes         show output in bytes
  -k, --kibibytes     show output in kibibytes (default)
  -m, --mebibytes     show output in mebibytes
  -g, --gibibytes     show output in gibibytes
  -h, --human         show human-readable output
  -w, --wide          wide output
  -t, --total         show total for RAM + swap
|   
|   Flags:
|     -b, --bytes               show output in bytes
|     -g, --gibibytes           show output in gibibytes
|     -H, --human               show human-readable output
|     -k, --kibibytes           show output in kibibytes
|     -m, --mebibytes           show output in mebibytes
|     -t, --total               show total for RAM + swap
|     -w, --wide                wide output
|   
+-- gbc
|   Usage: omni gbc [flags]
|   Description: Alias for 'git branch-clean'.
Delete merged branches.

Examples:
  omni gbc --dry-run
|   
|   Flags:
|         --dry-run             Show branches that would be deleted
|   
+-- gh
|   Usage: omni gh
|   Description: Convenience wrappers around common gh (GitHub CLI) operations.
|   
|   +-- actions-rerun
|   |   Usage: omni gh actions-rerun <run-id>
|   |   Description: Re-run a workflow run
|   |   
|   +-- issue-mine
|   |   Usage: omni gh issue-mine
|   |   Description: List issues assigned to you
|   |   
|   +-- pr-approve
|   |   Usage: omni gh pr-approve <number>
|   |   Description: Approve a pull request
|   |   
|   +-- pr-checkout
|   |   Usage: omni gh pr-checkout <number>
|   |   Description: Check out a pull request
|   |   
|   +-- pr-diff
|   |   Usage: omni gh pr-diff <number>
|   |   Description: Show diff for a pull request
|   |   
|   \-- repo-clone-org
|       Usage: omni gh repo-clone-org <org> [flags]
|       Description: Clone all repositories in an organization
|       
|       Flags:
|             --limit int           max repos to clone
|       
+-- git
|   Usage: omni git
|   Description: Git shortcut commands for common operations.
|   
|   +-- amend
|   |   Usage: omni git amend
|   |   Description: Amend the last commit without editing the message.
Equivalent to: git commit --amend --no-edit

Examples:
  omni git amend
|   |   
|   +-- blame-line
|   |   Usage: omni git blame-line <file> [flags]
|   |   Description: Show blame for a specific line range in a file.

Examples:
  omni git blame-line main.go --start 10 --end 20
|   |   
|   |   Flags:
|   |         --end int             End line number
|   |         --start int           Start line number
|   |   
|   +-- branch-clean
|   |   Usage: omni git branch-clean [flags]
|   |   Description: Delete local branches that have been merged into the current branch.
Skips main, master, and develop branches.

Examples:
  omni git branch-clean
  omni git bc --dry-run
|   |   
|   |   Flags:
|   |         --dry-run             Show branches that would be deleted
|   |   
|   +-- diff-words
|   |   Usage: omni git diff-words
|   |   Description: Show word-level diff instead of line-level.
Equivalent to: git diff --word-diff

Examples:
  omni git diff-words
  omni git diff-words HEAD~1
|   |   
|   +-- fetch-all
|   |   Usage: omni git fetch-all
|   |   Description: Fetch all remotes with prune.
Equivalent to: git fetch --all --prune

Examples:
  omni git fetch-all
  omni git fa
|   |   
|   +-- log-graph
|   |   Usage: omni git log-graph [flags]
|   |   Description: Show a pretty git log with graph visualization.
Equivalent to: git log --oneline --graph --decorate --all

Examples:
  omni git log-graph
  omni git lg -n 20
|   |   
|   |   Flags:
|   |     -n, --count int           Number of commits to show
|   |   
|   +-- pull-rebase
|   |   Usage: omni git pull-rebase
|   |   Description: Pull from remote with rebase.
Equivalent to: git pull --rebase

Examples:
  omni git pull-rebase
  omni git pr
|   |   
|   +-- push
|   |   Usage: omni git push [flags]
|   |   Description: Push to the remote repository.

Examples:
  omni git push
  omni git push --force
|   |   
|   |   Flags:
|   |         --force               Force push (with lease)
|   |   
|   +-- quick-commit
|   |   Usage: omni git quick-commit [flags]
|   |   Description: Stage all changes and commit with a message.
Equivalent to: git add -A && git commit -m "message"

Examples:
  omni git quick-commit -m "fix bug"
  omni git qc -m "add feature"
|   |   
|   |   Flags:
|   |     -a, --all                 Stage all changes before commit
|   |     -m, --message string      Commit message (required)
|   |   
|   +-- stash-staged
|   |   Usage: omni git stash-staged [flags]
|   |   Description: Stash only staged changes, leaving unstaged changes in the working directory.

Examples:
  omni git stash-staged
  omni git stash-staged -m "WIP: feature"
|   |   
|   |   Flags:
|   |     -m, --message string      Stash message
|   |   
|   +-- status
|   |   Usage: omni git status
|   |   Description: Show short git status.
Equivalent to: git status -sb

Examples:
  omni git status
  omni git st
|   |   
|   \-- undo
|       Usage: omni git undo
|       Description: Undo the last commit, keeping changes staged.
Equivalent to: git reset --soft HEAD~1

Examples:
  omni git undo
|       
+-- gops
|   Usage: omni gops [PID] [flags]
|   Description: Display information about running Go processes.

Uses google/gops to detect Go processes and show their version
and build information.

Examples:
  omni gops           # list all Go processes
  omni gops -j        # output as JSON
  omni gops 1234      # show info for specific PID
|   
|   Flags:
|     -j, --json                output as JSON
|   
+-- gqc
|   Usage: omni gqc [flags]
|   Description: Alias for 'git quick-commit'.
Stage all changes and commit with a message.

Examples:
  omni gqc -m "fix bug"
|   
|   Flags:
|     -a, --all                 Stage all changes before commit
|     -m, --message string      Commit message (required)
|   
+-- grep
|   Usage: omni grep [options] PATTERN [FILE...] [flags]
|   Description: Search for PATTERN in each FILE.
When FILE is '-', read standard input.
With no FILE, read '.' if recursive; otherwise, read standard input.
|   
|   Flags:
|     -A, --after-context int   print NUM lines of trailing context
|     -B, --before-context int  print NUM lines of leading context
|     -C, --context int         print NUM lines of output context
|     -c, --count               only print a count of matching lines per FILE
|     -E, --extended-regexp     interpret PATTERN as an extended regular expression
|     -l, --files-with-matches  only print FILE names containing matches
|     -L, --files-without-match  only print FILE names not containing matches
|     -F, --fixed-strings       interpret PATTERN as fixed strings
|     -i, --ignore-case         ignore case distinctions in patterns and data
|     -v, --invert-match        select non-matching lines
|     -n, --line-number         prefix each line of output with line number
|     -x, --line-regexp         match only whole lines
|     -m, --max-count int       stop after NUM matches
|         --no-filename         suppress the file name prefix on output
|     -o, --only-matching       show only nonempty parts of lines that match
|     -q, --quiet               suppress all normal output
|     -r, --recursive           search directories recursively
|     -H, --with-filename       print file name with output lines
|     -w, --word-regexp         match only whole words
|   
+-- gunzip
|   Usage: omni gunzip [OPTION]... [FILE]... [flags]
|   Description: Decompress FILEs in gzip format.

Equivalent to gzip -d.

Examples:
  omni gunzip file.txt.gz      # decompress
  omni gunzip -k file.txt.gz   # keep original
|   
|   Flags:
|     -f, --force               force overwrite
|     -k, --keep                keep original files
|     -c, --stdout              write to stdout
|     -v, --verbose             verbose mode
|   
+-- gzip
|   Usage: omni gzip [OPTION]... [FILE]... [flags]
|   Description: Compress or decompress FILEs using gzip format.

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
|   
|   Flags:
|     -9, --best int            compress better
|     -d, --decompress          decompress
|     -1, --fast int            compress faster
|     -f, --force               force overwrite
|     -k, --keep                keep original files
|     -c, --stdout              write to stdout
|     -v, --verbose             verbose mode
|   
+-- hash
|   Usage: omni hash [OPTION]... [FILE]... [flags]
|   Description: Print or check cryptographic hashes (checksums).

With no FILE, or when FILE is -, read standard input.

  -a, --algorithm ALG  hash algorithm: md5, sha1, sha256 (default), sha512, crc32, crc64
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
|   
|   Flags:
|     -a, --algorithm string    hash algorithm (md5, sha1, sha256, sha512, crc32, crc64)
|     -b, --binary              read in binary mode
|     -c, --check               read checksums from FILE and check them
|         --quiet               don't print OK for verified files
|     -r, --recursive           hash files recursively
|         --status              don't output anything, use status code
|     -w, --warn                warn about improperly formatted lines
|   
+-- head
|   Usage: omni head [option]... [file]... [flags]
|   Description: Print the first 10 lines of each FILE to standard output.
With more than one FILE, precede each with a header giving the file name.
With no FILE, or when FILE is -, read standard input.

Numeric shortcuts are supported: -80 is equivalent to -n 80.
|   
|   Flags:
|     -c, --bytes int           print the first NUM bytes of each file
|     -n, --lines int           print the first NUM lines instead of the first 10
|     -q, --quiet               never print headers giving file names
|     -v, --verbose             always print headers giving file names
|   
+-- help
|   Usage: omni help [command]
|   Description: Help provides help for any command in the application.
Simply type omni help [path to command] for full details.
|   
+-- hex
|   Usage: omni hex
|   Description: Hexadecimal encoding and decoding utilities.

Subcommands:
  encode    Encode text to hexadecimal
  decode    Decode hexadecimal to text

Examples:
  omni hex encode "hello"
  omni hex decode "68656c6c6f"
  echo "test" | omni hex encode
|   
|   +-- decode
|   |   Usage: omni hex decode [HEX]
|   |   Description: Decode hexadecimal string back to text.

Accepts hex strings with or without separators (spaces, colons, dashes).

Examples:
  omni hex decode "68656c6c6f"         # Output: hello
  omni hex decode "68:65:6c:6c:6f"     # With colons
  omni hex decode "68 65 6c 6c 6f"     # With spaces
  echo "74657374" | omni hex decode    # Read from stdin
|   |   
|   \-- encode
|       Usage: omni hex encode [TEXT] [flags]
|       Description: Encode text to hexadecimal representation.

Each byte is converted to its two-character hex representation.

Examples:
  omni hex encode "hello"              # Output: 68656c6c6f
  omni hex encode --upper "hello"      # Output: 68656C6C6F
  echo "test" | omni hex encode        # Read from stdin
  omni hex encode file.txt             # Read from file
|       
|       Flags:
|         -u, --upper               use uppercase hex letters
|       
+-- html
|   Usage: omni html
|   Description: HTML utilities for formatting, encoding, and decoding.

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
|   
|   +-- decode
|   |   Usage: omni html decode [TEXT]
|   |   Description: HTML decode text by unescaping HTML entities.

Converts HTML entities like &lt;, &gt;, &amp;, &quot; back to their original characters.

Examples:
  omni html decode "&lt;script&gt;"
  omni html decode "Tom &amp; Jerry"
  echo "&lt;div&gt;" | omni html decode
|   |   
|   +-- encode
|   |   Usage: omni html encode [TEXT]
|   |   Description: HTML encode text by escaping special characters.

Converts characters like <, >, &, ", and ' to their HTML entity equivalents.

Examples:
  omni html encode "<script>alert('xss')</script>"
  omni html encode "Tom & Jerry"
  echo "<div>" | omni html encode
|   |   
|   +-- fmt
|   |   Usage: omni html fmt [FILE] [flags]
|   |   Description: Format HTML with proper indentation.

  -i, --indent=STR     indentation string (default "  ")
  --sort-attrs         sort attributes alphabetically

Examples:
  omni html fmt file.html
  omni html fmt "<div><p>text</p></div>"
  cat file.html | omni html fmt
  omni html fmt --sort-attrs file.html
|   |   
|   |   Flags:
|   |     -i, --indent string       indentation string
|   |         --sort-attrs          sort attributes alphabetically
|   |   
|   +-- minify
|   |   Usage: omni html minify [FILE]
|   |   Description: Minify HTML by removing unnecessary whitespace and comments.

Examples:
  omni html minify file.html
  cat file.html | omni html minify
|   |   
|   \-- validate
|       Usage: omni html validate [FILE]
|       Description: Validate HTML syntax.

Exit codes:
  0  Valid HTML
  1  Invalid HTML or error

  --json    output result as JSON

Examples:
  omni html validate file.html
  omni html validate "<div><p>text</p></div>"
  omni html validate --json file.html
|       
+-- id
|   Usage: omni id [OPTION]... [USER] [flags]
|   Description: Print user and group information for the specified USER,
or (when USER omitted) for the current user.

  -g, --group   print only the effective group ID
  -G, --groups  print all group IDs
  -n, --name    print a name instead of a number, for -ugG
  -r, --real    print the real ID instead of the effective ID, with -ugG
  -u, --user    print only the effective user ID
|   
|   Flags:
|     -g, --group               print only the effective group ID
|     -G, --groups              print all group IDs
|     -n, --name                print a name instead of a number
|     -r, --real                print the real ID instead of the effective ID
|     -u, --user                print only the effective user ID
|   
+-- join
|   Usage: omni join [OPTION]... FILE1 FILE2 [flags]
|   Description: For each pair of input lines with identical join fields, write a line to
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
|   
|   Flags:
|     -1, --1 int               join on this FIELD of file 1
|     -2, --2 int               join on this FIELD of file 2
|     -a, --a int               also print unpairable lines from file FILENUM (1 or 2)
|     -e, --e string            replace missing fields with EMPTY
|     -i, --i                   ignore case when comparing fields
|     -t, --t string            use CHAR as input and output field separator
|     -v, --v int               print only unpairable lines from file FILENUM (1 or 2)
|   
+-- jq
|   Usage: omni jq [OPTION]... FILTER [FILE]... [flags]
|   Description: jq is a lightweight command-line JSON processor.

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
|   
|   Flags:
|     -c, --compact-output      compact output
|     -n, --null-input          don't read any input
|     -r, --raw-output          output raw strings
|     -s, --slurp               read all inputs into array
|     -S, --sort-keys           sort object keys
|         --tab                 use tabs for indentation
|   
+-- json
|   Usage: omni json
|   Description: JSON utilities for formatting, minifying, and validating JSON data.

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
|   
|   +-- fmt
|   |   Usage: omni json fmt [FILE]... [flags]
|   |   Description: Format JSON with proper indentation and line breaks.

  -i, --indent=STR   indentation string (default "  ")
  -t, --tab          use tabs for indentation
  -s, --sort-keys    sort object keys alphabetically
  -e, --escape-html  escape HTML characters (<, >, &)

Examples:
  omni json fmt file.json              # beautify with 2-space indent
  omni json fmt -t file.json           # use tabs
  omni json fmt -s file.json           # sort keys
  echo '{"b":2,"a":1}' | omni json fmt -s
|   |   
|   |   Flags:
|   |     -e, --escape-html         escape HTML characters
|   |     -i, --indent string       indentation string
|   |     -s, --sort-keys           sort object keys
|   |     -t, --tab                 use tabs for indentation
|   |   
|   +-- fromcsv
|   |   Usage: omni json fromcsv [FILE] [flags]
|   |   Description: Convert CSV data to JSON array of objects.

  --no-header          first row is data, not headers
  -d, --delimiter=STR  field delimiter (default ",")
  -a, --array          always output as array (even for single row)

Examples:
  omni json fromcsv file.csv
  cat file.csv | omni json fromcsv
  omni json fromcsv -d ";" file.csv    # semicolon delimiter
  omni json fromcsv --no-header file.csv
|   |   
|   |   Flags:
|   |     -a, --array               always output as array
|   |     -d, --delimiter string    field delimiter
|   |         --no-header           first row is data, not headers
|   |   
|   +-- fromtoml
|   |   Usage: omni json fromtoml [FILE] [flags]
|   |   Description: Convert TOML data to JSON format.

  -m, --minify    output minified JSON (no indentation)

Examples:
  omni json fromtoml file.toml
  cat file.toml | omni json fromtoml
  omni json fromtoml -m file.toml     # minified output
  omni json fromtoml file.toml > output.json
|   |   
|   |   Flags:
|   |     -m, --minify              output minified JSON
|   |   
|   +-- fromxml
|   |   Usage: omni json fromxml [FILE] [flags]
|   |   Description: Convert XML data to JSON format.

  --attr-prefix=STR    prefix for attributes in JSON (default "-")
  --text-key=STR       key for text content (default "#text")

Examples:
  omni json fromxml file.xml
  cat file.xml | omni json fromxml
  omni json fromxml --attr-prefix=@ file.xml
|   |   
|   |   Flags:
|   |         --attr-prefix string  prefix for attributes in JSON
|   |         --text-key string     key for text content
|   |   
|   +-- fromyaml
|   |   Usage: omni json fromyaml [FILE] [flags]
|   |   Description: Convert YAML data to JSON format.

  -m, --minify    output minified JSON (no indentation)

Examples:
  omni json fromyaml file.yaml
  cat file.yaml | omni json fromyaml
  omni json fromyaml -m file.yaml     # minified output
  omni json fromyaml file.yaml > output.json
|   |   
|   |   Flags:
|   |     -m, --minify              output minified JSON
|   |   
|   +-- keys
|   |   Usage: omni json keys [FILE]
|   |   Description: List all keys (paths) in a JSON object recursively.

Examples:
  omni json keys file.json
  echo '{"a":{"b":1}}' | omni json keys
|   |   
|   +-- minify
|   |   Usage: omni json minify [FILE]... [flags]
|   |   Description: Remove all unnecessary whitespace from JSON.

Examples:
  omni json minify file.json
  cat file.json | omni json minify
  omni json minify -s file.json        # also sort keys
|   |   
|   |   Flags:
|   |     -s, --sort-keys           sort object keys
|   |   
|   +-- stats
|   |   Usage: omni json stats [FILE]
|   |   Description: Display statistics about JSON data including type, depth, size, etc.

Examples:
  omni json stats file.json
  echo '[1,2,3]' | omni json stats
|   |   
|   +-- tocsv
|   |   Usage: omni json tocsv [FILE] [flags]
|   |   Description: Convert JSON array of objects to CSV format.

Nested objects are flattened with dot notation (e.g., address.city).

  --no-header          don't include header row
  -d, --delimiter=STR  field delimiter (default ",")
  --no-quotes          don't quote fields

Examples:
  omni json tocsv file.json
  echo '[{"name":"John","age":30}]' | omni json tocsv
  omni json tocsv -d ";" file.json     # semicolon delimiter
  omni json tocsv --no-header file.json
|   |   
|   |   Flags:
|   |     -d, --delimiter string    field delimiter
|   |         --no-header           don't include header row
|   |         --no-quotes           don't quote fields
|   |   
|   +-- tostruct
|   |   Usage: omni json tostruct [FILE] [flags]
|   |   Description: Convert JSON data to a Go struct definition.

  -n, --name=NAME      struct name (default "Root")
  -p, --package=PKG    package name (default "main")
  --inline             inline nested structs
  --omitempty          add omitempty to all fields

Examples:
  omni json tostruct file.json
  echo '{"name":"test","count":1}' | omni json tostruct
  omni json tostruct -n User -p models file.json
  omni json tostruct --omitempty file.json
|   |   
|   |   Flags:
|   |         --inline              inline nested structs
|   |     -n, --name string         struct name
|   |         --omitempty           add omitempty to all fields
|   |     -p, --package string      package name
|   |   
|   +-- toxml
|   |   Usage: omni json toxml [FILE] [flags]
|   |   Description: Convert JSON data to XML format.

  -r, --root=NAME      root element name (default "root")
  -i, --indent=STR     indentation string (default "  ")
  --item-tag=NAME      tag for array items (default "item")
  --attr-prefix=STR    prefix for attributes (default "-")

Examples:
  omni json toxml file.json
  echo '{"name":"John"}' | omni json toxml
  omni json toxml -r person file.json   # custom root
  omni json toxml --item-tag=entry file.json
|   |   
|   |   Flags:
|   |         --attr-prefix string  prefix for attributes
|   |     -i, --indent string       indentation string
|   |         --item-tag string     tag for array items
|   |     -r, --root string         root element name
|   |   
|   +-- toyaml
|   |   Usage: omni json toyaml [FILE]
|   |   Description: Convert JSON data to YAML format.

Examples:
  omni json toyaml file.json
  echo '{"name":"test"}' | omni json toyaml
  omni json toyaml file.json > output.yaml
|   |   
|   \-- validate
|       Usage: omni json validate [FILE]...
|       Description: Validate JSON syntax without outputting the data.

Exit codes:
  0  Valid JSON
  1  Invalid JSON or error

  --json    output result as JSON

Examples:
  omni json validate file.json
  omni json validate --json file.json
  echo '{"valid": true}' | omni json validate
|       
+-- jwt
|   Usage: omni jwt
|   Description: JWT (JSON Web Token) utilities.

Subcommands:
  decode    Decode and inspect a JWT token

Examples:
  omni jwt decode "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  echo $TOKEN | omni jwt decode
  omni jwt decode --header token.txt
|   
|   \-- decode
|       Usage: omni jwt decode [TOKEN] [flags]
|       Description: Decode and inspect a JWT token.

Displays the header and payload of a JWT token. Does NOT verify the signature
(use a proper JWT library for that). Useful for debugging and inspection.

Examples:
  omni jwt decode "eyJhbGciOiJIUzI1NiIs..."
  omni jwt decode --header "eyJhbGciOiJIUzI1NiIs..."
  omni jwt decode --payload "eyJhbGciOiJIUzI1NiIs..."
  omni jwt decode --json "eyJhbGciOiJIUzI1NiIs..."
  echo $TOKEN | omni jwt decode
|       
|       Flags:
|         -H, --header              show only header
|             --json                output as JSON
|         -p, --payload             show only payload
|             --raw                 output raw JSON without formatting
|       
+-- kconfig
|   Usage: omni kconfig
|   Description: Show current kubeconfig context and cluster info.

Examples:
  omni kconfig
|   
+-- kcs
|   Usage: omni kcs [context]
|   Description: Switch to a different kubectl context.
Without arguments, lists available contexts.

Examples:
  omni kcs              # list contexts
  omni kcs production   # switch to production
|   
+-- kdebug
|   Usage: omni kdebug <pod> [flags]
|   Description: Run an ephemeral debug container in a pod.
Equivalent to: kubectl debug -it <pod> --image=<image>

Examples:
  omni kdebug mypod
  omni kdebug mypod --image=nicolaka/netshoot
|   
|   Flags:
|         --image string        Debug container image (default: busybox)
|     -n, --namespace string    Namespace
|   
+-- kdp
|   Usage: omni kdp <selector> [flags]
|   Description: Delete pods by label selector.
Equivalent to: kubectl delete pods -l <selector>

Examples:
  omni kdp app=nginx
  omni kdp app=nginx -n default --force
|   
|   Flags:
|         --force               Force delete
|     -n, --namespace string    Namespace
|   
+-- kdrain
|   Usage: omni kdrain <node> [flags]
|   Description: Drain a node for maintenance.
Equivalent to: kubectl drain <node>

Examples:
  omni kdrain mynode
  omni kdrain mynode --ignore-daemonsets
|   
|   Flags:
|         --delete-emptydir     Delete emptydir data
|         --ignore-daemonsets   Ignore daemonsets
|   
+-- keb
|   Usage: omni keb <pod> [flags]
|   Description: Exec into a pod with bash (falls back to sh).
Equivalent to: kubectl exec -it <pod> -- /bin/bash

Examples:
  omni keb mypod
  omni keb mypod -n default
  omni keb mypod -c mycontainer
|   
|   Flags:
|     -c, --container string    Container name
|     -n, --namespace string    Namespace
|   
+-- kga
|   Usage: omni kga [flags]
|   Description: Get all common resources in a namespace.
Equivalent to: kubectl get pods,svc,deploy,... -o wide

Examples:
  omni kga
  omni kga -n kube-system
  omni kga -A
|   
|   Flags:
|     -A, --all-namespaces      All namespaces
|     -n, --namespace string    Namespace
|   
+-- kge
|   Usage: omni kge [flags]
|   Description: Get events sorted by last timestamp.
Equivalent to: kubectl get events --sort-by='.lastTimestamp'

Examples:
  omni kge
  omni kge -n kube-system
  omni kge -A
|   
|   Flags:
|     -A, --all-namespaces      All namespaces
|     -n, --namespace string    Namespace
|   
+-- kill
|   Usage: omni kill [OPTION]... PID... [flags]
|   Description: Send the specified signal to the specified processes.

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
|   
|   Flags:
|     -l, --list                list signal names
|     -s, --signal string       specify the signal to be sent
|     -v, --verbose             report successful signals
|   
+-- klf
|   Usage: omni klf <pod> [flags]
|   Description: Follow logs for a pod with timestamps.
Equivalent to: kubectl logs -f --timestamps <pod>

Examples:
  omni klf mypod
  omni klf mypod -n default -c mycontainer
  omni klf mypod --tail 100
|   
|   Flags:
|     -c, --container string    Container name
|     -n, --namespace string    Namespace
|         --tail int            Lines to show from end of logs
|   
+-- kns
|   Usage: omni kns [namespace]
|   Description: Switch the default namespace for the current context.
Without arguments, lists all namespaces.

Examples:
  omni kns           # list namespaces
  omni kns default   # switch to default namespace
|   
+-- kpf
|   Usage: omni kpf <pod|svc/name> <local:remote> [flags]
|   Description: Quick port forward to a pod or service.
Equivalent to: kubectl port-forward <target> <local>:<remote>

Examples:
  omni kpf mypod 8080:80
  omni kpf svc/myservice 3000:80 -n default
|   
|   Flags:
|     -n, --namespace string    Namespace
|   
+-- krr
|   Usage: omni krr <deployment> [flags]
|   Description: Restart a deployment using rollout restart.
Equivalent to: kubectl rollout restart deployment/<name>

Examples:
  omni krr mydeployment
  omni krr mydeployment -n default
|   
|   Flags:
|     -n, --namespace string    Namespace
|   
+-- krun
|   Usage: omni krun <name> --image=<image> [-- command] [flags]
|   Description: Run a one-off pod that auto-deletes after completion.
Equivalent to: kubectl run <name> --image=<image> --rm -it --restart=Never

Examples:
  omni krun test --image=busybox -- sh
  omni krun curl --image=curlimages/curl -- curl google.com
|   
|   Flags:
|         --image string        Container image (required)
|     -n, --namespace string    Namespace
|   
+-- kscale
|   Usage: omni kscale <deployment> <replicas> [flags]
|   Description: Scale a deployment to the specified number of replicas.

Examples:
  omni kscale mydeployment 3
  omni kscale mydeployment 0 -n default
|   
|   Flags:
|     -n, --namespace string    Namespace
|   
+-- ksuid
|   Usage: omni ksuid [OPTION]... [flags]
|   Description: Generate KSUIDs (K-Sortable Unique IDentifiers).

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
|   
|   Flags:
|     -n, --count int           generate N KSUIDs
|   
+-- ktn
|   Usage: omni ktn [flags]
|   Description: Show top nodes by resource usage.
Equivalent to: kubectl top nodes

Examples:
  omni ktn
  omni ktn --sort-by cpu
|   
|   Flags:
|         --sort-by string      Sort by (cpu or memory)
|   
+-- ktp
|   Usage: omni ktp [flags]
|   Description: Show top pods by resource usage.
Equivalent to: kubectl top pods

Examples:
  omni ktp
  omni ktp -n default
  omni ktp --sort-by cpu
|   
|   Flags:
|     -A, --all-namespaces      All namespaces
|     -n, --namespace string    Namespace
|         --sort-by string      Sort by (cpu or memory)
|   
+-- kubectl
|   Usage: omni kubectl
|   Description: Kubernetes command-line tool integrated into omni.

This is a full integration of kubectl, supporting all kubectl commands and flags.
You can use 'omni kubectl' or the shorter alias 'omni k'.

Examples:
  omni kubectl get pods
  omni k get pods -A
  omni k describe node mynode
  omni k logs -f mypod
  omni k exec -it mypod -- /bin/sh
  omni k apply -f manifest.yaml
|   
+-- kwp
|   Usage: omni kwp [flags]
|   Description: Watch pods continuously.
Equivalent to: kubectl get pods -w

Examples:
  omni kwp
  omni kwp -n default
  omni kwp -l app=nginx
|   
|   Flags:
|     -A, --all-namespaces      All namespaces
|     -n, --namespace string    Namespace
|     -l, --selector string     Label selector
|   
+-- less
|   Usage: omni less [OPTION]... [FILE] [flags]
|   Description: Display file contents one screen at a time with scrolling support.

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
|   
|   Flags:
|     -N, --LINE-NUMBERS        show line numbers
|     -S, --chop-long-lines     truncate long lines
|     -i, --ignore-case         case-insensitive search
|     -X, --no-init             don't clear screen on start
|     -F, --quit-if-one-screen  quit if content fits on one screen
|     -R, --raw-control-chars   show raw control characters
|   
+-- lint
|   Usage: omni lint [OPTION]... [FILE|DIR]... [flags]
|   Description: Lint Taskfiles for cross-platform portability.

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
|   
|   Flags:
|         --fix                 auto-fix issues where possible
|     -q, --quiet               only show errors, not warnings
|         --strict              enable strict mode (more warnings become errors)
|   
+-- ln
|   Usage: omni ln [OPTION]... TARGET LINK_NAME [flags]
|   Description: Create a link to TARGET with the name LINK_NAME.
Create hard links by default, symbolic links with --symbolic.

  -s, --symbolic     make symbolic links instead of hard links
  -f, --force        remove existing destination files
  -n, --no-dereference  treat LINK_NAME as a normal file if it is a symlink
  -v, --verbose      print name of each linked file
  -b, --backup       make a backup of each existing destination file
  -r, --relative     create symbolic links relative to link location
|   
|   Flags:
|     -b, --backup              make a backup of each existing destination file
|     -f, --force               remove existing destination files
|     -n, --no-dereference      treat LINK_NAME as a normal file if it is a symlink
|     -r, --relative            create symbolic links relative to link location
|     -s, --symbolic            make symbolic links instead of hard links
|     -v, --verbose             print name of each linked file
|   
+-- loc
|   Usage: omni loc [PATH]... [flags]
|   Description: Count lines of code, comments, and blanks by programming language.

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
|   
|   Flags:
|     -e, --exclude stringSlice  directories to exclude
|         --hidden              include hidden files
|   
+-- logger
|   Usage: omni logger [flags]
|   Description: Configure omni command logging by outputting shell export statements.

Usage with eval to set environment variables:
  eval "$(omni logger --path /path/to/omni.log)"

To disable logging:
  eval "$(omni logger --disable)"

To view all log files:
  omni logger --viewer

Environment variables set:
  OMNI_LOG_ENABLED - Set to "true" to enable logging
  OMNI_LOG_PATH    - Path to the log file
|   
|   Flags:
|     -d, --disable             Disable logging (unset environment variables)
|     -p, --path string         Path to the log file
|     -s, --status              Show current logging status
|     -v, --viewer              View all log files sorted by time
|   
+-- ls
|   Usage: omni ls [file...] [flags]
|   Description: List information about the FILEs (the current directory by default).
Sort entries alphabetically if none of -tSU is specified.
|   
|   Flags:
|     -a, --all                 do not ignore entries starting with .
|     -A, --almost-all          do not list implied . and ..
|     -F, --classify            append indicator (*/=>@|) to entries
|     -d, --directory           list directories themselves, not their contents
|     -H, --human-readable      with -l, print sizes in human readable format
|     -i, --inode               print the index number of each file
|     -l, --long                use a long listing format
|     -U, --no-sort             do not sort; list entries in directory order
|     -1, --one                 list one file per line
|     -R, --recursive           list subdirectories recursively
|     -r, --reverse             reverse order while sorting
|     -S, --size                sort by file size, largest first
|     -t, --time                sort by modification time, newest first
|   
+-- lsof
|   Usage: omni lsof [OPTIONS] [flags]
|   Description: List information about files and network connections opened by processes.

Shows network connections for all processes by default. Use filters to narrow down results.

Options:
  -p PID      show files for specific process ID
  -u USER     show files for specific user
  -i          show network connections only
  -i:PORT     show connections using specific port
  -c CMD      filter by command name prefix
  -t          show TCP connections only
  -U          show UDP connections only
  -4          show only IPv4
  -6          show only IPv6

Examples:
  omni lsof                    # show all network connections
  omni lsof -i                 # show network files only
  omni lsof -i:80              # show connections on port 80
  omni lsof -i:443 -t          # TCP connections on port 443
  omni lsof -p 1234            # show files for PID 1234
  omni lsof -u root            # show files for root user
  omni lsof -c nginx           # show files for nginx processes
  omni lsof -j                 # output as JSON
|   
|   Flags:
|     -c, --command string      filter by command name prefix
|     -e, --established         show only established connections
|     -4, --ipv4                show only IPv4
|     -6, --ipv6                show only IPv6
|     -l, --listen              show only listening sockets
|     -i, --network             show network connections only
|     -n, --no-headers          don't print headers
|     -p, --pid int             show files for specific process ID
|         --port int            filter by port number (use with -i)
|     -t, --tcp                 show TCP connections only
|     -U, --udp                 show UDP connections only
|     -u, --user string         show files for specific user
|   
+-- md5sum
|   Usage: omni md5sum [OPTION]... [FILE]... [flags]
|   Description: Print or check MD5 (128-bit) checksums.

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
|   
|   Flags:
|     -b, --binary              read in binary mode
|     -c, --check               read checksums from FILE and check them
|         --quiet               don't print OK for verified files
|         --status              don't output anything, use status code
|     -w, --warn                warn about improperly formatted lines
|   
+-- mkdir
|   Usage: omni mkdir [directory...] [flags]
|   Description: Create the DIRECTORY(ies), if they do not already exist.
|   
|   Flags:
|     -p, --parents             no error if existing, make parent directories as needed
|   
+-- more
|   Usage: omni more [OPTION]... [FILE] [flags]
|   Description: Display file contents one screen at a time.

more is a simpler pager than less - it's designed to show content
and quit when reaching the end.

Navigation:
  Space, Enter    Scroll down one page
  q               Quit

Examples:
  omni more file.txt
  omni more -n file.txt     # with line numbers
  cat file.txt | omni more  # from stdin
|   
|   Flags:
|     -n, --line-numbers        show line numbers
|   
+-- move
|   Usage: omni move
|   Description: move is an alias for the mv command.

Usage:
  omni move SOURCE DEST

See 'omni mv --help' for full options.
|   
+-- mv
|   Usage: omni mv [source...] [destination]
|   Description: Rename SOURCE to DEST, or move SOURCE(s) to DIRECTORY.
|   
+-- nanoid
|   Usage: omni nanoid [OPTION]... [flags]
|   Description: Generate NanoIDs - compact, URL-safe, unique string identifiers.

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
|   
|   Flags:
|     -a, --alphabet string     custom alphabet
|     -n, --count int           generate N NanoIDs
|     -l, --length int          length of NanoID
|   
+-- netstat
|   Usage: omni netstat [OPTIONS] [flags]
|   Description: Display network connections. Alias for ss command.

Shows information about network connections including state,
local/remote addresses, and optionally process information.

Examples:
  omni netstat -a         # show all connections
  omni netstat -l         # show listening sockets
  omni netstat -t         # show TCP connections
  omni netstat -p         # show process info
  omni netstat -an        # all connections, numeric
|   
|   Flags:
|     -a, --all                 display all sockets
|     -l, --listening           display listening sockets only
|     -n, --numeric             don't resolve service names
|     -p, --processes           show process using socket
|     -t, --tcp                 display TCP sockets
|     -u, --udp                 display UDP sockets
|   
+-- nl
|   Usage: omni nl [OPTION]... [FILE]... [flags]
|   Description: Write each FILE to standard output, with line numbers added.

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
|   
|   Flags:
|     -b, --body-numbering string  use STYLE for numbering body lines
|     -i, --line-increment int  line number increment
|     -n, --number-format string  insert line numbers according to FORMAT
|     -s, --number-separator string  add STRING after line number
|     -w, --number-width int    use N columns for line numbers
|     -v, --starting-line-number int  first line number
|   
+-- note
|   Usage: omni note [flags]
|   Description: Save quick notes to a JSON file in your user Documents directory.

By default notes are saved to:
  Windows: %USERPROFILE%\Documents\omni-notes.json
  Linux:   $HOME/Documents/omni-notes.json
  macOS:   $HOME/Documents/omni-notes.json

Examples:
  omni note add "buy milk"
  omni note add deploy production at 10pm
  omni note remove 1
  omni note remove 1770847088806767400
  omni note list
  omni note --list
  omni note --list --json
  omni note --file ./notes.json "local test note"
|   
|   Flags:
|     -n, --limit int           show only last N notes (used with --list)
|         --list                list saved notes instead of adding a new one
|   
|   +-- add
|   |   Usage: omni note add <TEXT...>
|   |   Description: Add a new note entry
|   |   
|   +-- list
|   |   Usage: omni note list [flags]
|   |   Description: List note entries
|   |   
|   |   Flags:
|   |     -n, --limit int           show only last N notes
|   |   
|   \-- remove
|       Usage: omni note remove <INDEX_OR_ID>
|       Description: Remove a note entry by index or ID
|       
+-- paste
|   Usage: omni paste [OPTION]... [FILE]... [flags]
|   Description: Write lines consisting of the sequentially corresponding lines from
each FILE, separated by TABs, to standard output.

With no FILE, or when FILE is -, read standard input.

  -d, --delimiters=LIST   reuse characters from LIST instead of TABs
  -s, --serial            paste one file at a time instead of in parallel
  -z, --zero-terminated   line delimiter is NUL, not newline
|   
|   Flags:
|     -d, --delimiters string   reuse characters from LIST instead of TABs
|     -s, --serial              paste one file at a time instead of in parallel
|     -z, --zero-terminated     line delimiter is NUL, not newline
|   
+-- path
|   Usage: omni path
|   Description: Clean, resolve, and manipulate file paths using OS-native conventions.
|   
|   +-- abs
|   |   Usage: omni path abs [path...]
|   |   Description: Abs returns an absolute representation of path. Relative paths are resolved from the current working directory.
|   |   
|   \-- clean
|       Usage: omni path clean [path...]
|       Description: Clean returns the shortest path name equivalent to path by purely lexical processing. It applies OS-native separators.
|       
+-- pgrep
|   Usage: omni pgrep [OPTIONS] PATTERN [flags]
|   Description: List processes matching a pattern. Alias for pkill -l.

Options:
  -x          match process name exactly
  -f          match against full command line
  -n          select only the newest matching process
  -o          select only the oldest matching process
  -c          count matching processes
  -u USER     match only processes owned by USER
  -P PID      match only processes with parent PID
  -i          case insensitive matching

Examples:
  omni pgrep firefox           # find firefox processes
  omni pgrep -f "node server"  # match full command line
  omni pgrep -c python         # count python processes
|   
|   Flags:
|     -c, --count               count matching processes
|     -x, --exact               match exactly
|     -f, --full                match against full command line
|     -i, --ignore-case         case insensitive matching
|     -n, --newest              select only the newest process
|     -o, --oldest              select only the oldest process
|     -P, --parent int          match only processes with parent PID
|     -u, --user string         match only processes owned by user
|   
+-- pipe
|   Usage: omni pipe {CMD}, {CMD}, ... | CMD | CMD [flags]
|   Description: Chain multiple omni commands together, passing output from one to the next.

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
|   
|   Flags:
|     -s, --sep string          command separator
|         --var string          variable name for output substitution (default: OUT)
|     -v, --verbose             show intermediate results
|   
+-- pipeline
|   Usage: omni pipeline STAGE [STAGE...] [flags]
|   Description: Streaming text processing engine with built-in transform stages.

Each stage is a quoted string describing a transform. Stages are connected
via io.Pipe goroutines for memory-efficient, line-by-line processing.

Available stages:
  grep PATTERN       Filter lines matching regex pattern (-i, -v)
  grep-v PATTERN     Filter lines NOT matching pattern
  contains SUBSTR    Filter lines containing literal substring (-i)
  replace OLD NEW    Replace all occurrences of OLD with NEW
  head [N]           Output first N lines (default 10)
  take [N]           Alias for head
  tail [N]           Output last N lines (default 10)
  skip N             Skip first N lines
  sort               Sort lines (-r reverse, -n numeric, -rn both)
  uniq               Remove consecutive duplicate lines (-i)
  cut -dDELIM -fN    Extract fields (-d delimiter, -f fields)
  tr FROM TO         Translate characters
  sed s/PAT/REPL/g   Regex substitution
  rev                Reverse each line
  nl                 Number each line
  tee FILE           Copy output to file and next stage
  tac                Reverse line order
  wc                 Count lines/words/chars (-l, -w, -c)

Examples:
  omni pipeline 'grep error' 'sort' 'uniq' 'head 10' < log.txt
  omni pipeline -f access.log 'grep 404' 'cut -d" " -f1' 'sort' 'uniq'
  omni pipeline -v 'grep -i warning' 'sort -rn' 'head 5'
|   
|   Flags:
|     -f, --file string         input file (default: stdin)
|     -v, --verbose             show stage names before processing
|   
+-- pkill
|   Usage: omni pkill [OPTIONS] PATTERN [flags]
|   Description: Send a signal to processes matching a pattern.

By default, sends SIGTERM to matching processes. Use -l to list
matching processes without killing them (pgrep behavior).

Options:
  -signal     signal to send (default: TERM)
  -x          match process name exactly
  -f          match against full command line
  -n          select only the newest matching process
  -o          select only the oldest matching process
  -c          count matching processes
  -l          list matching processes (pgrep mode)
  -u USER     match only processes owned by USER
  -P PID      match only processes with parent PID
  -i          case insensitive matching

Examples:
  omni pkill firefox           # kill all firefox processes
  omni pkill -9 chrome         # send SIGKILL to chrome
  omni pkill -l python         # list python processes (pgrep)
  omni pkill -f "node server"  # match full command line
  omni pkill -n -l java        # show newest java process
  omni pkill -u root httpd     # kill httpd owned by root
  omni pkill -c nginx          # count nginx processes
|   
|   Flags:
|     -c, --count               count matching processes
|     -x, --exact               match exactly
|     -f, --full                match against full command line
|     -i, --ignore-case         case insensitive matching
|     -l, --list                list matching processes (pgrep mode)
|     -n, --newest              select only the newest process
|     -o, --oldest              select only the oldest process
|     -P, --parent int          match only processes with parent PID
|     -s, --signal string       signal to send (default: TERM)
|     -u, --user string         match only processes owned by user
|     -v, --verbose             verbose output
|   
+-- printf
|   Usage: omni printf FORMAT [ARG...] [flags]
|   Description: Format and print data using printf-style format specifiers.

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
|   
|   Flags:
|     -n, --no-newline          do not append a trailing newline
|   
+-- project
|   Usage: omni project
|   Description: Project analyzer that inspects any codebase directory and outputs
a structured report covering project type detection, language identification,
dependency parsing, git status, and documentation checks.

Subcommands:
  info     Full project overview (type, deps, git, docs)
  deps     Dependency analysis only
  docs     Documentation status check
  git      Git repository info
  health   Health score (0-100) with grade

Examples:
  omni project info
  omni project info --json
  omni project deps /path/to/project
  omni project health --markdown
  omni project git
|   
|   +-- deps
|   |   Usage: omni project deps [path]
|   |   Description: Analyze project dependencies. Supports Go (go.mod), Node.js (package.json),
Python (requirements.txt, pyproject.toml), Rust (Cargo.toml), Java (pom.xml,
build.gradle), Ruby (Gemfile), PHP (composer.json), and .NET (*.csproj).

Examples:
  omni project deps
  omni project deps --json
  omni project deps /path/to/project
|   |   
|   +-- docs
|   |   Usage: omni project docs [path]
|   |   Description: Check documentation status of the project: README, LICENSE, CHANGELOG,
CONTRIBUTING, CI/CD configs, linter configs, and more.

Examples:
  omni project docs
  omni project docs --json
  omni project docs /path/to/project
|   |   
|   +-- git
|   |   Usage: omni project git [path]
|   |   Description: Show git repository information including branch, remote, status,
recent commits, tags, and contributor count.

Examples:
  omni project git
  omni project git --json
  omni project git -n 20
|   |   
|   +-- health
|   |   Usage: omni project health [path]
|   |   Description: Compute a project health score from 0 to 100 based on presence of
README, LICENSE, CI/CD, tests, linter config, and other best practices.

Grade scale: A (90+), B (80+), C (70+), D (60+), F (<60)

Examples:
  omni project health
  omni project health --json
  omni project health --markdown
|   |   
|   \-- info
|       Usage: omni project info [path]
|       Description: Analyze the project directory and show a complete overview including
project type, languages, dependencies, git info, and documentation status.

Examples:
  omni project info
  omni project info --json
  omni project info --markdown
  omni project info /path/to/project
|       
+-- ps
|   Usage: omni ps [OPTION]... [flags]
|   Description: Display information about active processes.

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
|   
|   Flags:
|     -a, --all                 show processes for all users
|     -f, --full                full-format listing
|         --go                  show only Go processes
|     -l, --long                long format
|         --no-headers          don't print header line
|     -p, --pid int             show process with specified PID
|         --sort string         sort by column (pid, cpu, mem, time)
|     -u, --user string         show processes for specified user
|   
+-- pwd
|   Usage: omni pwd
|   Description: Print the full filename of the current working directory.
|   
+-- random
|   Usage: omni random [OPTION]... [flags]
|   Description: Generate random numbers, strings, or bytes using crypto/rand.

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
|   
|   Flags:
|     -c, --charset string      custom character set
|     -n, --count int           number of values to generate
|     -l, --length int          length of random strings
|         --max int64           maximum value for integers
|         --min int64           minimum value for integers
|     -s, --separator string    separator between values
|     -t, --type string         type: int, float, string, alpha, hex, password, bytes, custom
|   
+-- readlink
|   Usage: omni readlink [OPTION]... FILE... [flags]
|   Description: Print value of a symbolic link or canonical file name.

  -f, --canonicalize            canonicalize by following every symlink
  -e, --canonicalize-existing   canonicalize, all components must exist
  -m, --canonicalize-missing    canonicalize without requirements on existence
  -n, --no-newline              do not output the trailing delimiter
  -q, --quiet                   suppress most error messages
  -s, --silent                  suppress most error messages
  -v, --verbose                 report error messages
  -z, --zero                    end each output line with NUL, not newline
|   
|   Flags:
|     -f, --canonicalize        canonicalize by following every symlink
|     -e, --canonicalize-existing  canonicalize, all components must exist
|     -m, --canonicalize-missing  canonicalize without requirements on existence
|     -n, --no-newline          do not output the trailing delimiter
|     -q, --quiet               suppress most error messages
|     -s, --silent              suppress most error messages
|     -v, --verbose             report error messages
|     -z, --zero                end each output line with NUL, not newline
|   
+-- realpath
|   Usage: omni realpath [path...]
|   Description: Print the resolved absolute file name; all components must exist.
|   
+-- remove
|   Usage: omni remove
|   Description: remove is an alias for the rm command.

Usage:
  omni remove FILE [FILE...]

See 'omni rm --help' for full options.
|   
+-- repo
|   Usage: omni repo
|   Description: Repository analysis tools for generating structured context
about any codebase, optimized for LLM consumption.

Subcommands:
  analyze   Generate comprehensive repository context

Examples:
  omni repo analyze .
  omni repo analyze /path/to/project
  omni repo analyze github.com/owner/repo
  omni repo analyze . --compact
  omni repo analyze . --json
  omni repo analyze . -o context.md
  omni repo analyze . --sections=tree,deps,keys
|   
|   \-- analyze
|       Usage: omni repo analyze [path|url] [flags]
|       Description: Analyze a repository and produce structured Markdown or JSON context
optimized for LLM consumption. Includes directory tree, key file contents,
dependencies, entry points, architecture patterns, API surface, git info,
test patterns, and CI/CD configuration.

Supports local paths and remote GitHub repositories (cloned to temp dir).

Sections: overview, tree, keys, deps, api, git, tests, ci

Examples:
  omni repo analyze .                          # Local
  omni repo analyze /path/to/project           # Local absolute
  omni repo analyze github.com/owner/repo      # Remote (clones to temp)
  omni repo analyze . --compact                # Shorter output
  omni repo analyze . --json                   # JSON format
  omni repo analyze . -o context.md            # Write to file
  omni repo analyze . --sections=tree,deps     # Specific sections only
|       
|       Flags:
|             --compact             shorter output for smaller context windows
|         -o, --output string       write output to file
|             --sections string     comma-separated section filter (overview,tree,keys,deps,api,git,tests,ci)
|       
+-- rev
|   Usage: omni rev [FILE]...
|   Description: Reverse the characters in each line of FILE(s) or standard input.

Examples:
  echo "hello" | omni rev     # olleh
  omni rev file.txt           # reverse each line in file
|   
+-- rg
|   Usage: omni rg [OPTIONS] PATTERN [PATH...] [flags]
|   Description: Recursively search current directory for a regex pattern.

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
|   
|   Flags:
|     -A, --after-context int   show N lines after match
|     -B, --before-context int  show N lines before match
|     -b, --byte-offset         show byte offset of each line (not yet implemented)
|         --color string        when to use colors: auto, always, never
|         --colors stringSlice  custom color specification (e.g., 'path:fg:magenta')
|         --column              show column numbers
|     -C, --context int         show N lines before and after match
|     -c, --count               only show count of matches per file
|     -l, --files-with-matches  only show file names with matches
|     -F, --fixed-strings       treat pattern as literal string
|     -L, --follow              follow symbolic links
|     -g, --glob stringSlice    include/exclude files matching GLOB (prefix with ! to exclude)
|         --hidden              search hidden files and directories
|     -i, --ignore-case         case insensitive search
|     -v, --invert-match        show non-matching lines
|         --json-stream         output results as streaming NDJSON (one JSON object per line)
|     -n, --line-number         show line numbers
|     -m, --max-count int       limit matches per file
|         --max-depth int       limit directory traversal depth
|     -U, --multiline           enable multiline matching
|     -H, --no-heading          don't group matches by file name
|         --no-ignore           don't respect gitignore files
|     -o, --only-matching       show only matching part of line
|         --passthru            show all lines, highlighting matches
|     -q, --quiet               quiet mode, exit on first match
|     -r, --replace string      replace matches with STRING
|     -S, --smart-case          smart case (insensitive if pattern is all lowercase)
|         --stats               show search statistics
|     -j, --threads int         number of worker threads (default: CPU count)
|         --trim                trim leading/trailing whitespace from each line
|     -t, --type stringSlice    only search files of TYPE (go, js, py, etc.)
|     -T, --type-not stringSlice  exclude files of TYPE
|     -w, --word-regexp         only match whole words
|   
+-- rm
|   Usage: omni rm [file...] [flags]
|   Description: Remove the FILE(s).

Protected paths (system directories, SSH keys, credentials, etc.) cannot be
deleted without explicit override flags. Use --force for non-critical
protected paths, or --no-preserve-root for critical system paths.
|   
|   Flags:
|     -f, --force               ignore nonexistent files and arguments, never prompt
|         --no-preserve-root    do not treat protected paths specially (dangerous)
|     -r, --recursive           remove directories and their contents recursively
|   
+-- rmdir
|   Usage: omni rmdir [directory...] [flags]
|   Description: Remove the DIRECTORY(ies), if they are empty.

Protected paths (system directories, credentials, etc.) cannot be
deleted without the --no-preserve-root flag.
|   
|   Flags:
|         --no-preserve-root    do not treat protected paths specially (dangerous)
|   
+-- scaffold
|   Usage: omni scaffold
|   Description: scaffold provides code generation utilities for scaffolding applications.

Subcommands:
  cobra init    Initialize a new Cobra CLI application
  cobra add     Add a new command to an existing Cobra application
  cobra config  Manage cobra generator configuration
  handler       Generate HTTP handler
  repository    Generate database repository
  test          Generate tests for a Go source file
  mcp           Generate MCP server with tools, resources, and debug logging

Configuration:
  Default values can be set in ~/.cobra.yaml (compatible with cobra-cli).
  Command-line flags override config file values.

Examples:
  omni scaffold cobra init myapp --module github.com/user/myapp
  omni scaffold cobra add serve --parent root
  omni scaffold cobra config --show
  omni scaffold handler user --method GET,POST --framework chi
  omni scaffold repository user --entity User --table users
  omni scaffold test internal/cli/foo/foo.go
  omni scaffold mcp myserver --transport sse --addr :9090
|   
|   +-- cobra
|   |   Usage: omni scaffold cobra
|   |   Description: Generate Cobra CLI applications and commands.

Configuration file (~/.cobra.yaml):
  author: Your Name <email@example.com>
  license: MIT
  useViper: true
  useService: false
  full: false

Subcommands:
  init       Initialize a new Cobra CLI application
  add        Add a new command to an existing application
  add-tools  Add cmdtree and aicontext to an existing project
  config     Manage generator configuration
|   |   
|   |   +-- add
|   |   |   Usage: omni scaffold cobra add <command-name> [flags]
|   |   |   Description: Add a new command to an existing Cobra CLI application.

Creates a new command file in the cmd/{appName}/ directory with cmd_ prefix.

Examples:
  omni scaffold cobra add serve
  omni scaffold cobra add serve --parent root
  omni scaffold cobra add list --parent user --description "List all users"
|   |   |   
|   |   |   Flags:
|   |   |     -d, --description string  command description
|   |   |         --dir string          project directory (defaults to current directory)
|   |   |     -p, --parent string       parent command
|   |   |   
|   |   +-- add-tools
|   |   |   Usage: omni scaffold cobra add-tools [flags]
|   |   |   Description: Add cmdtree and aicontext commands to an existing Cobra CLI project.

Creates:
  - cmd/{appName}/cmd_cmdtree.go   Command tree utility

With --aicontext:
  - cmd/{appName}/cmd_aicontext.go AI context generator for coding agents

Examples:
  omni scaffold cobra add-tools
  omni scaffold cobra add-tools --aicontext
  omni scaffold cobra add-tools --dir /path/to/project
|   |   |   
|   |   |   Flags:
|   |   |         --aicontext           include aicontext command for AI coding agents
|   |   |         --dir string          project directory (defaults to current directory)
|   |   |   
|   |   +-- config
|   |   |   Usage: omni scaffold cobra config [flags]
|   |   |   Description: Manage the cobra generator configuration file.

The configuration file is stored at ~/.cobra.yaml and is compatible
with cobra-cli's configuration format.

Available options in config file:
  author: Your Name <email@example.com>
  license: MIT | Apache-2.0 | BSD-3
  useViper: true | false
  useService: true | false
  full: true | false

Examples:
  omni scaffold cobra config --show
  omni scaffold cobra config --init
  omni scaffold cobra config --init --author "John Doe" --license MIT
|   |   |   
|   |   |   Flags:
|   |   |     -a, --author string       author name for config
|   |   |         --full                set full in config
|   |   |         --init                create a new configuration file
|   |   |     -l, --license string      license type for config
|   |   |         --service             set useService in config
|   |   |         --show                show current configuration
|   |   |         --viper               set useViper in config
|   |   |   
|   |   \-- init
|   |       Usage: omni scaffold cobra init <directory> [flags]
|   |       Description: Initialize a new Cobra CLI application with all necessary scaffolding.

Configuration:
  Reads defaults from ~/.cobra.yaml if present.
  Command-line flags override config file values.

Creates (basic mode):
  - cmd/{name}/{name}.go       Entry point + root command
  - cmd/{name}/cmd_version.go  Version command
  - cmd/{name}/cmd_cmdtree.go  Command tree utility
  - go.mod                     Go module
  - README.md                  Documentation
  - Taskfile.yml               Task runner
  - .gitignore                 Git ignore file
  - .editorconfig              Editor configuration
  - LICENSE                    License file (optional)

With --aicontext:
  - cmd/{name}/cmd_aicontext.go AI context generator for coding agents

With --viper:
  - internal/config/config.go  Viper configuration

With --service:
  - internal/parameters/config.go  Service parameters
  - internal/service/service.go    Service handler (uses inovacc/config)

With --full (includes all above plus):
  - .goreleaser.yaml               GoReleaser configuration
  - .golangci.yml                  GolangCI-Lint configuration (v2)
  - tools.go                       Build tool dependencies
  - .github/workflows/build.yml    GitHub Actions build workflow
  - .github/workflows/test.yml     GitHub Actions test workflow
  - .github/workflows/release.yaml GitHub Actions release workflow

Examples:
  omni scaffold cobra init myapp --module github.com/user/myapp
  omni scaffold cobra init ./apps/cli --module github.com/user/cli --viper
  omni scaffold cobra init myapp --module github.com/user/myapp --license MIT --author "John Doe"
  omni scaffold cobra init myapp --module github.com/user/myapp --full --service
  omni scaffold cobra init myapp --module github.com/user/myapp --aicontext
|   |       
|   |       Flags:
|   |             --aicontext           include aicontext command for AI coding agents
|   |         -a, --author string       author name
|   |         -d, --description string  application description
|   |             --full                full project with goreleaser, workflows, etc.
|   |         -l, --license string      license type (MIT, Apache-2.0, BSD-3)
|   |         -m, --module string       Go module path (required)
|   |         -n, --name string         application name (defaults to directory name)
|   |             --service             include service pattern with inovacc/config
|   |             --viper               include viper for configuration
|   |       
|   +-- handler
|   |   Usage: omni scaffold handler <name> [flags]
|   |   Description: Generate an HTTP handler with the specified name.

Supports multiple frameworks: stdlib, chi, gin, echo

  -p, --package      Package name (default: "handler")
  -d, --dir          Output directory (default: "internal/handler")
  -m, --method       HTTP methods: GET,POST,PUT,DELETE,PATCH (default: GET,POST,PUT,DELETE)
  --path             URL path pattern
  --middleware       Include middleware support
  -f, --framework    Framework: stdlib, chi, gin, echo (default: stdlib)

Examples:
  omni scaffold handler user
  omni scaffold handler user --method GET,POST --framework chi
  omni scaffold handler user --dir handlers --package handlers
  omni scaffold handler product --middleware --framework gin
|   |   
|   |   Flags:
|   |     -d, --dir string          output directory
|   |     -f, --framework string    framework: stdlib, chi, gin, echo
|   |     -m, --method string       HTTP methods (comma-separated)
|   |         --middleware          include middleware support
|   |     -p, --package string      package name
|   |         --path string         URL path pattern
|   |   
|   +-- mcp
|   |   Usage: omni scaffold mcp <name> [flags]
|   |   Description: Generate an MCP (Model Context Protocol) server with tools, resources, and debug logging.

Generates:
  - internal/mcp/server.go       Server setup with transport selection
  - internal/mcp/tools.go        Example greet tool
  - internal/mcp/resources.go    Example info resource
  - internal/mcp/debug.go        Logger with debug/trace levels
  - cmd/{appName}/cmd_mcp.go     Cobra command with mcp serve

Transports:
  stdio        Standard I/O (default, for CLI integration)
  sse          Server-Sent Events (legacy, HTTP-based)
  http-stream  Streamable HTTP (recommended for remote servers)

Examples:
  omni scaffold mcp myserver
  omni scaffold mcp myserver --transport sse --addr :9090
  omni scaffold mcp myserver --transport http-stream
  omni scaffold mcp myserver --module github.com/user/myapp
|   |   
|   |   Flags:
|   |         --addr string         listen address (for sse/http-stream)
|   |     -m, --module string       Go module path (auto-detected from go.mod)
|   |         --transport string    transport type: stdio, sse, http-stream
|   |   
|   +-- repository
|   |   Usage: omni scaffold repository <name> [flags]
|   |   Description: Generate a database repository with the specified name.

  -p, --package      Package name (default: "repository")
  -d, --dir          Output directory (default: "internal/repository")
  -e, --entity       Entity struct name (default: capitalized name)
  -t, --table        Database table name (default: lowercase name + "s")
  --db               Database type: postgres, mysql, sqlite (default: postgres)
  --interface        Generate interface (default: true)

Examples:
  omni scaffold repository user
  omni scaffold repository user --entity User --table users
  omni scaffold repository product --db mysql
  omni scaffold repository order --interface=false
|   |   
|   |   Flags:
|   |         --db string           database type: postgres, mysql, sqlite
|   |     -d, --dir string          output directory
|   |     -e, --entity string       entity struct name
|   |         --interface           generate interface
|   |     -p, --package string      package name
|   |     -t, --table string        database table name
|   |   
|   \-- test
|       Usage: omni scaffold test <file.go> [flags]
|       Description: Generate test stubs for exported functions in a Go source file.

Parses the input file and generates test functions for all exported
functions and methods.

  --table           Generate table-driven tests (default: true)
  --parallel        Add t.Parallel() calls
  --mock            Generate mock setup
  --benchmark       Include benchmark tests
  --fuzz            Include fuzz tests (Go 1.18+)

Examples:
  omni scaffold test internal/cli/foo/foo.go
  omni scaffold test pkg/service/user.go --parallel
  omni scaffold test handler.go --table=false
  omni scaffold test service.go --benchmark --mock
|       
|       Flags:
|             --benchmark           include benchmark tests
|             --fuzz                include fuzz tests
|             --mock                generate mock setup
|             --parallel            add t.Parallel() calls
|             --table               generate table-driven tests
|       
+-- sed
|   Usage: omni sed [OPTION]... {script} [FILE]... [flags]
|   Description: Sed is a stream editor. A stream editor is used to perform basic text
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
|   
|   Flags:
|     -e, --expression stringSlice  add the script to the commands to be executed
|     -i, --in-place            edit files in place
|         --in-place-suffix string  backup suffix for in-place edit
|     -n, --quiet               suppress automatic printing of pattern space
|     -r, --r                   use extended regular expressions (alias)
|     -E, --regexp-extended     use extended regular expressions
|   
+-- seq
|   Usage: omni seq [OPTION]... LAST or seq [OPTION]... FIRST LAST or seq [OPTION]... FIRST INCREMENT LAST [flags]
|   Description: Print numbers from FIRST to LAST, in steps of INCREMENT.

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
|   
|   Flags:
|     -w, --equal-width         equalize width with leading zeros
|     -f, --format string       use printf style FORMAT
|     -s, --separator string    use STRING to separate numbers
|   
+-- sha256sum
|   Usage: omni sha256sum [OPTION]... [FILE]... [flags]
|   Description: Print or check SHA256 (256-bit) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Examples:
  omni sha256sum file.txt           # compute hash
  omni sha256sum -c checksums.txt   # verify checksums
|   
|   Flags:
|     -b, --binary              read in binary mode
|     -c, --check               read checksums from FILE and check them
|         --quiet               don't print OK for verified files
|         --status              don't output anything, use status code
|     -w, --warn                warn about improperly formatted lines
|   
+-- sha512sum
|   Usage: omni sha512sum [OPTION]... [FILE]... [flags]
|   Description: Print or check SHA512 (512-bit) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Examples:
  omni sha512sum file.txt           # compute hash
  omni sha512sum -c checksums.txt   # verify checksums
|   
|   Flags:
|     -b, --binary              read in binary mode
|     -c, --check               read checksums from FILE and check them
|         --quiet               don't print OK for verified files
|         --status              don't output anything, use status code
|     -w, --warn                warn about improperly formatted lines
|   
+-- shuf
|   Usage: omni shuf [OPTION]... [FILE] [flags]
|   Description: Write a random permutation of the input lines to standard output.

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
|   
|   Flags:
|     -e, --echo                treat each ARG as an input line
|     -n, --head-count int      output at most COUNT lines
|     -i, --input-range string  treat each number LO through HI as an input line
|     -r, --repeat              output lines can be repeated
|     -z, --zero-terminated     line delimiter is NUL
|   
+-- sleep
|   Usage: omni sleep NUMBER[SUFFIX]...
|   Description: Pause for NUMBER seconds. SUFFIX may be:
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
|   
+-- snowflake
|   Usage: omni snowflake [OPTION]... [flags]
|   Description: Generate Snowflake IDs - distributed, time-sortable unique identifiers.

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
|   
|   Flags:
|     -n, --count int           generate N Snowflake IDs
|     -w, --worker int64        worker ID (0-1023)
|   
+-- sort
|   Usage: omni sort [option]... [file]... [flags]
|   Description: Write sorted concatenation of all FILE(s) to standard output.

With no FILE, or when FILE is -, read standard input.
|   
|   Flags:
|     -c, --check               check for sorted input; do not sort
|     -d, --dictionary-order    consider only blanks and alphanumeric characters
|     -t, --field-separator string  use SEP instead of non-blank to blank transition
|     -f, --ignore-case         fold lower case to upper case characters
|     -b, --ignore-leading-blanks  ignore leading blanks
|     -k, --key string          sort via a key
|     -n, --numeric-sort        compare according to string numerical value
|     -o, --output string       write result to FILE instead of standard output
|     -r, --reverse             reverse the result of comparisons
|     -s, --stable              stabilize sort by disabling last-resort comparison
|     -u, --unique              with -c, check for strict ordering; without -c, output only the first of an equal run
|   
+-- split
|   Usage: omni split [OPTION]... [FILE [PREFIX]] [flags]
|   Description: Output pieces of FILE to PREFIXaa, PREFIXab, ...;
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
|   
|   Flags:
|     -b, --bytes string        put SIZE bytes per output file
|     -l, --lines int           put NUMBER lines per output file
|     -d, --numeric-suffixes    use numeric suffixes
|     -a, --suffix-length int   generate suffixes of length N
|         --verbose             print diagnostic for each output file
|   
+-- sql
|   Usage: omni sql [FILE] [flags]
|   Description: SQL utilities for formatting, minifying, and validating SQL.

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
|   
|   Flags:
|     -i, --indent string       indentation string
|     -u, --uppercase           uppercase keywords
|   
|   +-- fmt
|   |   Usage: omni sql fmt [FILE] [flags]
|   |   Description: Format SQL with proper indentation and keyword capitalization.

  -i, --indent=STR     indentation string (default "  ")
  -u, --uppercase      uppercase keywords (default: true)
  -d, --dialect=NAME   SQL dialect: mysql, postgres, sqlite (default: generic)

Examples:
  omni sql fmt file.sql
  omni sql fmt "select * from users where id = 1"
  cat file.sql | omni sql fmt
  omni sql fmt --indent "    " file.sql
|   |   
|   |   Flags:
|   |     -d, --dialect string      SQL dialect (mysql, postgres, sqlite, generic)
|   |     -i, --indent string       indentation string
|   |     -u, --uppercase           uppercase keywords
|   |   
|   +-- minify
|   |   Usage: omni sql minify [FILE]
|   |   Description: Minify SQL by removing unnecessary whitespace.

Examples:
  omni sql minify file.sql
  cat file.sql | omni sql minify
|   |   
|   \-- validate
|       Usage: omni sql validate [FILE] [flags]
|       Description: Validate SQL syntax without outputting the query.

Exit codes:
  0  Valid SQL
  1  Invalid SQL or error

  --json    output result as JSON

Examples:
  omni sql validate file.sql
  omni sql validate "SELECT * FROM users"
  omni sql validate --json file.sql
|       
|       Flags:
|         -d, --dialect string      SQL dialect
|       
+-- sqlite
|   Usage: omni sqlite
|   Description: sqlite provides commands for working with SQLite databases.

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
|   
|   +-- check
|   |   Usage: omni sqlite check <database>
|   |   Description: Verify database integrity
|   |   
|   +-- columns
|   |   Usage: omni sqlite columns <database> <table>
|   |   Description: Show table columns
|   |   
|   +-- dump
|   |   Usage: omni sqlite dump <database> [table]
|   |   Description: Export database as SQL
|   |   
|   +-- import
|   |   Usage: omni sqlite import <database> <sql-file>
|   |   Description: Import SQL file into database
|   |   
|   +-- indexes
|   |   Usage: omni sqlite indexes <database>
|   |   Description: List all indexes
|   |   
|   +-- query
|   |   Usage: omni sqlite query <database> <sql> [flags]
|   |   Description: Execute SQL query against a SQLite database.

Query logging can be enabled with the omni logger command:
  eval "$(omni logger --path /path/to/logs)"

With logging enabled, queries and results are recorded to log files.
Use --log-data to include result data in logs (use with caution for large results).

Examples:
  omni sqlite query mydb.sqlite "SELECT * FROM users"
  omni sqlite query mydb.sqlite "SELECT * FROM users" --header
  omni sqlite query mydb.sqlite "SELECT * FROM users" --json
  omni sqlite query mydb.sqlite "SELECT * FROM users" --log-data
|   |   
|   |   Flags:
|   |     -H, --header              show column headers
|   |         --log-data            include result data in logs (use with caution for large results)
|   |     -s, --separator string    column separator
|   |   
|   +-- schema
|   |   Usage: omni sqlite schema <database> [table]
|   |   Description: Show table schema
|   |   
|   +-- stats
|   |   Usage: omni sqlite stats <database>
|   |   Description: Display database statistics
|   |   
|   +-- tables
|   |   Usage: omni sqlite tables <database>
|   |   Description: List all tables in the database
|   |   
|   \-- vacuum
|       Usage: omni sqlite vacuum <database>
|       Description: Optimize database
|       
+-- ss
|   Usage: omni ss [OPTIONS] [flags]
|   Description: Display socket statistics. Similar to the Linux ss command.

Shows information about TCP, UDP, and Unix sockets including state,
local/remote addresses, and optionally process information.

Options:
  -a          display all sockets
  -l          display listening sockets only
  -t          display TCP sockets
  -u          display UDP sockets
  -x          display Unix sockets
  -p          show process using socket
  -n          don't resolve service names
  -4          display only IPv4 sockets
  -6          display only IPv6 sockets
  -s          print summary statistics
  -e          show extended socket info
  --state     filter by state (established, listen, time_wait, etc.)

Examples:
  omni ss -l              # show listening sockets
  omni ss -t              # show TCP sockets
  omni ss -tl             # show TCP listening sockets
  omni ss -tlp            # show TCP listening sockets with process info
  omni ss -a              # show all sockets
  omni ss -s              # show summary statistics
  omni ss --state listen  # filter by state
  omni ss -4              # show only IPv4
  omni ss -j              # output as JSON
|   
|   Flags:
|     -a, --all                 display all sockets
|     -e, --extended            show extended socket info
|     -4, --ipv4                display only IPv4 sockets
|     -6, --ipv6                display only IPv6 sockets
|     -l, --listening           display listening sockets only
|         --no-header           don't print headers
|     -n, --numeric             don't resolve service names
|     -p, --processes           show process using socket
|         --state string        filter by state (established, listen, time_wait, etc.)
|     -s, --summary             print summary statistics
|     -t, --tcp                 display TCP sockets
|     -u, --udp                 display UDP sockets
|     -x, --unix                display Unix sockets
|   
+-- stat
|   Usage: omni stat [file...]
|   Description: Display file or file system status.
|   
+-- strings
|   Usage: omni strings [OPTION]... [FILE]... [flags]
|   Description: Print the sequences of printable characters in files.

  -n, --bytes=MIN   print sequences of at least MIN characters (default 4)
  -t, --radix=TYPE  print offset in TYPE: d=decimal, o=octal, x=hex

Examples:
  omni strings binary.exe         # extract strings from binary
  omni strings -n 8 file.bin      # strings of at least 8 chars
  omni strings -t x program       # show hex offsets
|   
|   Flags:
|     -n, --bytes int           minimum string length
|     -t, --radix string        print offset (d/o/x)
|   
+-- tac
|   Usage: omni tac [OPTION]... [FILE]... [flags]
|   Description: Write each FILE to standard output, last line first.

With no FILE, or when FILE is -, read standard input.

  -b, --before             attach the separator before instead of after
  -r, --regex              interpret the separator as a regular expression
  -s, --separator=STRING   use STRING as the separator instead of newline
|   
|   Flags:
|     -b, --before              attach the separator before instead of after
|     -r, --regex               interpret the separator as a regular expression
|     -s, --separator string    use STRING as the separator instead of newline
|   
+-- tagfixer
|   Usage: omni tagfixer [PATH] [flags]
|   Description: Fix and standardize struct tags in Go files.

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
|   
|   Flags:
|     -a, --analyze             analyze mode - generate report only
|     -c, --case string         target case type (camel, pascal, snake, kebab)
|     -d, --dry-run             preview changes without writing
|         --json                output as JSON
|     -r, --recursive           process directories recursively
|     -t, --tags string         comma-separated tag types to fix
|     -v, --verbose             verbose output
|   
|   \-- analyze
|       Usage: omni tagfixer analyze [PATH] [flags]
|       Description: Analyze Go files to understand current struct tag patterns.

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
|       
|       Flags:
|             --json                output as JSON
|         -r, --recursive           process directories recursively
|         -t, --tags string         comma-separated tag types to analyze
|         -v, --verbose             verbose output (show all files)
|       
+-- tail
|   Usage: omni tail [option]... [file]... [flags]
|   Description: Print the last 10 lines of each FILE to standard output.
With more than one FILE, precede each with a header giving the file name.
With no FILE, or when FILE is -, read standard input.

Numeric shortcuts are supported: -80 is equivalent to -n 80.
|   
|   Flags:
|     -c, --bytes int           output the last NUM bytes
|     -f, --follow              output appended data as the file grows
|     -n, --lines int           output the last NUM lines, instead of the last 10
|     -q, --quiet               never output headers giving file names
|         --sleep-interval duration  with -f, sleep for approximately N seconds between iterations
|     -v, --verbose             always output headers giving file names
|   
+-- tar
|   Usage: omni tar [OPTION]... [FILE]... [flags]
|   Description: Manipulate tape archive files.

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
|   
|   Flags:
|     -c, --create              create a new archive
|     -C, --directory string    change to directory DIR
|     -x, --extract             extract files from an archive
|     -f, --file string         use archive file ARCHIVE
|     -z, --gzip                filter through gzip
|         --json                output as JSON (for list mode)
|     -t, --list                list the contents of an archive
|         --strip-components int  strip N leading path components
|     -v, --verbose             verbosely list files processed
|   
+-- task
|   Usage: omni task [TASK...] [flags]
|   Description: A task runner that executes tasks defined in Taskfile.yml.

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
|   
|   Flags:
|         --allow-external      allow external (non-omni) commands
|     -d, --dir string          working directory
|         --dry-run             print commands without executing
|     -f, --force               force run even if up-to-date
|     -l, --list                list available tasks
|     -s, --silent              suppress output
|         --summary             show task summary
|     -t, --taskfile string     path to Taskfile.yml
|     -v, --verbose             verbose output
|   
+-- terraform
|   Usage: omni terraform
|   Description: Terraform CLI wrapper integrated into omni.

Provides access to all Terraform commands. You can use 'omni terraform'
or the shorter alias 'omni tf'.

Examples:
  omni terraform init
  omni tf plan
  omni tf apply -auto-approve
  omni tf destroy
|   
|   +-- apply
|   |   Usage: omni terraform apply [plan-file] [flags]
|   |   Description: Apply the changes required to reach the desired state.

Examples:
  omni tf apply
  omni tf apply plan.tfplan
  omni tf apply -auto-approve
|   |   
|   |   Flags:
|   |         --auto-approve        Skip interactive approval
|   |         --var stringToString  Set variables
|   |         --var-file stringSlice  Variable files
|   |   
|   +-- console
|   |   Usage: omni terraform console
|   |   Description: Launch an interactive console for evaluating expressions.

Examples:
  omni tf console
|   |   
|   +-- destroy
|   |   Usage: omni terraform destroy [flags]
|   |   Description: Destroy all remote objects managed by Terraform.

Examples:
  omni tf destroy
  omni tf destroy -auto-approve
|   |   
|   |   Flags:
|   |         --auto-approve        Skip interactive approval
|   |         --var stringToString  Set variables
|   |         --var-file stringSlice  Variable files
|   |   
|   +-- fmt
|   |   Usage: omni terraform fmt [flags]
|   |   Description: Reformat configuration files to a canonical format.

Examples:
  omni tf fmt
  omni tf fmt -check
  omni tf fmt -recursive
|   |   
|   |   Flags:
|   |         --check               Check if formatted
|   |         --diff                Display diffs
|   |         --recursive           Process subdirectories
|   |   
|   +-- get
|   |   Usage: omni terraform get [flags]
|   |   Description: Download and install modules for the configuration.

Examples:
  omni tf get
  omni tf get -update
|   |   
|   |   Flags:
|   |         --update              Update modules
|   |   
|   +-- graph
|   |   Usage: omni terraform graph [flags]
|   |   Description: Generate a visual representation of dependencies.

Examples:
  omni tf graph
  omni tf graph | dot -Tpng > graph.png
|   |   
|   |   Flags:
|   |         --draw-cycles         Draw cycles
|   |         --plan string         Plan file
|   |   
|   +-- import
|   |   Usage: omni terraform import <address> <id> [flags]
|   |   Description: Import existing infrastructure into Terraform state.

Examples:
  omni tf import aws_instance.example i-1234567890
|   |   
|   |   Flags:
|   |         --var stringToString  Set variables
|   |         --var-file stringSlice  Variable files
|   |   
|   +-- init
|   |   Usage: omni terraform init [flags]
|   |   Description: Initialize a new or existing Terraform working directory.

Examples:
  omni tf init
  omni tf init -upgrade
  omni tf init -reconfigure
|   |   
|   |   Flags:
|   |         --reconfigure         Reconfigure backend
|   |         --upgrade             Upgrade modules and plugins
|   |   
|   +-- output
|   |   Usage: omni terraform output [name] [flags]
|   |   Description: Read an output variable from the state file.

Examples:
  omni tf output
  omni tf output instance_ip
  omni tf output -json
|   |   
|   |   Flags:
|   |         --json                JSON output
|   |   
|   +-- plan
|   |   Usage: omni terraform plan [flags]
|   |   Description: Create an execution plan showing what Terraform will do.

Examples:
  omni tf plan
  omni tf plan -out=plan.tfplan
  omni tf plan -var "region=us-east-1"
|   |   
|   |   Flags:
|   |         --destroy             Create destroy plan
|   |     -o, --out string          Write plan to file
|   |         --var stringToString  Set variables
|   |         --var-file stringSlice  Variable files
|   |   
|   +-- providers
|   |   Usage: omni terraform providers
|   |   Description: Show the providers required for this configuration.

Examples:
  omni tf providers
|   |   
|   +-- refresh
|   |   Usage: omni terraform refresh [flags]
|   |   Description: Update the state file with real-world infrastructure.

Examples:
  omni tf refresh
|   |   
|   |   Flags:
|   |         --var stringToString  Set variables
|   |         --var-file stringSlice  Variable files
|   |   
|   +-- show
|   |   Usage: omni terraform show [plan-file] [flags]
|   |   Description: Show a human-readable output from a plan file or state.

Examples:
  omni tf show
  omni tf show plan.tfplan
  omni tf show -json
|   |   
|   |   Flags:
|   |         --json                JSON output
|   |   
|   +-- state
|   |   Usage: omni terraform state
|   |   Description: Commands for managing Terraform state.
|   |   
|   |   +-- list
|   |   |   Usage: omni terraform state list [addresses...]
|   |   |   Description: List resources in the Terraform state.

Examples:
  omni tf state list
  omni tf state list aws_instance.example
|   |   |   
|   |   +-- mv
|   |   |   Usage: omni terraform state mv <source> <destination> [flags]
|   |   |   Description: Move a resource from one address to another.

Examples:
  omni tf state mv aws_instance.old aws_instance.new
|   |   |   
|   |   |   Flags:
|   |   |         --dry-run             Dry run
|   |   |   
|   |   +-- rm
|   |   |   Usage: omni terraform state rm <addresses...> [flags]
|   |   |   Description: Remove resources from the Terraform state.

Examples:
  omni tf state rm aws_instance.example
|   |   |   
|   |   |   Flags:
|   |   |         --dry-run             Dry run
|   |   |   
|   |   \-- show
|   |       Usage: omni terraform state show <address>
|   |       Description: Show the attributes of a single resource in the state.

Examples:
  omni tf state show aws_instance.example
|   |       
|   +-- taint
|   |   Usage: omni terraform taint <address>
|   |   Description: Mark a resource instance as not fully functional.

Examples:
  omni tf taint aws_instance.example
|   |   
|   +-- test
|   |   Usage: omni terraform test
|   |   Description: Execute Terraform test files.

Examples:
  omni tf test
|   |   
|   +-- untaint
|   |   Usage: omni terraform untaint <address>
|   |   Description: Remove the taint state from a resource instance.

Examples:
  omni tf untaint aws_instance.example
|   |   
|   +-- validate
|   |   Usage: omni terraform validate
|   |   Description: Validate the configuration files.

Examples:
  omni tf validate
|   |   
|   +-- version
|   |   Usage: omni terraform version
|   |   Description: Show the current Terraform version.

Examples:
  omni tf version
|   |   
|   \-- workspace
|       Usage: omni terraform workspace
|       Description: Commands for managing Terraform workspaces.
|       
|       +-- delete
|       |   Usage: omni terraform workspace delete <name> [flags]
|       |   Description: Delete a workspace.

Examples:
  omni tf workspace delete old-workspace
|       |   
|       |   Flags:
|       |         --force               Force delete
|       |   
|       +-- list
|       |   Usage: omni terraform workspace list
|       |   Description: List all available workspaces.

Examples:
  omni tf workspace list
|       |   
|       +-- new
|       |   Usage: omni terraform workspace new <name>
|       |   Description: Create a new workspace.

Examples:
  omni tf workspace new development
|       |   
|       +-- select
|       |   Usage: omni terraform workspace select <name>
|       |   Description: Select a workspace to use.

Examples:
  omni tf workspace select production
|       |   
|       \-- show
|           Usage: omni terraform workspace show
|           Description: Show the name of the current workspace.

Examples:
  omni tf workspace show
|           
+-- testcheck
|   Usage: omni testcheck [directory] [flags]
|   Description: Scan a directory for Go packages and report which have tests.

By default, only shows packages WITHOUT tests. Use --all to show all packages.

Examples:
  omni testcheck ./pkg/cli/           # Check packages in pkg/cli
  omni testcheck .                    # Check current directory
  omni testcheck --all ./pkg/         # Show all packages
  omni testcheck --summary ./pkg/     # Show only summary
  omni testcheck --json ./pkg/        # Output as JSON
|   
|   Flags:
|     -a, --all                 show all packages (default shows only missing)
|     -s, --summary             show only summary
|     -v, --verbose             show test file names
|   
+-- time
|   Usage: omni time
|   Description: The time utility executes and times the specified utility. After the
utility finishes, time writes to the standard error stream, the total
time elapsed.

Note: Since omni doesn't execute external commands, this command
provides timing utilities and can measure internal operations.

Examples:
  omni time sleep 2    # Time a sleep operation
  omni time            # Just show current time info
|   
+-- toml
|   Usage: omni toml
|   Description: TOML utilities for validation and formatting.

Subcommands:
  validate    Validate TOML syntax
  fmt         Format/beautify TOML

Examples:
  omni toml validate config.toml
  omni toml fmt config.toml
|   
|   +-- fmt
|   |   Usage: omni toml fmt [FILE] [flags]
|   |   Description: Format and beautify TOML.

Parses TOML and outputs it with consistent formatting.

Examples:
  omni toml fmt config.toml
  cat config.toml | omni toml fmt
  omni toml fmt --indent 4 config.toml
|   |   
|   |   Flags:
|   |     -i, --indent int          indentation width
|   |   
|   \-- validate
|       Usage: omni toml validate [FILE...]
|       Description: Validate TOML syntax for one or more files.

Checks that the input is valid TOML.

Examples:
  omni toml validate config.toml
  omni toml validate *.toml
  cat config.toml | omni toml validate
  omni toml validate --json config.toml
|       
+-- top
|   Usage: omni top [flags]
|   Description: Display system processes sorted by CPU or memory usage.

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
|   
|   Flags:
|         --go                  show only Go processes
|     -n, --num int             number of processes to show
|         --sort string         sort by column: cpu, mem, pid
|   
+-- touch
|   Usage: omni touch [file...]
|   Description: Update the access and modification times of each FILE to the current time. A FILE argument that does not exist is created empty.
|   
+-- tr
|   Usage: omni tr [OPTION]... SET1 [SET2] [flags]
|   Description: Translate, squeeze, and/or delete characters from standard input,
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
|   
|   Flags:
|     -c, --complement          use the complement of SET1
|     -d, --delete              delete characters in SET1, do not translate
|     -s, --squeeze-repeats     replace repeated characters with single occurrence
|     -t, --truncate-set1       first truncate SET1 to length of SET2
|   
+-- tree
|   Usage: omni tree [path] [flags]
|   Description: Display a tree visualization of directory contents.

Examples:
  omni tree                          # current directory
  omni tree /path/to/dir             # specific directory
  omni tree -a                       # show hidden files
  omni tree -L 3                     # limit depth to 3
  omni tree -d 3                     # limit depth to 3 (alias for -L)
  omni tree -i "node_modules,.git"   # ignore patterns
  omni tree --dirs-only              # show only directories
  omni tree -s                       # show statistics
  omni tree --json                   # output as JSON
  omni tree --json-stream            # streaming NDJSON output
  omni tree -t 8                     # use 8 parallel workers
  omni tree --max-files 10000        # cap at 10000 items
  omni tree --compare a.json b.json  # compare two snapshots
|   
|   Flags:
|     -a, --all                 show hidden files
|         --compare stringSlice  compare two JSON tree snapshots
|         --date                show modification dates
|     -d, --depth int           maximum depth to scan (-1 for unlimited)
|         --detect-moves        detect moved files when comparing (default true)
|         --dirs-only           show only directories
|         --hash                show SHA256 hash for files
|     -i, --ignore string       patterns to ignore (comma-separated)
|     -j, --json                output as JSON format
|         --json-stream         streaming NDJSON output (one JSON object per line)
|     -L, --level int           maximum depth level (alias for --depth)
|         --max-files int       maximum number of files to scan (0 = unlimited)
|         --max-hash-size int64  skip hashing files larger than N bytes (0 = unlimited)
|         --no-color            disable colored output
|         --no-dir-slash        don't add trailing slash to directory names
|         --size                show file sizes
|     -s, --stats               show statistics
|     -t, --threads int         number of parallel workers (0 = auto, 1 = sequential)
|   
+-- ulid
|   Usage: omni ulid [OPTION]... [flags]
|   Description: Generate ULIDs (Universally Unique Lexicographically Sortable Identifiers).

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
|   
|   Flags:
|     -n, --count int           generate N ULIDs
|     -l, --lower               output in lowercase
|   
+-- uname
|   Usage: omni uname [flags]
|   Description: Print certain system information. With no OPTION, same as -s.

  -a, --all                print all information
  -s, --kernel-name        print the kernel name
  -n, --nodename           print the network node hostname
  -r, --kernel-release     print the kernel release
  -v, --kernel-version     print the kernel version
  -m, --machine            print the machine hardware name
  -p, --processor          print the processor type
  -i, --hardware-platform  print the hardware platform
  -o, --operating-system   print the operating system
|   
|   Flags:
|     -a, --all                 print all information
|     -i, --hardware-platform   print the hardware platform
|     -s, --kernel-name         print the kernel name
|     -r, --kernel-release      print the kernel release
|     -v, --kernel-version      print the kernel version
|     -m, --machine             print the machine hardware name
|     -n, --nodename            print the network node hostname
|     -o, --operating-system    print the operating system
|     -p, --processor           print the processor type
|   
+-- uniq
|   Usage: omni uniq [option]... [input [output]] [flags]
|   Description: Filter adjacent matching lines from INPUT (or standard input),
writing to OUTPUT (or standard output).

With no options, matching lines are merged to the first occurrence.

Note: 'uniq' does not detect repeated lines unless they are adjacent.
You may want to sort the input first, or use 'sort -u' without 'uniq'.
|   
|   Flags:
|     -D, --all-repeated        print all duplicate lines
|     -w, --check-chars int     compare no more than N characters in lines
|     -c, --count               prefix lines by the number of occurrences
|     -i, --ignore-case         ignore differences in case when comparing
|     -d, --repeated            only print duplicate lines, one for each group
|     -s, --skip-chars int      avoid comparing the first N characters
|     -f, --skip-fields int     avoid comparing the first N fields
|     -u, --unique              only print unique lines
|     -z, --zero-terminated     line delimiter is NUL, not newline
|   
+-- unxz
|   Usage: omni unxz [OPTION]... [FILE]... [flags]
|   Description: Decompress FILEs in xz format.

Equivalent to xz -d.

Note: Full decompression requires external library.
|   
|   Flags:
|     -f, --force               force overwrite
|     -k, --keep                keep original files
|     -c, --stdout              write to stdout
|     -v, --verbose             verbose mode
|   
+-- unzip
|   Usage: omni unzip [OPTION]... ZIPFILE [flags]
|   Description: Extract files from a zip archive.

  -l, --list        list contents without extracting
  -v, --verbose     verbose output
  -d, --directory   extract files into directory
      --strip-components=N  strip N leading path components

Examples:
  omni unzip archive.zip              # extract to current directory
  omni unzip -d /dest archive.zip     # extract to specific directory
  omni unzip -l archive.zip           # list contents
  omni unzip -v archive.zip           # verbose extraction
|   
|   Flags:
|     -d, --directory string    extract files into directory
|         --json                output as JSON (for list mode)
|     -l, --list                list contents without extracting
|         --strip-components int  strip N leading path components
|     -v, --verbose             verbose output
|   
+-- uptime
|   Usage: omni uptime [OPTION]... [flags]
|   Description: Print the current time, how long the system has been running,
how many users are currently logged on, and the system load averages
for the past 1, 5, and 15 minutes.

  -p, --pretty   show uptime in pretty format
  -s, --since    system up since, in yyyy-mm-dd HH:MM:SS format
|   
|   Flags:
|     -p, --pretty              show uptime in pretty format
|     -s, --since               system up since
|   
+-- url
|   Usage: omni url
|   Description: URL encoding and decoding utilities.

Subcommands:
  encode    URL encode text
  decode    URL decode text

Examples:
  omni url encode "hello world"
  omni url decode "hello%20world"
  echo "test string" | omni url encode
  omni url encode --component "a=b&c=d"
|   
|   +-- decode
|   |   Usage: omni url decode [TEXT] [flags]
|   |   Description: URL decode percent-encoded text.

By default uses path decoding. Use --component for query string decoding.

Examples:
  omni url decode "hello%20world"         # Output: hello world
  omni url decode --component "a%3Db"     # Output: a=b
  echo "test%20string" | omni url decode  # Read from stdin
|   |   
|   |   Flags:
|   |     -c, --component           use query component decoding
|   |         --json                output as JSON
|   |   
|   \-- encode
|       Usage: omni url encode [TEXT] [flags]
|       Description: URL encode text for safe use in URLs.

By default uses path encoding. Use --component for query string encoding
which is more aggressive (encodes more characters).

Examples:
  omni url encode "hello world"           # Output: hello%20world
  omni url encode --component "a=b&c=d"   # Output: a%3Db%26c%3Dd
  echo "test" | omni url encode           # Read from stdin
  omni url encode file.txt                # Read from file
|       
|       Flags:
|         -c, --component           use query component encoding (more aggressive)
|             --json                output as JSON
|       
+-- uuid
|   Usage: omni uuid [OPTION]... [flags]
|   Description: Generate random UUIDs (Universally Unique Identifiers).

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
|   
|   Flags:
|     -n, --count int           generate N UUIDs
|     -x, --no-dashes           output without dashes
|     -u, --upper               output in uppercase
|     -v, --version int         UUID version (4 or 7)
|   
+-- vault
|   Usage: omni vault
|   Description: HashiCorp Vault CLI for secrets management.

Environment variables:
  VAULT_ADDR       Vault server address (default: https://127.0.0.1:8200)
  VAULT_TOKEN      Authentication token
  VAULT_NAMESPACE  Vault namespace

Examples:
  omni vault status
  omni vault login -method=token
  omni vault kv get secret/myapp
  omni vault kv put secret/myapp key=value
|   
|   +-- delete
|   |   Usage: omni vault delete <path>
|   |   Description: Delete a secret at the given path.

Examples:
  omni vault delete secret/data/myapp
|   |   
|   +-- kv
|   |   Usage: omni vault kv
|   |   Description: Interact with Vault's KV secrets engine (v2).
|   |   
|   |   +-- delete
|   |   |   Usage: omni vault kv delete <path> [flags]
|   |   |   Description: Soft delete a secret from the KV secrets engine.

Examples:
  omni vault kv delete secret/myapp
  omni vault kv delete -versions=1,2,3 secret/myapp
|   |   |   
|   |   |   Flags:
|   |   |         --versions string     Comma-separated versions to delete
|   |   |   
|   |   +-- destroy
|   |   |   Usage: omni vault kv destroy <path> [flags]
|   |   |   Description: Permanently destroy versions of a secret in the KV secrets engine.

Examples:
  omni vault kv destroy -versions=1,2,3 secret/myapp
|   |   |   
|   |   |   Flags:
|   |   |         --versions string     Comma-separated versions to destroy (required)
|   |   |   
|   |   +-- get
|   |   |   Usage: omni vault kv get <path> [flags]
|   |   |   Description: Retrieve a secret from the KV secrets engine.

Examples:
  omni vault kv get secret/myapp
  omni vault kv get -mount=kv myapp
  omni vault kv get -version=2 secret/myapp
  omni vault kv get -field=password secret/myapp
|   |   |   
|   |   |   Flags:
|   |   |         --field string        Print only this field
|   |   |         --version int         Version to retrieve (0 = latest)
|   |   |   
|   |   +-- list
|   |   |   Usage: omni vault kv list [path]
|   |   |   Description: List secrets at a path in the KV secrets engine.

Examples:
  omni vault kv list secret/
  omni vault kv list -mount=kv myapp/
|   |   |   
|   |   +-- metadata
|   |   |   Usage: omni vault kv metadata
|   |   |   Description: Manage metadata for secrets in the KV secrets engine.
|   |   |   
|   |   |   +-- delete
|   |   |   |   Usage: omni vault kv metadata delete <path>
|   |   |   |   Description: Permanently delete all versions and metadata for a secret.

Examples:
  omni vault kv metadata delete secret/myapp
|   |   |   |   
|   |   |   \-- get
|   |   |       Usage: omni vault kv metadata get <path>
|   |   |       Description: Retrieve metadata for a secret.

Examples:
  omni vault kv metadata get secret/myapp
|   |   |       
|   |   +-- put
|   |   |   Usage: omni vault kv put <path> [key=value...]
|   |   |   Description: Write a secret to the KV secrets engine.

Examples:
  omni vault kv put secret/myapp key=value
  omni vault kv put -mount=kv myapp username=admin password=secret
|   |   |   
|   |   \-- undelete
|   |       Usage: omni vault kv undelete <path> [flags]
|   |       Description: Restore deleted versions of a secret in the KV secrets engine.

Examples:
  omni vault kv undelete -versions=1,2,3 secret/myapp
|   |       
|   |       Flags:
|   |             --versions string     Comma-separated versions to restore (required)
|   |       
|   +-- list
|   |   Usage: omni vault list <path>
|   |   Description: List secrets at the given path.

Examples:
  omni vault list secret/metadata/
|   |   
|   +-- login
|   |   Usage: omni vault login [token] [flags]
|   |   Description: Authenticate to Vault using various methods.

Methods:
  token     - Direct token authentication (default)
  userpass  - Username/password authentication
  approle   - AppRole authentication

Examples:
  omni vault login                           # Prompts for token
  omni vault login s.xxxxx                   # Direct token
  omni vault login -method=userpass -username=admin
  omni vault login -method=approle -role-id=xxx -secret-id=xxx
|   |   
|   |   Flags:
|   |         --method string       Auth method (token, userpass, approle)
|   |         --no-store            Don't save token to file
|   |         --password string     Password for userpass auth
|   |         --path string         Mount path for auth method
|   |         --role-id string      Role ID for approle auth
|   |         --secret-id string    Secret ID for approle auth
|   |         --username string     Username for userpass auth
|   |   
|   +-- read
|   |   Usage: omni vault read <path> [flags]
|   |   Description: Read a secret from Vault at the given path.

Examples:
  omni vault read secret/data/myapp
  omni vault read -field=password secret/data/myapp
|   |   
|   |   Flags:
|   |         --field string        Print only this field
|   |   
|   +-- status
|   |   Usage: omni vault status
|   |   Description: Show the status of the Vault server.

Examples:
  omni vault status
|   |   
|   +-- token
|   |   Usage: omni vault token
|   |   Description: Token management operations.
|   |   
|   |   +-- lookup
|   |   |   Usage: omni vault token lookup
|   |   |   Description: Display information about the current token.

Examples:
  omni vault token lookup
|   |   |   
|   |   +-- renew
|   |   |   Usage: omni vault token renew [flags]
|   |   |   Description: Renew the current token's lease.

Examples:
  omni vault token renew
  omni vault token renew -increment=3600
|   |   |   
|   |   |   Flags:
|   |   |         --increment int       Lease increment in seconds
|   |   |   
|   |   \-- revoke
|   |       Usage: omni vault token revoke
|   |       Description: Revoke the current token.

Examples:
  omni vault token revoke
|   |       
|   \-- write
|       Usage: omni vault write <path> [key=value...]
|       Description: Write a secret to Vault at the given path.

Examples:
  omni vault write secret/data/myapp key=value
  omni vault write secret/data/myapp username=admin password=secret
|       
+-- video
|   Usage: omni video [URL] [flags]
|   Description: Video downloader supporting YouTube and other video platforms.

Subcommands:
  download      Download video(s) from URL
  channel       Download all videos from a YouTube channel (incremental)
  info          Show video metadata as JSON
  list-formats  List available download formats
  search        Search YouTube
  interactive   Interactive menu (download/info/formats/nerd stats)
  extractors    List supported sites
  auth          Extract YouTube cookies from Chrome for authenticated downloads

Examples:
  omni video auth
  omni video download "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video channel "https://www.youtube.com/@GithubAwesome" --limit 5
  omni video info "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video list-formats "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video download -f worst "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video search "golang tutorial"
  omni video interactive
  omni video "https://www.youtube.com/watch?v=dQw4w9WgXcQ" --interactive
  omni video extractors
|   
|   Flags:
|         --cookies string      Netscape cookie file path for interactive mode
|         --interactive         start interactive mode (same as 'video interactive')
|         --proxy string        HTTP/SOCKS proxy URL for interactive mode
|     -v, --verbose             verbose output for interactive mode
|   
|   +-- auth
|   |   Usage: omni video auth
|   |   Description: Launch Chrome with your existing profile to extract YouTube cookies.

Cookies are saved to a well-known cache location and will be
auto-loaded for all subsequent video commands. This enables
authenticated downloads (age-restricted, members-only, etc.).

Requires Google Chrome with an active YouTube login.

Examples:
  omni video auth
|   |   
|   +-- channel
|   |   Usage: omni video channel <URL> [flags]
|   |   Description: Download all videos from a YouTube channel with SQLite tracking.

Downloads are stored in ~/Downloads/<ChannelName>/ with a channel.db
database that tracks downloaded videos. Re-running the same command
skips already-downloaded videos (incremental).

Each video is downloaded in "complete" mode (best quality + .md sidecar).

Flags:
  --limit=N             Max videos to download (-1 = all)
  --cookies=FILE        Netscape cookie file
  --proxy=URL           HTTP/SOCKS proxy
  -q, --quiet           Suppress progress output
  --no-progress         Disable progress bar
  -R, --retries=N       Number of retries (default 3)
  -v, --verbose         Verbose output

Examples:
  omni video channel "https://www.youtube.com/@GithubAwesome"
  omni video ch "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw" --limit 5
  omni video channel "https://www.youtube.com/c/ChannelName" --cookies cookies.txt
|   |   
|   |   Flags:
|   |         --cookies string      Netscape cookie file path
|   |         --cookies-from-browser  auto-load cookies extracted by 'video auth'
|   |         --limit int           max videos to download (-1 = all)
|   |         --no-progress         disable progress bar
|   |         --proxy string        HTTP/SOCKS proxy URL
|   |     -q, --quiet               suppress output
|   |         --rate-limit string   rate limit (e.g., 1M, 500K)
|   |     -R, --retries int         number of retries
|   |     -v, --verbose             verbose output
|   |   
|   +-- download
|   |   Usage: omni video download <URL> [flags]
|   |   Description: Download video from a URL.

Flags:
  -f, --format=SPEC       Format selector (default "best")
  -o, --output=TEMPLATE   Output filename template
  -q, --quiet             Suppress progress output
  --no-progress           Disable progress bar
  --rate-limit=RATE       Rate limit (e.g., "1M", "500K")
  -R, --retries=N         Number of retries (default 3)
  -c, --continue          Resume partial downloads
  --no-part               Don't use .part files
  --cookies=FILE          Netscape cookie file
  --proxy=URL             HTTP/SOCKS proxy
  --write-info-json       Write .info.json file
  --write-subs            Write subtitle files
  --no-playlist           Download single video, not playlist
  --playlist-start=N      Start index (1-based)
  --playlist-end=N        End index
  -v, --verbose           Verbose output
  --complete              Best quality + write .md with description and link

Format selectors:
  best          Best quality with video+audio (default)
  worst         Worst quality with video+audio
  bestvideo     Best video-only stream
  bestaudio     Best audio-only stream
  FORMAT_ID     Specific format by ID
  best[height<=720]   Best format with height <= 720

Output template variables:
  %(id)s, %(title)s, %(ext)s, %(uploader)s, %(upload_date)s,
  %(channel)s, %(format_id)s, %(resolution)s

Examples:
  omni video download "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video dl -f worst "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video dl -o "%(title)s-%(format_id)s.%(ext)s" URL
  omni video dl --rate-limit 1M URL
  omni video dl -c URL           # resume partial download
  omni video dl --complete URL   # best quality + markdown sidecar
|   |   
|   |   Flags:
|   |         --complete            download best quality and write .md file with description and link
|   |     -c, --continue            resume partial downloads
|   |         --cookies string      Netscape cookie file path
|   |         --cookies-from-browser  auto-load cookies extracted by 'video auth'
|   |     -f, --format string       format selector
|   |         --no-part             don't use .part files
|   |         --no-playlist         download single video only
|   |         --no-progress         disable progress bar
|   |     -o, --output string       output filename template
|   |         --playlist-end int    playlist end index
|   |         --playlist-start int  playlist start index (1-based)
|   |         --proxy string        HTTP/SOCKS proxy URL
|   |     -q, --quiet               suppress output
|   |         --rate-limit string   rate limit (e.g., 1M, 500K)
|   |     -R, --retries int         number of retries
|   |     -v, --verbose             verbose output
|   |         --write-info-json     write .info.json file
|   |         --write-subs          write subtitle files
|   |   
|   +-- extractors
|   |   Usage: omni video extractors
|   |   Description: List all registered video extractors.

Examples:
  omni video extractors
|   |   
|   +-- info
|   |   Usage: omni video info <URL> [flags]
|   |   Description: Extract and display video metadata in JSON format.

Examples:
  omni video info "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video info URL | jq '.title'
|   |   
|   |   Flags:
|   |         --cookies string      Netscape cookie file path
|   |         --cookies-from-browser  auto-load cookies extracted by 'video auth'
|   |         --proxy string        HTTP/SOCKS proxy URL
|   |     -v, --verbose             verbose output
|   |   
|   +-- interactive
|   |   Usage: omni video interactive [URL] [flags]
|   |   Description: Start an interactive menu for video operations.

The menu lets you:
  - input or change URL
  - list formats
  - show metadata JSON
  - inspect nerd stats
  - list supported extractors
  - download with best or custom format selector

Examples:
  omni video interactive
  omni video interactive "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video "https://www.youtube.com/watch?v=dQw4w9WgXcQ" --interactive
|   |   
|   |   Flags:
|   |         --complete            download best quality and write .md file with description and link
|   |     -c, --continue            resume partial downloads
|   |         --cookies string      Netscape cookie file path
|   |         --cookies-from-browser  auto-load cookies extracted by 'video auth'
|   |     -f, --format string       format selector used by 'download (best)'
|   |         --no-part             don't use .part files
|   |         --no-playlist         download single video only
|   |         --no-progress         disable progress bar
|   |     -o, --output string       output filename template
|   |         --playlist-end int    playlist end index
|   |         --playlist-start int  playlist start index (1-based)
|   |         --proxy string        HTTP/SOCKS proxy URL
|   |     -q, --quiet               suppress output
|   |         --rate-limit string   rate limit (e.g., 1M, 500K)
|   |     -R, --retries int         number of retries
|   |     -v, --verbose             verbose output
|   |         --write-info-json     write .info.json file
|   |         --write-subs          write subtitle files
|   |   
|   +-- list-formats
|   |   Usage: omni video list-formats <URL> [flags]
|   |   Description: List all available download formats for a video.

Examples:
  omni video list-formats "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video formats URL --json
|   |   
|   |   Flags:
|   |         --cookies string      Netscape cookie file path
|   |         --cookies-from-browser  auto-load cookies extracted by 'video auth'
|   |         --json                output as JSON
|   |         --proxy string        HTTP/SOCKS proxy URL
|   |   
|   \-- search
|       Usage: omni video search <QUERY> [flags]
|       Description: Search YouTube and display results.

Examples:
  omni video search "golang tutorial"
  omni video search "how to cook pasta"
|       
|       Flags:
|         -v, --verbose             verbose output
|       
+-- watch
|   Usage: omni watch [OPTION]... COMMAND [flags]
|   Description: Execute a command repeatedly, displaying its output.

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
|   
|   Flags:
|     -b, --beep                beep if command has a non-zero exit
|     -g, --chgexit             exit when output changes
|     -c, --color               interpret ANSI color sequences
|     -d, --differences         highlight differences between updates
|     -e, --errexit             exit if command has a non-zero exit
|     -n, --interval float64    seconds to wait between updates
|     -t, --no-title            turn off header
|         --only-changes        only display output when it changes
|     -p, --precise             attempt run command in precise intervals
|   
+-- wc
|   Usage: omni wc [option]... [file]... [flags]
|   Description: Print newline, word, and byte counts for each FILE, and a total line if
more than one FILE is specified. A word is a non-zero-length sequence of
characters delimited by white space.

With no FILE, or when FILE is -, read standard input.

The options below may be used to select which counts are printed, always in
the following order: newline, word, character, byte, maximum line length.
|   
|   Flags:
|     -c, --bytes               print the byte counts
|     -m, --chars               print the character counts
|     -l, --lines               print the newline counts
|     -L, --max-line-length     print the maximum display width
|     -w, --words               print the word counts
|   
+-- which
|   Usage: omni which [OPTION]... COMMAND... [flags]
|   Description: Write the full path of COMMAND(s) to standard output.

  -a, --all   print all matching executables in PATH, not just the first

Examples:
  omni which go              # /usr/local/go/bin/go
  omni which python python3  # locate multiple commands
  omni which -a python       # show all python executables
|   
|   Flags:
|     -a, --all                 print all matches
|   
+-- whoami
|   Usage: omni whoami
|   Description: Print the user name associated with the current effective user ID.
|   
+-- xargs
|   Usage: omni xargs [OPTION]... [COMMAND [INITIAL-ARGS]] [flags]
|   Description: Read items from standard input, delimited by blanks or newlines, and
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
|   
|   Flags:
|     -I, --I string            replace occurrences of REPLACE-STR
|     -d, --delimiter string    input items are separated by DELIM
|     -n, --max-args int        use at most MAX arguments per command line
|     -P, --max-procs int       run at most MAX processes at a time
|     -r, --no-run-if-empty     if there are no arguments, do not run
|     -0, --null                input items are separated by a null character
|     -t, --verbose             print commands before executing them
|   
+-- xml
|   Usage: omni xml [FILE] [flags]
|   Description: XML utilities for formatting, validation, and conversion.

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
|   
|   Flags:
|     -i, --indent string       indentation string
|     -m, --minify              minify XML (remove whitespace)
|   
|   +-- fmt
|   |   Usage: omni xml fmt [FILE] [flags]
|   |   Description: Format and beautify XML.

Reads XML from a file, argument, or stdin and outputs formatted XML
with proper indentation. Use --minify to remove whitespace.

Examples:
  omni xml fmt file.xml
  omni xml fmt "<root><item>value</item></root>"
  cat file.xml | omni xml fmt
  omni xml fmt --minify file.xml
  omni xml fmt --indent "    " file.xml
|   |   
|   |   Flags:
|   |     -i, --indent string       indentation string
|   |     -m, --minify              minify XML (remove whitespace)
|   |   
|   +-- fromjson
|   |   Usage: omni xml fromjson [FILE] [flags]
|   |   Description: Convert JSON data to XML format.

  -r, --root=NAME      root element name (default "root")
  -i, --indent=STR     indentation string (default "  ")
  --item-tag=NAME      tag for array items (default "item")
  --attr-prefix=STR    prefix for attributes (default "-")

Examples:
  omni xml fromjson file.json
  echo '{"name":"John"}' | omni xml fromjson
  omni xml fromjson -r person file.json
|   |   
|   |   Flags:
|   |         --attr-prefix string  prefix for attributes
|   |     -i, --indent string       indentation string
|   |         --item-tag string     tag for array items
|   |     -r, --root string         root element name
|   |   
|   +-- tojson
|   |   Usage: omni xml tojson [FILE] [flags]
|   |   Description: Convert XML data to JSON format.

  --attr-prefix=STR    prefix for attributes in JSON (default "-")
  --text-key=STR       key for text content (default "#text")

Examples:
  omni xml tojson file.xml
  cat file.xml | omni xml tojson
  omni xml tojson --attr-prefix=@ file.xml
|   |   
|   |   Flags:
|   |         --attr-prefix string  prefix for attributes in JSON
|   |         --text-key string     key for text content
|   |   
|   \-- validate
|       Usage: omni xml validate [FILE...]
|       Description: Validate XML syntax for one or more files.

Checks that the input is well-formed XML.

Examples:
  omni xml validate file.xml
  omni xml validate *.xml
  cat file.xml | omni xml validate
  omni xml validate --json file.xml
|       
+-- xxd
|   Usage: omni xxd [OPTIONS] [FILE] [flags]
|   Description: Make a hex dump of a file or standard input, or reverse it.

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
|   
|   Flags:
|     -b, --bits                binary digit dump (bits instead of hex)
|     -c, --cols int            format <cols> octets per line (default 16)
|     -g, --groupsize int       separate output with <bytes> spaces (default 2)
|     -i, --include             output in C include file style
|     -l, --len int             stop after <len> octets
|     -p, --plain               output plain hex dump (no addresses or ASCII)
|     -r, --reverse             reverse operation: convert hex dump to binary
|     -s, --seek int            start at <seek> bytes offset
|     -u, --uppercase           use uppercase hex letters
|   
+-- xz
|   Usage: omni xz [OPTION]... [FILE]... [flags]
|   Description: Compress or decompress FILEs using xz format.

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
|   
|   Flags:
|     -d, --decompress          decompress
|     -f, --force               force overwrite
|     -k, --keep                keep original files
|     -l, --list                list compressed file info
|     -c, --stdout              write to stdout
|     -v, --verbose             verbose mode
|   
+-- xzcat
|   Usage: omni xzcat [FILE]...
|   Description: Decompress and print FILEs to stdout.

Equivalent to xz -dc.

Note: Full decompression requires external library.
|   
+-- yaml
|   Usage: omni yaml
|   Description: YAML utilities for validation and formatting.

Subcommands:
  validate    Validate YAML syntax
  fmt         Format/beautify YAML
  k8s         Format YAML with Kubernetes conventions
  tostruct    Convert YAML to Go struct definition

Examples:
  omni yaml validate config.yaml
  omni yaml fmt config.yaml
  omni yaml fmt --sort-keys config.yaml
  omni yaml k8s deployment.yaml
  omni yaml tostruct config.yaml
|   
|   +-- fmt
|   |   Usage: omni yaml fmt [FILE] [flags]
|   |   Description: Format and beautify YAML.

Parses YAML and outputs it with consistent formatting.
Supports multi-document YAML files.

Examples:
  omni yaml fmt config.yaml
  omni yaml fmt --sort-keys config.yaml
  omni yaml fmt --remove-empty config.yaml
  omni yaml fmt -i config.yaml              # in-place edit
  cat config.yaml | omni yaml fmt
|   |   
|   |   Flags:
|   |     -i, --in-place            modify file in place
|   |         --indent int          indentation width
|   |         --json                output as JSON instead of YAML
|   |         --remove-empty        remove empty/null values
|   |         --sort-keys           sort keys alphabetically
|   |   
|   +-- k8s
|   |   Usage: omni yaml k8s [FILE] [flags]
|   |   Description: Format YAML with Kubernetes-specific key ordering.

Orders keys according to Kubernetes conventions:
  - Top level: apiVersion, kind, metadata, spec, status
  - Metadata: name, namespace, labels, annotations

Supports multi-document YAML files (---, multiple resources).

Examples:
  omni yaml k8s deployment.yaml
  omni yaml k8s --remove-empty deployment.yaml
  omni yaml k8s -i deployment.yaml           # in-place edit
  cat manifest.yaml | omni yaml k8s
|   |   
|   |   Flags:
|   |     -i, --in-place            modify file in place
|   |         --indent int          indentation width
|   |         --remove-empty        remove empty/null values
|   |   
|   +-- tostruct
|   |   Usage: omni yaml tostruct [FILE] [flags]
|   |   Description: Convert YAML data to a Go struct definition.

  -n, --name=NAME      struct name (default "Root")
  -p, --package=PKG    package name (default "main")
  --inline             inline nested structs
  --omitempty          add omitempty to all fields

Examples:
  omni yaml tostruct config.yaml
  cat config.yaml | omni yaml tostruct
  omni yaml tostruct -n Config -p models config.yaml
  omni yaml tostruct --omitempty config.yaml
|   |   
|   |   Flags:
|   |         --inline              inline nested structs
|   |     -n, --name string         struct name
|   |         --omitempty           add omitempty to all fields
|   |     -p, --package string      package name
|   |   
|   \-- validate
|       Usage: omni yaml validate [FILE...] [flags]
|       Description: Validate YAML syntax for one or more files.

Checks that the input is valid YAML. Supports multi-document YAML files.

Examples:
  omni yaml validate config.yaml
  omni yaml validate *.yaml
  omni yaml validate --strict config.yaml
  cat config.yaml | omni yaml validate
  omni yaml validate --json config.yaml
|       
|       Flags:
|             --strict              fail on unknown fields
|       
+-- yes
|   Usage: omni yes [STRING]...
|   Description: Repeatedly output a line with all specified STRING(s), or 'y'.

Examples:
  omni yes              # outputs 'y' forever
  omni yes hello        # outputs 'hello' forever
  omni yes | head -5    # outputs 5 'y' lines
|   
+-- yq
|   Usage: omni yq [OPTION]... FILTER [FILE]... [flags]
|   Description: yq is a lightweight command-line YAML processor.

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
|   
|   Flags:
|     -c, --compact-output      compact output
|     -I, --indent int          indentation level
|     -n, --null-input          don't read any input
|     -o, --output-format string  output format (yaml or json)
|     -r, --raw-output          output raw strings
|   
+-- zcat
|   Usage: omni zcat [FILE]...
|   Description: Decompress and print FILEs to stdout.

Equivalent to gzip -dc.

Examples:
  omni zcat file.txt.gz        # print decompressed content
  omni zcat file.gz | grep x   # decompress and grep
|   
\-- zip
    Usage: omni zip [OPTION]... ZIPFILE FILE... [flags]
    Description: Create a zip archive from files and directories.

  -v, --verbose     verbose output
  -r, --recursive   recurse into directories (default for directories)
  -C, --directory   change to directory before adding files

Examples:
  omni zip archive.zip file1.txt file2.txt   # create zip
  omni zip archive.zip dir/                   # zip directory
  omni zip -v archive.zip file.txt           # verbose output
    
    Flags:
      -C, --directory string    change to directory before adding
      -r, --recursive           recurse into directories
      -v, --verbose             verbose output
    
```

