#Message Templates

```yaml
grp_welcome:
  en:
    msg_template: |
      ✨ Welcome to 𝘾𝙊𝙉𝙉𝙀𝘾𝙏𝙀𝘿 Chat!
    parse_mode: "HTML"
    include_media: false
    media_type: ""
    media_id: ""
    alt_med_url: ""
    alt_med_path: ""

    disabled: false

    ##optional
    btnconf:
      btns: [["btn1[https://t.me]"]]

    ##පහල තියෙන field 4 ම තනි message එකක් යවද්දි නෙමේ messaging session එකක් වගේ ඉද්දි ඕන වෙන්නෙ
    continue_media: false
    supercontinue: false
    media_skip: false
    contin_skip_text: false
```

මේ උඩ තියෙන්නෙ එක Message එකකට අදාලව template එක. උඩ තියෙන් එකනම් send වෙන්නෙ group එකට User කෙනෙක් join වෙද්දි, example එකක් විදියට ඕක සලකමු.

**`msg_template`**

මේකෙ තමා අදාල Message එකේ content එක තියෙන්නෙ මේකට varibles හම්බෙනවා ඒ ඒ templates එකට අනුව හම්බෙන vaibles මෙතනින් බලාගන්න.
varibles example එකක් ගත්තොත්

- template - `Hello {{.Name}}`
- render message(send to user) - `Hello kamal` (kamal කියන්නෙ bot එවලෙ use කරන User කියලා හිතමු)

[ඒ ඒ template එකට අදාල varibles මෙතනින් බලන්න](./varibles.md)

**`parse_mode`**

Telegram වලින් parse mode තුනක් දෙනවා (HTML, Markdownv1, Markdownv2) මේ තුනෙන් කැමති එකක් use කරන්න පුලුවන්. message template එක ඒ parse mode එකට අදාලව හදලා තියෙන්න ඕනෙ. parse mode එකක් නැතුව නෝර්මල් text නම් මේක හිස් තියන්න

- `HTML`

**`include_media`**

Photo එකක් හරි Video එකක් හරි message එකත් එක්ක යවනවනම් අනිවාරෙන් මේක true වෙන්න ඕනෙ. මේක true නම් විතරයි පහල හතරම අවශ්‍ය වෙන්නෙ.

    - media_type
    - media_id
    - alt_med_url
    - alt_med_path

**`media_type`**

මේක අනිවාරෙන් `photo` or `video` වෙන media type support නැහැ

**`media_id`**

Telegram වලට ඒ ඒ file එකට අදාලව uniq id එකක් තියේ මේක rawdata bot ට Photo එකක් send කරලා ගන්න පුලුවන් නැත්තන් මේක හිස්ව තියලා alt_med_path හරි alt_med_url හරි දුන්නොත් bot start වෙන වෙලාවෙ auto ම දාගන්නවා.

**`alt_med_url`**

Send වෙන්න ඕනෙ photo එකෙ download link එකක්

**`alt_med_path`**

Send වෙන්න ඕනෙ photo eka local තියෙනවනන් ඒකෙ path එක

## මේ පහලින් තියෙන field ඔක්කොම අදාල වෙන්නෙ Message session වලට.

Message session එකක් කියන්නෙ Message කිහිපයක් එකිනෙකට connect වෙන අවස්තා උදා cmd:= /help, /getinfo, /configure, /status, /buildconf

ඒ වගේම /start messages, welcome messages වගේ එව්ව තනි message එකක් විදියට සලකන්න පුලුවන්.

**`continue_media`**

මේකෙන් කියන්නෙ දැන් තියෙන message template එකේ media(photo/video) එක දිගටම continue කරන්න කියලා. ඒ කියන්නෙ ඊලගට එන template වල media එකක් තිබ්බෙ නැතත් මේ media එක දාන්න කියලා (session ඉවර වෙනකම්/ close button press කරනකන් Or timeout).

උදාහරණ අවස්ථා - හිතන්න user /help cmd එක දුන්නාම යන්න ඕන message template එකට photo එකක් media විදියට දීලා සහ `continue_media` true කරලා තිබ්බොත්, ඊලගට User about button එක එබුවොත් ඊලගට about එකට අදාල message template එකේ media එකක් නැතත් අර message එකේ Photo එක add වෙලා තමා send වෙන්නෙ.

මේක දිගටම මෙහෙම යනවා ආයෙ media තියෙන message template එකක් හම්බෙනකම්. මෙහෙම Media තියෙන Message එකක් හම්බුනාම ඒ Message එකේ media එක තමා ඒ template එකෙන් send වෙන්නෙ ඒ template එකේ `continue_media` false වෙලා තිබ්බොත් එතනින් එහාට media එක සෙන්ඩ් වෙන්නෙ නැ. ඒ වගේම true වෙලා තිබ්බොත් එතනින් එහාට send වෙන්නෙ ඒ template එකෙ media එක

**`supercontinue`**

මේකත වැඩ කරන්නෙ `continue_media` වගේම තමා පොඩි වෙනසයි තියෙන්නෙ `continue_media` එකට ආයෙ template එකක් හම්බුනොත් සහ ඒකෙ `continue_media` false වෙලා thibboth එතනින් එහාට media (media නැති post වලට) send වෙන්නෙ නැ. හැබැයි මේක දුන්නාම එහෙම වෙන්නෙ නැ. එතනින් එහාටත් Media සෙන්ඩ් වෙනවා. send වෙන media එක වෙන්නෙ supercontinue true කරලා තිබ්බ template එකේ media එක

**`media_skip`**

`supercontinue` දීලා තියෙද්දි message template එකක් media නැතුව send කරගන්න ඕනෙනම් මේක true දෙන්න.

**`contin_skip_text`**

මේක `continue_media` හෝ `supercontinue` true දීලා තියෙන template එකක දීලා තිබ්බොත් එතනින් එහාට buttons නැතුව එන normal text messages වලට media add කරන්නෙ නැ.

**`btnconf`**

මේ field එක හැම template එකටම අදාල නැහැ. විශේශ කිහිපයකට විතරයි. මේ field එක Use කරලා custom url Buttons add කරන්න පුලුවන්.
මේ field එක support එක කරන template
