# k8s-aks-dns-ingress

[![Travis CI](https://travis-ci.org/jessfraz/k8s-aks-dns-ingress.svg?branch=master)](https://travis-ci.org/jessfraz/k8s-aks-dns-ingress)

An ingress controller.

## Installation

#### Binaries

- **darwin** [386](https://github.com/jessfraz/k8s-aks-dns-ingress/releases/download/v0.0.0/k8s-aks-dns-ingress-darwin-386) / [amd64](https://github.com/jessfraz/k8s-aks-dns-ingress/releases/download/v0.0.0/k8s-aks-dns-ingress-darwin-amd64)
- **freebsd** [386](https://github.com/jessfraz/k8s-aks-dns-ingress/releases/download/v0.0.0/k8s-aks-dns-ingress-freebsd-386) / [amd64](https://github.com/jessfraz/k8s-aks-dns-ingress/releases/download/v0.0.0/k8s-aks-dns-ingress-freebsd-amd64)
- **linux** [386](https://github.com/jessfraz/k8s-aks-dns-ingress/releases/download/v0.0.0/k8s-aks-dns-ingress-linux-386) / [amd64](https://github.com/jessfraz/k8s-aks-dns-ingress/releases/download/v0.0.0/k8s-aks-dns-ingress-linux-amd64) / [arm](https://github.com/jessfraz/k8s-aks-dns-ingress/releases/download/v0.0.0/k8s-aks-dns-ingress-linux-arm) / [arm64](https://github.com/jessfraz/k8s-aks-dns-ingress/releases/download/v0.0.0/k8s-aks-dns-ingress-linux-arm64)
- **solaris** [amd64](https://github.com/jessfraz/k8s-aks-dns-ingress/releases/download/v0.0.0/k8s-aks-dns-ingress-solaris-amd64)
- **windows** [386](https://github.com/jessfraz/k8s-aks-dns-ingress/releases/download/v0.0.0/k8s-aks-dns-ingress-windows-386) / [amd64](https://github.com/jessfraz/k8s-aks-dns-ingress/releases/download/v0.0.0/k8s-aks-dns-ingress-windows-amd64)

## Usage

```console
$ k8s-aks-dns-ingress -h
k8s-aks-dns-ingress
An ingress controller.
Version: v0.0.0
  -alsologtostderr
        log to standard error as well as files
  -azureconfig string
        Azure service principal configuration file (eg. path to azure.json, defaults to the value of 'AZURE_AUTH_LOCATION' env var
  -d    run in debug mode
  -domain string
        Root domain name to use for the creating the DNS record sets, defaults to the value of 'DOMAIN_NAME_ROOT' env var
  -interval string
        Controller resync period (default "30s")
  -kubeconfig string
        Path to kubeconfig file with authorization and master location information (default is $HOME/.kube/config)
  -log_backtrace_at value
        when logging hits line file:N, emit a stack trace
  -log_dir string
        If non-empty, write log files in this directory
  -logtostderr
        log to standard error instead of files
  -namespace string
        Kubernetes namespace to watch for ingress (default is to watch all namespaces)
  -region string
        Azure region, defaults to the value of 'AZURE_REGION' env var
  -resource string
        Azure resource name, defaults to the value of 'AZURE_RESOURCE_NAME' env var
  -resource-group string
        Azure resource group name, defaults to the value of 'AZURE_RESOURCE_GROUP' env var
  -stderrthreshold value
        logs at or above this threshold go to stderr
  -v value
        log level for V logs
  -version
        print version and exit
  -vmodule value
        comma-separated list of pattern=N settings for file-filtered logging
```
