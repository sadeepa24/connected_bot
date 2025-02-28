#Sing Box Config

පහල තියෙන්නෙ සම්පූර්ණ sing box config එකකට example එකක්.

Inbounds

- port 80 ws, without tls
- port 443 ws, with tls
- port 2096 ws, with tls

Outbounds

- direct
- wg1
- wg2

direct, wg1, wg2 මේ outbound තුන Bot /configure වලින් මේවා මාරු කරන්න පුලුවන්.
routings rules add කරන විදිය මේකෙන් අදහසක් ගන්නත් පුලුවන්.

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
      }
    ],
    "final": "cf"
  },

  "inbounds": [
    {
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
        "certificate_path": "fullchain.pem",
        "key_path": "privkey.pem"
      },
      "transport": {
        "type": "ws"
      }
    },
    {
      "type": "vless",
      "tag": "vless_2096",
      "info": "tls එක්ක මොකක් හරි host එකක් connect කරනවනන් මේ Inbound එක තෝරගන්න",
      "support_info": [
        "use cdn.connectedbot.site to connect via cloudflare cdn",
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
      "users": [],
      "tls": {
        "enabled": true,
        "certificate_path": "fullchain.pem",
        "key_path": "privkey.pem"
      },
      "transport": {
        "type": "ws"
      }
    },
    {
      "info": "tls නැතුව ws (Host in Http Header) විදියට connect කරගන්නවනම් මේ inbound එක use කරන්න. ",
      "type": "vless",
      "tag": "vless_80",
      "support_info": [
        "use cdn.connectedbot.site to connect via cloudflare cdn",
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
  "endpoints": [],
  "outbounds": [
    {
      "type": "direct",
      "tag": "direct",
      "id": 1,
      "info": "normal එදිනෙදා use කරන්න. ( torrent download කරනවනන් වෙන Outbound එක්කට switch කරන්න)"
    },
    {
      "type": "wireguard",
      "tag": "wg1",
      "id": 2,
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
      "id": 3,
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
      {
        "action": "sniff",
        "timeout": "400ms"
      },
      {
        "protocol": "dns",
        "outbound": "direct"
      },
      { "type": "botrule", "outbound": "wg1", "action": "route" },
      { "type": "botrule", "outbound": "wg2", "action": "route" },
      { "action": "reject", "protocol": "bittorrent" },
      { "type": "botrule", "outbound": "direct", "action": "route" }
    ],
    "final": "direct",
    "auto_detect_interface": true
  }
}
```
