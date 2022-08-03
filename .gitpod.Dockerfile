# Test with:
# > docker build -f .gitpod.Dockerfile -t gitpod-dockerfile-test .

FROM gitpod/workspace-go:latest

# Useful tools
RUN \
  sudo apt-get update && \
  sudo apt-get install -y iputils-ping dnsutils mc rsync

# Install Microsoft AZ CLI tool
RUN \
  curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# Install `kubectl` and `helm`
RUN \
  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
  sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
  rm kubectl && \
  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install kubelogin, the azure auth is getting obsolete
RUN \
    curl -LO 'https://github.com/Azure/kubelogin/releases/download/v0.0.18/kubelogin-linux-amd64.zip' && \
    unzip kubelogin-linux-amd64.zip && \
    rm -f kubelogin-linux-amd64.zip && \
    sudo mv bin/linux_amd64/kubelogin /usr/local/bin/

# Install the latest official Tailscale version
USER root
RUN curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/focal.gpg | apt-key add - \
  && curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/focal.list | tee /etc/apt/sources.list.d/tailscale.list \
  && apt-get update \
  && apt-get install -y tailscale \
  && update-alternatives --set ip6tables /usr/sbin/ip6tables-nft
USER gitpod
