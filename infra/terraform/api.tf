resource "openstack_blockstorage_volume_v3" "api_server_disk" {
  count                = var.vms_enabled ? var.api_server_count : 0
  name                 = "api-server-volume"
  volume_type          = "ceph-ssd"
  size                 = 10
  image_id             = local.image_id
  enable_online_resize = false
}

resource "openstack_networking_port_v2" "api_server_port" {
  count              = var.api_server_count
  network_id         = openstack_networking_network_v2.private_network.id
  security_group_ids = [openstack_networking_secgroup_v2.default_group.id]
}

resource "openstack_compute_instance_v2" "api_server" {
  count                = var.vms_enabled ? var.api_server_count : 0
  name        = "api-server-${count.index}"
  flavor_name = "d1.ram4cpu4"
  key_pair    = openstack_compute_keypair_v2.ssh.name

  user_data = file("${path.module}/resources/cloud-init.yml")

  network {
    port = openstack_networking_port_v2.api_server_port[count.index].id
  }

  block_device {
    uuid                  = openstack_blockstorage_volume_v3.api_server_disk[count.index].id
    boot_index            = 0
    source_type           = "volume"
    destination_type      = "volume"
    delete_on_termination = false
  }
}