# About

This is a server [LSB init script](https://wiki.debian.org/LSBInitScripts) to make sure the TeamWork server binary starts automatically with every server reboot.

# Installation

Copy it as root/sudo into /etc/init.d as teamwork-server.sh and edit lines 15-25 with correct values for your environment.

Then install it via update-rc.d to start automatically on boot:

```sh
# cp teamwork-server-model.sh /etc/init.d/teamwork-server.sh
# chmod 755 /etc/init.d/teamwork-server.sh
# update-rc.d teamwork-server.sh defaults
```