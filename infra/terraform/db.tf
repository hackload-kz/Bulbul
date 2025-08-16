resource "openstack_blockstorage_volume_v3" "postgres_disk" {
  name                 = "postgres-volume"
  volume_type          = "ceph-ssd"
  size                 = 50
  image_id             = local.image_id
  enable_online_resize = false
}

resource "openstack_networking_port_v2" "postgres_port" {
  network_id         = openstack_networking_network_v2.private_network.id
  security_group_ids = [openstack_networking_secgroup_v2.default_group.id]
}

resource "openstack_compute_instance_v2" "postgres_server" {
  name        = "postgres"
  flavor_name = "d1.ram8cpu4"
  key_pair    = openstack_compute_keypair_v2.ssh.name

  user_data = file("${path.module}/resources/cloud-init.yml")

  network {
    port = openstack_networking_port_v2.postgres_port.id
  }

  block_device {
    uuid                  = openstack_blockstorage_volume_v3.postgres_disk.id
    boot_index            = 0
    source_type           = "volume"
    destination_type      = "volume"
    delete_on_termination = false
  }
}

resource "openstack_blockstorage_volume_v3" "nats_disk" {
  name                 = "nats-volume"
  volume_type          = "ceph-ssd"
  size                 = 25
  image_id             = local.image_id
  enable_online_resize = false
}

resource "openstack_blockstorage_volume_v3" "valkey_disk" {
  name                 = "valkey-volume"
  volume_type          = "ceph-ssd"
  size                 = 25
  image_id             = local.image_id
  enable_online_resize = false
}

resource "openstack_networking_port_v2" "valkey_port" {
  network_id         = openstack_networking_network_v2.private_network.id
  security_group_ids = [openstack_networking_secgroup_v2.default_group.id]
}

resource "openstack_compute_instance_v2" "valkey_server" {
  name        = "valkey"
  flavor_name = "d1.ram8cpu4"
  key_pair    = openstack_compute_keypair_v2.ssh.name

  user_data = file("${path.module}/resources/cloud-init.yml")

  network {
    port = openstack_networking_port_v2.valkey_port.id
  }

  block_device {
    uuid                  = openstack_blockstorage_volume_v3.valkey_disk.id
    boot_index            = 0
    source_type           = "volume"
    destination_type      = "volume"
    delete_on_termination = false
  }
}