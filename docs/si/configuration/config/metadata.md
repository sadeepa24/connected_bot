```json
{
  "channel_id": -1002431173286,
  "groupd_id": -1002351503478,
  "bandwidth": "6000GB",
  "login_limit": 5,
  "max_config_count": 10,
  "refresh_rate": 2,
  "group_link": "https://t.me/connected_chat",
  "channel_link": "https://t.me/just_connected",
  "bot_link": "https://t.me/just_connected_bot",
  "group_name": "𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 Chat",
  "channel_name": "𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿",
  "bot_name": "𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 Bot",
  "admin": 111111111,
  "default_domain": "origin.connectedbot.site",
  "default_publicip": "172.236.149.167",
  "store_path": "./store.json",
  "config_folder": "./configs",
  "help_cmd": {
    "info_pages": 1,
    "tuto_pages": 1,
    "cmd_pages": 7,
    "builder_pages": 1
  },
  "inline_posts": [],
  "warn_rate": 24
}
```

### **`channel_id`**

channel එකේ ID එක, මේ [bot](https://t.me/RawDataBot) ට channel එකෙන් message එකක් forward කරලා ගන්න.

### **`groupd_id`**

group එකේ ID එක, මේ [bot](https://t.me/RawDataBot) ට group එකෙන් message එකක් forward කරලා ගන්න.

### **`bandwidth`**

server එකේ bandwidth එක, මේක string විදියට දෙන්න.

- `6000GB`
- `2000GB`

### **`login_limit`**

user කෙනෙක් config එකක් හදද්දි එයාගෙ config එකට login limit විදියට දෙන්න පුලුවන් max value එක.

### **`max_config_count`**

user කෙනෙකුට හදන්න පුලුවන් උපරිම config ගණන

### **`refresh_rate`**

database එක refrsh වෙන්නෙ පැය කීයකට වාරයක්ද කියන එක. (2 - 6 අතර ප්‍රමාණයක් දෙන්න දශම දෙන්න එපා.)

### **`group_link`**

group එකේ ලින්ක් එක

### **`channel_link`**

channel එකේ ලින්ක් එක

### **`bot_link`**

bot link එක

### **`group_name`**

group එකේ නම

### **`channel_name`**

channel එකේ නම

### **`bot_name`**

bot ගෙ නම

### **`admin`**

bot අදාල admin ගෙ telegram id එක. මේ දෙන id එක අදාල කෙනාට තමා ඔක්කොම control කරන්න පුලුවන්.

### **`default_domain`**

default domain එක මේක තමා vpn connect කරන්න ගන්නෙ

### **`default_publicip`**

default ip එක මේක තමා ip එකෙන් vpn connect කරනවනම් ගන්නෙ

### **`store_path`**

store config path එක මේක configure කරන විදිය [මෙතනින්](../../other_config/store/index.md) බලන්න

### **`config_folder`**

working directory එකේ folder එකක් හදලා ඒකෙ Path එක දෙන්න.
මේ folder එක ඇතුලෙ තමා user ලා buildconf command එක use කරලා හදන config save වෙන්නෙ server change කරද්දි මේකත් මාරු කරන්න නැත්තන් ඔක්කොම හදලා තිබ්බ config default වෙනවා.

### **`help_cmd`**

help cmd එකට අදාල විස්තර ඔක්කොම තියෙන්නෙ මෙතන
ඔයාලට ඇති තරම් help cmd එකේ pages හදන්න පුලුවන්.

```json
{
  "info_pages": 1,
  "tuto_pages": 1,
  "cmd_pages": 7,
  "builder_pages": 1
}
```

උදාහරණයක් විදියට info_pages කියන එකට 5 ක් දුන්නොත් /help දුන්නහම එන Bot Info button එකෙන් next next දීලා Pages පහක් යනකන් ගියහැකි ඒ ඒ Pages වලට පෙන්නන message එක
templates file එකේ දාන්න ඕනෙ. Info page අනුවනන් help_info1, help_info2, help_info3... ඔය විදියට අනික් 3 ටත් ඔහොමයි

### **`inline_posts`**

මේ array එකට කැමති නම් දෙන්න පුලුවන් මේ array එකට chan_post කියලා නමක් දුන්නොත්. ඒ නමින් template එකක් හදන්න ඕනෙ templates file එකේ ඊට පස්සෙ ඒ post එක
bot ගෙ inline mode එකෙන් share කරන්න පුලුවන්.

### **`warn_rate`**

This feature is new from v1.2.0

user කෙනෙකුට config use කරන්නෙ නැතුව ඉන්න පුලුවන් උපරිම පැය ගණන, default 24 ( මේ පැය ගනන ගියාට පස්සෙ User ට temp limit එකක් වැටෙන ඒ එයාටම අයින් කරන්න පුලුවන් අයින් කරද්දි එයාට විතරක් මේ rate එක අඩු වෙනවා ඒ කියන්නෙ 24 එක 12 වෙනවා. මෙහෙම අඩු වෙවී ගිහින් 0 වුනොත් මාසෙම limit)
