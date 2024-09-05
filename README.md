# The Society of Engineering Website

## 🧞 Commands

All commands are run from the root of the project, from a terminal:

| Command                   | Action                                           |
| :------------------------ | :----------------------------------------------- |
| `pnpm install`             | Installs dependencies                            |
| `pnpm run dev`             | Starts local dev server at `localhost:4321`      |
| `pnpm run build`           | Build your production site to `./dist/`          |
| `pnpm run preview`         | Preview your build locally, before deploying     |
| `pnpm run astro ...`       | Run CLI commands like `astro add`, `astro check` |
| `pnpm run astro -- --help` | Get help using the Astro CLI                     |

## Backend

Located in the `backend` directory, the backend is built with Go and uses Clover as a database. To run the backend, navigate to the `backend` directory and run `go run .`.

### Deployment

1. Build the frontend
- Install JavaScript dependencies with `pnpm install`
- Build the frontend with `pnpm run build`
- Host the contents of the `dist` directory on a static file server
2. Build the backend
- Navigate to the `backend` directory
- The build output can be run by `./contact-form`
3. Proxy the backend to the /api endpoint
- Use a reverse proxy to forward requests to the `/api` endpoint to the backend server 
Nginx example:
```nginx
server {
 listen 443 ssl http2;
 listen [::]:443 ssl http2;
 add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always; # HSTS: Optional
 ssl_certificate /path/to/fullchain.crt;
 ssl_certificate_key /path/to/private.key;
 server_name societyofengineering.alphabyte.pw;
 root /path/to/dist/folder/contents; # Path to the frontend build output
 location /api {
        proxy_pass http://127.0.0.1:6583; # Backend server address
        proxy_set_header STE-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
 }
}
```

- The backend can be run using SystemD 
```systemd
[Unit]
Description=Society of Engineering website contact form
Documentation=https://github.com/1alphabyte/SofE-website/tree/main/backend
After=network.target

[Service]
Type=simple
User=user
WorkingDirectory=/path/to/folder/containing/contact-form
Environment="PORT=6583"
ExecStart=/path/to/contact-form
Restart=on-failure

[Install]
WantedBy=multi-user.target
```