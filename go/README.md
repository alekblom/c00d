# c00d

**Code from anywhere.** A self-hosted web IDE in a single binary.

## Quick Start

```bash
# Download
curl -L https://c00d.com/download/linux-amd64 -o c00d
chmod +x c00d

# Run
./c00d

# Open http://localhost:3000
```

## Installation

### Download Binary

| Platform | Command |
|----------|---------|
| Linux (amd64) | `curl -L https://c00d.com/download/linux-amd64 -o c00d && chmod +x c00d` |
| Linux (arm64) | `curl -L https://c00d.com/download/linux-arm64 -o c00d && chmod +x c00d` |
| macOS (Intel) | `curl -L https://c00d.com/download/darwin-amd64 -o c00d && chmod +x c00d` |
| macOS (Apple Silicon) | `curl -L https://c00d.com/download/darwin-arm64 -o c00d && chmod +x c00d` |
| Windows | Download from [c00d.com/download](https://c00d.com/download) |

### Build From Source

```bash
git clone https://github.com/alekblom/c00d.git
cd c00d
go build -o c00d
```

## Usage

```bash
# Basic usage (serves current directory on port 3000)
./c00d

# Custom port
./c00d -port 8080

# Specific directory
./c00d -path /var/www/myproject

# With config file
./c00d -config /path/to/config.yaml
```

## Configuration

Create `config.yaml`:

```yaml
port: 3000
password: your-secure-password
base_path: /var/www/myproject
data_dir: .c00d           # Where to store SQLite database

ai:
  provider: c00d          # c00d, anthropic, openai, ollama
  license_key: ""         # c00d Pro license for unlimited AI
  api_key: ""             # For anthropic/openai
  model: claude-sonnet-4-20250514
  ollama_url: http://localhost:11434

editor:
  theme: vs-dark
  font_size: 14
  tab_size: 4

security:
  allowed_ips: []         # Restrict access to specific IPs
  require_https: false
  log_ips: true           # Log all IP addresses (default: true)
```

See `config.example.yaml` for all options.

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/auth` | GET/POST/DELETE | Authentication |
| `/api/files` | GET | List directory |
| `/api/file` | GET/POST/DELETE/PATCH | File operations |
| `/api/terminal` | POST | Execute commands |
| `/api/git` | POST | Git operations |
| `/api/search` | POST | Search file contents |
| `/api/ai` | POST | AI chat |
| `/api/iplogs` | GET | View IP access logs |
| `/api/config` | GET | Get editor/AI config |

### Git Operations

```bash
# Get status
curl -b cookies.txt -d '{"action":"status"}' localhost:3000/api/git

# Stage/unstage files
curl -b cookies.txt -d '{"action":"stage","file":"path/to/file"}' localhost:3000/api/git
curl -b cookies.txt -d '{"action":"stage_all"}' localhost:3000/api/git

# Commit and push
curl -b cookies.txt -d '{"action":"commit","message":"Your message"}' localhost:3000/api/git
curl -b cookies.txt -d '{"action":"push"}' localhost:3000/api/git

# View diff
curl -b cookies.txt -d '{"action":"diff","staged":true}' localhost:3000/api/git
```

### File Search

```bash
# Search for text
curl -b cookies.txt -d '{"query":"TODO","max_results":50}' localhost:3000/api/search

# Regex search in specific files
curl -b cookies.txt -d '{"query":"func\\s+\\w+","is_regex":true,"file_glob":"*.go"}' localhost:3000/api/search
```

### IP Logs

```bash
# View recent IP logs
curl -b cookies.txt 'localhost:3000/api/iplogs?limit=100'

# View unique IPs with request counts
curl -b cookies.txt 'localhost:3000/api/iplogs?view=unique'
```

## Features

- **Monaco Editor** - Same editor as VS Code
- **File Browser** - Navigate, create, edit, rename, delete files
- **Terminal** - Run commands in your browser
- **Git Integration** - Stage, commit, push, pull, diff from the UI
- **File Search** - Search file contents with regex support
- **AI Assistant** - Code explanation, bug fixes, improvements
- **Password Protection** - Secure your instance
- **Session Persistence** - Sessions survive server restarts
- **IP Logging** - Track all access for security auditing
- **Mobile Ready** - Works on phone and tablet
- **Single Binary** - No dependencies, just download and run

## AI Providers

| Provider | Setup | Cost |
|----------|-------|------|
| **c00d API** | Just works | Free: 20/day, Pro: $9/mo unlimited |
| **Anthropic** | Add API key | Pay Anthropic directly |
| **OpenAI** | Add API key | Pay OpenAI directly |
| **Ollama** | Run locally | Free |

### Using Ollama (Free, Private)

```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a model
ollama pull codellama

# Configure c00d
```

```yaml
ai:
  provider: ollama
  model: codellama
  ollama_url: http://localhost:11434
```

## Security

For production use:

1. **Set a password** - Add `password: your-password` to config
2. **Use HTTPS** - Put behind nginx/caddy with SSL
3. **Firewall** - Restrict access to trusted IPs
4. **Don't expose publicly** without authentication
5. **IP logging** - Enabled by default, view via `/api/iplogs`

### Session Persistence

Sessions are stored in SQLite and persist across server restarts. No need to re-login after updates or reboots.

### IP Logging

All requests are logged with:
- IP address (handles X-Forwarded-For for proxies)
- Endpoint and HTTP method
- User agent
- Timestamp

Disable with `security.log_ips: false` in config.

Example nginx config:

```nginx
server {
    listen 443 ssl;
    server_name code.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

## Docker

```dockerfile
FROM alpine:latest
COPY c00d /usr/local/bin/
COPY config.yaml /etc/c00d/
EXPOSE 3000
CMD ["c00d", "-config", "/etc/c00d/config.yaml"]
```

```bash
docker build -t c00d .
docker run -p 3000:3000 -v $(pwd):/code c00d -path /code
```

## Pro License

For unlimited AI without API keys:

1. Get a license at [c00d.com/pro](https://c00d.com/pro) ($9/month)
2. Add to config:

```yaml
ai:
  provider: c00d
  license_key: your-license-key
```

## License

MIT License - use it however you want.

## Links

- Website: [c00d.com](https://c00d.com)
- GitHub: [github.com/alekblom/c00d](https://github.com/alekblom/c00d)
- Pro: [c00d.com/pro](https://c00d.com/pro)
