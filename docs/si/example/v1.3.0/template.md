#Message Template

මේ පහල තියෙන්නෙ Connected channel එකේ use කරපු Template එක ඒක messages වෙනස් කරලා use කරන්න පුලුවන්. මේකෙ කිසිම Media එකක් (photo/video) එකක් Use කරන්නෙ නැ. කැමතිනම් add කරගන්න වෙනස් කරගන්න ඕනෙ එකක් කරගන්න පුලුවන්.

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
grp_comeback:
  en:
    msg_template: |
      Hi {{if .Username}}@{{.Username}}{{else}}<a href="tg://user?id={{.TgId}}">{{.Name}}</a>{{end}} 👋, it's great to see you back in here! 🎉
      {{if .IsInChannel}}
      <b>🌟 You're already part of our amazing channel! 🚀</b>
      🎯 <i>Simply start again our bot to continue your journey with us.</i>
      We’re thrilled to have you on board! 💬
      {{end}}
    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
chan_welcome:
  en:
    msg_template: "Hi {{.Name}} Welcome Message Template user who did not started bot and joined group and chan both "
    parse_mode: ""
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: true
chan_comeback:
  en:
    msg_template: "Hi {{.Name}} Hello welcome back to our group You are still not in our channel if you want to use again service please rejoin channel  "
    parse_mode: ""
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: true
dm_welcome: #welcome message for inbox both channel and group
  en:
    msg_template: |
      Hi {{.Name}} You joind our {{.Chat}}
    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: true
dm_verified:
  en:
    msg_template: |
      🎉 <b>Hi {{.Name}}, Congratulations! </b>

      ✨ You are now a verified user. 🎖️
      🚀 Enjoy full access to our amazing service—start exploring and make the most out of it!

      💡 If you need help or have questions, don’t hesitate to reach out. Welcome aboard! 🌟
      💡send /start get started
    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
dm_verified_again:
  en:
    msg_template: |
      <b>🎉 Hi {{.Name}}, Welcome Back!

      ✨ You’ve been verified again and can now access the service. 🎖️
      🚀 Dive in and enjoy all the features we offer—happy exploring!
      💡 Need assistance? We're here to help. Glad to have you back!</b>

      💡 Send /start get started again
    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
chat_mem_left:
  en:
    msg_template: |
      <b>✨ Goodbye, {{.Name}}! ✨</b>

      We're sad to see you go 😢, but if you ever change your mind, you're always welcome to rejoin 🌟.

      Stay amazing, {{.Username}}! 💫 (Your ID: {{.TgId}})
      Take care and see you soon! 🚀"

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false

########## Message Releted To start command
start_monthlimited:
  en:
    msg_template: |
      Hi <b>{{.Name}}</b>, this month is limited for you. Your limit will end in <b>{{.LimitendIn}}</b> days.
    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwOj1naT6nb3yLQgEiqVbbLCpKRdqOCQAC_70xG9TUSFcW58STzcK_zQEAAwIAA3kAAzYE"
    continue_media: false
    disabled: false
    alt_med_path: "./res/bot_logo.png"
start restricted:
  en:
    msg_template: |
      Hi <b>{{.Name}}</b> You Are 🔴 Restricted 🔴 By admin, you may need to contact admin
    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwOj1naT6nb3yLQgEiqVbbLCpKRdqOCQAC_70xG9TUSFcW58STzcK_zQEAAwIAA3kAAzYE"
    continue_media: false
    disabled: false
    alt_med_path: "./res/bot_logo.png"
start_newuser: ## Completly New user start the bot
  en:
    msg_template: |
      👋 <b>Hi {{.Name}},</b> Welcome to <b>𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 BOT</b>! 🚀

      ✨ <b>Introducing 𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 BOT</b> ✨

      With <b>𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 BOT</b>, you can:

      🔧 <b>Create and Manage Configs</b>
      🔄 <b>Configure Inbounds and Outbounds</b>
      📄 <b>View Usage Stats</b>
      🛠️ <b>Build Custom Sing-box Configurations</b>
      🎉 <b>Claim Exciting Events</b> a
      🎁 <b>And More...</b>

      <code>════════════════════════</code>

      👨‍💻 <b>Developer</b>: <a href="https://t.me/DaRker_WoLF">Wolf</a>

      <code>════════════════════════</code>

      Need help for getting started? Just send /help 📩

      ⚠️ <b>Note:</b> You must join our group and channel to use this bot. Make sure you’re subscribed to both! 😊

    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwOj1naT6nb3yLQgEiqVbbLCpKRdqOCQAC_70xG9TUSFcW58STzcK_zQEAAwIAA3kAAzYE"
    continue_media: false
    disabled: false
    alt_med_path: "./res/bot_logo.png"
start_newuser_verified: ## Verified User start the bot first time
  en:
    msg_template: |
      👋 <b>Hi {{.Name}},</b> Welcome to <b>𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 BOT</b>! 🚀

      ✨ <b>Introducing 𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 BOT</b> ✨

      With <b>𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 BOT</b>, you can:
      🔧 <b>Create and Manage Configurations</b>
      🔄 <b>Configure Inbounds and Outbounds</b>
      📊 <b>Track and View Usage Statistics</b>
      🛠️ <b>Build Custom Sing-box Configurations</b>
      🎉 <b>Participate in Exciting Events</b>
      🎁 <b>And Explore More Features...</b>

      <code>════════════════════════</code>

      👨‍💻 <b>Developer:</b> <a href="https://t.me/DaRker_WoLF">Wolf</a>

      <code>════════════════════════</code>

      💡 Need guidance? Just send /help to get started! 🚀
      💬 You're already in our group and channel, so you can start now 🎉

    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwOj1naT6nb3yLQgEiqVbbLCpKRdqOCQAC_70xG9TUSFcW58STzcK_zQEAAwIAA3kAAzYE"
    continue_media: false
    disabled: false
    alt_med_path: "./res/bot_logo.png"
start_newuser_unverified: ## UnVerified User start the bot first time
  en:
    msg_template: |
      👋 <b>Hi {{.Name}},</b> Welcome to <b>𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 BOT</b>! 🚀

      ✨ <b>Introducing 𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 BOT</b> ✨

      With <b>𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 BOT</b>, you can:

      🔧 <b>Create and Manage Configs</b>
      🔄 <b>Configure Inbounds and Outbounds</b>
      📄 <b>View Usage Stats</b>
      🛠️ <b>Build Custom Sing-box Configurations</b>
      🎉 <b>Claim Exciting Events</b>
      🎁 <b>And More...</b>

      <code>════════════════════════</code>

      👨‍💻 <b>Developer</b>: <a href="https://t.me/DaRker_WoLF">Wolf</a>

      <code>════════════════════════</code>

      💡 Need help getting started? Just send <code>/help</code> 📩

      ⚠️ <b>Note:</b> You must join our group and channel to use this bot. Make sure you’re subscribed to both! 😊

    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwOj1naT6nb3yLQgEiqVbbLCpKRdqOCQAC_70xG9TUSFcW58STzcK_zQEAAwIAA3kAAzYE"
    continue_media: false
    disabled: false
    alt_med_path: "./res/bot_logo.png"
start_regular: ##Regular Message when user Alredy started and verified
  en:
    msg_template: |
      👋 <b>Welcome back, {{.Name}}!</b>

      Here's a quick summary of your account:

      - 💡 <b>Total Quota:</b> {{.CalculatedQuota}}
      - 📊 <b>Current Month Usage:</b> {{.MUsage}}
      - ⏳ <b>All-time Usage:</b> {{.Alltime}}
      <tg-spoiler>Only Shows Usage Status Until Last Db refresh Please Use /status or /getinfo to get realtime usage</tg-spoiler>

      You’re doing great! Keep an eye on your usage to stay within your limits. 😊

      Need help? Send /help anytime! 📩
    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwOj1naT6nb3yLQgEiqVbbLCpKRdqOCQAC_70xG9TUSFcW58STzcK_zQEAAwIAA3kAAzYE"
    continue_media: false
    # btnconf:
    #   btns: [["btn1[https://t.me]"]]
    disabled: false
    alt_med_path: "./res/bot_logo.png"
start_removed: # when removed user start the bot
  en:
    msg_template: |
      👋 <b>Hi {{.Name}},</b>

      {{if .IsInChannel}}
      ⚠️ You are not in our <b>group</b>. Please join our <a href="https://t.me/yourchatgroup">chat group</a> to become a verified user.
      {{else if .IsinGroup}}
      ⚠️ You are not subscribed to our <b>channel</b>. Please subscribe to our <a href="https://t.me/yourchannel">channel</a> to become a verified user.
      {{else}}
      ⚠️ You need to join <b>both</b> our <a href="https://t.me/yourchannel">channel</a> and <a href="https://t.me/yourchatgroup">group</a> to use this bot.
      {{end}}

      ℹ️ <b>Current Status:</b>
      - In Group: {{.IsinGroup}}
      - In Channel: {{.IsInChannel}}

    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwOj1naT6nb3yLQgEiqVbbLCpKRdqOCQAC_70xG9TUSFcW58STzcK_zQEAAwIAA3kAAzYE"
    continue_media: false
    disabled: false
    alt_med_path: "./res/bot_logo.png"

##Create Command

create_out_info:
  en:
    msg_template: |
      📡 <b>Selected Outbound Information</b>

      - 🌐 <b>Outbound Name:</b> {{.OutName}}
      - 🔄 <b>Outbound Type:</b> {{.OutType}}
      - 🛠️ <b>Outbound Info:</b> {{.OutInfo}}
      - ⏱️ <b>Latency:</b> {{.Latency}}ms

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
create_available_quota:
  en:
    msg_template: |
      💡 You have an available quota of <b>{{.Quota}}</b>.
      ❓ Please specify how much quota you want to dedicate to this config.
      ⚙️ Enter your desired quota amount and proceed!
    parse_mode: HTML
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
create_result:
  en:
    msg_template: |
      You have Success Fully Created Config use /getinfo => Configs  to see them & export according to you'r package and relevent Inbound
      db.Config Info Available Here
    parse_mode: HTML
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false

# Command Refer
refer_home:
  en:
    msg_template: |
      👋 Hello <b>{{.Name}}</b>!

      🌟 You have referred 📦 <b>{{.Refred}}</b> users so far, and 📦 <b>{{.Verified}}</b> of them are verified.
      🎁 <b>Important:</b> To claim your gift, you need verified users.

      🛡️ <i>(A verified user is someone who has joined both the group and the channel.)</i>

      🚀 Keep inviting and growing your network to unlock exciting rewards!

    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: ""
    continue_media: false
    disabled: false
    alt_med_path: "./res/bot_logo.png"
refer_share:
  en:
    msg_template: |
      🚀 Hey Everyone!

      I’m <b>{{.Name}}</b> and I’ve got something awesome for you! 🎉
      🔗 Start This through My referral link: <a href="{{.Botlink}}">Start bot</a>
      💡 This bot lets you easily create and manage VPN configs—simple, fast, and convenient! Don’t miss out, join now and experience it yourself! 🚀
    parse_mode: "HTML"
    include_media: false
    media_type: photo
    media_id: AgACAgUAAxkBAAEvZapnOFjKbTYqbIAgTIy0xf-VZQQ0PwACUcAxG8dwyVV7OMp6tw3cUwEAAwIAA3MAAzYE
    continue_media: false
    disabled: false
    alt_med_path: "./res/bot_logo.png"

#Command Cap
setcap_warn:
  en:
    msg_template: |
      <b> Warning: You are about to cap your quota. Once you do this, it cannot be undone until the period is over. If you want to continue, press continue. You can set the cap quota between {{.CapRange}} </b><b> Warning: You are going to cap you'r quota When You do this you can't undo this until it's over, if you want to continue press continue, You can set cap quota between {{.CapRange}} </b>
    parse_mode: HTML
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
setcap_already:
  en:
    msg_template: |
      <b>
      ⏳ Your cap will end on <b>{{.EndDate}}</b>.
      🔔 Be sure to check your usage and take necessary action before the end date!
      </b>
    parse_mode: HTML
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
setcap_reply:
  en:
    msg_template: |
      🚀 Hi!
      Please provide the quota you'd like to cap, ensuring it's between {{.CapRange}}.
      This quota will apply until period over, and changes can't be reverted during this period.
      💡 Tip: Be precise while setting your cap to align with your needs.
    parse_mode: ""
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
setcap_get:
  en:
    msg_template: |
      📢 <b>Important Notice</b>

      capble range {{.CapRange}}

      ⚠️ You are requested to send a new capped quota.
      🔒 <b>Note:</b> The new capped quota must be in range {{.CapRange}}<b></b>.

      🛠️ All your configurations' quotas will be adjusted based on the new main quota.
      ⏳ This change will remain in effect for <b>days you enter</b>.
      ❌ <b>During this period, you cannot undo the change.</b>

      📆 After days you selected, your main quota will be automatically updated.

      ✨ Please proceed carefully to ensure your requirements are met!

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false

#Command Gift
gift_reciver:
  en:
    msg_template: |
      🎉 Hi <b>{{.Name}}</b>!
      🎁 You have received a <b>{{.Gift}}</b> gift from <b>{{.FromUser}}</b>!
      ✨ Enjoy your gift and make the most of it.

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
gift_send:
  en:
    msg_template: |
      ❓ Hi! You currently have <b>{{.LeftQuota}}</b>.
      📩 How much of your quota would you like to send?
    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false

#Command Status
status_home:
  en:
    msg_template: |
      ✅ <b>Your Usage</b> ✅
      📊 <b>Usage For Last {{.UsageDuration}} </b>
      ⬇️ <b>Total Download:</b> {{.TDownload}}
      ⬆️ <b>Total Upload:</b> {{.TUpload}}

      📅 <b>All Time Usage</b>
      ⬆️ <b>Uploaded Total:</b> {{.MUpload}}
      ⬇️ <b>Downloaded Total:</b> {{.MDownload}}
      🕒 <b>Overall:</b> {{.Alltime}} | 📆 <b>This Month:</b> {{.MonthAll}}

      ✨ <i>If you want config-specific usage, click the buttons below.</i> ✨

    parse_mode: HTML
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwFPpnYv0SstubUhwsqM5YMmThlsVsUgACEsMxG4p3GVfhFKnckNKn3QEAAwIAA3kAAzYE"
    continue_media: false
    disabled: false
    alt_med_path: "./res/status.jpg"
status_callback: #THis is Callback Answere media sending and html parsing unsupported
  en:
    msg_template: |
      🚀 Your Usage Summary 🚀
      ⬆️ Uploaded Total: {{.MUpload}}
      ⬇️ Downloaded Total: {{.MDownload}}

      📊 Usage for the Last {{.UsageDuration}}
      ⬇️ Total Download: {{.TDownload}}
      ⬆️ Total Upload: {{.TUpload}}
    parse_mode: ""
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false

## Command Getinfo
getinfo_home:
  en:
    msg_template: |
      🔰 This feature allows you to <b> access detailed 🗓 information about your account and configurations.

      Simply choose what you'd like to explore! 〽️

      ✨ What would you like to do?

      ⚙️ View User Information
      ⚙️ View Configuration Information

      🚀 Select an option to proceed! 🚀 </b>
    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwFPxnYv0SYmTbaBM9WcDl7HhdZXfwkgACE8MxG4p3GVdFztHVaMXHPgEAAwIAA3kAAzYE"
    continue_media: true
    supercontinue: true
    contin_skip_text: true
    disabled: false
    alt_med_path: "./res/getinfo.jpg"
getinfo_user:
  en:
    msg_template: |
      👤 <b>Name:</b> {{.Name}}  
      📛 <b>Username:</b> @{{.Username}}  
      🆔 <b>Telegram ID:</b> {{.TgId}}  

      💼 <b>Account Details:</b>  
      🎯 <b>Dedicated:</b> {{.Dedicated}}  
      📦 <b>Total Quota:</b> {{.TQuota}}  
      🔋 <b>Left Quota:</b> {{.LeftQuota}}  
      🛠️ <b>Configurations Count:</b> {{.ConfCount}}  
      📈 <b>Total Usage:</b> {{.TUsage}}  
      📊 <b>Usage Percentage:</b> {{.UsagePercentage}}  
      📅 <b>Joined:</b> {{.Joined}}  
      🌀 <b>Empty Cycle:</b> {{.NonUseCycle}}  
      {{if .Isdisuser}}  
      📉 <b>Distribution Ending In:</b> {{.Disendin}}  
      {{end}}  
      ⏳ <b>Usage Resets In:</b> {{.UsageResetIn}} Days  
      🧮 <b>Tempory-Limitation-Ratio:</b> {{.TempLimitRate}}  (If this value is zero, it means you're totally limited for the month)
        
      🔍 <b>Account Flags:</b>  
      🔒 <b>Capped:</b> {{.Iscapped}}  
      🛡️ <b>Templimited:</b> {{.IsTemplimited}}  
      ✅ <b>Verified:</b> {{.IsVerified}}  
      🎁 <b>Gift:</b> {{.Isgifted}}  
      🚫 <b>Distributed User:</b> {{.Isdisuser}}  
      📆 <b>Is Month Limited:</b> {{.IsMonthLimited}}  
      📍 <b>Joined Place:</b> {{.JoinedPlace}}  

      🎫 <b>Gift Quota:</b> {{.GiftQuota}}  
      📊 <b>Alltime Usage:</b> {{.AlltimeUsage}}  
      ⭐ <b>Points:</b> {{.Points}}  
      ⏸️ <b>Paused:</b> {{.Paused}}  

      📆 <b>Joined:</b> {{.Joined}}  
      {{if .Iscapped}}  
      📅 <b>Cap Days:</b> {{.CapDays}}
      📉 <b>Capped Quota:</b> {{.CappedQuota}}  
      ⚠️ <b>Cap Ending In:</b> {{.CapEndin}}  
      {{end}}
    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
    alt_med_path: "./res/getinfo.jpg"
getinfo_config:
  en:
    msg_template: |
      <b>📋 {{.Name}} - Configuration Summary</b>

      <b>🌐 Total Quota:</b> {{.TotalQuota}}

      {{if .Active}}<b>🟢 Status:</b> <i>Online</i>{{else}}<b>🔴 Status:</b> <i>Offline</i>{{end}}

      <b>🛠 Configuration Information</b>
      ▫️ <b>Name:</b> {{.ConfigName}}  
      ▫️ <b>UUID:</b> <code>{{.ConfigUUID}}</code>
      ▫️ <b>Password:</b> {{.ConfigPassword}} (for inbounds which require password ex - trojan)

      <b>📊 Usage Overview</b>
      ▫️ <b>Used Percentage:</b> {{.UsedPresenTage}}%  
      ▫️ <b>Upload:</b> {{.ConfigUpload}}  
      ▫️ <b>Download:</b> {{.ConfigDownload}}  
      ▫️ <b>Total Usage:</b> {{.ConfigUsage}}

      📌 <b>Login Limit:</b> {{.Loginlimit}}

      <b>📅 Last {{.UsageDuration}} Usage</b>
      🔼 <b>Upload:</b> {{.ConfigUploadtd}}  
      🔽 <b>Download:</b> {{.ConfigDownloadtd}}  
      📈 <b>Usage:</b> {{.ConfigUsagetd}}

      <b>🔄 Outbound Information</b>
      ▫️ <b>Name:</b> {{.OutName}}  
      ▫️ <b>Type:</b> {{.OutType}}  
      ▫️ <b>Details:</b> {{.OutInfo}}

      {{if .Active}}⚡ <b>Latency:</b> {{.Latency}}ms  
      🟢 <b>Online Clients:</b> {{.Online}}

      <b>🌍 Connected IPs</b>
      {{- range $key, $value := .IpMap }}
      ▫️ <b>IP:</b> <code>{{$key}}</code> | <b>Connections:</b> {{$value}}
      {{- end }}
      {{end}}
      <b>⏳ Usage Resets In:</b> {{.ResetDays}} Days

    parse_mode: HTML
    include_media: false
    media_type: ""
    media_skip: false
    media_id: ""
    continue_media: false
    disabled: false
  sin:
    msg_template: sinhala test
    parse_mode: HTML
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
getinfo_out:
  en:
    msg_template: |
      <b>
      🌐✨ Outbound Information

      📛 Name: {{.OutName}}
      ℹ️ Info: {{.OutInfo}}
      ⚡ Latency: {{.Latency}}
      📦 Type: {{.OutType}}

      💡 🔍 Use this information to select the 🌟 outbound that exactly matches your needs for the best performance and seamless connectivity! 🚀
      </b>
    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwFPxnYv0SYmTbaBM9WcDl7HhdZXfwkgACE8MxG4p3GVdFztHVaMXHPgEAAwIAA3kAAzYE"
    continue_media: true
    supercontinue: true
    contin_skip_text: true
    disabled: false
#new
getinfo_allin:
  en:
    msg_template: |
      <b>🔗 You have added {{.InboundCount}} inbound{{if ne .InboundCount 1}}s{{end}} to this config:</b>

      {{- range .AllIn }}
      <b>📥 Inbound Configuration</b>

      ▫️ <b>Name:</b> {{.Name}}  
      📦 <b>Type:</b> {{.Type}}  
      🔌 <b>Listen Port:</b> {{.Listenport}}  
      🔐 <b>TLS Enabled:</b> {{.Tlsenabled}}
      🌐 <b>Domain:</b> {{.Domain}}  
      🌍 <b>Public IP:</b> {{.PublicIp}}

      🛠 <b>Description:</b> {{.Custom_info}} 

      <b>🔗 Proto URL:</b>  
      <code><pre>{{.ProtoUrl}}</pre></code>
      {{- end }}
      <b>📌 Choose an inbound to see more details.</b>
    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwFPxnYv0SYmTbaBM9WcDl7HhdZXfwkgACE8MxG4p3GVdFztHVaMXHPgEAAwIAA3kAAzYE"
    continue_media: true
    supercontinue: true
    contin_skip_text: true
    disabled: false
    alt_med_path: "./res/getinfo.jpg"
getinfo_confin:
  en:
    msg_template: |
      <b>📌 Selected Inbound Info</b>

      ▫️ <b>Name:</b> {{.Name}}  
      ▫️ <b>Tag:</b> {{.Tag}}  
      📦 <b>Type:</b> {{.Type}}  
      📍 <b>Listen Address:</b> {{.ListenAddres}}  
      🔌 <b>Listen Port:</b> {{.Listenport}}  
      🔐 <b>TLS Enabled:</b> {{.Tlsenabled}}  
      🔀 <b>Transport Type:</b> {{.TransPortType}}  
      📁 <b>Transport Path:</b> {{.TransPortPath}}  
      🌐 <b>Domain:</b> {{.Domain}}  
      🌍 <b>Public IP:</b> {{.PublicIp}}
      🛠 <b>Description:</b> {{.Custom_info}} 

      <b>🔗 Proto URL:</b>  
      <code><pre>{{.ProtoUrl}}</pre></code>

    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwFPxnYv0SYmTbaBM9WcDl7HhdZXfwkgACE8MxG4p3GVdFztHVaMXHPgEAAwIAA3kAAzYE"
    continue_media: true
    supercontinue: true
    contin_skip_text: true
    disabled: false
    alt_med_path: "./res/getinfo.jpg"
getinfo_in_export:
  #same as confin
  en:
    msg_template: |
      <b>📌 Selected Inbound Info</b>

      ▫️ <b>Name:</b> {{.Name}}  
      ▫️ <b>Tag:</b> {{.Tag}}  
      📦 <b>Type:</b> {{.Type}}  
      📍 <b>Listen Address:</b> {{.ListenAddres}}  
      🔌 <b>Listen Port:</b> {{.Listenport}}  
      🔐 <b>TLS Enabled:</b> {{.Tlsenabled}}  
      🔀 <b>Transport Type:</b> {{.TransPortType}}  
      📁 <b>Transport Path:</b> {{.TransPortPath}}  
      🌐 <b>Domain:</b> {{.Domain}}  
      🌍 <b>Public IP:</b> {{.PublicIp}}
      🛠 <b>Description:</b> {{.Custom_info}}  

      🔒 <b>SNI:</b> {{.Sni}}  
      🛰 <b>Host:</b> {{.Host}}  
      🖥 <b>Server:</b> {{.Server}}

      <b>🔗 Proto URL:</b>  
      <code><pre>{{.ProtoUrl}}</pre></code>

    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwFPxnYv0SYmTbaBM9WcDl7HhdZXfwkgACE8MxG4p3GVdFztHVaMXHPgEAAwIAA3kAAzYE"
    continue_media: true
    supercontinue: true
    contin_skip_text: true
    disabled: false
    alt_med_path: "./res/getinfo.jpg"
#common
inbound_info:
  en:
    msg_template: |
      <b>📥 Inbound Information</b>  

      🔖 <b>Name:</b> {{.Name}}  
      🏷 <b>Tag:</b> {{.Tag}}  
      📦 <b>Type:</b> {{.Type}}  
      📡 <b>Support Info:</b> {{range .Support}}<code>{{.}}</code> {{end}}

      <b>📍 Listening Info</b>
      🌐 <b>Address:</b> {{.ListenAddres}}  
      🔌 <b>Port:</b> {{.Listenport}}  
      🔐 <b>TLS Enabled:</b> {{.Tlsenabled}}

      <b>🔀 Transport Info</b>
      🚚 <b>Transport Type:</b> {{.TransPortType}}  
      🛤 <b>Transport Path:</b> <code>{{.TransPortPath}}</code>  

      <b>🛠 Extra Info</b>
      📝 <b>Description:</b> {{.Custom_info}}  
      🌍 <b>Domain:</b> {{.Domain}}  
      📡 <b>Public IP:</b> {{.PublicIp}}

    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwFPxnYv0SYmTbaBM9WcDl7HhdZXfwkgACE8MxG4p3GVdFztHVaMXHPgEAAwIAA3kAAzYE"
    continue_media: true
    supercontinue: true
    contin_skip_text: true
    disabled: false
    alt_med_path: "./res/getinfo.jpg"

#Command Configure
configure_home:
  en:
    msg_template: |
      ⚙️ Here, you can adjust and customize various settings, including modifying <b>inbounds📥 and outbounds 📤 </b> , to perfectly suit your needs.

      ✨ Take full control of your configurations and unlock the bot's full potential!😁

      <b> 💡 If you're unsure where to start, use the /help ✅ command and watch the available tutorials. </b>

    parse_mode: "HTML"
    include_media: false
    media_type: photo
    media_id: AgACAgUAAxkBAAEwFPZnYv0QOWtlTkwlrRviRwMAAfRDPFYAAhDDMRuKdxlXK89bhFgCwkUBAAMCAAN5AAM2BA
    continue_media: true
    supercontinue: true
    contin_skip_text: true
    disabled: false
    alt_med_path: "./res/configure.jpg"
conf_configure:
  en:
    msg_template: |
      <b>👋 You have selected the config 🪀 {{.ConfName}} 🪀</b>

      ⭕️ In this section, you can 🛠 <b>manage and modify</b> 🛠 your configuration settings.

      😼 Simply choose what you'd like to adjust!
      ✨ <b>What would you like to do?</b>

      <b>1️⃣ Change Inbound 📥</b>  
      <b>2️⃣ Change Outbound 📤</b>  
      <b>3️⃣ Change Name 〽️</b>  
      <b>4️⃣ Delete Configuration 🗑</b>  
      <b>5️⃣ Change Quota 🌊</b>

      🔒 <i>Always close the session after making changes—it's best practice. 🫢</i>

      🚀 Select an option to proceed!

      <b>📄 Configuration Details</b>  

      ▫️ <b>Name:</b> {{.ConfName}}  
      ▫️ <b>🌐 Available Quota:</b> {{.TotalQuota}}  
      ▫️ <b>{{if .Active}}🟢 Online{{else}}🔴 Offline{{end}}</b>
      ▫️ <b>🔐 Login Limit:</b> {{.Loginlimit}}
    parse_mode: "HTML"
    include_media: false
    media_type: photo
    media_id: AgACAgUAAxkBAAEwDQhnYcTjo1CNV6dteoj-mIUth5nVpAACa70xG4p3EVeTGl_WV_9blwEAAwIAA3MAAzYE
    continue_media: false
    disabled: false
conf_name_change:
  en:
    msg_template: |
      ❓ are you sure you want to change your name to <b>{{.NewName}}</b>?
      ⚠️ Please confirm if this is correct before proceeding.

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
conf_quota_change:
  en:
    msg_template: |
      📥 Please provide the new quota for config {{.ConfName}}.
      The new quota should be below: {{.AvblQuota}}

      Enter the new quota to proceed. 🚀
    parse_mode: HTML
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
conf_in_change:
  en:
    msg_template: |
      {{if .ActionRemove}}
      <b>❌ You are about to remove this inbound from the config.</b>
      {{else}}
      <b>✅ You are about to add this inbound to the config.</b>
      {{end}}

      <b>📥 Inbound Information</b>  
      🔖 <b>Name:</b> {{.Name}}  
      🏷 <b>Tag:</b> {{.Tag}}  
      📦 <b>Type:</b> {{.Type}}  
      📡 <b>Support Info:</b> {{range .Support}}<code>{{.}}</code> {{end}}

      <b>📍 Listening Info</b>  
      🌐 <b>Address:</b> {{.ListenAddres}}  
      🔌 <b>Port:</b> {{.Listenport}}  
      🔐 <b>TLS Enabled:</b> {{.Tlsenabled}}

      <b>🔀 Transport Info</b>  
      🚚 <b>Transport Type:</b> {{.TransPortType}}  
      🛤 <b>Transport Path:</b> <code>{{.TransPortPath}}</code>

      <b>🛠 Extra Info</b>  
      📝 <b>Description:</b> {{.Custom_info}}  
      🌍 <b>Domain:</b> {{.Domain}}  
      📡 <b>Public IP:</b> {{.PublicIp}}

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
conf_out_change:
  en:
    msg_template: |
      🚀 <b>Selected Outbound Info:</b>

      🔹 <b>OutName:</b> {{.OutName}}
      🌐 <b>OutType:</b> {{.OutType}}
      📝 <b>OutInfo:</b> {{.OutInfo}}
      ⏱️ <b>Latency:</b> {{.Latency}}ms

      💡 Please verify that everything is correct before proceeding!

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false

##Events Templates
event_home:
  en:
    msg_template: |
      👋 Hi <b>{{.Name}}</b>!

      ✅ There are <b>{{.AvblCount}}</b> events available for you.
      🏆 You have successfully completed <b>{{.Completed}}</b> events so far!

      🌟 Keep up the great work and aim for more achievements!

    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: ""
    alt_med_path: "./res/events.jpg"
    continue_media: false
    disabled: false

#points
points_home:
  en:
    msg_template: |
      👋 Hi <b>{{.Name}}</b>!

      ⭐ You currently have <b>{{.Count}}</b> available points.

      🎯 Keep earning and using your points wisely!

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false

#distribute Command
distribute_group: #msg sent to group when user distribute the quota
  en:
    msg_template: |
      👋 <b>Hello Everyone!</b>
      👏 Let's all give a big round of applause to <b>{{.Name}}</b>! 🎉
      🎁 {{.Name}} have generously distributed his/her quota, dedicating <b>{{.Disquota}}</b> to the group.
      🙏 A heartfelt thank you for his/her kindness and support! 💖

    parse_mode: "HTML"
    include_media: false
    media_type: photo
    media_id: AgACAgUAAxkBAAEvZapnOFjKbTYqbIAgTIy0xf-VZQQ0PwACUcAxG8dwyVV7OMp6tw3cUwEAAwIAA3MAAzYE
    continue_media: false
    disabled: true

##Templates for builder

builder_home:
  en:
    msg_template: |
      👋 Hi <b>{{.Name}}</b>, welcome to the Config Builder!

      🔧 You’ve created <b>{{.ConfCount}}</b> configs so far. Great job!

      ⚡️ Some features of config builder are not available yet, but don't worry—they will be developed soon!

      Stay tuned for more updates! 🚀

    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwE6FnYt1K_DrxnaFjifGaUZTs-1WhWgAC6MIxG4p3GVfonvtJDsoLMAEAAwIAA3MAAzYE"
    continue_media: false
    disabled: false
    alt_med_path: "./res/builder.jpg"

##ParserCommon
com_unverified:
  en:
    msg_template: |
      ⚠️ Hi <b>{{.Name}}</b>,

      🔑 You need to be a verified user to use this feature.

      ✅ To verify, please join both the bot group and channel.
      💬 Once verified, you'll be able to access all features!

    parse_mode: "HTML"
    include_media: false
    alt_med_path: "./res/chan_logo.png"
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwE6FnYt1K_DrxnaFjifGaUZTs-1WhWgAC6MIxG4p3GVfonvtJDsoLMAEAAwIAA3MAAzYE"
    continue_media: false
    disabled: false

chan_share:
  en:
    msg_template: |
      <b>
      ⛔️ Groups 🫠 පීරලා Config හොයලා එපා වෙලාද❔
      ⛔️ හම්බෙන 🥴 එව්වත් වැඩ නැද්ද ❔
      ⛔️ වැඩ කලත් උපරිම දවස් තුනයිද😅 ❔
      ⛔️ Lk ip  Config 😁 එකක් තිබ්බොත් කොහොමද❔
      ⛔️ Torrent Download 🥺 කරන්න  බැරිද❔
      ⛔️ SING-BOX Try  කරන්න කැමති 😕   වුණත් Configs නැද්ද❔

      💥 INTRODUCING 💫 𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 💫

      🌚 What We Offer ❔

      🔌 Create Your Own Configurations
      ⚙️ Effortless Configs Management
      🛠 Real-Time Usage and Connection Status
      🗳 Custom Sing-Box Configurations 🛠
      ✨ Exciting Events and Offers Await!


      ⚡️VPN Info ⚡️

      🎯Easily Create & Mange Configs using Our Bot 🧩
      🎯 See Usage Status With Our Bot  📊
      🎯 Support For Torrent Downloading 📥
      🎯 Latest Sing-Box Core Running on Server Side 🚀
      🎯 Optimized Routing For Low Latancy ⚡️
      🎯 Can Use 50+ SurfShark Premium WireGuard Outbound
      🎯 Guaranteed:😼 Your configuration works  As long as Our service is live 🗿
      🎯 Your configs works until your usage ends and renews automatically in 30 days! 🎉

        And More! 📂 Customize features to perfectly match your needs. 📝


      🎉  EVERYTHING YOU NEED, ALL IN ONE PLACE! FOR FREE! 🎉</b>
    parse_mode: "HTML"
    include_media: true
    media_type: "photo"
    media_id: ""
    alt_med_path: "./res/chan_logo.png"
    continue_media: false
    disabled: false

restricted:
  en:
    msg_template: |
      You are Restricted User May Need To contact admin
    parse_mode: "HTML"
    include_media: false
    alt_med_path: "./res/chan_logo.png"
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwE6FnYt1K_DrxnaFjifGaUZTs-1WhWgAC6MIxG4p3GVfonvtJDsoLMAEAAwIAA3MAAzYE"
    continue_media: false
    disabled: false

## help Command
help_home:
  en:
    msg_template: <b> Hi {{.Name}} What Do You Want Know ⁉️ </b>
    parse_mode: "HTML"
    include_media: false
    media_type: photo
    media_id: AgACAgUAAxkBAAEwOj1naT6nb3yLQgEiqVbbLCpKRdqOCQAC_70xG9TUSFcW58STzcK_zQEAAwIAA3kAAzYE
    continue_media: true
    disabled: false
    supercontinue: true
    alt_med_path: "./res/bot_logo.png"
help_info1:
  en:
    msg_template: |
      About Bandwidth

      මුලින්ම කියන්න ඕන කාටවත් constant quota එකක් හම්බෙන්නෙ නැ. ඔයාගෙ configs වැඩ නැතුව යන්නෙ ඔයාට හම්බෙලා තියෙන quota එක ඉවර වුනොත් විතරයි.
      ඔයාගේ account එකට මුලු bandwidth quota එකක් හම්බෙනවා. මෙක constant එකක් නෙමෙ ගොඩක් දේවල් හින්දා වෙනස් වෙන්න පුලුවන්.

      Ex :=
      when a user,
      join both group and channel
      leave from group or channel
      distribute his quota
      setcap to his quota
      vps bandwidth change

      ඔයාගෙ main quota එක වෙනස් වෙන ratio එකෙන්ම config වලට දීලා තියෙන quota ත් වෙනස් වෙනවා බැරි වෙලාහරි usage එක config quota එකට වඩා වැඩි වුනොත් config එක off වෙනවා.
      ඔයාගෙ usage ගොඩක් වැඩිව තිබිලා පස්සෙ ඊට වඩා quota එක අඩු වුනොත් කෝම හරි මාසෙ අගට usage එක quota එකට වඩා වැඩි වුනොත් ඊලග usage reset වෙන දවසෙ ඉදන් next 30 days වලට
      ඒ excess usage add වෙනව.

      ඔයා කාටහරි gift send කරලා තියෙනවනන් ඒ Gift send කරපු quota ත් main quota වෙනස් වන ratio වලින්ම වෙනස් වෙනවා අඩු හෝ වැඩි.

      ඒ වගෙම ඔයා Usage reset day එක වෙනකොට තමන්ගෙ quota එකෙන් 75% Use කරලා තිබ්බෙ නැත්තන් next 30 days ඔයාට කිසිම දෙයක් Use කරන්න වෙන්නෙ නැ. ඒ නිස bandwidth waste කරන්න එපා.
      ඔයාට හම්බෙන bandiwdth වැඩිනම් use කරන කෙනෙකුට gift කරන්න නැත්තන් /setcap use කරන්න.

      ඔයා මේ service use කරන්නෙ නැත්තන් ඔයාට හම්බෙන එව්ව distribute කරලා දාන්න.

    parse_mode: ""
    media_skip: true
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
help_info2:
  en:
    msg_template: Hello {{.Name}} this is help info page  2
    parse_mode: ""
    include_media: false
    media_type: photo
    media_id: AgACAgUAAxkBAAEvZapnOFjKbTYqbIAgTIy0xf-VZQQ0PwACUcAxG8dwyVV7OMp6tw3cUwEAAwIAA3MAAzYE
    continue_media: false
    disabled: false
help_info3:
  en:
    msg_template: Hello {{.Name}} this is help info page  3
    parse_mode: ""
    include_media: false
    media_type: photo
    media_id: AgACAgUAAxkBAAEvZapnOFjKbTYqbIAgTIy0xf-VZQQ0PwACUcAxG8dwyVV7OMp6tw3cUwEAAwIAA3MAAzYE
    continue_media: false
    disabled: false
help_info4:
  en:
    msg_template: "Hello {{.Name}} this is help info page 4 "
    parse_mode: ""
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false

help_cmd1:
  en:
    msg_template: |

      <b> <u>🟢 /start 🟢 </u> </b>

      🔹bot alive ද කියලා බලාගන්න මේ command එක use කරලා.

      <b> <u>🟢 /create 🟢</u> </b>

      🔹 මේ command එක use කරලා තම server එකේ තියෙන inbounds වලට අදාලව ඔයාට config එකක් හදාගන්නෙ මේකෙන් ඉල්ලන දේවල් හරියටම දෙන්න
      outbounds තෝරද්දි noraml use වලට direct දෙන්න, හැබැයි torrent traffic auto m වෙන safe outbound එකකින් තමා route වෙන්නෙ

      inbound තෝරද්දි ඔයාගෙ පැකේජ් වලට අනුව ගැලපෙන එක තෝරන්න පස්සෙ /configure use කරලා වෙනස් කරන්නත් පුලුවන් ඕන්නම්
      ඔයාට හම්බෙන මුලු quota එකෙන් config එකට වෙනම quota එකක් වෙන් කරහැකි මේ වෙන් කරන quota එක constant එකක් නෙමෙ මේක group එකට user la join වෙද්දි leave වෙද්දි  කව්රුහරි quota distribute කරද්දි වගේ වෙනස් වෙනවා.
      ඔයා මේ වෙන් කරන quota එක ඉවර උනොත් config එක වැඩ නැතුව යනවා.

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
help_cmd2:
  en:
    msg_template: |
      <b> 🟢 /configure 🟢

      🔹 මේ cmd එකෙන් තමා server එකේ create කරලා තියෙන ඔයාගෙ configs වල ප්‍රදාන වශයෙන් Inbound 📥, outbound 📤 සහ තවත් settings වෙනස් කරන්නෙ. </b>

      මේ command එක send කරහම් මුලින්ම ඔයා create කරල තියෙන configs පෙන්නනවා. ඒකෙන් ඕන එක select 🎯 කරගන්න
      select කරගත්තට පසේ ඔයාට හම්බෙනවා

      (මෙතන Inbound, outbound කියලා සඳහන් කරන්නෙ server එකට අදාලව ඔයගේ device එකට අදාලව නෙමෙ.)

      <b>
      🔹Change Inbound
      🔹Change Outbound
      🔹Chane Name
      🔹Change Quota
      🔹Delete Config
       </b>

       <b><u>◽️ Change Inbound</u>

      🔹 මේකෙදි සරලව වෙන්නෙ ඔයාගෙ inbound එක change කරන එක,</b>

      🔺ඒ කියන්නෙ ඔයාගෙ config එකේ uuid එකෙන් 443 port එකෙන් වැඩ කරන inbound එකක නම් තියෙන්නෙ ඔයාට මේ option එක use කරලා server එකේ තියෙන වෙන inbound එක්කට මාරු කරගන්න පුලුවන්.

      🔺හැබැයි ඔයාගෙ එතකොට තියෙන config එක වැඩ නැතුව යනවා. මොකද inbound change වෙන නිසා.

      <b><u>◽️Change Outbound</u>

      🔹 මෙතනදි වෙන්නෙ ඉතින් Outbound වෙනස් වෙන එක, මේක වෙනස් කරා කියලා ඔයාගෙ දැන් තියෙන config එක වැඩ නැතුව යන්නෙ නැ. හැබැයි ip addr වෙනස් වෙනවා.</b>

      🔺ඔයා Outbound එකක් change කරහම ඔයාගෙ දැනට තියෙ connections close වෙන්නෙ නැ එව්ව එහෙමම පරන outbound එක හරහා read write වෙන්නෙ.

      🔺Outbound එක වෙනස් කරපු වෙලේ ඉදම් යන connection තමා අලුත් outbound එකෙන් route වෙන්නෙ.

      🔺vpn එක disconnect කරලා ආයෙ connect කලොත් ඔක්කොම අලුත් outbound එකෙන් යන්නෙ (තේරුනෙ නැත්තන් වැඩිය හිතනන යන්න එපා)

      🔺regular use එකට direct දාගෙන use කරන්න ඒකෙ තමා අඩුම ping එකක් දෙන්නෙ මොකද server එකෙන් කෙලින්ම destination යන නිසා.

      🔺ඔයාට මොකක් හරි විශේශ හෙතුවක් තියෙනන් අනික් outbound එකක් use කරන්න, outbound විශාල ප්‍රමානයක් දීල තියෙනවා ගොඩක් එව්වා surfshark premium wireguard

      🔺lk outbound එක්කුත් තියේ education block දේවල් වලට මේක use කරලා බලන්න ලොකු ping එකක් සෙට් වෙන්නෙ නැ 100 -115 වගේ.

      🔸ඔයාලට මේ outbound වල latancy බලාගන්න පුලුවන් ඒ ඔන්නම් Ping කරලත් බලන්න පුලුවන් /getinfo command එක හරහා

      🔺ඒ වගේම මේ Outbound නිසා ඔයාලට torrent download කරන්න පුලුවන්. හැබැයි torrent download කරද්දි අනිවාරෙන් surfshark or direct outbound use කරන්න cloudflare warp එව්වා use කරන්න එපා.. එව්වලින් torrent download කරන්නත් බැ.

      <b><u>◽️Change Name </u>

      🔹 මේකෙනම් විස්තර කරන්න දෙයක් නැනෙ නම වෙනස් කරන එකනෙ කරන්නෙ. </b>

      <b><u>◽️Change Quota</u>

      🔹  මේ option එකෙන් පුලුවන් ඔයා දැනට config එකට දීලා තියෙන quota එක change කරන්න. හැබැයි ඔයාගෙ usage එකට වඩා අඩු quota එකක් අලුත් quota එක විදියට දෙන්න බැ  </b>

      <b><u>◽️ Delete Config </u>

      🔹 මේකෙන් ඔයාගෙ config එක සම්පූර්නයෙන් delete වෙලා යනවා. හැබැයි ඔයාගෙ usage එක නම් එහෙමම හිටිනවා.
      </b>

    parse_mode: "HTML"
    media_skip: true
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
help_cmd3:
  en:
    msg_template: |
      Information Gathering

      <b> <u>🟢 /status 🟢 </u> </b>
      🔹 මේකෙන් ඔයාගෙ usage එකේ status ගන්න පුලුවන්.

      <b><u>🟢 /getinfo 🟢 </u> </b>
      🔹 මේකෙන් ඔයාට ගොඩක් විස්තර ගන්න පුලුවන්.

      <b>
      🔹UserInfo
      🔹Configs
      🔹Check Outbound
      </b>

      <b><u>◽️UserInfo </u>
      🔹 මේ option එක use කරලා ඔයාගෙ Information ගන්න පුලුවන්. Usage, Limitation, Acc Flags, etc.. </b>

      <b><u>◽️Configs </u>

      🔹මේකෙන් ඔයාගෙ ඔක්කොම config වල details ගන්න පුලුවන්.
      අදාල config එකට අදාලව ඔයාට හම්බෙනවා, </b>

      දැනට use කරලා තියෙන ප්‍රමාණය, config එකෙ quota එක, අවසාන පැය කීපයක් usage එක
      දැනට config එකෙන් connect වෙලා ඉන්න clients ගණන, client කෙනෙකුට තියෙන connection ගණන

      ඒ වගේම ඔයට හම්බෙනවා තව option දෙකකුත්

      - Close Conns
      - Refresh

      ◽️ Close Conns
      මේකෙන් ඔයාගෙ config එකේ දැනට තියෙන්න ඔක්කොම tcp connection close කරලා දානවා.

      ◽️ 2 Refresh
      මේකෙන් Usage Info msg එක Update කරහැකි, කරලම බලන්න
    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
help_cmd4:
  en:
    msg_template: |
      <b><u>🟢/buildconf 🟢</u></b>

      🔹 මේකෙන් ඔයාට කැමති විදියට sing-box configuration හදාගන්න පුලුවන්. ඒ වගේම මේක server එකට කිසිම සම්බන්දයක් නැ නිකන් ඔයගෙ local config editor එකක් විතරයි
      builder එක ගැන වැඩිදුර විස්තර බලාගන්න  Builder Help වලට යන්න.

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
help_cmd5:
  en:
    msg_template: |
      <b><u>🟢 /setcap 🟢</u></b>

      මේකෙන් ඔයාගෙ සම්පූර්ණ ලැබෙන quota එක cap කරන්න පුලුවන්. ඒ cap එක remove වෙන්න දාපු දවසෙ ඉදන් 30 days යනවා.
      පුලුවන්නම් මේක දාගන්න ඔයාට ඕන Gb 30 cap එක සෙට් කරන්න 30 GB
      මොකද bandwidth waste වෙනවා නැත්තන්. කාටහරි use කරන්න තියෙන එක ඒ වගේම ඔය මාසෙට අදාල තම්න්ගෙ quota එකෙන් 75% use කරලා නැත්තන්
      ඔයාට් ඊලග මාසෙ service එක use කරන්න වෙන්නෙ නැ.

      <b><u>🟢 /distribute 🟢</u></b>

      මේකෙන් ඔයාගෙ සේරම quota එක group members අතරෙ බෙදෙනවා. ආයෙ ඔයගෙ quota එක හම්බෙන්නෙ usage reset වෙන දවසට.
      ඔයා service එක use කරන්නෙ නැත්තන් අනිවාරෙන් මේක කරන්න නැත්තන් Monthlimited වැදුනොත් මාසයක් use කරන්න බැ මුකුත්.

      <b><u> 🟢/sendgift🟢 </u></b>

      මේකෙන් කාටහරි Quota gift එකක් යවන්න පුලුවන්.
      ඔයාට bandwidth වැඩී වගේනම් ඒක use කරන කාටහරි දෙන්න, බොරුවට තියාගෙන Usage reset වෙද්දි limit වද්දගන්න එපා.
      ඔයා කාටහරි gift එකක් දුන්නොත් ඔයාගෙ info වල ඒක සෘණ විදියට තමා පෙන්නන්නෙ.
      මේ යවන gift quota එක constant එකක් නෙමේ ඒ කියන්නෙ ඔයාගෙ main quota අඩු වෙන ratio වලින්ම මේකෙ ගාණත් අඩුවෙනවා.

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
help_cmd6:
  en:
    msg_template: |
      <b><u>🟢 /contact 🟢</u></b>
      මේකෙන් Admin contact කරන්න පුලුවන් 2 min chat time එකක් හම්බෙනවා admin online නම් reply කරනවා නැත්තන් ඔන්ලයින් ඉන්න වෙලාවක reply mg එකක් දායි.
      chat session එක cancel කරන්න /cancel send කරන්න

      <b><u>🟢 /sugess 🟢</u></b>
      මේකෙන් ඔයාලට මොනවහරි දෙයක් sugess කරන්න තියෙනවනන් ඒක එවන්න පුලුවන්.

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
help_cmd7:
  en:
    msg_template: |
      <b><u>🟢 /points 🟢</u></b>
      මේකෙන් ඔයාට දැනට තියෙන points ගාන බලාගන්න පුලුවන්.

      <b><u>🟢 /refer 🟢</u></b>
      මේකෙන් ඔයාගෙ reffral ගැන විස්තර බලාගන්න වගේම reffral link එක gen කරගන්නත් පුලුවන්
      reflink format := https://t.me/connected_test_bot?start=<you'r telegramid>

      ඒ වගේම reffral වුණ User ගාන අනුව ඔයාලට points redeeme කරගන්න පුලුවන්. reffral එකක් හැදෙන්නෙ bot start කරන user අනිවාරෙන් group එකෙයි channel එකෙයි ඉන්න ඕන් හෝ ජොයින් වෙන්න ඕන.

      <b><u>🟢 /events 🟢</u></b>
      ඔයාල ලග තියෙන Points වලින් events claim කරන්න පුලුවන් දැනට නම් එක event එකයි තියෙන්නෙ. ඉස්ස්‍රහට තව දාන්නම්කො.

    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false

help_builder1:
  en:
    msg_template: |
      Not Available Yet 🫠
    parse_mode: ""
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
help_builder2:
  en:
    msg_template: "Hello {{.Name}} this is help Builder Help page 2"
    parse_mode: ""
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
help_builder3:
  en:
    msg_template: "Hello {{.Name}} this is help Builder Help page 3"
    parse_mode: ""
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
help_builder4:
  en:
    msg_template: "Hello {{.Name}} this is help Builder Help page 4"
    parse_mode: ""
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false
help_builder5:
  en:
    msg_template: "මෙව්ව dev කරන්නත් පුලුවන් මේ හෑලි ලියන එකනෙ එපා වෙන්නෙ"
    parse_mode: ""
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false

help_tutorial1:
  en:
    msg_template: |
      Create Simple Sing Box Config Using Connected BOT
      video credit @Ghost_Broo
    parse_mode: ""
    include_media: false
    media_type: "video"
    media_id: "BAACAgUAAxkBAAEwMwFnaDt0x_qoSyGQM6vlArviFawmzAAC_CEAAtTUQFfVnr64ZKuPXzYE"
    continue_media: false
    disabled: false
    alt_med_path: "./res/tutorial.mp4"
help_tutorial2:
  en:
    msg_template: "මෙව්ව dev කරන්නත් පුලුවන් මේ හෑලි ලියන එකනෙ එපා වෙන්නෙ"
    parse_mode: ""
    include_media: false
    media_type: ""
    media_id: ""
    continue_media: false
    disabled: false

help_about:
  en:
    msg_template: |
      <b>🤖 About This Bot 🤖 </b>

      〽️ This bot is crafted to provide a seamless and efficient experience 😏, leveraging advanced features for managing configurations and usage 🏂.

      <b>🛠 Technology Stack: 🛠 </b>

      🔹 Built Using Go Lang 🐹.
      🔹 Latest Sing-Box Core 📦  with custom enhancements for optimal functionality 🚀.

      <b>👨‍💻 Developed By: Dark Wolf</b>

      - Special thanks to <b>Ghost</b> for Surfshark account and <b>Shay C</b> for the bot logo.

      ⚪️ Intuitive user interface with real-time usage statistics 🛰
      ⚪️ Advanced configuration management  ⚙️
      ⚪️ Optimized  for high performance 🚀
      ⚪️ Highly Customizable 🛠 Config Creation with  Config Builder ⚙️
      ⚪️ Built For Fun & Community 😌

      <b>
      <tg-spoiler>This project was born out of a passion for innovation and a desire to share something unique with the community.</tg-spoiler>

      🌟 Your Feedback Matters
      🗳 We value your input! Feel free to share your suggestions📨 or feedback using /suggest or /contact 💭 </b>
    parse_mode: "HTML"
    include_media: false
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwOj1naT6nb3yLQgEiqVbbLCpKRdqOCQAC_70xG9TUSFcW58STzcK_zQEAAwIAA3kAAzYE"
    continue_media: false
    disabled: false

state_info:
  en:
    msg_template: |
      {{- if .IsPaused }}<b>⏯ Resume Service?</b>
      You’ve currently <b>paused</b> your service.

      🔄 Resuming will:
      • Reactivate your configuration
      • Restart no-usage cycle tracking

      Do you want to resume now?{{else}}
      <b>⏸ Pause Service?</b>
      You’re about to <b>pause</b> your active configuration.

      ⚠️ Pausing will:
      • Deduct <b>1 point</b> from your balance
      • Temporarily disable the service
      • <b>Stop no-usage cycle counting</b> until resumed

      You can resume anytime later.{{end}}

      <b>👤 Your Current Status</b>
      ▫️ 📆 <b>No-Usage Cycle:</b> {{.NonUseCycle}}  
      ▫️ 🌊 <b>Remaining Quota:</b> {{.LeftQuota}}  
      ▫️ 📈 <b>Total Usage:</b> {{.TUsage}}  
      ▫️ ⭐ <b>Points:</b> {{.Points}}

      <b>❓ Would you like to continue?</b>

    parse_mode: "HTML"
    include_media: false
    alt_med_path: "./res/chan_logo.png"
    media_type: "photo"
    media_id: "AgACAgUAAxkBAAEwE6FnYt1K_DrxnaFjifGaUZTs-1WhWgAC6MIxG4p3GVfonvtJDsoLMAEAAwIAA3MAAzYE"
    continue_media: false
    disabled: false
```
