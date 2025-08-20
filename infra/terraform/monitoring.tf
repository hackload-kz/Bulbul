
resource "openstack_networking_secgroup_v2" "monitoring_group" {
  name        = "bulbul-monitoring"
  description = "Allow"
}

resource "openstack_networking_secgroup_rule_v2" "monitoring_ssh_rule" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.monitoring_group.id
}

resource "openstack_networking_secgroup_rule_v2" "monitoring_grafana_rule" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 3000
  port_range_max    = 3000
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.monitoring_group.id
}

resource "openstack_networking_secgroup_rule_v2" "monitoring_internal_rule" {
  direction         = "ingress"
  ethertype         = "IPv4"
  remote_ip_prefix  = "192.168.0.0/24"
  security_group_id = openstack_networking_secgroup_v2.monitoring_group.id
}

resource "openstack_blockstorage_volume_v3" "monitoring_disk" {
  count                = var.vms_enabled ? 1 : 0
  name                 = "monitoring-volume"
  volume_type          = "ceph-ssd"
  size                 = 30
  image_id             = local.image_id
  enable_online_resize = false
}

resource "openstack_networking_port_v2" "monitoring_port" {
  network_id         = openstack_networking_network_v2.private_network.id
  security_group_ids = [openstack_networking_secgroup_v2.monitoring_group.id]
}

resource "openstack_compute_instance_v2" "monitoring_server" {
  count = var.vms_enabled ? 1 : 0
  name        = "monitoring"
  flavor_name = "d1.ram4cpu4"
  key_pair    = openstack_compute_keypair_v2.ssh.name

  user_data = file("${path.module}/resources/cloud-init.yml")

  network {
    port = openstack_networking_port_v2.monitoring_port.id
  }

  block_device {
    uuid                  = openstack_blockstorage_volume_v3.monitoring_disk[0].id
    boot_index            = 0
    source_type           = "volume"
    destination_type      = "volume"
    delete_on_termination = false
  }
}

resource "openstack_networking_floatingip_associate_v2" "monitoring_fip_association" {
  floating_ip = openstack_networking_floatingip_v2.monitoring_fip.address
  port_id     = openstack_networking_port_v2.monitoring_port.id
}
