```json
{
  "rules": [
    {
      "type": "botrule",
      "outbound": "direct-no",
      "action": "route"
    }
  ],

  "final": "direct",
  "auto_detect_interface": true
}
```

ඔය උඩ තියෙන්නෙ route object එකක්, ඒකෙ rules වලට අදාලව ඔයාලට rules දාගන්න පුලුවන්

සෑම Outbound එකකටම අදාලව botrule එකක් add කරන්න ඕනෙ. නැත්තන් tg bot ඃඅරහා realtme outbound change කරන්න බැ.

හිතන්න ඔයා ගාව wg-out කියලා tag තියෙන Outbound එකක් තියේ කියලා එහෙනන් අනිවාරෙන ඒකට අදාලව `{ "type": "botrule", "outbound": "wg-out", "action": "route" }` මේ rule එකත් උඩ object එකේ ඇඩ් කරලා තියෙන්න ඕන.

මීට අමතරව ඔයාට sing-box වල තියෙන rules කැමති විදියකට Use කරන්න පුලුවන්.

### පහල තියෙන්නෙ old version අදාලව (v1.0.0)

for v.1.0.0

```json
{
  "rules": [{ "type": "botrule", "outbound": "direct" }],
  "final": "direct",
  "auto_detect_interface": true
}
```

ඔය උඩ තියෙන්නෙ route object එකක්, ඒකෙ rules වලට අදාලව ඔයාලට rules දාගන්න පුලුවන්

හැම Outbound එකකටම (block ඇරෙන්න) අදාලව botrule එකක් add කරන්න ඕනෙ. නැත්තන් outbound change කරන්න බැ.

හිතන්න ඔයා ගාව wg-out කියලා tag තියෙන Outbound එකක් තියේ කියලා එහෙනන් අනිවාරෙන ඒකට අදාලව `{ "type": "botrule", "outbound": "wg-out" }` මේ rule එකත් උඩ object එකේ ඇඩ් කරලා තියෙන්න ඕන.

මීට අමතරව ඔයාට sing-box වල තියෙන rules කැමති විදියකට Use කරන්න පුලුවන්.
