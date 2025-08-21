# Infra

Configure `~/.ssh/config`:

```
Host 192.168.0.*
  User ubuntu
  ProxyJump ubuntu@91.147.93.57
  Port 22
  IdentityFile ~/.ssh/id_ed25519_hackload
```

Create ansible vault password file `infra/.ansible_password` and put there
vault password.

Init environment variables:

```
cd infra
source ./init.sh
```

Create resources using terraform:

```
cd infra/terraform
terraform apply
```

Run ansible playbook:

```
cd infra
ansible-playbook playbook.yml
```