# VoidFS

VoidFS is a small web file manager for a Linux server. It is meant for the awkward case where you have a VPS but only a browser, such as an old iPad that cannot install an SFTP client.

It browses the server filesystem, uploads and downloads files, and includes a code editor. Login goes through Linux PAM, so VoidFS uses an existing server account instead of keeping a separate password database.

## What works

- Browse the filesystem from `/` or a configured root directory
- Open, edit, and save text files
- Upload and download files
- Create folders
- Rename and delete files or folders
- Jump through directories with clickable breadcrumbs
- Switch between full file view, split view, and full editor view
- Sign in with a Linux account through PAM
- Use the interface from tablet and mobile browsers

## Requirements

- Linux with PAM
- Go 1.22 or newer
- PAM development headers when building from source

On Ubuntu or Debian:

```bash
sudo apt update
sudo apt install -y golang-go libpam0g-dev
```

## Install

Clone the repository and build the binary:

```bash
git clone https://github.com/YOUR_USERNAME/voidfs.git
cd voidfs
go build -o bin/voidfs ./cmd/server
```

Install the PAM service definition:

```bash
sudo install -m 0644 configs/pam-voidfs /etc/pam.d/voidfs
```

Start VoidFS:

```bash
./bin/voidfs
```

It listens on port `8787` by default:

```text
http://SERVER_IP:8787
```

Log in with the Linux account configured by `APP_ALLOWED_USER`. The default is `root`.

## Configuration

VoidFS reads configuration from environment variables.

| Variable | Default | Purpose |
|---|---:|---|
| `APP_ADDR` | `:8787` | HTTP listen address |
| `APP_ROOT_DIR` | `/` | Highest directory visible in the file manager |
| `APP_ALLOWED_USER` | `root` | Linux user allowed to sign in |
| `APP_SESSION_SECRET` | `change-me` | HMAC key used to sign session cookies |
| `APP_MAX_UPLOAD_BYTES` | `10485760` | Maximum upload size in bytes |
| `APP_MAX_EDIT_BYTES` | `1048576` | Largest file that can be opened in the editor |

Example:

```bash
export APP_ADDR=127.0.0.1:8787
export APP_ROOT_DIR=/srv
export APP_ALLOWED_USER=deploy
export APP_SESSION_SECRET="$(openssl rand -hex 32)"
./bin/voidfs
```

VoidFS does not load `.env` files itself. Export the variables in your shell or provide them through systemd, Docker, or your process manager.

## Run as a systemd service

Create `/etc/systemd/system/voidfs.service`:

```ini
[Unit]
Description=VoidFS web file manager
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/voidfs
ExecStart=/opt/voidfs/bin/voidfs
Environment=APP_ADDR=127.0.0.1:8787
Environment=APP_ROOT_DIR=/
Environment=APP_ALLOWED_USER=root
Environment=APP_SESSION_SECRET=replace-this-with-a-long-random-value
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Then enable it:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now voidfs
sudo systemctl status voidfs
```

## Put HTTPS in front of it

VoidFS accepts a Linux password on the login form. Do not expose it over plain HTTP on the public internet.

Bind VoidFS to localhost:

```bash
APP_ADDR=127.0.0.1:8787 ./bin/voidfs
```

Then proxy it through Caddy, Nginx, or another HTTPS reverse proxy. A minimal Caddy config looks like this:

```caddyfile
files.example.com {
    reverse_proxy 127.0.0.1:8787
}
```

## Security notes

VoidFS runs with the permissions of its process. If you run it as `root` with `APP_ROOT_DIR=/`, a browser session can read, edit, rename, or delete almost anything on the server.

For a safer setup:

- Run VoidFS as a dedicated Linux user.
- Set `APP_ROOT_DIR` to a directory such as `/srv/files`.
- Use a random `APP_SESSION_SECRET`.
- Serve it through HTTPS.
- Keep it behind a VPN or firewall when possible.
- Back up anything you cannot afford to delete.

PAM authentication checks who may log in. It does not reduce the filesystem privileges of a process running as root.

## Development

Run the server:

```bash
go run ./cmd/server
```

Run tests:

```bash
go test ./...
```

Build the binary:

```bash
./scripts/build.sh
```

The frontend uses server-rendered HTML, plain JavaScript, and a vendored CodeMirror 5 build. There is no Node.js build step.

## Project layout

```text
cmd/server/           application entry point
internal/auth/        PAM authentication
internal/config/      environment configuration
internal/editor/      read and save text files
internal/files/       filesystem operations
internal/server/      routes, sessions, middleware
internal/upload/      multipart uploads
internal/views/       HTML templates
web/static/           CSS, JavaScript, CodeMirror assets
tests/                integration and service tests
```

## License

VoidFS is available under the [MIT License](LICENSE).
