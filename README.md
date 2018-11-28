# Kubernetes Admission Control and Archiving Showcase

Deploys a custom Kubernetes validating admission controller. 
The controller will review deployments for namespaces with the label `admission-webhook: enabled`.

The controller then uses an OPA (OpenPolicyAgent) sidecar to decide what to do with the deployment.

That step of indirection allows us to use the
[OPA data API](https://www.openpolicyagent.org/docs/rest-api.html#data-api)
instead of the Query API. The data API is much more convenient to manage.

OPA can not only allow or deny an admission, it can also provide more advanced policies, most importantly control taking
actions upon deployments.

The policies define a set of processors that can be called out-of-band if the deployment is allowed. The deployment
won't be slowed down by steps that can take a lot of time. This approach makes it easy to fulfil auditing or archiving
requirements.

In the future, things like max/min resource limits, number of instances and so on could also be controlled by
central OPA policies.

## Prerequisites

- GNU Make
- Kubectl
- Docker
- Go > 1.11 (Because we are using [Go modules](https://github.com/golang/go/wiki/Modules))
- OpenSSL
- [Base64](https://www.fourmilab.ch/webtools/base64/)
- Kubernetes Cluster

    The Makefile assumes that `docker build` will install the image in the target cluster's registry.
    This is the case for Docker Desktop and Minikube, but not for remote clusters.

- [OPA](https://www.openpolicyagent.org/) (Only needed for manual local testing)

## How to Use

1. Build & Deploy the Web hook

    ```bash
    make deploy
    ```

2. Try to deploy an application that does not meet the policy

    ```bash
    kubectl apply -f test/deployments/invalid.yaml
    ```

3. You should get the following error message

    ```text
    Error from server (Forbidden): error when creating "test/deployments/invalid.yaml": admission webhook "test-validating-webhook.az82.de" denied the request: No explicit image version for the container hello-kubernetes, Invalid Git repository annotation, Invalid Git commit hash annotation
    ```

4. Try to deploy an application that meets the policy

    ```bash
    kubectl apply -f test/deployments/valid.yaml
    ```

5. [Inspect the policies](policies). You can then try to create a deployment that fulfils the policies or try to tweak the policies.

## Cleaning up

- Undeploy everything

    ```bash
    make undeploy
    ```

- Clean the workspace

    ```bash
    make clean
    ```

## See also

- [Kubernetes Admission Control Documentation](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
- [OPA Documentation](https://www.openpolicyagent.org/docs/)

## Pitfalls

Kubernetes currently does not support creating config maps recursively from a directory. That means that policies
stored in a config map cannot be organized in directories.
See 
[kubernetes/kubernetes#62421](https://github.com/kubernetes/kubernetes/issues/62421)
for reference
