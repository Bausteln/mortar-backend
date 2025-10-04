FROM mcr.microsoft.com/devcontainers/go:dev-1.25-bookworm

ARG TARGETOS
ARG TARGETARCH

ENV KUBECOLORVERSION="0.4.0"

# Install kubecolor
RUN curl -L -o kubecolor.tar.gz https://github.com/kubecolor/kubecolor/releases/download/v${KUBECOLORVERSION}/kubecolor_${KUBECOLORVERSION}_linux_${TARGETARCH}.tar.gz && tar -xvzf kubecolor.tar.gz && install -c -m 0755 kubecolor /usr/local/bin

# Install bash-complete to make kubectl autocomplete work
RUN apt-get update && apt-get install bash-completion make vim -y && rm -rf /var/lib/apt/lists/*
