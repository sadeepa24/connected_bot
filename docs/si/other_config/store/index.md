#Config Store

Store File එක සම්පූර්ණයෙන්ම අදාල වෙන්නෙ /buildconf command එකත් එක්ක මේක vpn එක්ක කිසිම සම්බන්දයක් නැහැ. user ලට config build කරද්දි prebuild කරපු කොටස් එයාලගෙ config වලට add කරගන්න මේක තියෙන්නෙ.

sing-box latest version සමග මෙයත් support කරනවා

## Structure :

```yaml
- type: routerule
  name: testroute rule
  wizard: true
  item: >
    {
      "inbound": "test",
      "auth_user": ["<<send example user1 name>>", "user2", "user3", "user4", "user5"],
      "action": "reject",
      "method": "default",
      "no_drop": true
    }
```

**`type`**

#### Boilerplate Type 6 ක් තියෙනවා, පහල විදියට

- [fullconfig](https://sing-box.sagernet.org/configuration/)
- [inbound](https://sing-box.sagernet.org/configuration/inbound/)
- [outbound](https://sing-box.sagernet.org/configuration/outbound/)
- [endpoint](https://sing-box.sagernet.org/configuration/endpoint/)
- [dnsrule](https://sing-box.sagernet.org/configuration/dns/rule/)
- [routerule](https://sing-box.sagernet.org/configuration/route/rule/)

**`name`**

අදාල BoilerPlate එකේ නම

**`wizard`**

මේක true නම් විතරයි user ගෙන් input අරගෙන් parameter ඇතුලෙ තියෙන දේවල් replace වෙන්නෙ. <<>>,
parameter නැතුව static config එකක් නම් මේක false තියෙන්න ඕනෙ

**`item`**

BoilerPlate එක,

<<>> මෙහෙම එකක් තියෙනවා කියන්නෙ ඒක ඇතුලෙ තියෙන එක User ට send කරනවා. user දෙන Input එකෙන් එතන replace වෙනවා.

## Example

```yaml
- type: fullconfig
  name: testconfig
  wizard: true
  item: >
    {
      "dns": {
        "servers": [
          {
            "tag": "newdnssrv",
            "address": "1.1.1.1",
            "address_strategy": "prefer_ipv4",
            "address_fallback_delay": "300ms",
            "detour": "direct"
          }
        ],
        "rules": [
          {
            "inbound": "test",
            "action": "reject",
            "no_drop": true
          }
        ]
      },
      "inbounds": [
        {
          "type": "vless",
          "tag": "<<send the inbound tag name>>",
          "users": [
            {
              "name": "testname",
              "uuid": "<<send your' vless uuid>>"
            }
          ]
        }
      ],
      "route": {
        "rules": [
          {
            "inbound": "test",
            "auth_user": ["user1", "user2", "user3", "user4", "user5"],
            "action": "reject",
            "no_drop": true
          }
        ],
        "final": "<<send the final tag name>>"
      },
      "experimental": {
        "clash_api": {
          "external_controller": "127.0.0.1"
        }
      }
    }

- type: inbound
  name: testinbound
  wizard: true
  item: >
    {
      "type": "trojan",
      "tag": "new",
      "listen": "0.0.0.0",
      "listen_port": <<send port>>,
      "tcp_fast_open": true,
      "tls": {
        "enabled": true,
        "insecure": true
      },
      "multiplex": {
        "enabled": true
      },
      "transport": {
        "type": "ws",
        "path": "/",
        "headers": {
          "host": "<<send hosthname>>"
        }
      }
    }
- type: inbound
  name: testinbound2
  wizard: false
  item: >
    {
      "type": "trojan",
      "tag": "new",
      "listen": "0.0.0.0",
      "listen_port": 443,
      "tcp_fast_open": true,
      "tls": {
        "enabled": true,
        "insecure": true
      },
      "multiplex": {
        "enabled": true
      },
      "transport": {
        "type": "ws",
        "path": "/"
      }
    }

- type: outbound
  name: testout
  wizard: true
  item: >
    {
      "type": "vless",
      "tag": "bt",
      "server": "origin.connectedbot.site",
      "server_port": 443,
      "uuid": "3d4ae362-c28c-42db-8e4d-a92b5892e2d9",
      "tls": {
        "enabled": true,
        "server_name": "<< send sni >>",
        "insecure": true
      },
      "transport": {
        "type": "ws",
        "headers": {
          "host": "<<send websocket hostname>>"
        }
      }
    }

- type: routerule
  name: testroute rule
  wizard: true
  item: >
    {
      "inbound": "test",
      "auth_user": ["<<send example user1 name>>", "user2", "user3", "user4", "user5"],
      "action": "reject",
      "method": "default",
      "no_drop": true
    }
```
