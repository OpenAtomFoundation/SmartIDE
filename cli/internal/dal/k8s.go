/*
SmartIDE - CLI
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

	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

type K8sDO struct {
	k_id         int
	k_kubeconfg  sql.NullString
	k_context    string
	k_namespace  string
	k_deployment string
	k_pvc        string
	k_created    time.Time
	k_is_del     bool
}

func GetK8sInfoList() (k8sInfoList []workspace.K8sInfo, err error) {
	db := getDb()
	defer db.Close()

	rows, err := db.Query(`select k_id, k_kubeconfg, k_context, k_namespace, k_deployment,k_pvc,k_created 
							from k8s
							where k_is_del = 0`)
	for rows.Next() {
		do := K8sDO{}

		switch errSql := rows.Scan(&do.k_id, &do.k_kubeconfg, &do.k_context, &do.k_namespace, do.k_deployment, do.k_pvc, &do.k_created); errSql {
		/* case sql.ErrNoRows:
		   common.SmartIDELog.Warning() //TODO */
		case nil:

			k8sInfo := workspace.K8sInfo{}
			k8sInfo.KubeConfigFilePath = do.k_kubeconfg.String
			k8sInfo.Context = do.k_namespace

			k8sInfo.Namespace = do.k_namespace
			k8sInfo.DeploymentName = do.k_deployment
			k8sInfo.PVCName = do.k_pvc
			k8sInfo.CreatedTime = do.k_created

			k8sInfoList = append(k8sInfoList, k8sInfo)

		default:
			err = errSql
		}
		//   fmt.Println(uid, username, department, created)
	}

	return k8sInfoList, err
}

func GetK8sInfoById(id int) (K8sInfo *workspace.K8sInfo, err error) {
	return GetK8sInfo(id, "")
}

//

func RemoveK8s(id int, context string) error {
	db := getDb()
	defer db.Close()

	// 数据校验
	var count int
	var row *sql.Row
	if len(context) > 0 {
		row = db.QueryRow("select count(1) from k8s where k_context=? and k_is_del = 0", context)
	} else if id > 0 {
		row = db.QueryRow("select count(1) from k8s where k_id=? and k_is_del = 0", id)
	}
	switch err := row.Scan(&count); err {
	case sql.ErrNoRows:
		msg := fmt.Sprintf("k8s （%v | %v）", id, context)
		common.SmartIDELog.WarningF(i18nInstance.Common.Warn_dal_record_not_exit_condition, msg) // 没有查询到数据
	case nil:
		if count <= 0 {
			return errors.New(i18nInstance.Common.Warn_dal_record_not_exit)
		} else if count > 1 {
			return errors.New(i18nInstance.Common.Err_dal_record_repeat)
		}
	default:
		panic(err)
	}

	//
	var stmt *sql.Stmt
	var err error
	if len(context) > 0 {
		stmt, err = db.Prepare("update k8s set k_is_del=1 where k_context=? and k_is_del = 0")
	} else if id > 0 {
		stmt, err = db.Prepare("update k8s set k_is_del=1 where k_id=? and k_is_del = 0")
	}
	if err != nil {
		return err
	}
	res, err := stmt.Exec(id, context)
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

func InsertOrUpdateK8sInfo(K8sInfo workspace.K8sInfo) (id int, err error) {

	var single *workspace.K8sInfo

	if K8sInfo.ID > 0 {
		single, err = GetK8sInfoById(K8sInfo.ID)
		if err != nil {
			return id, err
		}
	} else {
		single, err = GetK8sInfo(0, K8sInfo.Context)
		if err != nil {
			return id, err
		}
	}

	//1. init
	db := getDb()
	defer db.Close()

	//2. insert or update
	if single != nil { //2.1. update
		stmt, err := db.Prepare(`update k8s set
		 k_kubeconfig=?, k_context=?, k_namespace=?,k_deployment=? ,k_pvc=?  
		where k_id=? or k_context=?`)
		if err != nil {
			return id, err
		}
		_, err = stmt.Exec(K8sInfo.KubeConfigFilePath, K8sInfo.Context, K8sInfo.Namespace,
			K8sInfo.DeploymentName, K8sInfo.PVCName,
			K8sInfo.ID, K8sInfo.Context)
		if err != nil {
			return -1, err
		}

		id = single.ID

	} else { //2.2. insert
		stmt, err := db.Prepare(`INSERT INTO k8s(k_kubeconfig, k_context, k_namespace, k_deployment, k_pvc)  
                                        values(?, ?, ?, ?, ?)`)
		if err != nil {
			return id, err
		}
		res, err := stmt.Exec(K8sInfo.KubeConfigFilePath, K8sInfo.Context, K8sInfo.Namespace,
			K8sInfo.DeploymentName, K8sInfo.PVCName)
		if err != nil {
			return id, err
		}
		id64, err := res.LastInsertId()
		id = int(id64)
		return id, err
	}

	return id, err
}

func GetK8sInfo(id int, context string) (K8sInfo *workspace.K8sInfo, err error) {

	db := getDb()
	defer db.Close()

	do := K8sDO{}

	var row *sql.Row
	if len(context) > 0 {
		row = db.QueryRow("select k_id, k_kubeconfig, k_context, k_namespace,k_deployment,k_pvc, k_created from k8s where k_context=? and k_is_del = 0", context)
	} else if id > 0 {
		row = db.QueryRow("select k_id, k_kubeconfig, k_context, k_namespace,k_deployment,k_pvc ,k_created from k8s where k_id=? and k_is_del = 0", id)
	} else {
		return K8sInfo, nil
	}

	switch err := row.Scan(&do.k_id, &do.k_kubeconfg, &do.k_context, &do.k_namespace, &do.k_deployment, &do.k_pvc, &do.k_created); err {
	case sql.ErrNoRows:
		msg := fmt.Sprintf("deployment (%v | %v)", context, id)
		common.SmartIDELog.WarningF(i18nInstance.Common.Warn_dal_record_not_exit_condition, msg) // 不存在
	case nil:
		tmp := workspace.K8sInfo{}
		K8sInfo = &tmp
		K8sInfo.ID = do.k_id
		K8sInfo.KubeConfigFilePath = do.k_kubeconfg.String
		K8sInfo.Context = do.k_context
		K8sInfo.Namespace = do.k_namespace
		K8sInfo.DeploymentName = do.k_deployment
		K8sInfo.PVCName = do.k_pvc
		K8sInfo.CreatedTime = do.k_created

	default:
		panic(err)
	}

	return K8sInfo, err
}
