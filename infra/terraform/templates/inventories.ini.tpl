[load_balancer]
lb-1 ansible_host=${load_balancer_public_ip} ansible_user=ubuntu

[api_servers]
%{ for server in api_servers ~}
${server.name} ansible_host=${server.ip} ansible_user=ubuntu
%{ endfor ~}

[postgres]
${postgres_ip}

[valkey]
${valkey_ip}

[all:vars]
ansible_ssh_common_args='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ProxyCommand="ssh -W %h:%p -q ubuntu@${load_balancer_public_ip}"'
