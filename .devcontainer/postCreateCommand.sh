#!/bin/bash

mkdir -p "$HOME/.kube"
kubectl completion bash > /home/vscode/.kube/completion.bash.inc
printf "
source /usr/share/bash-completion/bash_completion
source "$HOME/.kube/completion.bash.inc"
complete -F __start_kubectl k
" >> "$HOME/.bashrc"

printf "
source <(kubectl completion zsh)
complete -F __start_kubectl k
" >> "$HOME/.zshrc"

(
  set -x; cd "$(mktemp -d)" &&
  OS="$(uname | tr '[:upper:]' '[:lower:]')" &&
  ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" &&
  KREW="krew-${OS}_${ARCH}" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" &&
  tar zxvf "${KREW}.tar.gz" &&
  ./"${KREW}" install krew
)

sudo ln -s "$HOME/.krew/bin/kubectl-krew" /usr/local/bin/kubectl-krew

kubectl krew install view-serviceaccount-kubeconfig

sudo ln -s "$HOME/.krew/bin/kubectl-view_serviceaccount_kubeconfig" /usr/local/bin/kubectl-view_serviceaccount_kubeconfig

# Setup stuff
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.30.0/kind-linux-amd64
chmod +x ./kind
./kind create cluster --name motar

helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
helm install crossplane \
--namespace crossplane-system \
--create-namespace crossplane-stable/crossplane \
--wait


git clone git@gitlab.bausteln.ch:net-core/kubernetes-projects/crossplane.git
