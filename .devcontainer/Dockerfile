FROM mcr.microsoft.com/devcontainers/javascript-node:bookworm

COPY example-welcome-message.txt /usr/local/etc/vscode-dev-containers/first-run-notice.txt

# Install latest pnpm
USER node
ENV SHELL=bash
RUN curl -fsSL https://get.pnpm.io/install.sh | sh -