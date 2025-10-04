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
./kind delete clusters mortar
./kind create cluster --name mortar

helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
helm install crossplane \
--namespace crossplane-system \
--create-namespace crossplane-stable/crossplane \
--wait

helm install \
  cert-manager oci://quay.io/jetstack/charts/cert-manager \
  --version v1.18.2 \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true \
  --wait

kubectl apply -f manifests/self-signed-cert.yaml

kubectl create namespace proxy-rules
kubectl apply -f crossplane/functions --wait
kubectl wait --for=condition=healthy --timeout=300s provider/provider-kubernetes
kubectl apply -f crossplane/rp --wait

SA=$(kubectl -n crossplane-system get sa -o name | grep provider-kubernetes | sed -e 's|serviceaccount\/|crossplane-system:|g')
kubectl create clusterrolebinding provider-kubernetes-admin-binding --clusterrole cluster-admin --serviceaccount="${SA}"

kubectl apply -f manifests/test-rule.yaml
