/*
 * @Author: jason chen
 * @Date: 2021-11-08
 * @Description: sqlite data access layer
 */

package dal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	_ "github.com/logoove/sqlite"
	"gopkg.in/yaml.v2"
	//_ "github.com/mattn/go-sqlite3"
)

// key, 只能 16、24 或者 32位
const aesDecryptKey string = "smartide@1O24!QAZWSXedcrfv"

// 多语言
var i18nInstance *i18n.I18nSource = i18n.GetInstance()

// workspace orm
type workspaceDo struct {
	w_id                       int
	w_name                     string
	w_workingdir               sql.NullString
	w_docker_compose_file_path sql.NullString
	w_mode                     string
	w_config_file              sql.NullString
	w_git_clone_repo_url       sql.NullString
	w_git_auth_type            string
	w_branch                   string
	r_id                       sql.NullInt32
	k_id                       sql.NullInt32
	w_is_del                   bool

	w_json sql.NullString

	w_config_content       sql.NullString
	w_link_compose_content sql.NullString
	w_temp_compose_content sql.NullString

	w_created time.Time
}

// 插入工作空间的数据
func InsertOrUpdateWorkspace(workspaceInfo workspace.WorkspaceInfo) (affectId int64, err error) {
	//1. init
	db := getDb()
	defer db.Close()

	//2. 是否数据已经存在
	isExit := false
	remoteID := -1
	remoteHost := ""
	if (workspaceInfo.Remote != workspace.RemoteInfo{}) {
		remoteID = workspaceInfo.Remote.ID
		remoteHost = workspaceInfo.Remote.Addr
	}
	if workspaceInfo.ID > 0 { // 用户录入workspaceid的情况
		affectId = int64(workspaceInfo.ID)
		isExit = true
	} else { // 用户有可能会不输入workspaceid，继续使用原有的参数
		originWorkspace, err := GetSingleWorkspaceByParams(workspaceInfo.Mode, workspaceInfo.WorkingDirectoryPath, workspaceInfo.GitCloneRepoUrl, remoteID, remoteHost)
		common.CheckError(err)
		if originWorkspace.IsNotNil() {
			affectId = int64(originWorkspace.ID)
			isExit = true
		}
	}

	//3. insert or update
	jsonBytes, err := json.Marshal(workspaceInfo.Extend) // 扩展字段序列化
	common.CheckError(err)

	if workspaceInfo.ConfigYaml.IsNil() {
		return -1, errors.New("配置文件数据为空！") //TODO
	}
	if workspaceInfo.TempDockerCompose.IsNil() && workspaceInfo.Mode != workspace.WorkingMode_K8s {
		return -1, errors.New("生成docker-compose数据为空！") //TODO
	}

	configStr, _ := workspaceInfo.ConfigYaml.ToYaml()
	linkComposeStr, _ := workspaceInfo.LinkDockerCompose.ToYaml()
	tempComposeStr, _ := workspaceInfo.TempDockerCompose.ToYaml()
	if !isExit { // insert
		remoteId := sql.NullInt32{} // 可能是个空值
		k8sId := sql.NullInt32{}    // 可能是个空值

		//3.1. 插入到remote表中
		if workspaceInfo.Mode != workspace.WorkingMode_K8s {
			if (workspaceInfo.Remote != workspace.RemoteInfo{}) {
				tmpId, err := InsertOrUpdateRemote(workspaceInfo.Remote)
				common.CheckError(err)
				if tmpId > 0 {
					remoteId = sql.NullInt32{
						Int32: int32(tmpId),
						Valid: true,
					}
				}
			}
		} else {
			tmpId, err := InsertOrUpdateK8sInfo(workspaceInfo.K8sInfo)
			common.CheckError(err)
			if tmpId > 0 {
				k8sId = sql.NullInt32{
					Int32: int32(tmpId),
					Valid: true,
				}
			}
		}

		//3.2.
		stmt, err := db.Prepare(`INSERT INTO workspace(w_name, w_workingdir, w_docker_compose_file_path, w_config_file, r_id,k_id,
												w_mode, w_git_clone_repo_url, w_git_auth_type, w_branch,
												w_json, w_config_content, w_link_compose_content, w_temp_compose_content)  
						VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return -1, err
		}

		res, err := stmt.Exec(workspaceInfo.Name, workspaceInfo.WorkingDirectoryPath, workspaceInfo.TempDockerComposeFilePath, workspaceInfo.ConfigFilePath, remoteId, k8sId,
			workspaceInfo.Mode, workspaceInfo.GitCloneRepoUrl, workspaceInfo.GitRepoAuthType, workspaceInfo.Branch,
			string(jsonBytes), configStr, linkComposeStr, tempComposeStr)
		if err != nil {
			return -1, err
		}
		return res.LastInsertId()
	} else { // update
		stmt, err := db.Prepare(`update workspace 
								set w_name=?, w_workingdir=?, w_docker_compose_file_path=?, w_config_file=?,
									w_mode=?, w_git_clone_repo_url=?, w_git_auth_type=?, w_branch=?,
									w_json=?, w_config_content=?, w_link_compose_content=?, w_temp_compose_content=?
								where w_id=?`)
		if err != nil {
			return -1, err
		}
		_, err = stmt.Exec(workspaceInfo.Name, workspaceInfo.WorkingDirectoryPath, workspaceInfo.TempDockerComposeFilePath, workspaceInfo.ConfigFilePath,
			workspaceInfo.Mode, workspaceInfo.GitCloneRepoUrl, workspaceInfo.GitRepoAuthType, workspaceInfo.Branch,
			string(jsonBytes), configStr, linkComposeStr, tempComposeStr,
			affectId)
		if err != nil {
			return -1, err
		}
	}

	return affectId, err
}

//
func GetWorkspaceList() (workspaces []workspace.WorkspaceInfo, err error) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query(`select w_id, w_name, w_workingdir, w_docker_compose_file_path, w_mode, w_config_file,
									w_git_clone_repo_url, w_git_auth_type, w_branch, r_id, w_is_del, 
									w_json, w_config_content, w_link_compose_content, w_temp_compose_content, 
									w_created 
							from workspace 
							where w_is_del = 0
							order by w_created desc`)
	for rows.Next() {
		do := workspaceDo{}
		switch errSql := rows.Scan(&do.w_id, &do.w_name, &do.w_workingdir, &do.w_docker_compose_file_path, &do.w_mode, &do.w_config_file,
			&do.w_git_clone_repo_url, &do.w_git_auth_type, &do.w_branch, &do.r_id, &do.w_is_del,
			&do.w_json, &do.w_config_content, &do.w_link_compose_content, &do.w_temp_compose_content,
			&do.w_created); errSql {
		/* case sql.ErrNoRows:
		common.SmartIDELog.Warning() //TODO */
		case nil:

			workspaceInfo := workspace.WorkspaceInfo{}
			err = workspaceDataMap(&workspaceInfo, do)

			workspaces = append(workspaces, workspaceInfo)

		default:
			err = errSql
		}
		//   fmt.Println(uid, username, department, created)
	}

	return workspaces, err
}

//
func GetSingleWorkspaceByParams(workingMode workspace.WorkingMode, workingDir string, gitCloneUrl string, remoteId int, remoteHost string) (workspaceInfo workspace.WorkspaceInfo, err error) {
	db := getDb()
	defer db.Close()

	// 查询参数
	params := workspaceDo{}
	params.w_mode = string(workingMode)
	if workingDir != "" {
		params.w_workingdir = sql.NullString{String: workingDir, Valid: true}
	}
	if gitCloneUrl != "" {
		params.w_git_clone_repo_url = sql.NullString{String: gitCloneUrl, Valid: true}
	}
	if remoteId <= 0 {
		remoteInfo, err := getRemote(remoteId, remoteHost)
		common.CheckError(err)
		if remoteInfo.ID > 0 {
			params.r_id = sql.NullInt32{Int32: int32(remoteInfo.ID), Valid: true}
		}
	} else {
		params.r_id = sql.NullInt32{Int32: int32(remoteId), Valid: true}
	}

	// sql
	var row *sql.Row
	if workingMode == workspace.WorkingMode_Remote {
		row = db.QueryRow(`select w_id, w_name, w_workingdir, w_docker_compose_file_path, w_mode, w_config_file,
								w_git_clone_repo_url, w_git_auth_type, w_branch, r_id, w_is_del, 
								w_json, w_config_content, w_link_compose_content, w_temp_compose_content, 
								w_created 
							from workspace 
							where w_workingdir=? 
							and w_mode=? 
							and w_git_clone_repo_url=? 
							and r_id = ?
							and w_is_del = 0`,
			params.w_workingdir, workingMode, params.w_git_clone_repo_url, params.r_id)
	} else {
		row = db.QueryRow(`select w_id, w_name, w_workingdir, w_docker_compose_file_path, w_mode, w_config_file,
								w_git_clone_repo_url, w_git_auth_type, w_branch, r_id, w_is_del, 
								w_json, w_config_content, w_link_compose_content, w_temp_compose_content, 
								w_created 
							from workspace 
							where w_workingdir=? 
							and w_mode=? 
							and w_git_clone_repo_url=? 
							and w_is_del = 0`,
			params.w_workingdir, workingMode, params.w_git_clone_repo_url)
	}

	// 赋值
	do := workspaceDo{}
	switch err := row.Scan(&do.w_id, &do.w_name, &do.w_workingdir, &do.w_docker_compose_file_path, &do.w_mode, &do.w_config_file,
		&do.w_git_clone_repo_url, &do.w_git_auth_type, &do.w_branch, &do.r_id, &do.w_is_del,
		&do.w_json, &do.w_config_content, &do.w_link_compose_content, &do.w_temp_compose_content,
		&do.w_created); err {
	case sql.ErrNoRows:
		msg := fmt.Sprintf("（%v，%v，%v）", workingDir, workingMode, gitCloneUrl)
		common.SmartIDELog.WarningF(i18nInstance.Common.Warn_dal_record_not_exit_condition, msg)
	case nil:
		err = workspaceDataMap(&workspaceInfo, do)
		return workspaceInfo, err
	default:
		panic(err)
	}

	// return
	return workspaceInfo, err
}

// 赋值
func workspaceDataMap(workspaceInfo *workspace.WorkspaceInfo, do workspaceDo) error {

	// 基本信息
	workspaceInfo.ID = do.w_id
	workspaceInfo.Name = do.w_name
	workspaceInfo.WorkingDirectoryPath = do.w_workingdir.String
	workspaceInfo.ConfigFilePath = do.w_config_file.String
	workspaceInfo.TempDockerComposeFilePath = do.w_docker_compose_file_path.String

	// 初始化配置文件
	if do.w_mode == string(workspace.WorkingMode_Remote) {
		workspaceInfo.ConfigYaml = *config.NewConfigRemote(do.w_workingdir.String, do.w_config_file.String, do.w_config_content.String)

	} else {
		workspaceInfo.ConfigYaml = *config.NewConfig(do.w_workingdir.String, do.w_config_file.String, do.w_config_content.String)

	}

	// 连接的docker-compose文件
	if do.w_link_compose_content.String != "" {
		err := yaml.Unmarshal([]byte(do.w_link_compose_content.String), &workspaceInfo.LinkDockerCompose)
		if err != nil {
			return err
		}
	}

	// 生成的docker-compose文件
	if do.w_temp_compose_content.String != "" {
		err := yaml.Unmarshal([]byte(do.w_temp_compose_content.String), &workspaceInfo.TempDockerCompose)
		if err != nil {
			return err
		}
	}

	// 扩展属性
	if do.w_json.String != "" {
		err := json.Unmarshal([]byte(do.w_json.String), &workspaceInfo.Extend)
		common.CheckError(err)
	}

	// 类型
	if do.w_mode == string(workspace.WorkingMode_Local) {
		workspaceInfo.Mode = workspace.WorkingMode_Local
	} else if do.w_mode == string(workspace.WorkingMode_Remote) {
		workspaceInfo.Mode = workspace.WorkingMode_Remote
	} else if do.w_mode == string(workspace.WorkingMode_K8s) {
		workspaceInfo.Mode = workspace.WorkingMode_K8s
	} else {
		panic("w_mode != string(WorkingMode_Local)")
	}

	// git 验证方式
	if do.w_git_auth_type == string(workspace.GitRepoAuthType_SSH) {
		workspaceInfo.GitRepoAuthType = workspace.GitRepoAuthType_SSH
	} else {
		workspaceInfo.GitRepoAuthType = workspace.GitRepoAuthType_HTTPS
	}

	// git 相关
	workspaceInfo.GitCloneRepoUrl = do.w_git_clone_repo_url.String
	workspaceInfo.Branch = do.w_branch

	// 远程主机信息
	rid := int(do.r_id.Int32)
	if rid >= 0 {
		workspaceInfo.Remote, _ = GetRemoteById(rid)
	}
	kid := int(do.k_id.Int32)
	if int(kid) >= 0 {
		workspaceInfo.K8sInfo, _ = GetK8sInfoById(kid)
	}

	// 其他
	workspaceInfo.CreatedTime = do.w_created

	return nil
}

// 获取工作空间的详情数据
func GetSingleWorkspace(workspaceid int) (workspaceInfo workspace.WorkspaceInfo, err error) {

	db := getDb()
	defer db.Close()

	do := workspaceDo{}
	row := db.QueryRow(`select w_id, w_name, w_workingdir, w_docker_compose_file_path, w_mode, w_config_file, 
								w_git_clone_repo_url, w_git_auth_type, w_branch, r_id, k_id,
								w_json, w_config_content, w_link_compose_content, w_temp_compose_content, 
								w_created 
					    from workspace 
						where w_id=? and w_is_del = 0`, workspaceid)
	switch err := row.Scan(&do.w_id, &do.w_name, &do.w_workingdir, &do.w_docker_compose_file_path, &do.w_mode, &do.w_config_file,
		&do.w_git_clone_repo_url, &do.w_git_auth_type, &do.w_branch, &do.r_id, &do.k_id,
		&do.w_json, &do.w_config_content, &do.w_link_compose_content, &do.w_temp_compose_content,
		&do.w_created); err {
	case sql.ErrNoRows:
		common.SmartIDELog.WarningF(i18nInstance.Common.Warn_dal_record_not_exit_condition, "workspaceid ("+strconv.Itoa(workspaceid)+")") // 没有查询到数据
	case nil:
		err = workspaceDataMap(&workspaceInfo, do)
	default:
		panic(err)
	}

	return workspaceInfo, err
}

//
func RemoveWorkspace(workspaceId int) error {
	db := getDb()
	defer db.Close()

	// 数据校验
	var count int
	row := db.QueryRow("select count(1) from workspace where w_id=? and w_is_del = 0", workspaceId)
	switch err := row.Scan(&count); err {
	case sql.ErrNoRows:
		msg := fmt.Sprintf(" workspace (%v) ", workspaceId)
		common.SmartIDELog.WarningF(i18nInstance.Common.Warn_dal_record_not_exit_condition, msg)
	case nil:
		if count <= 0 {
			return errors.New(i18nInstance.Common.Warn_dal_record_not_exit)
		} else if count > 1 {
			return errors.New(i18nInstance.Common.Err_dal_update_count_too_much)
		}
	default:
		panic(err)
	}

	//
	stmt, err := db.Prepare("update workspace set w_is_del=1 where (w_id=?) and w_is_del = 0")
	if err != nil {
		return err
	}
	res, err := stmt.Exec(workspaceId)
	if err != nil {
		return err
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return err
	}

	//
	if affect != 1 {
		if affect <= 0 {
			return errors.New(i18nInstance.Common.Err_dal_update_fail) // 更新失败
		} else if affect > 1 {
			common.SmartIDELog.Warning(i18nInstance.Common.Err_dal_update_count_too_much) // 更新条目 > 1
		}
	}

	return nil
}
