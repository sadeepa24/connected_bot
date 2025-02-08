```json
{
  "out": {
    "type": "vless",
    "tag": "OnlyforTest",
    "server": "connected.bot",
    "server_port": 443,
    "uuid": "5dc65e91-5f05-455b-8178-4a04ea9f7de9",
    "transport": {
      "type": "ws",
      "path": "/",
      "headers": { "host": "connectebot" }
    }
  },
  "info": "test outbound, do not add this to you'r config",
  "reqirments": {
    "uuid": true,
    "server": true,
    "port": true,
    "server_name": true,
    "tag": true,
    "transporthost": true
  }
}
```

### **`out`**

මේක තමයි ඇත්තම sing box [Outbound](https://sing-box.sagernet.org/configuration/outbound/) object එක user config එකට Add වෙන.

### **`info`**

outbound object එක ගැන Details තියෙන්නෙ මේකෙ. මේක කියවලා User ට හිතාගන්න පුලුවන්නෙ ඒකෙන් වෙන දේ මොකක්ද කියලා.

### **`reqirments`**

මේකෙ තියෙන්නෙ outbound එක user ගෙ config එකට add කරන්න කලිම් user ගෙන් Inputs අරගෙන් එව්වා outbound එකටම add කරන එක `static` `true` දීලා තිබ්බොත් User ගෙන් කිසිම input එකක් ගන්නෙ නැහැ.

reqirments object

```json
{
  "uuid": true,
  "server": true,
  "port": true,
  "server_name": true,
  "tag": true,
  "transporthost": true
}
```

#### **`uuid`**

මේක true කලොත් user ගෙන් කිසිම uuid එකක් ඉල්ලලා ඒක outboudn එකට add කරනවා.

#### **`server`**

මේක true කලොත් user ගෙන් server addr එකක් ඉල්ලනවා. ඒක outbound object එකට add වෙනවා.

#### **`port`**

මේක true කලොත් user ගෙන් Port එකක් ඉල්ලනවා. ඉල්ලලා ඒක `port` field එක විදියට sing box outbound object එකට add වෙනවා.

#### **`tag`**

මේක true කලොත් user දෙන tag name එකෙන් Outbound එක add වෙන්නෙ.

#### **`transporthost`**

outbound එකේ transport host එක user input වලින් අරගෙන් add කරනවා.
