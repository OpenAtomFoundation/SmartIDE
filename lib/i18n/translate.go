package i18n

import (
	"embed"

	"encoding/json"

	"strings"

	"github.com/jeandeaual/go-locale"

	// 这里不要忘记引入驱动,如引入默认的json驱动

	_ "github.com/leansoftX/i18n/parser_json"
)

//
type I18nSource struct {
	Config struct {
		Info struct {
			Read_docker_compose string `json:"read_docker_compose"`
		} `json:"info"`
		Error struct {
			Services_not_exit         string `json:"services_not_exit"`
			File_not_exit             string `json:"file_not_exit"`
			Devcontainer_not_contains string `json:"devcontainer_not_contains"`
			Gitconfig_not_exit        string `json:"gitconfig_not_exit"`
		} `json:"error"`
	} `json:"config"`

	Main struct {
		Info struct {
			Help_short      string `json:"help_short"`
			Help_long       string `json:"help_long"`
			Help_flag_debug string `json:"help_flag_debug"`
			Usage_template  string `json:"usage_template"`
		} `json:"info"`
		Error struct {
			File_not_exit     string `json:"file_not_exit"`
			Version_not_build string `json:"version_not_build"`
		} `json:"error"`
	} `json:"main"`

	Help struct {
		Info struct {
			Help_short string `json:"help_short"`
			Help_long  string `json:"help_long"`
		} `json:"info"`
	} `json:"help"`

	VmStart struct {
		Info struct {
			Help_short string `json:"help_short"`
			Help_long  string `json:"help_long"`

			Info_starting              string `json:"info_starting"`
			Info_connect_remote        string `json:"info_connect_remote"`
			Info_git_clone             string `json:"info_git_clone"`
			Info_git_checkout_and_pull string `json:"info_git_checkout_and_pull"`
			Info_read_config           string `json:"info_read_config"`
			Info_create_network        string `json:"info_create_network"`
			Info_compose_up            string `json:"info_compose_up"`
			Info_warting_for_webide    string `json:"info_warting_for_webide"`
			Info_open_brower           string `json:"info_open_brower"`
		} `json:"info"`
	} `json:"vm_start"`

	Version struct {
		Info struct {
			Help_short string `json:"help_short"`
			Help_long  string `json:"help_long"`
			Template   string `json:"template"`
		} `json:"info"`
	} `json:"version"`

	Update struct {
		Info struct {
			Help_short         string `json:"help_short"`
			Help_long          string `json:"help_long"`
			Info_remove_repeat string `json:"info_remove_repeat"`
			Help_flag_version  string `json:"help_flag_version"`
			Help_flag_build    string `json:"help_flag_build"`
		} `json:"info"`

		Warn struct {
			Warn_rel_lastest string `json:"warn_rel_lastest"`
		} `json:"warn"`
	} `json:"update"`

	Init struct {
		Info struct {
			Help_short string `json:"help_short"`
			Help_long  string `json:"help_long"`
			Info_start string `json:"info_start"`
			Info_end   string `json:"info_end"`
		} `json:"info"`
	} `json:"init"`

	Start struct {
		Info struct {
			Help_short                   string `json:"help_short"`
			Help_long                    string `json:"help_long"`
			Info_start                   string `json:"info_start"`
			Info_end                     string `json:"info_end"`
			Info_running_container       string `json:"info_running_container"`
			Info_running_openbrower      string `json:"info_running_openbrower"`
			Info_docker_compose_filepath string `json:"info_docker_compose_filepath"`
			Info_ssh_tunnel              string `json:"info_ssh_tunnel"`
			Info_create_network          string `json:"info_create_network"`
			Info_open_in_brower          string `json:"info_open_in_brower"`
			Help_flag_filepath           string `json:"help_flag_filepath"`

			Info_saving_workspace string `json:"info_saving_workspace"`
		} `json:"info"`
		Error struct {
			Docker_Err          string `json:"docker_err"`
			DockerPs_Err        string `json:"dockerps_err"`
			Docker_Compose_Err  string `json:"docker_compose_err"`
			Docker_started      string `json:"docker_started"`
			Docker_compose_save string `json:"docker_compose_save"`
		} `json:"error"`
	} `json:"start"`

	Stop struct {
		Info struct {
			Help_short         string `json:"help_short"`
			Help_long          string `json:"help_long"`
			Info_start         string `json:"info_start"`
			Info_end           string `json:"info_end"`
			Help_flag_filepath string `json:"help_flag_filepath"`
		} `json:"info"`
		Error struct {
		} `json:"error"`
	} `json:"Stop"`

	Remove struct {
		Info struct {
			Help_short         string `json:"help_short"`
			Help_long          string `json:"help_long"`
			Info_start         string `json:"info_start"`
			Info_end           string `json:"info_end"`
			Help_flag_filepath string `json:"help_flag_filepath"`
		} `json:"info"`
		Error struct {
		} `json:"error"`
	} `json:"Remove"`

	ReStart struct {
		Info struct {
			Help_short         string `json:"help_short"`
			Help_long          string `json:"help_long"`
			Help_flag_filepath string `json:"help_flag_filepath"`
		} `json:"info"`
		Error struct {
		} `json:"error"`
	} `json:"restart"`

	New struct {
		Info struct {
			Help_short              string `json:"help_short"`
			Help_long               string `json:"help_long"`
			Help_info               string `json:"help_info"`
			Help_info_operation     string `json:"help_info_operation"`
			Help_flag_type          string `json:"help_flag_type"`
			Help_flag_projectFolder string `json:"help_flag_projectFolder"`
			Info_yaml_exist         string `json:"info_yaml_exist"`
			Info_noempty_is_comfirm string `json:"info_noempty_is_comfirm"`
			Info_type_no_exist      string `json:"info_type_no_exist"`
			Info_creating_project   string `json:"info_creating_project"`
		} `json:"info"`
		Error struct {
		} `json:"error"`
	} `json:"new"`

	Host struct {
		Info struct {
			Help_short string `json:"help_short"`
			Help_long  string `json:"help_long"`

			Help_get_short string `json:"help_get_short"`
			Help_get_long  string `json:"help_get_long"`

			Help_list_short string `json:"help_list_short"`
			Help_list_long  string `json:"help_list_long"`

			Info_host_table_header    string `json:"info_host_table_header"`
			Info_host_detail_template string `json:"info_host_detail_template"`

			Help_flag_hostid string `json:"help_flag_hostid"`
		} `json:"info"`
		Error struct {
		} `json:"error"`
	} `json:"host"`

	Common struct {
		Debug struct {
			Debug_key_public           string `json:"debug_key_public"`
			Debug_same_not_overwrite   string `json:"debug_same_not_overwrite"`
			Debug_auto_connect_gitrepo string `json:"debug_auto_connect_gitrepo"`
			Debug_empty_error          string `json:"debug_empty_error"`
		} `json:"debug"`
		Error struct {
			Err_sshremote_param_repourl_none string `json:"err_sshremote_param_repourl_none"`
			Err_password_none                string `json:"err_password_none"`

			Err_dal_record_repeat         string `json:"err_dal_record_repeat"`
			Err_dal_update_fail           string `json:"err_dal_update_fail"`
			Err_dal_update_count_too_much string `json:"err_dal_update_count_too_much"`
			Err_enum_error                string `json:"err_ernum_error"`
		} `json:"error"`
		Info struct {
			Info_privatekey_is_overwrite   string `json:"info_privatekey_is_overwrite"`
			Info_whether_overwrite         string `json:"info_whether_overwrite"`
			Info_gitrepo_cloned            string `json:"info_gitrepo_cloned"`
			Info_please_enter_password     string `json:"info_please_enter_password"`
			Info_canel_privatekey_password string `json:"info_canel_privatekey_password"`
			Info_port_is_binding           string `json:"info_port_is_binding"`
			Info_port_binding_result2      string `json:"info_port_binding_result2"`
			Info_port_binding_result       string `json:"info_port_binding_result"`
			Info_find_new_port             string `json:"info_find_new_port"`

			Info_webide_host_port string `json:"info_webide_host_port"`
			Info_ssh_host_port    string `json:"info_ssh_host_port"`

			Info_create_directory string `json:"info_create_directory"`

			Info_temp_created_docker_compose string `json:"info_temp_created_docker_compose"`
			Info_temp_created_config         string `json:"info_temp_created_config"`

			Info_table_header_containers string `json:"info_table_header_containers"`
		} `json:"info"`
		Warn struct {
			// 没有查询到对应的数据
			Warn_dal_record_not_exit_condition string `json:"warn_dal_record_not_exit_condition"`
			// 没有查询到对应的数据
			Warn_dal_record_not_exit string `json:"warn_dal_record_not_exit"`
			// 参数为空
			Warn_param_is_null string `json:"warn_param_is_null"`
		} `json:"warn"`
	} `json:"common"`
}

var instance *I18nSource

/* var I18nSource_EN string
var I18nSource_ZH string */

//go:embed language/*
var f embed.FS

// get internationalization source
// 获取当前系统的语言，动态加载对应的json文件并解析成结构体，方便在代码中调用
// 1. 新增，首先在 “lib/i18n/language” 的对应节点下新增，并同步在 “lib/i18n/language/translate.go” 中的 “I18nSource” 增加相应的属性；
// 2. 在代码中使用
//    var instanceI18nStart = i18n.GetInstance().Start
//    fmt.println(instanceI18nStart.Help_short)
func GetInstance() *I18nSource {
	if instance == nil {
		// locale
		currentLang, _ := locale.GetLocale()
		if strings.Index(currentLang, "zh-") == 0 { // 如果不是简体中文，就是英文
			currentLang = "zh_cn"
		} else {
			currentLang = "en_us"
		}

		// loading and parse json
		data, _ := f.ReadFile("language/" + currentLang + "/info.json")
		json.Unmarshal(data, &instance)
	}
	return instance
}
