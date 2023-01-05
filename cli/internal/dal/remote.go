/*
SmartIDE - Dev Containers
Copyright (C) 2023 leansoftX.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	aes4go "github.com/leansoftX/smartide-cli/pkg/aes"

	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

// remote orm
type remoteDO struct {
	r_id        int
	r_addr      string
	r_port      sql.NullInt32
	r_username  string
	r_auth_type string
	r_password  sql.NullString
	//r_is_del    bool
	r_created time.Time
}

func GetRemoteList() (remoteList []workspace.RemoteInfo, err error) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query(`select r_id, r_addr, r_port, r_auth_type, r_username, r_created
							from remote 
							where r_is_del = 0`)
	for rows.Next() {
		do := remoteDO{}

		switch errSql := rows.Scan(&do.r_id, &do.r_addr, &do.r_port, &do.r_auth_type, &do.r_username, &do.r_created); errSql {
		/* case sql.ErrNoRows:
		   common.SmartIDELog.Warning() //TODO */
		case nil:

			remoteInfo := workspace.RemoteInfo{}

			remoteInfo.ID = do.r_id
			remoteInfo.Addr = do.r_addr

			if do.r_port.Valid {
				remoteInfo.SSHPort = int(do.r_port.Int32)
			} else {
				remoteInfo.SSHPort = model.CONST_Container_SSHPort
			}

			remoteInfo.AuthType = workspace.RemoteAuthType(do.r_auth_type)
			remoteInfo.UserName = do.r_username
			remoteInfo.CreatedTime = do.r_created

			remoteList = append(remoteList, remoteInfo)

		default:
			err = errSql
		}
		//   fmt.Println(uid, username, department, created)
	}

	return remoteList, err
}

func GetRemoteById(remoteId int) (remoteInfo *workspace.RemoteInfo, err error) {
	return getRemote(remoteId, "", "")
}

func GetRemoteByHost(host string, userName string) (remoteInfo *workspace.RemoteInfo, err error) {
	return getRemote(0, host, userName)
}

func RemoveRemote(remoteId int, host string, userName string) error {
	db := getDb()
	defer db.Close()

	// 数据校验
	var exitCount, referenceCount int
	var row *sql.Row
	if len(host) > 0 {
		row = db.QueryRow(`select count(1) exitCount,
		   (select count(1) from workspace where w_is_del = 0 and r_id = remote.r_id) referenceCount 
		from remote 
		where r_addr=? and r_username = ? and r_is_del = 0`, host, userName)
	} else if remoteId > 0 {
		row = db.QueryRow("select count(1) exitCount,(select count(1) from workspace where w_is_del = 0 and r_id = remote.r_id) referenceCount from remote where r_id=? and r_is_del = 0", remoteId)
	}
	switch err := row.Scan(&exitCount, &referenceCount); err {
	case sql.ErrNoRows:
		msg := fmt.Sprintf("remote （%v | %v）", remoteId, host)
		common.SmartIDELog.WarningF(i18nInstance.Common.Warn_dal_record_not_exit_condition, msg) // 没有查询到数据
	case nil:
		if exitCount <= 0 { // 没有找到相关的记录
			return errors.New(i18nInstance.Common.Warn_dal_record_not_exit)
		} else if exitCount > 1 { // 存在多条记录
			return errors.New(i18nInstance.Common.Err_dal_record_repeat)
		}
	default:
		return err
	}

	// 是否被其他的workspace引用
	if referenceCount > 0 {
		return errors.New(i18nInstance.Common.Err_dal_remote_reference_by_workspace)
	}

	//
	var stmt *sql.Stmt
	var err error
	if len(host) > 0 {
		stmt, err = db.Prepare("update remote set r_is_del=1 where r_addr=? and r_is_del = 0")
	} else if remoteId > 0 {
		stmt, err = db.Prepare("update remote set r_is_del=1 where r_id=? and r_is_del = 0")
	}
	if err != nil {
		return err
	}
	res, err := stmt.Exec(remoteId, host)
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
			common.SmartIDELog.Warning(i18nInstance.Common.Err_dal_update_count_too_much)
		}
	}

	return nil
}

func InsertOrUpdateRemote(remoteInfo workspace.RemoteInfo) (id int, err error) {

	var single *workspace.RemoteInfo

	if remoteInfo.ID > 0 {
		single, err = GetRemoteById(remoteInfo.ID)
		if err != nil {
			return id, err
		}
	} else {
		single, err = GetRemoteByHost(remoteInfo.Addr, remoteInfo.UserName)
		if err != nil {
			return id, err
		}
	}

	//1. init
	db := getDb()
	defer db.Close()

	passwordEncrypt := ""
	if len(remoteInfo.Password) > 0 {
		passwordEncrypt = aes4go.Encrypt(remoteInfo.Password, aesDecryptKey)
	}

	//2. insert or update
	if single != nil { //2.1. update
		stmt, err := db.Prepare(`update remote
		set r_addr=?, r_port=?, r_username=?, r_auth_type=?, r_password=?  
		where r_id=? or r_addr=?`)
		if err != nil {
			return id, err
		}
		_, err = stmt.Exec(remoteInfo.Addr, remoteInfo.SSHPort, remoteInfo.UserName, remoteInfo.AuthType, passwordEncrypt,
			remoteInfo.ID, remoteInfo.Addr)
		if err != nil {
			return -1, err
		}

		id = single.ID

	} else { //2.2. insert
		stmt, err := db.Prepare(`INSERT INTO remote(r_addr, r_port, r_username, r_auth_type, r_password)  
                                        values(?, ?, ?, ?, ?)`)
		if err != nil {
			return id, err
		}
		res, err := stmt.Exec(remoteInfo.Addr, remoteInfo.SSHPort, remoteInfo.UserName, remoteInfo.AuthType, passwordEncrypt)
		if err != nil {
			return id, err
		}
		id64, err := res.LastInsertId()
		id = int(id64)
		return id, err
	}

	return id, err
}

func getRemote(remoteId int, host string, userName string) (remoteInfo *workspace.RemoteInfo, err error) {

	db := getDb()
	defer db.Close()

	do := remoteDO{}

	var row *sql.Row
	if len(host) > 0 {
		row = db.QueryRow(`select r_id, r_addr, r_port, r_username, r_auth_type, r_password, r_created 
		from remote 
		where r_addr=? and r_username = ? and r_is_del = 0`, host, userName)
	} else if remoteId > 0 {
		row = db.QueryRow(`select r_id, r_addr, r_port, r_username, r_auth_type, r_password, r_created 
		from remote where r_id=? and r_is_del = 0`, remoteId)
	} else {
		return
	}

	switch err := row.Scan(&do.r_id, &do.r_addr, &do.r_port, &do.r_username, &do.r_auth_type, &do.r_password, &do.r_created); err {
	case sql.ErrNoRows:
		msg := fmt.Sprintf("host (%v | %v)", host, remoteId)
		common.SmartIDELog.WarningF(i18nInstance.Common.Warn_dal_record_not_exit_condition, msg) // 不存在
	case nil:
		remoteInfo = &workspace.RemoteInfo{}
		remoteInfo.ID = do.r_id
		remoteInfo.Addr = do.r_addr
		remoteInfo.UserName = do.r_username
		if do.r_auth_type == string(workspace.RemoteAuthType_SSH) {
			remoteInfo.AuthType = workspace.RemoteAuthType_SSH
		} else if do.r_auth_type == string(workspace.RemoteAuthType_Password) {
			remoteInfo.AuthType = workspace.RemoteAuthType_Password
		} else {
			panic(do.r_auth_type + i18nInstance.Common.Err_enum_error) // 不能被识别
		}

		if int(do.r_port.Int32) > 0 {
			remoteInfo.SSHPort = int(do.r_port.Int32)
		} else {
			remoteInfo.SSHPort = 22
		}

		if do.r_password.Valid && len(do.r_password.String) > 0 {
			passwordDecrypt := aes4go.Decrypt(do.r_password.String, aesDecryptKey)
			remoteInfo.Password = passwordDecrypt
		}

		remoteInfo.CreatedTime = do.r_created
		//remote.CreatedTime = r_id

	default:
		panic(err)
	}

	return
}
