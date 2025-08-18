resource "local_file" "ansible_inventory" {
  count = var.vms_enabled ? 1 : 0
  content = templatefile("${path.module}/templates/inventories.ini.tpl", {
    load_balancer_public_ip  = openstack_networking_floatingip_v2.instance_fip.address
    load_balancer_private_ip = openstack_networking_port_v2.lb_port.all_fixed_ips[0]
    api_servers = [for i in range(var.api_server_count) : {
      name = "api-server-${i}"
      ip   = openstack_networking_port_v2.api_server_port[i].all_fixed_ips[0]
    }]

    postgres_ip = openstack_networking_port_v2.postgres_port.all_fixed_ips[0]
    valkey_ip   = openstack_networking_port_v2.valkey_port.all_fixed_ips[0]
  })
  filename = "${path.module}/../inventories.ini"
}
