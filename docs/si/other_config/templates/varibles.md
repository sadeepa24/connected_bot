# Template Variables

## **`grp_welcome`**

| Variable                               | Value Type | Info                                 |
| -------------------------------------- | ---------- | ------------------------------------ |
| [common user](./common.md#common-user) |            |                                      |
| `{{.IsInChannel}}`                     | bool       | Whether the user is in the channel   |
| `{{.IsBotStarted}}`                    | bool       | Whether the user has started the bot |
| `{{.GroupLink}}`                       | string     | Group link                           |
| `{{.ChanLink}}`                        | string     | Channel link                         |

## **`grp_comeback`**

| Variable                               | Value Type | Info                                |
| -------------------------------------- | ---------- | ----------------------------------- |
| [common user](./common.md#common-user) |            |                                     |
| `{{.IsInChannel}}`                     | bool       | Whether user is in channel or not   |
| `{{.IsBotStarted}}`                    | bool       | Whether user has started bot or not |
| `{{.GroupLink}}`                       | string     | Group link                          |
| `{{.ChanLink}}`                        | string     | Channel link                        |

## **`chan_welcome`**

| Variable                               | Value Type | Info                         |
| -------------------------------------- | ---------- | ---------------------------- |
| [common user](./common.md#common-user) |            |                              |
| `{{.Chat}}`                            | string     | Will be `group` or `channel` |
| `{{.GroupLink}}`                       | string     | Group link                   |
| `{{.ChanLink}}`                        | string     | Channel link                 |

## **`chan_comeback`**

| Variable                               | Value Type | Info         |
| -------------------------------------- | ---------- | ------------ |
| [common user](./common.md#common-user) |            |              |
| `{{.GroupLink}}`                       | string     | Group link   |
| `{{.ChanLink}}`                        | string     | Channel link |

## **`dm_welcome`**

| Variable                               | Value Type | Info                                    |
| -------------------------------------- | ---------- | --------------------------------------- |
| [common user](./common.md#common-user) |            |                                         |
| `{{.IsBotStarted}}`                    | bool       | Whether user has started the bot or not |
| `{{.IsInGroup}}`                       | bool       | Whether user is in the group or not     |
| `{{.GroupLink}}`                       | string     | Group link                              |
| `{{.ChanLink}}`                        | string     | Channel link                            |
| `{{.Chat}}`                            | string     | `channel` or `group`                    |

## **`dm_verified`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`dm_verified_again`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`chat_mem_left`**

| Variable                               | Value Type | Info                   |
| -------------------------------------- | ---------- | ---------------------- |
| [common user](./common.md#common-user) |            |                        |
| `{{.LeftQuota}}`                       | string     | User's remaining quota |

## **`start_monthlimited`**

| Variable                               | Value Type | Info                  |
| -------------------------------------- | ---------- | --------------------- |
| [common user](./common.md#common-user) |            |                       |
| `{{.LimitendIn}}`                      | int32      | Days until limit ends |

## **`start_restricted`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`start_newuser`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`start_newuser_verified`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`start_newuser_unverified`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`start_regular`**

| Variable                               | Value Type | Info                                    |
| -------------------------------------- | ---------- | --------------------------------------- |
| [common user](./common.md#common-user) |            |                                         |
| `{{.AddtionalQuota}}`                  | string     | Additional quota                        |
| `{{.CalculatedQuota}}`                 | string     | Calculated quota for user at the moment |
| `{{.MUsage}}`                          | string     | User's monthly usage                    |
| `{{.Alltime}}`                         | string     | User's all-time usage                   |

## **`start_removed`**

| Variable                               | Value Type | Info                              |
| -------------------------------------- | ---------- | --------------------------------- |
| [common user](./common.md#common-user) |            |                                   |
| `{{.IsInChannel}}`                     | bool       | Whether user is in channel or not |
| `{{.IsinGroup}}`                       | bool       | Whether user is in group or not   |

## **`inbound_info`**

| Variable                                     | Value Type | Info |
| -------------------------------------------- | ---------- | ---- |
| [common inbound](./common.md#common-inbound) |            |      |

## **`create_out_info`**

| Variable       | Value Type | Info             |
| -------------- | ---------- | ---------------- |
| `{{.OutName}}` | string     | Outbound Name    |
| `{{.OutType}}` | string     | Outbound Type    |
| `{{.OutInfo}}` | string     | Outbound Info    |
| `{{.Latency}}` | int32      | Outbound latency |

## **`create_available_quota`**

| Variable     | Value Type | Info            |
| ------------ | ---------- | --------------- |
| `{{.Quota}}` | string     | Available quota |

## **`create_result`**

| Variable          | Value Type | Info                                             |
| ----------------- | ---------- | ------------------------------------------------ |
| `{{.Name}}`       | string     | Config Name                                      |
| `{{.UUID}}`       | string     | Config UUID (for inbound which use uuid)         |
| `{{.Password}}`   | string     | Config Password (for inbound which use password) |
| `{{.Quota}}`      | string     | Config Quota                                     |
| `{{.LoginLimit}}` | int        | LoginLimit                                       |

## **`help_home`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`help_info1`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`help_cmd1`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`help_builder1`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`help_tutorial1`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`help_about`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`refer_home`**

| Variable                               | Value Type | Info                |
| -------------------------------------- | ---------- | ------------------- |
| [common user](./common.md#common-user) |            |                     |
| `{{.Refred}}`                          | string     | Referred by         |
| `{{.Verified}}`                        | string     | Verification status |

## **`refer_share`**

| Variable                               | Value Type | Info     |
| -------------------------------------- | ---------- | -------- |
| `{{.Botlink}}`                         | string     | Bot Link |
| [common user](./common.md#common-user) |            |          |

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

| Variable                               | Value Type | Info               |
| -------------------------------------- | ---------- | ------------------ |
| [common user](./common.md#common-user) |            |                    |
| `{{.Gift}}`                            | string     | Recived Gift Quota |
| `{{.FromUser}}`                        | string     | Sender Name        |

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

| Variable                               | Value Type | Info                                  |
| -------------------------------------- | ---------- | ------------------------------------- |
| [common user](./common.md#common-user) |            |                                       |
| `{{.Dedicated}}`                       | string     | Dedicated Quota info                  |
| `{{.TQuota}}`                          | string     | Total quota                           |
| `{{.Paused}}`                          | bool       | Service Paused or Not                 |
| `{{.LeftQuota}}`                       | string     | Remaining quota                       |
| `{{.ConfCount}}`                       | int16      | Configuration count                   |
| `{{.TUsage}}`                          | string     | Total usage                           |
| `{{.GiftQuota}}`                       | string     | Gifted quota                          |
| `{{.Joined}}`                          | string     | Join date                             |
| `{{.CapEndin}}`                        | string     | Cap end date                          |
| `{{.Disendin}}`                        | int32      | Days until distribution end           |
| `{{.UsageResetIn}}`                    | int32      | Days until usage reset                |
| `{{.AlltimeUsage}}`                    | string     | All-time usage                        |
| `{{.Iscapped}}`                        | bool       | Whether user is capped                |
| `{{.Isgifted}}`                        | bool       | Whether user is gifted                |
| `{{.Isdisuser}}`                       | bool       | Whether user is disabled              |
| `{{.IsMonthLimited}}`                  | bool       | Whether user is limited               |
| `{{.IsTemplimited}}`                   | bool       | Whether user is tempory limited       |
| `{{.TempLimitRate}}`                   | int16      | Current Temp Limitation Ratio         |
| `{{.JoinedPlace}}`                     | uint       | Join place                            |
| `{{.CapDays}}`                         | int32      | User Capped Days                      |
| `{{.UsagePercentage}}`                 | float64    | Usage Percentage for this month       |
| `{{.NonUseCycle}}`                     | int32      | Cycle count user didn't use in streak |

## **`getinfo_config`**

| Variable                               | Value Type       | Info                                                           |
| -------------------------------------- | ---------------- | -------------------------------------------------------------- |
| [common user](./common.md#common-user) |                  |                                                                |
| `{{.TotalQuota}}`                      | string           | User's Total Quota                                             |
| `{{.Active}}`                          | bool             | Whether config active or not                                   |
| `{{.ConfigName}}`                      | string           | Config Usage                                                   |
| `{{.ConfigUUID}}`                      | string           | Config UUID                                                    |
| `{{.ConfigPassword}}`                  | string           | Config Password                                                |
| `{{.ConfigUpload}}`                    | string           | Config Upload For this month                                   |
| `{{.ConfigDownload}}`                  | string           | Config Download For this month                                 |
| `{{.UsagePercentage}}`                 | string           | Config Usage As Percentage                                     |
| `{{.ConfigUploadtd}}`                  | string           | Config Upload For last db refresh(according to refresh rate)   |
| `{{.ConfigDownloadtd}}`                | string           | Config Download For last db refresh(according to refresh rate) |
| `{{.UsageDuration}}`                   | string           | elpsed time since last db refresh                              |
| `{{.ConfigUsage}}`                     | string           | Config Full Usage (down + up)                                  |
| `{{.ConfigUsagetd}}`                   | string           | Config Full Usage (down + up) since last db refresh            |
| `{{.UsedPresenTage}}`                  | float64          | Usage percentage as Float (from v1.2.0)                        |
| `{{.ResetDays}}`                       | int32            | Days to Renew Config (reset usages)                            |
| `{{.CrDate}}`                          | string           | Config Created Date                                            |
| `{{.Loginlimit}}`                      | int16            | Login Limit                                                    |
| `{{.Online}}`                          | int              | Realtime Connected Client(Ip)                                  |
| `{{.IpMap}}`                           | map[string]int16 | Realtime Connected Ip and Connection Count                     |
| [common out](./common.md#common-out)   |                  |                                                                |

## **`getinfo_out`**

| Variable                             | Value Type | Info |
| ------------------------------------ | ---------- | ---- |
| [common out](./common.md#common-out) |            |      |

## **`getinfo_allin`**

| Variable            | Value Type                                                 | Info                         |
| ------------------- | ---------------------------------------------------------- | ---------------------------- |
| `{{.InboundCount}}` | int                                                        | Inbound Count for the config |
| `{{.AllIn}}`        | slice of [common In Export](./common.md#common-confexport) | Slice of ExportInfo          |

## **getinfo_confin**

| Variable                                          | Value Type | Info |
| ------------------------------------------------- | ---------- | ---- |
| [common In Export](./common.md#common-confexport) |            |      |

## **inbound_info**

| Variable                                     | Value Type | Info |
| -------------------------------------------- | ---------- | ---- |
| [common Inbound](./common.md#common-inbound) |            |      |

## **`configure_home`**

| Variable                               | Value Type | Info                   |
| -------------------------------------- | ---------- | ---------------------- |
| [common user](./common.md#common-user) |            |                        |
| `{{.ConfCount}}`                       | int16      | Available Config Count |

## **`conf_configure`**

| Variable          | Value Type |
| ----------------- | ---------- | ---------------------------- |
| `{{.TotalQuota}}` | string     | User's Total Quota           |
| `{{.Active}}`     | bool       | Whether config active or not |
| `{{.ConfigName}}` | string     | Config Usage                 |
| `{{.Loginlimit}}` | int16      | Login Limit                  |
| `{{.CrDate}}`     | string     | Config Created Date          |

## **`conf_name_change`**

| Variable                               | Value Type | Info                         |
| -------------------------------------- | ---------- | ---------------------------- |
| [common user](./common.md#common-user) |            |                              |
| `{{.NewName}}`                         | string     | NewRecived Name (user Input) |

## **`conf_quota_change`**

| Variable         | Value Type |
| ---------------- | ---------- |
| `{{.AvblQuota}}` | string     |
| `{{.ConfName}}`  | string     |

## **`conf_in_change`**

| Variable                                     | Value Type | Info                           |
| -------------------------------------------- | ---------- | ------------------------------ |
| [common Inbound](./common.md#common-inbound) |            |                                |
| `{{.ActionRemove}}`                          | bool       | When Remove this value is true |

## **`conf_out_change`**

| Variable                             | Value Type | Info |
| ------------------------------------ | ---------- | ---- |
| [common out](./common.md#common-out) |            |      |

## **`event_home`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |
| `{{.AvblCount}}`                       | int16      |
| `{{.Completed}}`                       | int16      |

## **`points_home`**

| Variable                               | Value Type | Info         |
| -------------------------------------- | ---------- | ------------ |
| `{{.Count}}`                           | int64      | Points count |
| [common user](./common.md#common-user) |            |              |

## **`distribute_group`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |
| `{{.Disquota}}`                        | string     |

## **`builder_home`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |
| `{{.ConfCount}}`                       | int        |

## **`com_unverified`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |

## **`restricted`**

| Variable                               | Value Type | Info |
| -------------------------------------- | ---------- | ---- |
| [common user](./common.md#common-user) |            |      |
