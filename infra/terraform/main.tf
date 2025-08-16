terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 3.3.0"
    }
  }
}

locals {
  image_id = "73fc8398-34bf-46c4-91fc-53dca6f62d58" # Ubuntu-Server-24.04-LTS-amd64-202508
}

provider "openstack" {
  auth_url    = "https://auth.pscloud.io/v3/"
  region      = "kz-ala-1"
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

resource "openstack_networking_secgroup_v2" "default_group" {
  name        = "bulbul-0default"
  description = "Allow"
}

resource "openstack_networking_secgroup_rule_v2" "ssh_rule" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.default_group.id
}

resource "openstack_networking_secgroup_rule_v2" "http_rule" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.default_group.id
}

resource "openstack_networking_secgroup_rule_v2" "https_rule" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.default_group.id
}

resource "openstack_networking_secgroup_rule_v2" "internal_rule" {
  direction         = "ingress"
  ethertype         = "IPv4"
  remote_ip_prefix  = "192.168.0.0/24"
  security_group_id = openstack_networking_secgroup_v2.default_group.id
}