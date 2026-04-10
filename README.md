# c00d

**A web-first, self-hosted code editor. Code from anywhere - phone, tablet, or desktop.**

c00d is an open-source IDE you install on your server and access via browser. Edit files, run commands, and use AI assistance - all from any device.

## Why c00d?

- **Web-first** - Works on mobile, tablet, desktop. No app to install.
- **Self-hosted** - Your code stays on your server. Full control.
- **AI-powered** - Built-in AI assistant for code help (optional)
- **Zero config** - Just PHP. No complex setup.

## Features

- **Monaco Editor** - Same editor as VS Code
- **File Browser** - Navigate, create, rename, delete files
- **Integrated Terminal** - Run commands in your browser (with optional PTY for interactive CLI tools)
- **AI Assistant** - Explain code, fix bugs, generate tests
- **Auto-Updates** - Check for and install updates from within the IDE
- **Mobile Friendly** - Full functionality on any device
- **Password Protected** - Optional authentication
- **SQLite Storage** - Preferences and history saved locally

## Quick Start

```bash
# Download
curl -L https://c00d.com/download -o c00d.zip
unzip c00d.zip

# Configure (optional)
cp config.php config.local.php
nano config.local.php  # Set password, AI provider, etc.

# Access via browser
# Point your web server to the 'public' folder
# Open https://yourserver.com/c00d/
```

## Requirements

- PHP 8.0+
- SQLite extension (usually included)
- cURL extension (for AI features)

## AI Providers

c00d supports multiple AI backends. Choose what works for you:

| Provider | Setup | Cost |
|----------|-------|------|
| **c00d API** | Just works | Free: 20/day, Pro: $9/mo unlimited |
| **Claude CLI** | Install CLI, run `claude login` | Uses your Claude subscription |
| **Anthropic API** | Add API key | Pay Anthropic directly |
| **OpenAI API** | Add API key | Pay OpenAI directly |
| **Ollama** | Run Ollama locally | Free, fully private |

### Using Claude CLI (for Claude subscribers)

If you have a Claude Max or Pro subscription:

```bash
# Install Claude CLI on your server
npm install -g @anthropic-ai/claude-code

# Authenticate
claude login

# Configure c00d to use it
# In config.local.php:
'ai' => [
    'provider' => 'claude-cli',
]
```

### Using Ollama (free, private)

```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a model
ollama pull codellama

# Configure c00d
'ai' => [
    'provider' => 'ollama',
    'model' => 'codellama',
]
```

## Configuration

Create `config.local.php` (won't be overwritten on updates):

```php
<?php
return [
    // Restrict to a specific directory
    'base_path' => '/var/www/myproject',

    // Password protection (recommended!)
    'password' => 'your-secure-password',

    // AI settings
    'ai' => [
        'provider' => 'c00d',      // or: claude-cli, anthropic, openai, ollama
        'license_key' => '',        // c00d Pro license for unlimited AI
        'api_key' => '',            // For anthropic/openai providers
    ],

    // Editor
    'editor' => [
        'theme' => 'vs-dark',       // vs-dark, vs-light, hc-black
        'font_size' => 14,
        'tab_size' => 4,
    ],

    // Security
    'security' => [
        'allowed_ips' => [],        // Restrict access by IP
        'require_https' => true,
    ],
];
```

## File Structure

```
c00d/
├── config.php           # Default config
├── config.local.php     # Your settings (gitignored)
├── VERSION              # Current version number
├── public/
│   ├── index.php        # IDE interface
│   └── api.php          # API endpoint
├── src/
│   ├── Database.php     # SQLite handler
│   ├── IDE.php          # File/terminal operations
│   └── AI.php           # AI integration
├── terminal/            # Optional PTY terminal server
│   ├── server.js        # WebSocket terminal server
│   └── package.json     # Node.js dependencies
└── data/
    └── c00d.db          # SQLite database (auto-created)
```

## Terminal

c00d includes a hybrid terminal that automatically uses the best available method:

| Mode | What works | Setup needed |
|------|------------|--------------|
| **Full PTY** | Everything (`claude`, `vim`, `htop`, etc.) | Terminal server + connection |
| **Basic** | Simple commands (`ls`, `git`, `cat`, etc.) | None |

### Quick Start (Easiest)

The terminal server can be started directly from the IDE:

1. Open Terminal in c00d
2. Click **"▶ Start Terminal Server"**
3. If using **direct port** method, open port 3456 in your firewall

### Connection Methods

The IDE tries these in order:

1. **WebSocket Proxy** (`/ws/terminal`) - Requires web server config
2. **Direct Port** (`:3456`) - Just needs firewall port open
3. **Basic Mode** - Works everywhere, no interactive programs

### Option A: Direct Port (Recommended for simplicity)

Just open the port in your firewall:

```bash
# Ubuntu/Debian
sudo ufw allow 3456

# CentOS/RHEL
sudo firewall-cmd --add-port=3456/tcp --permanent
sudo firewall-cmd --reload
```

Then start the server from the IDE or manually:

```bash
cd terminal && npm install && node server.js
```

### Option B: WebSocket Proxy (Cleaner URLs)

**Nginx:**
```nginx
location /ws/terminal {
    proxy_pass http://127.0.0.1:3456;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_read_timeout 86400;
}
```

**Apache** (requires `mod_proxy_wstunnel`):
```apache
ProxyPass /ws/terminal ws://127.0.0.1:3456/
ProxyPassReverse /ws/terminal ws://127.0.0.1:3456/
```

### Running as a Service

**Using pm2 (recommended):**
```bash
npm install -g pm2
pm2 start /path/to/c00d/terminal/server.js --name c00d-terminal
pm2 save
pm2 startup
```

**Using systemd (recommended for production):**

A ready-to-use, hardened unit file is shipped in
[`contrib/c00d-terminal.service`](contrib/c00d-terminal.service). It runs the
PTY as a dedicated low-privilege user, confines it to your project directory
with POSIX ACLs, and uses systemd's `ProtectSystem=strict` +
`ProtectHome=read-only` to make the rest of the filesystem untouchable even
if the process is compromised.

```bash
# 1. Create a dedicated system user
sudo useradd --system --create-home --home-dir /home/c00d-term \
             --shell /sbin/nologin c00d-term

# 2. Grant it access to your project only (don't chown — use ACLs)
sudo setfacl -R -m u:c00d-term:rwX /path/to/your/project
sudo setfacl -R -d -m u:c00d-term:rwX /path/to/your/project

# 3. Install and start the service (edit paths inside first)
sudo cp contrib/c00d-terminal.service /etc/systemd/system/
sudo $EDITOR /etc/systemd/system/c00d-terminal.service
sudo systemctl daemon-reload
sudo systemctl enable --now c00d-terminal
```

**Do not** run the PTY as `www-data` or your web user — the PTY spawns a real
`bash` and inherits that user's full filesystem access, including every other
site they own. See [SECURITY.md](SECURITY.md) for the full privilege model.

### Configuration

In `config.local.php`:
```php
'terminal' => [
    'server_port' => 3456,  // Change if needed
],
```

## Security

- **Always set a password** for production use
- **Use HTTPS** - set `require_https` to true
- **Restrict IPs** if possible via `allowed_ips`
- **Protect data folder** - `.htaccess` included, but verify
- Don't expose on public internet without authentication
- **If you enable the PTY terminal, read [SECURITY.md](SECURITY.md)** — it
  documents the privilege model (the PTY is a real shell as whatever user
  runs `node server.js`, and it is NOT confined by `base_path`) and shows
  how to isolate it with a dedicated user + systemd hardening.

### Session Handling

Sessions use a cryptographically secure 256-bit random hash. Sessions persist across IP changes (useful for mobile users switching networks).

## Mobile Usage

c00d is designed to work on mobile devices:

- Responsive layout adapts to screen size
- Touch-friendly file browser
- On-screen keyboard works with terminal
- Monaco editor supports mobile editing

**Tip:** Add to home screen for app-like experience.

## Keyboard Shortcuts

All shortcuts use `Alt` to avoid conflicting with browser shortcuts (print, zoom, etc.)

| Shortcut | Action |
|----------|--------|
| `Alt+S` | Save file |
| `Alt+P` | Quick open (file search) |
| `Alt+B` | Toggle sidebar |
| `Alt+T` | Toggle terminal |
| `Alt+G` | Toggle Git panel |
| `Alt+A` | Toggle AI chat |

## Pro License

For unlimited AI access without API keys:

1. Go to [c00d.com/pro](https://c00d.com/pro)
2. Subscribe for $9/month
3. Add license key to config:

```php
'ai' => [
    'provider' => 'c00d',
    'license_key' => 'your-license-key',
]
```

## Contributing

Contributions welcome! Please:

1. Fork the repo
2. Create a feature branch
3. Submit a pull request

## License

MIT License - use it however you want.

The code is open source. The "c00d" name and logo are trademarks.

## Links

- **Website:** [c00d.com](https://c00d.com)
- **Docs:** [c00d.com/docs](https://c00d.com/docs)
- **Security:** [SECURITY.md](SECURITY.md) — read this before enabling the PTY
- **Pro License:** [c00d.com/pro](https://c00d.com/pro)
- **GitHub:** [github.com/alekblom/c00d](https://github.com/alekblom/c00d)
- **Issues:** [github.com/alekblom/c00d/issues](https://github.com/alekblom/c00d/issues)

---

Made with code by humans (and AI).
