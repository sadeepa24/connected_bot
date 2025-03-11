```json
{
  "disable_setwebhook": false,
  "http_path": "/",
  "full_url": "https://origin.connectedbot.site:88/",
  "secret": "3t6dsidt2hj6k2ns23jsah6ds52ndm32dchs",
  "req_reject_message": "Request From Unknown",
  "allowed_updates": [
    "message",
    "callback_query",
    "chat_member",
    "inline_query"
  ],
  "listen_option": {
    "server_name": "origin.connectedbot.site",
    "addr": "0.0.0.0:88",
    "cert": "./tls/origin.connectedbot.site/fullchain.pem",
    "key": "./tls/origin.connectedbot.site/privkey.pem",
    "allowd_cidr": ["149.154.160.0/20", "91.108.4.0/22"],
    "reject_message": "rejected conn"
  }
}
```

### **`http_path`**

telegram එකට webhook rejister කරන්නෙ මේ path එකෙන් මේ path එක හරියටම දෙන්න full url එකේ

### **`full_url`**

ඔයා උඩ දීපු settings අනුව full url එක
ex :=

88 න් listen කරලා ඔයාගෙ domain එක example.com කියලා හිතුවොත් path එක / දුන්නොත් full url වෙන්නෙ
https://example.com:88/

### **`secret`**

මේක තමයි එන request verify කරන්නෙ අනිවාරෙන් මේක දාන්න string එකක් මොකක් හරි ඔයාගෙ Uniq දේක hash එකක් වගෙ use කරන්න

### **`req_reject_message`**

අදාල නැති http req එකක් recive වුනහම. ඒකට අදාලව දෙන response message එක.

### **`allowed_updates`**

තියෙන ටික එහෙමම තියන්න, වෙනස් කරන්න ඕනෙ නැ

### **`disable_setwebhook`**

මේක true දාගන්න false දැම්මොත් webhook එක rejister කරන්නෙ නැ telegram වල.

## **listen_option**

### **`addr`**

telegram එකෙන් request එන්නෙ මේ listen වෙන socket එකට, telegram port හතරක support කරනවා මේකට 80, 88, 443, 8443, 80 & 443 use මෙතන use කරන්න එපා vpn inbound වලට එව්වා use කරන්න බරුව යනවා එතකොට. මෙතන දෙන port එකම හරියට `full_url` එකේ දාන්න ඕනෙ.

### **`cert`**

tls cert file path එක .pem format එකෙන්. letsencrypt වලින් නම් ගන්නෙ මේක ගොඩක් වෙලාවට තියෙන්නෙ මේ ලොකේශන් එකේ
/root/etc/letsencrypt/live/yourdomain.com/fullchain.pem

### **`key`**

tls Key path එක .pem format එකෙන්. letsencrypt වලින් නම් ගන්නෙ මේක ගොඩක් වෙලාවට තියෙන්නෙ මේ ලොකේශන් එකේ
/root/etc/letsencrypt/live/yourdomain.com/privkey.pem

### **`server_name`**

ඔයාගෙ ssl cert ගත්ත domain එක මෙතනට දෙන්න

### **`allowd_cidr`**

මෙතනට දාලා තියෙන ip rangers වලට විතරක් webhook එකට req කරන්න පුලුවන්. Telegram වලින් req එන්නෙ `149.154.160.0/20` `91.108.4.0/22` මේ range වලින් විතරයි ඒ range දෙක දාන්න මෙතනට මෙතනට මුකුත් range දාලා නැත්තන් Ip filter එක off විදියට වැඩ කරන්නෙ. ip addres එක reject වුනොත් එතනින් එහාට tls handhsake එකටත් කලින් connection terminate වෙනවා.

### **`reject_message`**

inbound connection එක allowd_cidr වල නැති ip addr එකකින් ආවොත් connection එක reject වෙද්දි connection එකට send කරන message එක.
