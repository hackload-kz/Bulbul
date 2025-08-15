terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 3.3.0"
    }
  }
}

provider "openstack" {
  auth_url    = "https://auth.pscloud.io/v3/"
  region      = "kz-ala-1"
}

variable "image_id" {
  default = "22e935a1-dffe-43d5-939f-98b5a2c92771"
}

resource "openstack_compute_keypair_v2" "ssh" {
  name       = "admin"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBExSo7X/+ReMQquqrzjQQ265bvYKRgS8z9Oo1j/7hCf admin"
}

resource "openstack_networking_network_v2" "private_network" {
  name           = "internal"
  admin_state_up = true
}

resource "openstack_networking_subnet_v2" "private_subnet" {
  name             = "subnet_name"
  network_id       = openstack_networking_network_v2.private_network.id
  cidr             = "192.168.0.0/24"
  dns_nameservers  = ["195.210.46.195", "195.210.46.132"]
  ip_version       = 4
  enable_dhcp      = true
}

resource "openstack_networking_floatingip_v2" "instance_fip" {
  pool = "FloatingIP Net"
}

resource "openstack_networking_router_v2" "router" {
  name                = "router_name"
  external_network_id = "83554642-6df5-4c7a-bf55-21bc74496109" # Floating IP network UUID
  admin_state_up      = true

  depends_on = [
    openstack_networking_network_v2.private_network
  ]
}

resource "openstack_networking_router_interface_v2" "router_interface" {
  router_id = openstack_networking_router_v2.router.id
  subnet_id = openstack_networking_subnet_v2.private_subnet.id
}

resource "openstack_networking_secgroup_v2" "security_group" {
  name        = "ssh"
  description = "Allow SSH"
}

resource "openstack_networking_secgroup_rule_v2" "ssh_rule" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.security_group.id
}

# resource "openstack_blockstorage_volume_v3" "disk" {
#   name                 = "volume_name"
#   volume_type          = "ceph-ssd" # Available: ceph-ssd, ceph-hdd, ceph-backup
#   size                 = 25
#   image_id             = var.image_id
#   enable_online_resize = true
# }

# resource "openstack_compute_instance_v2" "instance" {
#   name              = "instance_name"
#   flavor_name       = "d1.ram2cpu1"
#   key_pair          = openstack_compute_keypair_v2.ssh.name
#   security_groups   = [openstack_compute_secgroup_v2.security_group.name]

#   network {
#     uuid = openstack_networking_network_v2.private_network.id
#   }

#   block_device {
#     uuid                  = openstack_blockstorage_volume_v3.disk.id
#     boot_index            = 0
#     source_type           = "volume"
#     destination_type      = "volume"
#     delete_on_termination = false
#   }

#   depends_on = [
#     openstack_networking_network_v2.private_network,
#     openstack_blockstorage_volume_v3.disk
#   ]
# }

# resource "openstack_compute_floatingip_associate_v2" "instance_fip_association" {
#   floating_ip = openstack_networking_floatingip_v2.instance_fip.address
#   instance_id = openstack_compute_instance_v2.instance.id
#   fixed_ip    = openstack_compute_instance_v2.instance.access_ip_v4
# }