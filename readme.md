# Figma Comment Reporter

Automates exporting Figma comments to XLSX reports and emailing them on schedule.

## Features
- XLSX report generation
- Customizable report fields
- Scheduled email delivery
- YAML configuration

## Installation

1. Clone rep:
```bash
git clone https://github.com/Hikitak/figma-comment-reporter.git
cd figma-comment-reporter
```

2. Install dependencies:
```bash
go mod download
```

3. Build bin file:
```bash
go build -o bin/reporter ./cmd/reporter
```

## Configuration
Copy `config.yaml.example` to `config.yaml` and configure:
```yaml
figma:
  token: "your_figma_token"          # Figma personal access token
  file_keys:                         # Figma file keys
    - "abc123"
    - "def456"

schedule: "0 9 * * *"                # Cron schedule

email:
  smtp_host: "smtp.example.com"      # SMTP server
  smtp_port: 587                     # SMTP port
  smtp_username: "user@example.com"  # SMTP username
  smtp_password: "password"          # SMTP password
  from: "noreply@example.com"        # Sender email
  to:                                # Recipients
    - "user1@example.com"
    - "user2@example.com"
  subject: "Figma Comments Report"   # Email subject
  body: "Attached report"             # Email body

report:
  fields:                            # Custom report fields
    - name: "file_name"              # Field name
      display: "File Name"           # Column header
    # ... other fields
```
## Execution

```bash
./bin/reporter path/to/config.yaml
```

## Docker

Build image:

```bash
docker build -t figma-reporter .
```
Run container:

```bash
docker run -v $(pwd)/config.yaml:/root/config.yaml figma-reporter
```

## Report Field Configuration


Available fields:

- `file_name`: Figma file name
- `file_id`: Figma file ID
- `node_name`: Element name
- `node_id`: Element ID
- `message`: Comment text
- `author`: Comment author
- `created_at`: Creation time
- `status`: Status (open/resolved)
- `resolved_at`: Resolution time
- `link`: Comment link

Date format example:
```yaml
- name: "created_at"
  display: "Created At"
  format: "2006-01-02 15:04"
```

## Scheduling
Cron format:

```text
* * * * *
| | | | |
| | | | +----- Day of week (0-6) (0 = Sunday)
| | | +------- Month (1-12)
| | +--------- Day of month (1-31)
| +----------- Hour (0-23)
+------------- Minute (0-59)
```