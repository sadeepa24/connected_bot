#Sing Box Config

පහල තියෙන්නෙ සම්පූර්ණ sing box කොන්ෆිග් ekakata example එකක්.

Inbounds

- port 80 ws, without tls
- port 443 ws, with tls
- port 2096 ws, with tls

Outbounds

- block, dns-out, direct මේ තුන සියලුම config වල අනිවාරෙන් තියෙන්න ඕනෙ.
- wg1, wg2 මේ outbound දෙක wireguard Bot /configure වලින් මේවා මාරු කරන්න පුලුවන්.

මේ config එකම use කරනවනන් inbound වල tls path ටික වෙනස් කරන්න.

```json
      "tls": {
       "enabled": true,
        "certificate_path": "fullchain.pem", // change
        "key_path": "privkey.pem"  // change
      }
```

```json
{
  "log": { "disabled": true },
  "dns": {
    "servers": [
      {
        "tag": "cf",
        "address": "tcp://1.1.1.1"
      },
      {
        "tag": "block",
        "address": "rcode://success"
      }
    ]
  },
  "inbounds": [
    {
      "sniff": true,
      "sniff_override_destination": false,
      "sniff_timeout": "100ms",
      "type": "vless",
      "tag": "default",
      "info": "tls 443 port එක්ක මොකක් හරි host එකක් connect කරනවනන් මේ Inbound එක තෝරගන්න",
      "id": 1,
      "listen": "::",
      "listen_port": 443,
      "tcp_fast_open": true,
      "multiplex": {
        "enabled": false,
        "padding": false,
        "brutal": {}
      },
      "users": [],
      "tls": {
       "enabled": true,
        "certificate_path": "fullchain.pem",  // add you'r fullchain path
        "key_path": "privkey.pem"  // add you'r privkey path
      },
      "transport": {
        "type": "ws"
      }
    },
    {
      "sniff": true,
      "sniff_override_destination": false,
      "sniff_timeout": "100ms",
      "type": "vless",
      "tag": "vless_2096",
      "info": "tls එක්ක මොකක් හරි host එකක් connect කරනවනන් මේ Inbound එක තෝරගන්න",
      "support_info": [
        "you can also use this inbound to connect via cloudflare cdn",
        "special port for use cloudflare cdn"
      ],
      "id": 2,
      "listen": "::",
      "listen_port": 2096,
      "tcp_fast_open": true,
      "multiplex": {
        "enabled": false,
        "padding": false,
        "brutal": {}
      },
      "users": [
        { "name": "My", "uuid": "1c7a5143-bfeb-4cfd-b733-1f5e96edc949" }
      ],
      "tls": {
        "enabled": true,
        "certificate_path": "fullchain.pem",  // add you'r fullchain path
        "key_path": "privkey.pem"  // add you'r privkey path
      },
      "transport": {
        "type": "ws"
      }
    },
    {
      "sniff": true,
      "sniff_override_destination": false,
      "sniff_timeout": "100ms",
      "info": "tls නැතුව ws (Host in Http Header) විදියට connect කරගන්නවනම් මේ inbound එක use කරන්න. ",
      "type": "vless",
      "tag": "vless_80",
      "support_info": [
        "use this inbound to connect via cloudflare cdn",
        "use 443 port with tls when connecting using cloudflare",
        "you can add cloudfront or fastly cdn if you want"
      ],
      "id": 3,
      "listen": "::",
      "listen_port": 80,
      "tcp_fast_open": true,
      "multiplex": {
        "enabled": false,
        "padding": false,
        "brutal": {}
      },
      "users": [],
      "transport": {
        "type": "ws"
      }
    }
  ],
  "outbounds": [
    {
      "type": "direct",
      "tag": "direct",
      "id": 1,
      "info": "normal එදිනෙදා use කරන්න. ( torrent download කරනවනන් වෙන Outbound එක්කට switch කරන්න)"
    },
    { "type": "block", "tag": "block", "id": 2 },
    {
      "type": "dns",
      "tag": "dns-out",
      "id": 3
    },
    {
      "type": "wireguard",
      "tag": "wg1",
      "id": 4,
      "info": "cloudflare warp outbound, Cannot Download Torrent",
      "server": "162.159.192.1",
      "server_port": 2408,
      "local_address": [
        "172.16.0.2/32",
        "2606:4700:110:8538:5ed:5c4b:f9b:1316/128"
      ],
      "private_key": "QDYSUTLO9wza8vm8jodzo43EfXDBCeoB+bODa3faX18=",
      "peer_public_key": "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=",
      "reserved": [128, 110, 167],
      "mtu": 1280
    },

    {
      "type": "wireguard",
      "tag": "wg2",
      "id": 5,
      "server": "162.159.192.1",
      "info": "cloudflare warp outbound, Cannot Download Torrent",
      "server_port": 2408,
      "local_address": [
        "172.16.0.2/32",
        "2606:4700:110:8538:5ed:5c4b:f9b:1316/128"
      ],
      "private_key": "OJmT8jll9AKtXB1XTCjnfUZhC0gXFizA26Bf3ns8BmI=",
      "peer_public_key": "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=",
      "reserved": [245, 158, 70],
      "mtu": 1280
    }
  ],
  "route": {
    "rules": [
      { "type": "botrule", "outbound": "direct" }
      { "type": "botrule", "outbound": "wg1" },
      { "type": "botrule", "outbound": "wg2" },
      { "type": "botrule", "outbound": "block" },
      { "type": "", "outbound": "block", "protocol": "bittorrent" },
    ],
    "final": "direct",
    "auto_detect_interface": true
  }
}
```
