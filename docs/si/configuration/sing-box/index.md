#SING-BOX config

```json
{
  "log": {},
  "dns": {},
  "ntp": {},
  "inbounds": [],
  "outbounds": [],
  "route": {},
  "experimental": {}
}
```

### Fields

| Key            | Format                                                                    |
| -------------- | ------------------------------------------------------------------------- |
| `log`          | [Log](https://sing-box.sagernet.org/configuration/log/)                   |
| `dns`          | [DNS](https://sing-box.sagernet.org/configuration/dns/)                   |
| `ntp`          | [NTP](https://sing-box.sagernet.org/configuration/ntp/)                   |
| `inbounds`     | [Inbound](https://sing-box.sagernet.org/configuration/inbound/)           |
| `outbounds`    | [Outbound](https://sing-box.sagernet.org/configuration/outbound/)         |
| `route`        | [Route](https://sing-box.sagernet.org/configuration/route/)               |
| `experimental` | [Experimental](https://sing-box.sagernet.org/configuration/experimental/) |

original [sing box docs](https://sing-box.sagernet.org/configuration) වලට අනුව configs හදන්න වෙනසක් නැ අමතර fields ටිකක් add වෙනවා ඒ ටික දාන්නම්.
ඒ වාගෙම මේකෙ use කරන්නෙ singbox v1.10.1, singbox v1.11.0 වලින් config migration එකක් වුණා routing එහෙම පොඩ්ඩක් වෙනස් විදියට.

### **ඒ නිසා ඔයාලගෙ config හදද්දි v1.10.0 වලට support වෙන්න හදන්න ඕනෙ**

connected_bot එකේ singbox වල routing වෙනස් කරලා තියෙන්නෙ. ඒක වැඩ කරනන් විදිය පහලින් බලන්න

connection එක sniff වෙන්නෙ අදාල connection එකේ Outbound එක direct නම් විතරයි.
outbound direct වෙලා torrent traffic එකක් නම් එන්නෙ torrent කියලා tag එක තියෙන default outbound එකක් තිබ්බෙ නැත්තන් torrent traffic Block වෙනවා.

connection එක වෙන Outbound එකකින් නම් යන්නෙ (ex- warp, proxy, wireguard, or anyother) sniff වෙන්නෙ නැ. connection එක route කරනවා විතරයි. මෙහම කරන්නෙ memory usage අඩු කරගන්න.

### **අමතරව එකතු වුණ fields ටික පහලින් බලන්න. මේ fields අනිවාරෙන් add කරන්න ඕනෙ.**

- [Inbounds](./inbounds.md)
- [Outbounds](./outbounds.md)
- [Routing](./routing.md)
