#Config Store

Store File එක සම්පූර්ණයෙන්ම අදාල වෙන්නෙ /buildconf command එකත් එක්ක මේක vpn එක්ක කිසිම සම්බන්දයක් නැහැ. user ලට config build කරද්දි prebuild කරපු කොටස් එයාලගෙ config වලට add කරගන්න මේක තියෙන්නෙ.

පහල තියෙන්නෙ full config එකේ structure එක

මේ ඔක්කොම sing box v1.10.0 වලට අනුව තම හදලා තියෙන්නෙ.

```json
{
  "dnsrules": [],
  "dns_servers": [],
  "routerule": [],
  "outbounds": [],
  "ruleset": []
}
```

| Key         | Type                                   |
| ----------- | -------------------------------------- |
| dnsrules    | Array of [DnsRule](./dnsrules.md)      |
| dns_servers | Array of [DnsServer](./dns_servers.md) |
| routerule   | Array of [RouteRule](./routerule.md)   |
| outbounds   | Array of [Outbound](./outbounds.md)    |
| ruleset     | Array of [RuleSet](./ruleset.md)       |
