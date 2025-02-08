#Message Templates

```yaml
grp_welcome:
  en:
    msg_template: |
      ✨ Welcome to 𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 Chat!
      Hi {{if .Username}}@{{.Username}}{{else}}<a href="tg://user?id={{.TgId}}">{{.Name}}</a>{{end}} 👋, it's great to see you here! 🎉
      {{if .IsInChannel}}
      <b>🌟 You're already part of our amazing channel! 🚀</b>
      🎯 <i>Simply start our bot to continue your journey with us.</i>
      We’re thrilled to have you on board! 💬
      {{else}}
      <b>🚨 Important: You haven’t subscribed to our channel yet! 📢</b>
      🔗 Please <a href="{{.ChanLink}}">click here</a> to subscribe to unlock exciting content and get started.
      We promise it’ll be worth it! ✨
      {{end}}
      💡 Need help or have questions? Feel free to reach out—we’re here to assist! 🤝
      Enjoy your time in the 𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 community! 🌐
    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
```

            msg_template
          parse_mode" yaml:"parse_mode" `
         include_media" yaml:"include_media"`
          media_type" yaml:"media_type"`
            media_id" yaml:"media_id" `
               -" yaml:"-"`
        continue_media" yaml:"continue_media"`
           disabled" yaml:"disabled"`
           skip_text" yaml:"skip_text"`
        contin_skip_text" yaml:"contin_skip_text"`
        supercontinue" yaml:"supercontinue"`
        alt_med_url" yaml:"alt_med_url"`
       alt_med_path" yaml:"alt_med_path
    media_skip" yaml:"media_skip"`
