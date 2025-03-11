```json
{
  "dnsrules": [
    {
      "info": "only for testing do not add wee will add dnsrules soon",
      "rule": { "clash_mode": "direct", "server": "any", "tag": "testTag" },
      "reqirments": {
        "static": true
      }
    },

    {
      "info": "add this if you use our public domain as outbound server",
      "rule": {
        "server": "local",
        "tag": "connected_host_resolve",
        "domain_keyword": ["connected"]
      },
      "reqirments": {
        "static": true
      }
    }
  ],
  "dns_servers": [
    {
      "server": { "tag": "block", "address": "rcode://success" },
      "info": "defaul blocker"
    },
    {
      "server": { "tag": "cloudflare", "address": "tcp://1.1.1.1" },
      "info": "cloudflare tcp dns"
    },
    {
      "server": { "tag": "google", "address": "8.8.8.8" },
      "info": "google default root"
    },
    {
      "server": {
        "tag": "google_tls",
        "address": "tls://dns.google"
      },
      "info": "tls dns google need extrans resolver to resolve dns.google use local to resolve dns.google after adding to config add a resolver all dns server => resolver"
    },

    {
      "server": {
        "tag": "local",
        "address": "local",
        "detour": "direct"
      },
      "info": "you'r system dns"
    },

    {
      "server": {
        "tag": "cloudflare_https",
        "address": "https://1.1.1.1/dns-query"
      },
      "info": "cf https"
    },
    {
      "server": { "tag": "google_h3", "address": "h3://8.8.8.8/dns-query" },
      "info": "google 's h3 "
    },
    {
      "server": { "tag": "quic", "address": "quic://dns.adguard.com" },
      "info": "quic"
    }
  ],
  "routerule": [
    {
      "info": "test route rule, do not add this to you'r config",
      "rule": {
        "tag": "ads-block",
        "protocol": "dns",
        "outbound": "dns-out"
      }
    },
    {
      "info": "test route rule, do not add this to you'r config",
      "rule": {
        "tag": "srcipaddr",
        "source_ip_cidr": ["10.0.0.0/24", "192.168.0.1"],
        "outbound": "dns-out"
      },
      "reqirments": {
        "needip_cidr": true,
        "static": true
      }
    }
  ],
  "outbounds": [
    {
      "out": {
        "type": "vless",
        "tag": "OnlyforTest",
        "server": "connected.bot",
        "server_port": 443,
        "uuid": "5dc65e91-5f05-455b-8178-4a04ea9f7de9",
        "transport": {
          "type": "ws",
          "path": "/",
          "headers": { "host": "connectebot" }
        }
      },
      "info": "test outbound, do not add this to you'r config",
      "reqirments": {
        "tag": true
      }
    },
    {
      "out": {
        "type": "vless",
        "tag": "OnlyforTest1",
        "server": "connected.bot",
        "server_port": 443,
        "uuid": "5dc65e91-5f05-455b-8178-4a04ea9f7de9",
        "transport": {
          "type": "ws",
          "path": "/",
          "headers": { "host": "connectebot" }
        }
      },
      "info": "test route rule, do not add this to you'r config",
      "reqirments": {
        "uuid": false,
        "server": false,
        "port": false,
        "server_name": false,
        "tag": false,
        "transporthost": false
      }
    },
    {
      "out": {
        "type": "vless",
        "tag": "OnlyforTest2",
        "server": "connected.bot",
        "server_port": 443,
        "uuid": "5dc65e91-5f05-455b-8178-4a04ea9f7de9",
        "transport": {
          "type": "ws",
          "path": "/",
          "headers": { "host": "connectebot" }
        }
      },
      "info": "only for testings don't use",
      "reqirments": {
        "tag": true,
        "uuid": true
      }
    }
  ],
  "ruleset": [
    {
      "rule_set": {
        "type": "local",
        "tag": "rsettest",
        "format": "source",
        "path": "/testpath"
      },
      "info": "for testing do not use we will add many rule set soon"
    }
  ]
}
```
