# Security Model & Hardening Guide

This document describes c00d's privilege model and shows how to self-host it
securely. **If you enable the PTY terminal server, please read this.**

- [Threat model in one paragraph](#threat-model-in-one-paragraph)
- [Privilege model: what can c00d touch?](#privilege-model-what-can-c00d-touch)
- [`base_path` does not sandbox the PTY](#base_path-does-not-sandbox-the-pty)
- [Known limitations in `terminal/server.js`](#known-limitations-in-terminalserverjs)
- [Recommended production setup](#recommended-production-setup)
- [Reporting vulnerabilities](#reporting-vulnerabilities)

---

## Threat model in one paragraph

c00d is a remote-access IDE with two surfaces: (1) a PHP app that reads/writes
files under `base_path` and (2) an optional Node.js PTY server that spawns a
real `bash`. If an attacker gets a valid session, they get **everything the
Unix user running those processes can touch**. The PHP surface is confined by
`base_path`; the PTY surface is confined **only by Unix permissions of the
process owner**. Treat a c00d install as equivalent to giving a shell account
on the host.

## Privilege model: what can c00d touch?

c00d does **not** install as root, and it does **not** need root. But what it
can access depends entirely on which Unix user runs each process:

| Surface | Runs as | Confined to |
|---------|---------|-------------|
| PHP app (`public/*.php`) | Whatever user your web server / FPM pool runs as | `base_path` (enforced by PHP) |
| PTY server (`terminal/server.js`) | Whatever user starts `node server.js` | **Everything that user can read/write** — `base_path` is NOT enforced here |

Common footguns:

- Running `node server.js` as the same user as your web server (e.g. `www-data`,
  or your `username` in a cPanel/CWP user) means **the PTY can read every file
  that web user can read across every site they own**.
- Running it as `root` means the PTY is a root shell. Don't.
- On shared-user hosting (cPanel/CWP), `suphp` / PHP-FPM-per-user means the
  web surface runs as your Unix account, which is often fine — but your Unix
  account typically owns all your sites, all your mail, everything. The PTY
  inherits all of that.

## `base_path` does not sandbox the PTY

The PHP side honours `base_path`: `src/IDE.php` rejects paths that escape it.
The PTY side does not, and cannot, honour it. Once you spawn `bash`, the user
can `cd /` and interact with anything Unix permissions allow.

You will see this warning in `config.php`:

```php
// NOTE: PTY terminal bypasses base_path sandbox — it gives a real
// shell as whatever user runs node server.js.
'websocket_enabled' => false,
```

If `websocket_enabled => false` (the default), there is no PTY at all and
`base_path` is effectively the sandbox boundary.

If you turn it on, **you must isolate the Node process yourself**. See
[Recommended production setup](#recommended-production-setup) below.

## Known limitations in `terminal/server.js`

These are intentional simplifications; be aware of them before exposing the
PTY to the network.

### 1. The server binds to `0.0.0.0` — no `HOST` env var

`terminal/server.js` calls `new WebSocket.Server({ port: PORT })`, which
listens on every interface. There is no `HOST` environment variable to bind
it to `127.0.0.1` only. If you want loopback-only (recommended), use one of:

- A firewall rule (CSF/ufw/firewalld) blocking the port from outside, or
- A reverse proxy in front (Apache/nginx `ProxyPass ws://127.0.0.1:3456/`) with
  the PTY port closed externally, or
- systemd's `RestrictAddressFamilies=` + socket activation on `127.0.0.1:3456`.

### 2. `ALLOWED_ORIGINS` uses substring matching

The origin check is:

```js
const allowed = ALLOWED_ORIGINS.some(o => origin && origin.includes(o));
```

That's a **substring** match, not an exact match. `ALLOWED_ORIGINS=example.com`
will accept `https://example.com.attacker.net` and
`https://example.com-phishing.io`. To be safe today:

- Put the full scheme+host in the env: `ALLOWED_ORIGINS=https://c00d.example.com`
- Don't list a bare domain
- And/or front the PTY with a reverse proxy that enforces `Origin` itself

### 3. The PTY inherits the parent process environment

`pty.spawn('bash', [], { env: {...process.env, ...} })` means any environment
variable set for the Node process (including secrets you `export`ed into that
shell) becomes available to the spawned bash. Run the service with a minimal,
explicit environment — systemd's `Environment=` + `PassEnvironment=` is a
good fit for this.

### 4. No per-connection authentication at the WebSocket layer

The only gate is the `Origin` header check. c00d's PHP app is what issues and
holds the session; the WebSocket connection itself does not verify a c00d
session cookie. If you open the PTY port to the internet, the origin check
is all that stands between an attacker and a shell.

**Mitigations:** firewall-gate the port to your IP(s), front it with a reverse
proxy that forwards cookies and validates them, or leave it listening on
loopback only and use a proxy you control.

## Recommended production setup

The hardened layout below is what we'd install on our own box. It applies
defense in depth: firewall → IP allowlist → systemd sandbox → dedicated
low-privilege user → ACL-scoped filesystem access.

### 1. Dedicated Unix user for the PTY

Do **not** run `node terminal/server.js` as your web user. Create a system
user with no login shell and nothing in its home:

```bash
sudo useradd --system --create-home --home-dir /home/c00d-term \
             --shell /sbin/nologin c00d-term
```

### 2. Grant it access to one project only (POSIX ACLs)

Instead of `chown -R`, use ACLs so your own user still owns the files:

```bash
# One-time: grant access to the existing tree
sudo setfacl -R -m u:c00d-term:rwX /path/to/your/project

# Default ACL so new files inherit the grant
sudo setfacl -R -d -m u:c00d-term:rwX /path/to/your/project

# Verify it's blocked from everything else
sudo -u c00d-term ls /home/youruser        # should fail
sudo -u c00d-term touch /etc/passwd        # should fail
sudo -u c00d-term touch /path/to/your/project/test  # should succeed
```

(Requires a filesystem that supports POSIX ACLs — ext4, xfs, and btrfs all
do by default.)

### 3. Run the PTY as a sandboxed systemd service

A ready-to-use unit file is shipped in [`contrib/c00d-terminal.service`](contrib/c00d-terminal.service).
Copy it, edit the two paths marked `CHANGE ME`, then:

```bash
sudo cp contrib/c00d-terminal.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now c00d-terminal
sudo systemctl status c00d-terminal
```

The unit enables these kernel-level protections:

- `User=/Group=c00d-term` — dedicated low-priv account
- `ProtectSystem=strict` — `/` is read-only
- `ProtectHome=read-only` — `/home` and `/root` are read-only
- `ReadWritePaths=` — only the project dir and the service's own home are writable
- `PrivateTmp=true` — isolated `/tmp`
- `NoNewPrivileges=true` — cannot gain privileges via setuid
- `ProtectKernelTunables=true` / `ProtectKernelModules=true` / `ProtectControlGroups=true`
- `RestrictSUIDSGID=true`, `LockPersonality=true`, `RestrictRealtime=true`
- `MemoryMax=512M`, `TasksMax=256`, `LimitNOFILE=4096`

Even if the PTY is compromised, an attacker is confined to `c00d-term`'s
narrow view of the filesystem **and** everything outside the project dir
is read-only at the kernel level.

### 4. Keep the PTY port off the internet

Two layers:

- **Firewall**: don't open `3456/tcp` publicly. On CSF, just don't add it
  to `TCP_IN`. On ufw/firewalld, do not allow it.
- **Reverse proxy** (optional but cleaner): forward `wss://your.host/ws/terminal`
  to `ws://127.0.0.1:3456/`. Apache example:

  ```apache
  <IfModule mod_proxy_wstunnel.c>
      ProxyPass        /ws/terminal ws://127.0.0.1:3456/
      ProxyPassReverse /ws/terminal ws://127.0.0.1:3456/
  </IfModule>
  ```

  Nginx:

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

### 5. Lock the PHP app down too

In `config.local.php`:

```php
'security' => [
    'allowed_ips'   => ['1.2.3.4'],   // your IP(s); see below for multi-IP
    'require_https' => true,
],
```

And set a password (`'password' => '...'`) or rely on the auto-generated one
in `data/.generated_password`.

### Allowing multiple IPs (mobile, roaming)

`allowed_ips` accepts a list. For a home IP + office IP:

```php
'allowed_ips' => ['203.0.113.10', '198.51.100.25'],
```

For mobile carriers or travel, IPs are dynamic. You have three honest choices:

1. **Accept the risk and drop the IP allowlist** (set `allowed_ips => []`),
   rely entirely on a strong password + HTTPS + the PTY sandbox. Fine if
   your password is random and long.
2. **VPN home** from your phone/laptop and keep the strict allowlist — this
   is the most secure option if you already run a VPN.
3. **Use a CIDR allowlist for your carrier range** — messy, leaks to every
   other customer of that carrier, not recommended.

Option 1 is usually the right trade-off for a personal IDE: your password is
not guessable, the PTY is sandboxed by systemd, and everything is behind HTTPS.

## Reporting vulnerabilities

If you find a security issue in c00d, please do **not** open a public GitHub
issue. Email [mail@c00d.com](mailto:mail@c00d.com) with:

- A description of the issue and its impact
- Steps to reproduce
- Your affected version (from `VERSION`)

We'll acknowledge within a few days and coordinate a fix + advisory.
