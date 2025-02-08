# Dns Rule

```json
{

  ../Acctual sing box rule
  "rule": {
    "server": "local",
    "tag": "connected_host_resolve",
    "domain_keyword": ["connected"]
  },

  ../bot specific
  "info": "add this if you use our public domain as outbound server",
  "reqirments": {
    "static": true,
    "needport": true,
    "needportRange": true,
    "needdomain": true,
    "needprotocole": true,
    "needip_cidr": true,
    "needruleSet": true,
  }
}
```

### **`rule`**

මේක තමයි ඇත්තන sing box [dns rule](https://sing-box.sagernet.org/configuration/dns/rule/) එක user config එකට Add වෙන.

### **`info`**

rule එක ගැන Details තියෙන්නෙ මේකෙ. මේක කියවලා User ට හිතාගන්න පුලුවන්නෙ ඒකෙන් වෙන දේ මොකක්ද කියලා.

### **`reqirments`**

මේකෙ තියෙන්නෙ rule එක user ගෙ config එකට add කරන්න කලිම් user ගෙන් Inputs අරගෙන් එව්වා rule එකටම add කරන එක `static` `true` දීලා තිබ්බොත් User ගෙන් කිසිම input එකක් ගන්නෙ නැහැ.

reqirments object

```json
{
  "static": true,
  "needport": true,
  "needportRange": true,
  "needdomain": true,
  "needprotocole": true,
  "needip_cidr": true
}
```

## **`static`**

මේක true කලොත් user ගෙන් කිසිම input එකක් ගන්නෙ නැහැ.

### **`needport`**

මේක true කලොත් user ගෙන් Port එකක් ඉල්ලනවා. ඉල්ලලා ඒක `port` field එක විදියට sing box rule object එකට add වෙනවා.

### **`needportRange`**

මේක true කලොත් user ගෙන් Port range එකක් ඉල්ලනවා. ඉල්ලලා ඒක `port` field එක විදියට sing box rule object එකට add වෙනවා.

### **`needdomain`**

මේක true කලොත් user ගෙන් comma separated domain list එකක් ඉල්ලනවා. ඉල්ලලා ඒක `domain` field එක යටතට sing box rule object එකට add වෙනවා.

### **`needprotocole`**

මේක true කලොත් user ගෙන් comma separated protocole name list එකක් ඉල්ලනවා. ඉල්ලලා ඒක `domain` field එක යටතට sing box rule object එකට add වෙනවා.

### **`needip_cidr`**

මේක true කලොත් user ගෙන් comma separated ip_cidr list එකක් ඉල්ලනවා. ඉල්ලලා ඒක `ip_cidr` field එක යටතට sing box rule object එකට add වෙනවා.
