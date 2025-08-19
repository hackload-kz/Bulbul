output "load_balancer_public_ip" {
  value = var.vms_enabled ? openstack_networking_floatingip_v2.lb_fip.address : null
  description = "Load balancer public IP address"
}

output "load_balancer_private_ip" {
  value = var.vms_enabled ? openstack_networking_port_v2.lb_port.all_fixed_ips[0] : null
  description = "Load balancer private IP address"
}

output "monitoring_public_ip" {
  value = var.vms_enabled ? openstack_networking_floatingip_v2.monitoring_fip.address : null
  description = "Monitoring server public IP address (SSH jump host)"
}

output "monitoring_private_ip" {
  value = var.vms_enabled ? openstack_networking_port_v2.monitoring_port.all_fixed_ips[0] : null
  description = "Monitoring server private IP address"
}

output "api_servers" {
  value = var.vms_enabled ? [for i in range(var.api_server_count) : {
    name = "api-server-${i}"
    ip   = openstack_networking_port_v2.api_server_port[i].all_fixed_ips[0]
  }] : []
  description = "API servers with their private IPs"
}

output "postgres_ip" {
  value = var.vms_enabled ? openstack_networking_port_v2.postgres_port.all_fixed_ips[0] : null
  description = "PostgreSQL server private IP address"
}

output "valkey_ip" {
  value = var.vms_enabled ? openstack_networking_port_v2.valkey_port.all_fixed_ips[0] : null
  description = "Valkey server private IP address"
}