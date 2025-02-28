#Template Variables

## **`grp_welcome`**

| Variable            | Value Type | Info                                 |
| ------------------- | ---------- | ------------------------------------ |
| `{{.Name}}`         | string     | User's name                          |
| `{{.Username}}`     | string     | User's Telegram username             |
| `{{.TgId}}`         | int64      | User's Telegram ID                   |
| `{{.IsInChannel}}`  | bool       | Whether the user is in the channel   |
| `{{.IsBotStarted}}` | bool       | Whether the user has started the bot |
| `{{.GroupLink}}`    | string     | Group link                           |
| `{{.ChanLink}}`     | string     | Channel link                         |

## **`grp_comeback`**

| Variable            | Value Type | Info                                |
| ------------------- | ---------- | ----------------------------------- |
| `{{.Name}}`         | string     | user's Name                         |
| `{{.Username}}`     | string     | user Telegram UserName              |
| `{{.TgId}}`         | int64      | user Telegram ID                    |
| `{{.IsInChannel}}`  | bool       | Whether user is in channel or not   |
| `{{.IsBotStarted}}` | bool       | Whether user has started bot or not |
| `{{.GroupLink}}`    | string     | Group link                          |
| `{{.ChanLink}}`     | string     | Channel link                        |

## **`chan_welcome`**

| Variable         | Value Type | Info                         |
| ---------------- | ---------- | ---------------------------- |
| `{{.Name}}`      | string     | user's Name                  |
| `{{.Username}}`  | string     | user Telegram UserName       |
| `{{.TgId}}`      | int64      | user Telegram ID             |
| `{{.Chat}}`      | string     | Will be `group` or `channel` |
| `{{.GroupLink}}` | string     | Group link                   |
| `{{.ChanLink}}`  | string     | Channel link                 |

## **`chan_comeback`**

| Variable         | Value Type | Info                   |
| ---------------- | ---------- | ---------------------- |
| `{{.Name}}`      | string     | user's Name            |
| `{{.Username}}`  | string     | user Telegram UserName |
| `{{.TgId}}`      | int64      | user Telegram ID       |
| `{{.GroupLink}}` | string     | Group link             |
| `{{.ChanLink}}`  | string     | Channel link           |

## **`dm_welcome`**

| Variable            | Value Type | Info                                    |
| ------------------- | ---------- | --------------------------------------- |
| `{{.Name}}`         | string     | user's Name                             |
| `{{.Username}}`     | string     | user Telegram UserName                  |
| `{{.TgId}}`         | int64      | user Telegram ID                        |
| `{{.IsBotStarted}}` | bool       | Whether user has started the bot or not |
| `{{.IsInGroup}}`    | bool       | Whether user is in the group or not     |
| `{{.GroupLink}}`    | string     | Group link                              |
| `{{.ChanLink}}`     | string     | Channel link                            |
| `{{.Chat}}`         | string     | `channel` or `group`                    |

## **`dm_verified`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`dm_verified_again`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`chat_mem_left`**

| Variable         | Value Type | Info                   |
| ---------------- | ---------- | ---------------------- |
| `{{.Name}}`      | string     | user's Name            |
| `{{.Username}}`  | string     | user Telegram UserName |
| `{{.TgId}}`      | int64      | user Telegram ID       |
| `{{.LeftQuota}}` | string     | User's remaining quota |

## **`start_monthlimited`**

| Variable          | Value Type | Info                   |
| ----------------- | ---------- | ---------------------- |
| `{{.Name}}`       | string     | user's Name            |
| `{{.Username}}`   | string     | user Telegram UserName |
| `{{.TgId}}`       | int64      | user Telegram ID       |
| `{{.LimitendIn}}` | int32      | Days until limit ends  |

## **`start_restricted`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`start_newuser`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`start_newuser_verified`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`start_newuser_unverified`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`start_regular`**

| Variable               | Value Type | Info                                    |
| ---------------------- | ---------- | --------------------------------------- |
| `{{.Name}}`            | string     | user's Name                             |
| `{{.Username}}`        | string     | user Telegram UserName                  |
| `{{.TgId}}`            | int64      | user Telegram ID                        |
| `{{.AddtionalQuota}}`  | string     | Additional quota                        |
| `{{.CalculatedQuota}}` | string     | Calculated quota for user at the moment |
| `{{.MUsage}}`          | string     | User's monthly usage                    |
| `{{.Alltime}}`         | string     | User's all-time usage                   |

## **`start_removed`**

| Variable           | Value Type | Info                              |
| ------------------ | ---------- | --------------------------------- |
| `{{.Name}}`        | string     | user's Name                       |
| `{{.Username}}`    | string     | user Telegram UserName            |
| `{{.TgId}}`        | int64      | user Telegram ID                  |
| `{{.IsInChannel}}` | bool       | Whether user is in channel or not |
| `{{.IsinGroup}}`   | bool       | Whether user is in group or not   |

## **`create_select`**

| Variable            | Value Type | Info                                       |
| ------------------- | ---------- | ------------------------------------------ |
| `{{.Name}}`         | string     | user's Name                                |
| `{{.Username}}`     | string     | user Telegram UserName                     |
| `{{.TgId}}`         | int64      | user Telegram ID                           |
| `{{.CreaterCount}}` | int16      | Available creator count (always 1 for now) |

## **`create_conf_limit`**

| Variable     | Value Type | Info |
| ------------ | ---------- | ---- |
| `{{.Count}}` | int16      |

## **`create_in_info`**

| Variable              | Value Type | Info                          |
| --------------------- | ---------- | ----------------------------- |
| `{{.InName}}`         | string     | Inbound Name (tag name)       |
| `{{.InType}}`         | string     | Inbound Type `vless`          |
| `{{.InPort}}`         | int        | Port                          |
| `{{.InAddr}}`         | string     | Listen address                |
| `{{.InInfo}}`         | string     | Inbound Info                  |
| `{{.Domain}}`         | string     | Domain                        |
| `{{.PublicIp}}`       | string     | Public IP                     |
| `{{.TranstPortType}}` | string     | Transport Type                |
| `{{.TlsEnabled}}`     | bool       | Whether TLS is enabled or not |

## **`create_out_info`**

| Variable       | Value Type | Info             |
| -------------- | ---------- | ---------------- |
| `{{.OutName}}` | string     | Outbound Name    |
| `{{.OutType}}` | string     | Outbound Type    |
| `{{.OutInfo}}` | string     | Outbound Info    |
| `{{.Latency}}` | int32      | Outbound latency |

## **`create_available_quota`**

| Variable | Value Type | Info            |
| -------- | ---------- | --------------- |
| `Quota`  | string     | Available quota |

## **`create_result`**

| Variable          | Value Type | Info                          |
| ----------------- | ---------- | ----------------------------- |
| `Name`            | string     | user's Name                   |
| `Username`        | string     | user Telegram UserName        |
| `TgId`            | int64      | user Telegram ID              |
| `UUID`            | string     | Config UUID                   |
| `Domain`          | string     | Domain                        |
| `Transport`       | string     | Config Transport              |
| `ConfigName`      | string     | Config Name                   |
| `{{.TlsEnabled}}` | bool       | Whether TLS is enabled or not |
| `Port`            | int        | Port                          |
| `Path`            | string     | Path                          |
| `TransportType`   | string     | Transport Type                |

## **`help_home`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`help_info1`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`help_cmd1`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`help_builder1`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`help_tutorial1`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`help_about`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`refer_home`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |
| `{{.Refred}}`   | string     | Referred by            |
| `{{.Verified}}` | string     | Verification status    |

## **`refer_share`**

| Variable        | Value Type | Info                   |
| --------------- | ---------- | ---------------------- |
| `{{.Botlink}}`  | string     | Bot Link               |
| `{{.Name}}`     | string     | user's Name            |
| `{{.Username}}` | string     | user Telegram UserName |
| `{{.TgId}}`     | int64      | user Telegram ID       |

## **`setcap_already`**

| Variable       | Value Type |
| -------------- | ---------- |
| `{{.EndDate}}` | string     |

## **`setcap_warn`** (from v1.2.0)

| Variable         | Value Type | Info                              |
| ---------------- | ---------- | --------------------------------- |
| `{{.LeftQuota}}` | string     | User's remaining quota            |
| `{{.MinCap}}`    | string     | Minimum values that can be capped |
| `{{.CapRange}}`  | string     | Range within which to cap         |

## **`setcap_get`**

| Variable         | Value Type | Info                                     |
| ---------------- | ---------- | ---------------------------------------- |
| `{{.LeftQuota}}` | string     | User's remaining quota v1.2.0            |
| `{{.MinCap}}`    | string     | Minimum values that can be capped v1.2.0 |
| `{{.CapRange}}`  | string     | Range within which to cap v1.2.0         |

## **`setcap_reply`** removed from v1.2.0

| Variable       | Value Type | Info               |
| -------------- | ---------- | ------------------ |
| `{{.EndDate}}` | string     | Cap Quota End Data |

## **`gift_reciver`**

| Variable        | Value Type | Info               |
| --------------- | ---------- | ------------------ |
| `{{.Name}}`     | string     |
| `{{.Username}}` | string     | Common User Info   |
| `{{.TgId}}`     | int64      |
| `{{.Gift}}`     | string     | Recived Gift Quota |
| `{{.FromUser}}` | string     | Sender Name        |

## **`gift_send`**

| Variable         | Value Type |
| ---------------- | ---------- |
| `{{.LeftQuota}}` | string     |

## **`status_home`**

Overall Usage For all config

| Variable             | Value Type | Info                               |
| -------------------- | ---------- | ---------------------------------- |
| `{{.TDownload}}`     | string     | Total download since last refresh  |
| `{{.TUpload}}`       | string     | Total upload since last refresh    |
| `{{.UsageDuration}}` | string     | Time elapsed after last DB refresh |
| `{{.MDownload}}`     | string     | Total Download For Month           |
| `{{.MUpload}}`       | string     | Total Download For Month           |
| `{{.MonthAll}}`      | string     | Total Usage For Month              |
| `{{.Alltime}}`       | string     | All Time Usage                     |

## **`status_callback`**

This is Specific to selectedt config

| Variable             | Value Type           | Info                                         |
| -------------------- | -------------------- | -------------------------------------------- |
| `{{.TDownload}}`     | string               | Total download since last refresh            |
| `{{.TUpload}}`       | string               | Total upload since last refresh              |
| `{{.UsageDuration}}` | string               | Time elapsed after last DB refresh           |
| `{{.MDownload}}`     | string               | Total download for the current month         |
| `{{.MUpload}}`       | string               | Total upload for the current month           |
| `{{.Online}}`        | int                  | Number of online users                       |
| `{{.Ip}}`            | []netip.Addr         | List of IP addresses                         |
| `{{.ConnCount}}`     | []int64              | List of connection counts per IP address     |
| `{{.IpMap}}`         | map[netip.Addr]int64 | Mapping of IP addresses to connection counts |

## **`getinfo_home`**

| Variable      | Value Type |
| ------------- | ---------- |
| `no varibles` | -          |

## **`getinfo_user`**

| Variable               | Value Type | Info                                  |
| ---------------------- | ---------- | ------------------------------------- |
| `{{.Name}}`            | string     | user's Name                           |
| `{{.Username}}`        | string     | user Telegram UserName                |
| `{{.TgId}}`            | int64      | user Telegram ID                      |
| `{{.Dedicated}}`       | string     | Dedicated Quota info                  |
| `{{.TQuota}}`          | string     | Total quota                           |
| `{{.LeftQuota}}`       | string     | Remaining quota                       |
| `{{.ConfCount}}`       | int16      | Configuration count                   |
| `{{.TUsage}}`          | string     | Total usage                           |
| `{{.GiftQuota}}`       | string     | Gifted quota                          |
| `{{.Joined}}`          | string     | Join date                             |
| `{{.CapEndin}}`        | string     | Cap end date                          |
| `{{.Disendin}}`        | int32      | Days until distribution end           |
| `{{.UsageResetIn}}`    | int32      | Days until usage reset                |
| `{{.AlltimeUsage}}`    | string     | All-time usage                        |
| `{{.Iscapped}}`        | bool       | Whether user is capped                |
| `{{.Isgifted}}`        | bool       | Whether user is gifted                |
| `{{.Isdisuser}}`       | bool       | Whether user is disabled              |
| `{{.IsMonthLimited}}`  | bool       | Whether user is limited               |
| `{{.IsTemplimited}}`   | bool       | Whether user is tempory limited       |
| `{{.TempLimitRate}}`   | int16      | Current Temp Limitation Ratio         |
| `{{.JoinedPlace}}`     | uint       | Join place                            |
| `{{.CapDays}}`         | int32      | User Capped Days                      |
| `{{.UsagePercentage}}` | float64    | Usage Percentage for this month       |
| `{{.NonUseCycle}}`     | int32      | Cycle count user didn't use in streak |

## **`getinfo_usage`**

| Variable                | Value Type | Info                                                                     |
| ----------------------- | ---------- | ------------------------------------------------------------------------ |
| `{{.Name}}`             | string     | user's Name                                                              |
| `{{.Username}}`         | string     | user Telegram UserName                                                   |
| `{{.TgId}}`             | int64      | user Telegram ID                                                         |
| `{{.TotalQuota}}`       | string     | User's Total Quota                                                       |
| `{{.ConfigName}}`       | string     | Config Usage                                                             |
| `{{.ConfigType}}`       | string     | Config Type                                                              |
| `{{.ConfigUUID}}`       | string     | Config Type                                                              |
| `{{.ConfigUpload}}`     | string     | Config Upload For this month                                             |
| `{{.ConfigDownload}}`   | string     | Config Download For this month                                           |
| `{{.ConfigUploadtd}}`   | string     | Config Upload For last db refresh(aroud 2hr according to refresh rate)   |
| `{{.ConfigDownloadtd}}` | string     | Config Download For last db refresh(aroud 2hr according to refresh rate) |
| `{{.UsageDuration}}`    | string     | time from last db refresh                                                |
| `{{.ConfigUsage}}`      | string     | Config Full Usage (down + up)                                            |
| `{{.ConfigUsagetd}}`    | string     | Config Full Usage (down + up) For last db refresh                        |
| `{{.UsedPresenTage}}`   | float64    | Usage percentage (from v1.2.0)                                           |
| `{{.ResetDays}}`        | int32      | Days to Renew Config (reset usages)                                      |
| `{{.PublicIp}}`         | string     | Selected Config IP                                                       |
| `{{.PublicDomain}}`     | string     | Selected Config Domain                                                   |
| `{{.TransPortPath}}`    | string     | Selected Config's Transport Path                                         |
| `{{.InName}}`           | string     | Selected Inbound Name                                                    |
| `{{.InType}}`           | string     | Inbound Type                                                             |
| `{{.InPort}}`           | int        | Port                                                                     |
| `{{.InAddr}}`           | string     | Inbound Add                                                              |
| `{{.InInfo}}`           | string     | Inbound Info                                                             |
| `{{.TranstPortType}}`   | string     | Transport Type                                                           |
| `{{.Loginlimit}}`       | int16      | Login Limit                                                              |
| `{{.TlsEnabled}}`       | bool       | Whether Tls Enabled Or Not                                               |
| `{{.SupportInfo}}`      | []string   | Inbound Support Info                                                     |
| `{{.OutName}}`          | string     | Outbound Name                                                            |
| `{{.OutType}}`          | string     | Outbound Type                                                            |
| `{{.OutInfo}}`          | string     | Outbound Info                                                            |
| `{{.Latency}}`          | int32      | Outbound Latancy                                                         |
| `{{.Online}}`           | int        | Realtime Connected Client(Ip)                                            |
| `{{.IpMap}}`            | map        | Realtime Connected Ip and Connection Count                               |

## **`getinfo_out`**

| Variable       | Value Type | Info             |
| -------------- | ---------- | ---------------- |
| `{{.OutName}}` | string     | Outbound Name    |
| `{{.Info}}`    | string     | Outbound INfo    |
| `{{.Latency}}` | int32      | Outbound Latancy |
| `{{.Type}}`    | string     | Outbound Type    |

## **`getinfo_in`**

| Variable              | Value Type | Info                       |
| --------------------- | ---------- | -------------------------- |
| `{{.InName}}`         | string     | Inbound Name (tag name)    |
| `{{.InType}}`         | string     | Inbound Type `vless`       |
| `{{.InPort}}`         | int        | Port                       |
| `{{.InAddr}}`         | string     | Listen Address             |
| `{{.InInfo}}`         | string     | Inbound Info               |
| `{{.Domain}}`         | string     | Domain                     |
| `{{.PublicIp}}`       | string     | Public IP                  |
| `{{.TranstPortType}}` | string     | Transport Type             |
| `{{.TlsEnabled}}`     | bool       | Whether Tls Enabled or not |
| `{{.Support}}`        | []string   | Support Info array         |

## **`configure_home`**

| Variable         | Value Type | Info                   |
| ---------------- | ---------- | ---------------------- |
| `{{.Name}}`      | string     |
| `{{.Username}}`  | string     | Common User Info       |
| `{{.TgId}}`      | int64      |
| `{{.ConfCount}}` | int16      | Available Config Count |

## **`conf_configure`**

| Variable        | Value Type |
| --------------- | ---------- |
| `{{.ConfName}}` | string     |

## **`conf_name_change`**

| Variable        | Value Type | Info                         |
| --------------- | ---------- | ---------------------------- |
| `{{.Name}}`     | string     |
| `{{.Username}}` | string     | Common User Info             |
| `{{.TgId}}`     | int64      |
| `{{.NewName}}`  | string     | NewRecived Name (user Input) |

## **`conf_quota_change`**

| Variable         | Value Type |
| ---------------- | ---------- |
| `{{.AvblQuota}}` | string     |
| `{{.ConfName}}`  | string     |

## **`conf_in_change`**

| Variable              | Value Type | Info                          |
| --------------------- | ---------- | ----------------------------- |
| `{{.InName}}`         | string     | Inbound Name (tag name)       |
| `{{.InType}}`         | string     | Inbound Type `vless`          |
| `{{.InPort}}`         | int        | Port                          |
| `{{.InAddr}}`         | string     | Listen Address                |
| `{{.InInfo}}`         | string     | Inbound Info                  |
| `{{.Domain}}`         | string     | Domain                        |
| `{{.PublicIp}}`       | string     | Public IP                     |
| `{{.TranstPortType}}` | string     | Transport Type                |
| `{{.TlsEnabled}}`     | bool       | Whether TLS is enabled or not |
| `{{.Support}}`        | []string   | Support Info array            |

## **`conf_out_change`**

| Variable       | Value Type | Info             |
| -------------- | ---------- | ---------------- |
| `{{.OutName}}` | string     | Outbound Name    |
| `{{.OutType}}` | string     | Outbound Type    |
| `{{.OutInfo}}` | string     | Outbound Info    |
| `{{.Latency}}` | int32      | Outbound Latency |

## **`event_home`**

| Variable         | Value Type | Info             |
| ---------------- | ---------- | ---------------- |
| `{{.Name}}`      | string     |
| `{{.Username}}`  | string     | Common User Info |
| `{{.TgId}}`      | int64      |
| `{{.AvblCount}}` | int16      |
| `{{.Completed}}` | int16      |

## **`points_home`**

| Variable        | Value Type | Info             |
| --------------- | ---------- | ---------------- |
| `{{.Count}}`    | int64      | Points count     |
| `{{.Name}}`     | string     |
| `{{.Username}}` | string     | Common User Info |
| `{{.TgId}}`     | int64      |

## **`distribute_group`**

| Variable        | Value Type | Info             |
| --------------- | ---------- | ---------------- |
| `{{.Name}}`     | string     |
| `{{.Username}}` | string     | Common User Info |
| `{{.TgId}}`     | int64      |
| `{{.Disquota}}` | string     |

## **`builder_home`**

| Variable         | Value Type | Info             |
| ---------------- | ---------- | ---------------- |
| `{{.Name}}`      | string     |
| `{{.Username}}`  | string     | Common User Info |
| `{{.TgId}}`      | int64      |
| `{{.ConfCount}}` | int        |

## **`com_unverified`**

| Variable        | Value Type | Info             |
| --------------- | ---------- | ---------------- |
| `{{.Name}}`     | string     |                  |
| `{{.Username}}` | string     | Common User Info |
| `{{.TgId}}`     | int64      |

## **`restricted`**

| Variable        | Value Type | Info             |
| --------------- | ---------- | ---------------- |
| `{{.Name}}`     | string     |
| `{{.Username}}` | string     | Common User Info |
| `{{.TgId}}`     | int64      |

## **`overview`** ( removed from v1.2.0)

This is for admins

| Variable                  | Value Type |
| ------------------------- | ---------- |
| `{{.BandwidthAvailable}}` | string     |
| `{{.MonthTotal}}`         | string     |
| `{{.AllTime}}`            | string     |
| `{{.VerifiedUserCount}}`  | int64      |
| `{{.TotalUser}}`          | int32      |
| `{{.CappedUser}}`         | int64      |
| `{{.DistributedUser}}`    | int64      |
| `{{.Restricte}}`          | int64      |
| `{{.QuotaForEach}}`       | string     |
| `{{.LastRefresh}}`        | time.Time  |
