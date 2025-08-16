resource "openstack_blockstorage_volume_v3" "lb_disk" {
  name                 = "lb-volume"
  volume_type          = "ceph-ssd"
  size                 = 15
  image_id             = local.image_id
  enable_online_resize = false
}

# resource "openstack_compute_instance_v2" "lb" {
#   name              = "load-balancer"
#   flavor_name       = "d1.ram1cpu1"
#   key_pair          = openstack_compute_keypair_v2.ssh.name

#   user_data       = file("${path.module}/resources/cloud-init.yml")

#   network {
#     port = openstack_networking_port_v2.lb_port.id
#   }

#   block_device {
#     uuid                  = openstack_blockstorage_volume_v3.lb_disk.id
#     boot_index            = 0
#     source_type           = "volume"
#     destination_type      = "volume"
#     delete_on_termination = false
#   }
# }

# resource "openstack_networking_port_v2" "lb_port" {
#   network_id         = openstack_networking_network_v2.private_network.id
#   security_group_ids = [openstack_networking_secgroup_v2.default_group.id]
# }

# resource "openstack_networking_floatingip_associate_v2" "instance_fip_association" {
#   floating_ip = openstack_networking_floatingip_v2.instance_fip.address
#   port_id     = openstack_networking_port_v2.lb_port.id
# }