## Provider network
second interface on each worker must configure:

netplan:
```
    ens33:
      dhcp4: no
      dhcp6: no
      match:
        macaddress: 00:0c:29:91:56:29
      set-name: provider1
```

Better to use jumbo frames.
