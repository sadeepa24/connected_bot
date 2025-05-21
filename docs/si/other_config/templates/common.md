# Common Blocks

## **`common user`**

| Variable        | Value Type | Info                     |
| --------------- | ---------- | ------------------------ |
| `{{.Name}}`     | string     | User's name              |
| `{{.Username}}` | string     | User's Telegram username |
| `{{.TgId}}`     | int64      | User's Telegram ID       |

## **`common-out`**

| Variable       | Value Type | Info             |
| -------------- | ---------- | ---------------- |
| `{{.OutName}}` | string     | Outbound Name    |
| `{{.OutType}}` | string     | Outbound Type    |
| `{{.OutInfo}}` | string     | Outbound Info    |
| `{{.Latency}}` | int32      | Outbound Latancy |

## **`common-inbound`**

| Variable             | Value Type | Info                             |
| -------------------- | ---------- | -------------------------------- |
| `{{.Id}}`            | int16      | ID from JSON file                |
| `{{.Name}}`          | string     | Inbound name                     |
| `{{.Tag}}`           | string     | Inbound tag                      |
| `{{.Type}}`          | string     | Inbound type                     |
| `{{.Support}}`       | []string   | Supported protocols              |
| `{{.ListenAddres}}`  | string     | Address to listen on             |
| `{{.Listenport}}`    | int        | Port to listen on                |
| `{{.Tlsenabled}}`    | bool       | Whether TLS is enabled           |
| `{{.TransPortType}}` | string     | Transport type                   |
| `{{.TransPortPath}}` | string     | Transport path                   |
| `{{.Custom_info}}`   | string     | Custom information               |
| `{{.Domain}}`        | string     | Domain associated with inbound   |
| `{{.PublicIp}}`      | string     | Public IP address of the inbound |

## **`common-exportin`**

| Variable        | Value Type | Info                    |
| --------------- | ---------- | ----------------------- |
| `{{.ProtoUrl}}` | string     | Export Url Link         |
| `{{.Sni}}`      | string     | User Sni                |
| `{{.Host}}`     | string     | User Host               |
| `{{.Server}}`   | []string   | Selected or User server |

## **`common-confexport`**

| Variable                                       | Value Type | Info |
| ---------------------------------------------- | ---------- | ---- |
| [common in](./common.md#common-inbound)        |            |      |
| [common exportin](./common.md#common-exportin) |            |      |
