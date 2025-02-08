#BOT config
config should be in json format

```json
{
  "db_path": "./latest.db",
  "bot_token": "56754353:AAG-GSYKGsyluerdscEr3daSecfeEdx23",
  "bot_mainurl": "https://api.telegram.org/bot", //වෙනස් කරන්න එපා
  "sbox_path": "./sbox.json",
  "templates_path": "./templates.yaml",
  "log": {},
  "webhook_server": {},
  "metadata": {},
  "watchman": {}
}
```

| Key              | Description                                        |
| ---------------- | -------------------------------------------------- |
| `db_path`        | Database එකේ path එක                               |
| `bot_token`      | Botfather ගෙන් හදාගන්න token එක                    |
| `bot_mainurl`    | මේක වෙනස් කරන්න එපා (https://api.telegram.org/bot) |
| `sbox_path`      | sing box vpn config එකේ path එක                    |
| `templates_path` | Message Template file එකේ path එක                  |
| `log`            | [Configuration for logging](log.md)                |
| `webhook_server` | [Configuration for the webhook server](webhook.md) |
| `metadata`       | [Metadata information](metadata.md)                |
| `watchman`       | [Configuration for the watchman](watchman.md)      |

##Fields

### **`db_path`**

මෙතනට **database එකේ path එක** දෙන්න. අලුතෙන් run කරනවනන් database එක create වෙන්නෙ මේ දෙන path එකේ. server change කරද්දි backup db එකේ path මෙතනට දෙන්න ඕනෙ

### **`bot_token`**

telegram එකෙ [**bot father**](https://t.me/BotFather) ගෙන් හදන **Bot** ට අදාල **token එක** මෙතනට දෙන්න.

### **`bot_mainurl`**

මේක වෙනස් කරන්න එපා තියෙන එකම තියන්න. (https://api.telegram.org/bot)

### **`sbox_path`**

**sing-box vpn config එක** තියෙන path එක. මේ config එක හදන විදිය [මෙතනින්](../sing-box/index.md) බලන්න

### **`templates_path`**

මේකට දෙන්න ඕනෙ **message template yaml file එකේ path එක**.

මේ file එකේ තමා ඔක්කොම ඔයාට **customize** කරන්න පුලුවන් message තියෙන්නෙ. මේක **customize** කරන හැටි [මෙතනින්](../../other_config/templates/index.md) බලන්න

### **`log`**

Bot logs output වෙන්නේ මේ object එක අනුව මෙතනින් ඒක [බලන්න](./log.md)

### **`webhook_server`**

telegram updates එන්නෙ මේ **webhook එකට**, ඒකට අදාල configuration එක [මෙතනින්](webhook.md) බලන්න

### **`metadata`**

**bot run වෙන්න අවශ්‍ය ඔක්කොම detail** මේ Object එකට දෙන්න [මෙතනින්](metadata.md) බලන්න විස්තර

### **`watchman`**

[මේක වැඩිය සලකන්න එපා](watchman.md)
