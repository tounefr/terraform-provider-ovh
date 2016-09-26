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

* Example with working resources

```terraform
resource "ovh_vrack_publiccloud_attachment" "attach" {
  vrack_id   = "${var.vrack_id}"
  project_id = "${var.project_id}"
}

resource "ovh_publiccloud_private_network" "mynetwork" {
  project_id  = "${ovh_vrack_publiccloud_attachment.attach.project_id}"
  vlan_id     = 0
  name        = "terraform_testacc_private_net"
  regions     = ["GRA1", "BHS1"]
}

resource "ovh_publiccloud_private_network_subnet" "mysubnet" {
  project_id = "${ovh_publiccloud_private_network.mynetwork.project_id}"
  network_id = "${ovh_publiccloud_private_network.mynetwork.id}"
  region     = "GRA1"
  start      = "192.168.168.100"
  end        = "192.168.168.200"
  network    = "192.168.168.0/24"
  dhcp       = true
  no_gateway = false
}

resource "ovh_publiccloud_user" "myuser" {
  project_id  = "${ovh_publiccloud_private_network.mynetwork.project_id}"
  description = "my openstack user"
}
```
