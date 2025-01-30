package botapi

import (
	"github.com/sadeepa24/connected_bot/constbot"
)

var Testtemplts map[string]map[string]MgItem = map[string]map[string]MgItem{
	"welcome": {"sin": {
		Msgtmpl:   "hello",
		ParseMode: "HTML",
	}},

	"configusage": {"sin": {
		Msgtmpl:   "hello",
		ParseMode: "HTML",
	}},

	constbot.TmpConfigInfo: {
		"sin": {ParseMode: "HTML", Msgtmpl: "සින්හලැන්"},

		"en": {
			ParseMode: "HTML",
			Msgtmpl: `
				<b>{{.Name}} </b> You have tatal {{.TotalQuota}}

				Config Info
				Config Name := {{.ConfigName}}
				Config Type := {{.ConfigType}}
				Config UUID := {{.ConfigUUID}}

				ConfigUpload := {{.ConfigUpload }}
				ConfigDownload := {{.ConfigDownload}}
				ConfigUploadtd := {{.ConfigUploadtd}}
				ConfigDownloadtd := {{.ConfigDownloadtd}}
				ConfigUsage := {{.ConfigUsage}}
				ConfigUsagetd := {{.ConfigUsagetd}}

				InName := {{.InName}}
				InType := {{.InType}}
				InPort := {{.InPort}}
				InAddr := {{.InAddr}}
				InInfo := {{.InInfo}}
				TranstPortType  := {{.TranstPortType}}
				TlsEnabled := {{.TlsEnabled}}

				TotalQuota := {{.TotalQuota}}


				
				OutName = {{.OutName}}
				OutType = {{.OutType}}
				OutInfo = {{.OutInfo}}

				Days to Reset Usage = {{.ResetDays}}

				


	`,
		},
	},

	constbot.TmpConQuota: {
		"en": {
			Msgtmpl:   "Send New quota for config {{.AvblQuota}}",
			ParseMode: "HTML",
		},
	},

	constbot.TmpCrAvblQuota: {
		"en": {
			Msgtmpl:   "You have available quota {{.Quota}}",
			ParseMode: "HTML",
		},
	},
	constbot.TmpCrAlreadyHave: {
		"en": {
			Msgtmpl:   "You already have {{.Count}}",
			ParseMode: "HTML",
		},
	},
	constbot.TmpStTotal: {
		"en": {
			Msgtmpl: `Your usage 

			✅
			
			tddwn {{.TDownload}}  
			tdup {{.TUpload}}  
			totalUp {{.MUpload}}  
			totaldown {{.MDownload}}  
			Alltime {{.Alltime}}
			Month {{.MonthAll}}
			
			If you want config-specific usage, click the buttons below.`,
			ParseMode: "HTML",
		},
	},
	constbot.TmpStcallback: {
		"en": {
			Msgtmpl: "Your usage tddwn {{.TDownload}}  tdup {{.TUpload}}  totalUp {{.MUpload}}  totaldown {{.MDownload}}",
		},
	},

	constbot.TmpcapQuota: {
		"en": {
			Msgtmpl:   "you'r cap will end {{.EndDate}} ",
			ParseMode: "HTML",
		},
	},

	constbot.TmpcapWarn: {
		"en": {
			Msgtmpl: `

			you have total {{.Leftquota}} which can be capped this cap automatically reset after 30 days and you left quota after capping  will shared amoung all user
			also you can't undo this cap you have to wait, if the user quota goes down than your cap also your connection will offline 
			
			`,
		},
	},

	constbot.Tmpcapreply: {
		"en": {
			Msgtmpl: `
			
			Good now send me quota you need to cap {{.LeftQuota}}
			
			`,
		},
	},

	constbot.TmpGifSend: {
		"en": {
			Msgtmpl: "send how much do you need to send you have {{.LeftQuota}}",
		},
	},

	constbot.TmpChatmemLeft: {
		"en": {
			Msgtmpl: "Good Bye If you changed your mind feel free to rejoin {{.Name}} ",
		},
	},

	constbot.TmpGroupWelcome: {
		"en": {
			Msgtmpl: "Hello How are welcome to group {{.Name}}",
		},
	},

	constbot.TmpWelcomeInbox: {
		"en": {
			Msgtmpl: "Welcome Message Template user who did not started bot yet joined ",
		},
	},
	constbot.TmpGrpComeback: {
		"en": {
			Msgtmpl: "Hello welcome back to our group You are still not in our channel if you want to use again service please rejoin channel  ",
		},
	},

	constbot.TmpHelpHome: {
		"en": {
			Msgtmpl: "Hello {{.Name}} Help page home",
		},
	},

	constbot.TmpAbout: {
		"en": {
			Msgtmpl: "Hello {{.Name}} About page",
		},
	},

	constbot.TmpregularVerified: {
		"en": {
			Includemed: true,
			Mediatype:  "photo",
			MediaId:    "AgACAgUAAxkBAAEvY5dnN5SVlrEmRFUSDBm1PWmfJUBrtwACF8IxG-qvuVVy0oe6c5lPTAEAAwIAA20AAzYE",
			Msgtmpl:    "Hello {{.Name}} Verified user started all time - {{.Alltime}}, month usage  {{.MUsage}} ",
		},
	},

	constbot.TmpConfiConfigure: {
		"en": {
			Includemed:  true,
			Mediatype:   "photo",
			MediaId:     "AgACAgUAAxkBAAEvY5dnN5SVlrEmRFUSDBm1PWmfJUBrtwACF8IxG-qvuVVy0oe6c5lPTAEAAwIAA20AAzYE",
			Msgtmpl:     "Hello Verified user started all time - , month usage   ",
			ContinueMed: false,
		},
	},

	constbot.TmpHelpInfoPage + "1": {
		"en": {
			Msgtmpl: "Hello {{.Name}} hello this is info page 1 ",
		},
	},

	constbot.TmpHelpInfoPage + "2": {

		"en": {
			Includemed:  true,
			Mediatype:   "photo",
			MediaId:     "AgACAgUAAxkBAAEvY5dnN5SVlrEmRFUSDBm1PWmfJUBrtwACF8IxG-qvuVVy0oe6c5lPTAEAAwIAA20AAzYE",
			Msgtmpl:     "Hello {{.Name}} this is help info page  2",
			ContinueMed: false,
		},
	},

	constbot.TmpHelpInfoPage + "3": {
		"en": {
			Includemed:  true,
			Mediatype:   "photo",
			MediaId:     "AgACAgUAAxkBAAEvZapnOFjKbTYqbIAgTIy0xf-VZQQ0PwACUcAxG8dwyVV7OMp6tw3cUwEAAwIAA3MAAzYE",
			Msgtmpl:     "Hello {{.Name}} this is help info page  3",
			ContinueMed: true,
		},
	},
	constbot.TmpHelpInfoPage + "4": {
		"en": {
			Msgtmpl: "Hello {{.Name}} this is help info page 4 ",
		},
	},

	constbot.TmpHelpCmPage + "1": {
		"en": {
			Msgtmpl: "Hello {{.Name}} hello this is command page 1 ",
		},
	},

	constbot.TmpHelpCmPage + "2": {
		"en": {
			Msgtmpl: "Hello {{.Name}} this is help command page 2 ",
		},
	},

	constbot.TmpHelpCmPage + "3": {
		"en": {
			Msgtmpl: "Hello {{.Name}} this is help command page 3 ",
		},
	},

	constbot.TmpRefHome: {
		"en": {
			Msgtmpl: "Hello {{.Name}} You have {{.Refred}} users, and {{.Verified}} verified user you should have verified user to claim the gift, ( verified user = use who is in both group and channel)  ",
		},
	},
	constbot.TmpRefshare: {
		"en": {
			Msgtmpl: "Hello Everyone I am {{.Name}} You can Join with {{.Botlink}}  ",

			Includemed: true,
			Mediatype:  "photo",
			MediaId:    "AgACAgUAAxkBAAEvZapnOFjKbTYqbIAgTIy0xf-VZQQ0PwACUcAxG8dwyVV7OMp6tw3cUwEAAwIAA3MAAzYE",
		},
	},

	constbot.TmpUserInfo: {
		"en": {
			Msgtmpl: `
				Name {{.Name}}
				Username	= @{{.Username}}
				TgId		= {{.TgId}}
				
				Dedicated {{.Dedicated}}
				TQuota       {{.TQuota}}
				LeftQuota    {{.LeftQuota}}
				ConfCount   {{.ConfCount}}
				TUsage       {{.TUsage}}
				GiftQuota    {{.GiftQuota}}
				Joined       {{.Joined}}
				CapEndin     {{.CapEndin}}
				Disendin    {{.Disendin}}
				UsageResetIn{{.UsageResetIn}}

				Iscapped       {{.Iscapped}}
				Isgifted       {{.Isgifted}}
				Isdisuser      {{.Isdisuser}}
				IsMonthLimited {{.IsMonthLimited}}
				JoinedPlace {{.JoinedPlace}}

				{{if .Iscapped}}
    			CapEndin = {{.CapEndin}}
				{{end}}

   			 	{{if .Isgifted}}
        		GiftEndIn = {{.GiftEndIn}}
    			{{end}}
			

			`,
		},
	},

	constbot.TmpDisGroup: {
		"en": {
			Msgtmpl: "Hello Everyone user{{.Name}} is distributed his quota everyone clap hiom and thanks he dedicated {{.Disquota}}, රටක් දියුනු වෙන්න මේ වගෙ එව්න් තමා ඕන  ",

			Includemed: true,
			Mediatype:  "photo",
			MediaId:    "AgACAgUAAxkBAAEvZapnOFjKbTYqbIAgTIy0xf-VZQQ0PwACUcAxG8dwyVV7OMp6tw3cUwEAAwIAA3MAAzYE",
		},
	},

	constbot.TmpCrInInfo: {
		"en": {
			Msgtmpl: `selected inbound info is here if you want excatly this seeting for your config conform also you can change the inbound later 
			
			InName {{.InName}}
			InType {{.InType}}
			InPort {{.InPort}}
			InAddr {{.InAddr}}
			InInfo {{.InInfo}}
			TranstPortType {{.TranstPortType}}
			TlsEnabled {{.TlsEnabled}}
			Domain 	{{.Domain}}
			PublicIp {{.PublicIp}}
			
			`,
		},
	},


	constbot.TmpCrSendUID: {
		"en": {
			Msgtmpl: `

			this is tempory config structure we will provide good way to generate configs as you need till then use this

			UUID {{.UUID}}
			Domain {{.Domain}}
			Transport {{.Transport}}
			ConfigName {{.ConfigName}}
			TlsEnabled {{.TlsEnabled}}
			Port {{.Port}}



			`,
		},
	},

	constbot.TmpInchange: {
		"en": {
			Msgtmpl: `

			Info about inbound that you are goung to change 

			InName {{.InName}}
			InType {{.InType}}
			InPort {{.InPort}}
			InAddr {{.InAddr}}
			InInfo {{.InInfo}}
			TranstPortType {{.TranstPortType}}
			TlsEnabled {{.TlsEnabled}}
			Domain 	{{.Domain}}
			PublicIp {{.PublicIp}}


			`,
		},
	},
	constbot.TmpOutchange: {
		"en": {
			Msgtmpl: `

			Selected Outbound info

			OutName {{.OutName}}
			OutType {{.OutType}}
			OutInfo {{.OutInfo}}
			Latency {{.Latency}}ms


			`,
		},
	},
}

//_ = constbot

// constbot.TmpConQuota: {"en": "Send New quota for config {{.AvblQuota}}"},

// constbot.TmpCrAvblQuota: {"en": " You have Avalble quota {{.Quota}}"},

// constbot.TmpCrAlreadyHave: {"en": "you already have {{.Count}}"},

// constbot.TmpStTotal:    {"en": "your usage tddwn {{.TDownload}}  tdup {{.TUpload}}  totalUp {{.MUpload}}  totaldown {{.MDownload}}  Alltime {{.Alltime}}           if you want config specifc usage click below buttons      "},
// constbot.TmpStcallback: {"en": "your usage tddwn {{.TDownload}}  tdup {{.TUpload}}  totalUp {{.MUpload}}  totaldown {{.MDownload}} "},
