#Bot Config

පහල තියෙන්නෙ example full config එකක් ඒකෙ අනිවාරෙන්ම වෙනස් කරන්න ඕනෙ fields values තමා <> ඇතුලෙ තියෙන්නෙ ඔයා values දානකොට <> මේ tag දෙක අයින් කරන්න ඕනෙ < value > මෙහෙම තිබ්බොත් වැරදියි.

ඒ වගේම Bot Deploy කලාට පස්සෙ /template command එකෙන් මේ template edit කරන්න පුලුවන්.

```json
{
  "db_path": "<you'r db path>",
  "bot_token": "<you'r bot token>",
  "bot_mainurl": "https://api.telegram.org/bot",
  "sbox_path": "./sbox.json",
  "templates_path": "./templates.yaml",
  "metadata": {
    "channel_id": < channel id >,
    "groupd_id": < group id >,
    "bandwidth": "6000GB",
    "login_limit": 5,
    "max_config_count": 10,
    "refresh_rate": 2,
    "channel_link": "<channel link >",
    "group_link": "< group link >",
    "bot_link": "< bot link >",
    "group_name": "< Groupd Name>",
    "channel_name": "< Channel Name>",
    "bot_name": "< Bot Name>",
    "admin": < admin id>,
    "default_domain": "<Default Domain>",
    "default_publicip": "< Default Ip>",
    "store_path": "./store.json",
    "config_folder": "./configs",
    "help_cmd": {
      "info_pages": 1,
      "tuto_pages": 1,
      "cmd_pages": 1,
      "builder_pages": 1
    },
    "inline_posts": []
  },
  "log": {
    "level": "debug",
    "encoding": "console",
    "paths": ["stdout"]
  },
  "webhook_server": {
    "addr": "0.0.0.0:88",
    "cert": "<fullchain.pem  path>",
    "key": "<privkey.pem path>",
    "server_name": "<domain name >",
    "http_path": "/",
    "full_url": "https://<yourdomain>:88/",
    "secret": "< secret hash >",
    "allowed_updates": [
      "message",
      "callback_query",
      "chat_member",
      "inline_query"
    ],
    "disable_setwebhook": false
  },
  "watchman": {
    "del_buffer": 10
  }
}
```
