#Inbounds

inbound වල දැනට support කරන්නෙ vless protocole එක විතරයි. පුලුවන් වුනොත් ඉදිරියට vmess, trojan add කරන්නම්.

inbound object එකකට අමතර fields තුනක් add වෙනවා.

```json
{
  "type": "",
  "tag": "",

  ... // Listen Fields

  "users": [],
  "tls": {},
  "multiplex": {},
  "transport": {},

  //newly added fields
  "id": 1,
  "info": "information about this inbound",
  "support_info": ["supports cloudflare cdn"]
}
```

## default

users map එක ඒ විදිහටම හිස් ව තියන්න. ඒකට user ලා add කරන්නෙ නැ
ඉතුරු fields sing box docs වල තියෙන විදියට වෙනස් කරගන්න.

ඒ වගේම එක tag එකක් අනිවාරෙන් default කියලා තියෙන්න ඕනෙ. default inbound එකක්.

## specific to connectedbot

### **`id`**

inbound එකට අදාලව අනිවාරෙන් id එකක් තියෙන්න ඕනෙ. Intiger value එකක් මේක inbound list එකේ ඔක්කොටම ඒ ඒ Inbound වලට සුවිශේශි id එකක් තියෙන්න ඕනෙ.,
inbound දෙකක id සමාන වෙන්න බැහැ.

### **`info`**, **`support_info`**

මේ field දෙකම අත්‍ය්වශ්‍ය නැහැ. කැමතිනම් දාන්න පුලුවන් user ට Inbound එක ගැන details පෙන්නනෙ මෙතන දාන info වලින් කැමති දෙයක් දාගන්න පුලුවන්.
support Info array එකක් කරුණු වගෙ දේවල් මේකට දෙන්න. Inbound එක cdn support කරනවා. වගේ දෙවල්.
