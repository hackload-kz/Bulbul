[load_balancer]
lb-1 ansible_host=${load_balancer_private_ip} ansible_user=ubuntu

[api_servers]
%{ for server in api_servers ~}
${server.name} ansible_host=${server.ip} ansible_user=ubuntu
%{ endfor ~}

[postgres]
postgres_server ansible_host=${postgres_ip} ansible_user=ubuntu

[valkey]
valkey_server ansible_host=${valkey_ip} ansible_user=ubuntu

[monitoring]
monitoring_server ansible_host=${monitoring_public_ip} ansible_user=ubuntu

[elasticsearch]
elasticsearch_server ansible_host=${elasticsearch_ip} ansible_user=ubuntu

[workloads:children]
api_servers

[all:vars]
ansible_ssh_common_args='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ProxyCommand="ssh -W %h:%p -q ubuntu@${monitoring_public_ip}"'
ansible_python_interpreter=/usr/bin/python3