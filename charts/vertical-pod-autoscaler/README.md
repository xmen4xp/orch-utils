# Vertical POD Autoscaler(VPA)

This helm chart is created based on Kubernetes [vertical-pod-autoscaler](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler) with custom settings.

## Resources

Most of reasources such as service account, deployments, ... are generated with script provided by original
project.

To get resource definition, use the following command:

```bash
git clone https://github.com/kubernetes/autoscaler
cd autoscaler/vertical-pod-autoscaler
./hack/vpa-process-yamls.sh print
```

## Customization

The official guide uses a script to deploy VPA, the script contains additional step to generate TLS certificate
for `vpa-admission-controller` webhook.

This helm chart creates TLS certificate with `Certificate` CRD. The cert-manager will help creating TLS
certificate and place it to a Kubernetes secret.
See `templates/cert.yaml` for more information.

Another customization is the container parameter of `vpa-admission-controller` deployment. By default it uses
`caCert.pem`, `serverCert.pem`, and `serverKey.pem` files which is different to the name from what we created
with cert-manager.

To fix this, we add additional flags to set the correct path of certificate files.
See `templates/deployment.yaml` for more information.
