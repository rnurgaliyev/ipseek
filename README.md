# ipseek

This small server provides DDNS-like functionality to OpenStack VPNaaS IPsec
site connections. You would find this useful if you want to build an IPsec
Site-2-Site connection with a router, that does not have a static IP address.

IPsec peer ID and IP address are updated with the IP address, that is provided
in GET request to this server.

Update groups are configured in simple YAML file, so that multiple targets
can be updated at once.

The server is written in a way that it would be easy to add other services
that need dynamic address update for any developer who is familiar with Golang.

## Example configuration file

```
update_groups:
  - name: mygroup
    objects:

      - name: region-a
        type: openstack_ipsec_site_connection
        configuration:
          
          # URL to tokens endpoint of Identity v3 API (Keystone)
          authURL: "https://api.openstackprovider.net:5000/v3/auth/tokens"

          # URL to ipsec site connections endpoint of Network v2 API (Neutron) in region A
          url: "https://api.region-a.openstackprovider.net:9696/v2.0/vpn/ipsec-site-connections"

          user: "operator"
          domain: "Default"
          password: "secret"

          # ID of IPsec site connection that needs an update
          id: "ed236e07-625e-4a63-8d91-7b4ed59f2751"

      # Another object is configured in the same way
      - name: region-b
        type: openstack_ipsec_site_connection
        configuration:
          authURL: "https://api.openstackprovider.net:5000/v3/auth/tokens"
          url: "https://api.region-b.openstackprovider.net:9696/v2.0/vpn/ipsec-site-connections"
          user: "operator"
          domain: "Default"
          password: "secret"
          id: "29a49f7b-786b-4173-abb2-7e8cb80ed6c8"

```

## Example request to update IP address in VPNaaS configuration

```
http://x.x.x.x:8088/update?group=mygroup;address=192.0.2.1
```

## Running as Docker container on Raspberry Pi

```
docker run -dt -v /etc/ipseek.yml:/etc/ipseek.yml -p 8088:8088/tcp --name ipseek --restart unless-stopped imple/ipseek:latest 
```
