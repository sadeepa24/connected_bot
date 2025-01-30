package constbot

// command configure
const (
	TmpConfiConfigure string = "conf_configure" // msg before change config, para - {{.Name}}
 	TmplConfigureHome string = "configure_home"
 	TmpInchange string = "conf_in_change"
 	TmpOutchange string = "conf_out_change"
 	TmpConQuota string = "conf_quota_change"
	TmpNameChange string = "conf_name_change"      // msg when user change name of the configvia configurecommand, para - {{.Name}}

)

// CommandCreate
const (
	TmpCrAlreadyHave string = "create_conf_limit"
	TmpCrAvblQuota   string = "create_available_quota"
	TmpCrSendUID     string = "create_result"
	TmpCrInInfo      string = "create_in_info"
	TmpCrOutInfo     string = "create_in_info"
	TmplCrSelect     string = "create_select"
)

// CommandStatus
const (
	TmpStTotal    string = "status_home"
	TmpStcallback string = "status_callback"
)

// cap
const (
	TmpcapQuota string = "setcap_already"
	TmpcapWarn  string = "setcap_warn"
	Tmpcapreply string = "setcap_get"
)

// Gift
const (
	TmpGifSend  string = "gift_send"
	TmplRecived string = "gift_reciver"
)

// Chatmemupdate
const (
	TmpChatmemLeft         string = "chat_mem_left"
	TmpGroupWelcome        string = "grp_welcome"
	TmpWelcomeInbox        string = "dm_welcome"
	TmplInboxVerified      string = "dm_verified"
	TmplInboxVerifiedAgain string = "dm_verified_again"
	TmpGrpComeback         string = "grp_comeback"  //user who is not in channel join group
	TmpChanComeback        string = "chan_comeback" //user who is not in channel join group
	TmpChannelWelcome      string = "chan_welcome"
)

// common
const (
	TmplCommonUnverified string = "com_unverified"
)

// help
const (
	TmpHelpHome         string = "help_home"
	TmpHelpCmPage       string = "help_cmd"     // this does not use directly istead TmpPage + pagenum
	TmpHelpInfoPage     string = "help_info" // this does not use directly istead TmpPage + pagenum
	TmplHelpBuilderHelp string = "help_builder"
	TmplHelpTuto        string = "help_tutorial"
	TmpAbout            string = "help_about"
)

// start
const (
	TmpNewUsers             string = "start_newuser"
	TmpNewUsersVerified     string = "start_newuser_verified"
	TmplUserUnverifiedStart string = "start_newuser_unverified"

	TmpregularVerified string = "start_regular"

	TmplMonthLimited string = "start_monthlimited"
	TmpRemUserst     string = "start_removed"
	TmpRestrcistr string = "start_restricted"
)

// refer
const (
	TmpRefHome  string = "refer_home"
	TmpRefshare string = "refer_share"
)

// distribute
const (
	TmpDisGroup string = "distribute_group"
)

// points
const (
	TmplPoints string = "points_home"
)

// events
const (
	TmplEventHome string = "event_home"
)

const (
	TmplBuilderHome string = "builder_home"
)


//getinfo
const (
	TmplGetinfoHome string = "getinfo_home"
	TmpUserInfo string = "getinfo_user"
	TmplOutInfo string = "getinfo_out"
	TmpConfigInfo string = "getinfo_usage"
)
