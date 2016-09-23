Terraform Provider for OVH
======


First steps...

* Run tests

```bash
cd ./ovh
#TF_LOG=DEBUG
OVH_ENDPOINT=ovh-eu 
OVH_APPLICATION_KEY=.... 
OVH_APPLICATION_SECRET=.... 
OVH_VRACK=...
OVH_PUBLIC_CLOUD=...
TF_ACC=1 
OVH_CONSUMER_KEY=...
go test -v
```
