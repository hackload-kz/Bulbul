# Infra

Configure `~/.ssh/config`:

```
Host 192.168.0.*
  User ubuntu
  ProxyJump root@82.115.42.45
  Port 22
  IdentityFile ~/.ssh/id_ed25519_hackload
```
