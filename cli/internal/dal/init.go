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
	"log"
	"os"

	"github.com/leansoftX/smartide-cli/pkg/common"
)

func getDb() *sql.DB {
	dbInit()

	db, err := connection()
	common.CheckError(err)

	return db
}

// 创建数据库表
func createDataTables() {
	sql_table := `
CREATE TABLE IF NOT EXISTS "remote" (
   "r_id" INTEGER PRIMARY KEY AUTOINCREMENT,
   "r_addr" VARCHAR(256) NULL,
   "r_port" int default (22) NOT NULL,
   "r_username" VARCHAR(100) NULL,
   "r_auth_type" VARCHAR(25) NULL,
   "r_password" VARCHAR(100) NULL,
   "r_json" TEXT NULL,
   "r_is_del" BIT default (0),
   "r_created" TIMESTAMP default (datetime('now', 'localtime')) 
);
CREATE TABLE IF NOT EXISTS "workspace" (
	"w_id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"w_name" VARCHAR(256) NULL,
	"w_workingdir" VARCHAR(256) NULL,
	"w_config_file" VARCHAR(256) NULL,
	"w_docker_compose_file_path" VARCHAR(256) NULL,
	"w_mode" VARCHAR(10) NULL,
    "w_git_clone_repo_url" VARCHAR(200) NULL,
    "w_git_auth_type" VARCHAR(10) NULL,
	"w_git_username" VARCHAR(100) NULL,
	"w_git_password" VARCHAR(60) NULL,
    "w_git_auth_pat" VARCHAR(10) NULL,

    "w_branch" VARCHAR(50) NULL,
	"w_json" TEXT NULL,
	"w_config_content" text NULL,
	"w_link_compose_content" text NULL,
	"w_temp_compose_content" text NULL,

	"r_id" INTEGER NULL,
	"w_is_del" BIT default (0),
	"w_created" TIMESTAMP default (datetime('now', 'localtime')),
	FOREIGN KEY (r_id) REFERENCES remote(r_id)
 );
 CREATE TABLE IF NOT EXISTS "k8s" (
	"k_id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"k_kubeconfig" VARCHAR(500) NULL,
	"k_context" VARCHAR(50) NULL,
	"k_namespace" VARCHAR(50) NULL,
	"k_deployment" VARCHAR(50) NULL,
	"k_pvc" VARCHAR(50) NULL,
	"k_is_del" BIT default (0),
	"k_created" TIMESTAMP default (datetime('now', 'localtime')) 
 );

 `

	db, err := connection()
	if err != nil {
		db.Close()
		common.CheckError(err)
	}
	defer db.Close()

	_, err = db.Exec(sql_table)
	common.CheckError(err)

	// 新增的列
	db.Exec("ALTER TABLE remote ADD r_port int default (22);")
	db.Exec("ALTER TABLE remote ADD COLUMN r_json text;")

	db.Exec("ALTER TABLE workspace ADD COLUMN w_json text;")
	db.Exec("ALTER TABLE workspace ADD COLUMN w_config_file VARCHAR(256) NULL;")
	db.Exec("ALTER TABLE workspace ADD COLUMN w_config_content text NULL;")
	db.Exec("ALTER TABLE workspace ADD COLUMN w_link_compose_content text NULL;")
	db.Exec("ALTER TABLE workspace ADD COLUMN w_temp_compose_content text NULL;")
	db.Exec("ALTER TABLE workspace ADD COLUMN k_id INTEGER NULL;")
	db.Exec("ALTER TABLE workspace ADD COLUMN w_git_username VARCHAR(100) NULL;")
	db.Exec("ALTER TABLE workspace ADD COLUMN w_git_password VARCHAR(60) NULL;")

	db.Exec("ALTER TABLE k8s ADD COLUMN k_kubeconfig VARCHAR(500) NULL;")
}

// sqlite 数据库文件所在路径
var SqliteFilePath string = ".ide/.ide.db"

var isInit bool = false

func dbInit() {
	if !isInit {

		createDataTables()
		isInit = true

	}
}

func connection() (*sql.DB, error) {

	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	sqliteFilePath := common.PathJoin(dirname, SqliteFilePath)

	if !common.IsExist(sqliteFilePath) {
		os.MkdirAll(common.PathJoin(dirname, ".ide"), os.ModePerm) // create dir
		os.Create(sqliteFilePath)
	}

	db, err := sql.Open("sqlite3", sqliteFilePath)
	//defer db.Close()

	return db, err
}

/*

//插入数据
    fmt.Print("插入数据, ID=")
    stmt, err := db.Prepare("INSERT INTO userinfo(username, departname)  values(?, ?)")
    checkErr(err)
    res, err := stmt.Exec("astaxie", "研发部门")
    checkErr(err)
    id, err := res.LastInsertId()
    checkErr(err)
    fmt.Println(id)

    //更新数据
    fmt.Print("更新数据 ")
    stmt, err = db.Prepare("update userinfo set username=? where uid=?")
    checkErr(err)
    res, err = stmt.Exec("astaxieupdate", id)
    checkErr(err)
    affect, err := res.RowsAffected()
    checkErr(err)
    fmt.Println(affect)

    //查询数据
    fmt.Println("查询数据")
    rows, err := db.Query("SELECT * FROM userinfo")
    checkErr(err)
    for rows.Next() {
        var uid int
        var username string
        var department string
        var created string
        err = rows.Scan(&uid, &username, &department, &created)
        checkErr(err)
        fmt.Println(uid, username, department, created)
    }

    //删除数据
    fmt.Println("删除数据")
    stmt, err = db.Prepare("delete from userinfo where uid=?")
    checkErr(err)
    res, err = stmt.Exec(id)
    checkErr(err)
    affect, err = res.RowsAffected()
    checkErr(err)
    fmt.Println(affect)


*/
