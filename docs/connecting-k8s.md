# Connecting Different Kinds of Kubernetes Clusters

Currently the Stratos Kubernetes plugin supports the following four types of clusters:

1. CAASP (OIDC)
2. AWS EKS (AWS IAM auth)
2. Azure AKS 
4. Certificate based Kubernetes authentication

The following details, how to find the endpoint URL required to register the cluster in Stratos and what credentials are required to connect.

## CAASP (OIDC)
To connect a CAASP cluster to Stratos, download a `kubeconfig` from Velum.

1. To find the endpoint URL, inspect the file. The `server` property details the endpoint URL

```
apiVersion: v1
kind: Config
clusters:
- name: caasp
  cluster:
    server: https://kube-api-x1.devenv.caasp.suse.net:6443 <---Endpoint URL
    certificate-authority-data: 1c1MFpYSnVZV3dnUTBFd0hoY05NVGd4TURBMU1USXhNalU1V2hjTk1qZ3hNREF5TVRJeE1qVTVXakNCb1RFTApNQWtHQTFVRUJoTUNSRVV4RURBT0JnTlZCQWdNQjBKaGRtRnlhV0V4RWpBUUJnTlZCQWNNQ1U1MWNtVnRZbVZ5Clp6RWJNQmtHQTFVRUNnd1NVMVZUUlNCQmRYUnZaMl...
```
2. Specify the Endpoint URL when adding the endpoint to Stratos.
3. To connect to Kubernetes, select the `CAASP (OIDC)` option, and upload the `kubeconfig` file downloaded from Velum.

## Amazon EKS
The following details are required to connect to an EKS system:
- EKS Cluster endpoint URL. (To register the endpoint).

 This can be located in the generated configuration. See the following example.
To Connect the following details are required:
- Cluster Name (See the following example)
- AWS Access Key
- AWS Secret Key

### EKS Endpoint URL And Cluster Name
You can locate the EKS cluster endpoint URL and the cluster name, by inspecting the generated cluster configuration in your local `kubeconfig`. 

```
10:20 $ cat ~/.kube/config 
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa...QXR2N3dOQkt3eFhsYgpxZm5HRUs0WHRmSWNIcjJHSjhZOXdIa0lPRm0rR3Nvak1PaG1pK05wbER2YjVJc3BmMmxxbXdLd3RmRT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: https://40BCD34B7E297903DA2EAF19B6164521.sk1.us-east-1.eks.amazonaws.com
  name: arn:aws:eks:us-east-1:138384977974:cluster/BRSSCF

```
The endpoint URL is specified in the `server` property (i.e. `https://40BCD34B7E297903DA2EAF19B6164521.sk1.us-east-1.eks.amazonaws.com`), while the cluster name is the last part of the `name` property (i.e `BRSSCF`).

## Azure AKS 
To connect an AKS kubernetes instance, the following is required:
1. AKS Endpoint URL, which can be found from the AKS console or the generated kubernetes configuration.
2. To connect to the cluster, provide the `kubeconfig` file.

## Certificate based authentication (Minikube)

Minikube by default uses TLS certificates for authentication. To find the Minikube endpoint URL, locate the `minikube` entry in your local `kubeconfig` file. In the following example, the `minikube` endpoint URL is `https://192.168.99.100:8443`.

```
- cluster:
    certificate-authority: /home/user/.minikube/ca.crt
    server: https://192.168.99.100:8443
  name: minikube
```

To connect to the cluster, locate the relevant entry in the `users` section in your kubernetes config file.

```
users:
- name: minikube
  user:
    client-certificate: /home/user/.minikube/client.crt
    client-key: /home/user/.minikube/client.key

```
The two files specified under `client-certificate` and `client-key` are required to connect to the cluster.
Select the `Kubernetes Cert Auth` option in the connect dialog and select the two files to connect.