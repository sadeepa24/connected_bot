#Outbounds

sing-box වල තියෙන ඕනෙම Outbound එකක් support කරනවා. අනිවාරෙන් direct & block outbound එක config එකේ තියෙන්න ඕනෙ.

```json
{
  "type": "",
  "tag": "",

  ... // outbound fields according to outbound type


  //newly added fields
  "id": 1,
  "info": "information about this inbound",
}
```

## default

sing box docs වලට අනුව configure කරගන්න.

## specific to connectedbot

### **`id`**

outbound එකට අදාලව අනිවාරෙන් id එකක් තියෙන්න ඕනෙ. Intiger value එකක් මේක, outbound list එකේ ඔක්කොටම ඒ ඒ outbound වලට සුවිශේශි id එකක් තියෙන්න ඕනෙ.
outbound දෙකක id සමාන වෙන්න බැහැ.

outbound tag එකට අදාලව routing rules වල botrule එකක් add කරලා තියෙන්නෙ ඕනෙ නැත්තන් realtime outbound change කරන්න බැහැ.

### **`info`**,

inbound වල වගේම මේකෙත් outbound එක ගැන details තමා add කරන්න ඕනෙ.
