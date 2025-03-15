#Bot Config

පහල තියෙන්නෙ example full config ( v1.2.0) එකක් ඒකෙ අනිවාරෙන්ම වෙනස් කරන්න ඕනෙ fields values තමා <> ඇතුලෙ තියෙන්නෙ ඔයා values දානකොට <> මේ tag දෙක අයින් කරන්න ඕනෙ < value > මෙහෙම තිබ්බොත් වැරදියි.

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
    "warn_rate": 24,
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
    "inline_posts": ["chan_share"]
  },
  "log": {
    "level": "info",
    "encoding": "console",
    "paths": ["./bot.log", "stdout"]
  },
  "webhook_server": {
    "disable_setwebhook": false,
    "http_path": "/",
    "full_url": "https://<yourdomain>:88/",
    "secret": "< secret hash >",
    "req_reject_message": "conn rejected (add anything you want 😅😅)",
    "allowed_updates": [
      "message",
      "callback_query",
      "chat_member",
      "inline_query"
    ],
    "listen_option": {
      "addr": "0.0.0.0:88",
      "cert": "<fullchain.pem  path>",
      "key": "<privkey.pem path>",
      "server_name": "<domain name >",
      "reject_message": "conn rejected (add anything you want 😅😅)",
    }
  },
  "watchman": {
    "del_buffer": 10
  }
}
```
