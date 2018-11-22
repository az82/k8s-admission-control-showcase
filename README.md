# Kubernetes Admission Control and Archiving Showcase

Deploys a custom Kubernetes validating admission controller. 
That controller will review deployments for namespaces with the label `admission-webhook: enabled`.

At the moment, the implementation denies all requests.

## Prerequisites
- GNU Make
- Kubectl
- Docker
- Go 1.11
- OpenSSL
- [Base64](https://www.fourmilab.ch/webtools/base64/)
- Kubernetes Cluster
    
    The Makefile assumes that `docker build` will install the image in the target cluster's registry.
    This is the case for Docker Desktop and Minikube, but not for remote clusters.

## How to Use

1. Build & Deploy the Webhook

    ```bash
    make deploy
    ```


2. Try to deploy an application
    
    ```bash
    kubectl apply -f hello-world.yaml
    ```
    
3. You should get the following error message

    ```
    Error from server (Forbidden): error when creating “hello-world-deployment.yaml”: admission webhook “test-validating-webhook.az82.de” denied the request: You’re not getting in with these shoes.
    ```
    


