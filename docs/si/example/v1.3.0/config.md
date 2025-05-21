#Bot Config

පහල තියෙන්නෙ example full config ( v1.3.0) එකක් ඒකෙ අනිවාරෙන්ම වෙනස් කරන්න ඕනෙ fields values තමා <> ඇතුලෙ තියෙන්නෙ ඔයා values දානකොට <> මේ tag දෙක අයින් කරන්න ඕනෙ < value > මෙහෙම තිබ්බොත් වැරදියි.

```json
{
  "watchman": { "del_buffer": 10 },
  "db_path": "<you'r db path>",
  "bot_token": "<you'r bot token>",
  "usagedb_path": "<you'r log db path>",
  "templates_path": "./templates.yaml",
  "bot_mainurl": "https://api.telegram.org/bot",
  "metadata": {
    "channel_id": < channel id >,
    "group_id": < group id>,
    "bandwidth": "6000GB",
    "login_limit": 5,
    "max_config_count": 10,
    "max_build_conf": 4,
    "max_gift": 3,
    "refresh_rate": 2,
    "backup_rate": 3,
    "group_link": "< group link >",
    "channel_link": "< channel link >",
    "bot_link": "< bot link>",
    "group_name": "< group name>",
    "channel_name": "< channel name>",
    "bot_name": "< bot name >",
    "admin": < admin tg id>,
    "default_domain": "< you'r domain >",
    "default_publicip": "< server ip>",
    "store_path": "./store.yaml",
    "config_folder": "./configs",
    "help_cmd": {
      "info_pages": 1,
      "tuto_pages": 1,
      "cmd_pages": 7,
      "builder_pages": 1
    },
    "inline_posts": [
      {
        "description": "share this post with friends",
        "title": "Share this",
        "template_name": "chan_share"
      }
    ],
    "allowed_langs": ["en", "sin"],
    "default_lang": "en"
  },
  "webhook_server": {
    "http_path": "/",
    "allowed_updates": [
      "message",
      "callback_query",
      "chat_member",
      "inline_query"
    ],
    "full_url": "https://<you'rdomain>:88/",
    "secret": "< you'r secret (can be any string)>",
    "disable_setwebhook": false,
    "req_reject_message": "පලයන් යන්න පකයා.",
    "listen_option": {
      "addr": "0.0.0.0:88",
      "cert": "<you'r ssl key path>",
      "key": "<you'r priv key path>",
      "reject_message": "conn rejected (add anything you want 😅😅)",
      "server_name": "< you'r domain name>"
    }
  },
  "sbox_path": "./sbox.json",
  "log": { "paths": ["bot.log"], "level": "warn", "encoding": "console" }
}
```
