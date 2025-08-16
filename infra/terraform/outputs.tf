output "load_balancer_public_ip" {
  value = openstack_networking_floatingip_v2.instance_fip.address
}

output "load_balancer_private_ip" {
  value = openstack_networking_port_v2.lb_port.all_fixed_ips[0]
}