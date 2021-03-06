package controllers

import (
	"fmt"
	"strings"
	"yao/libs"
)

type Permission struct {
	Id int `db:"permission_id" json:"permission_id"`
	Name string `db:"permission_name" json:"permission_name"`
	Count int `db:"count" json:"count"`
}

func PMCreate(name string) (int64, error) {
	return libs.DBInsertGetId("insert into permissions values (null, ?, 0)", name)
}

func PMChangeName(id int, name string) error {
	_, err := libs.DBUpdateGetAffected("update permissions set permission_name=? where permission_id=?", name, id)
	return err
}

func PMDelete(id int) (int64, error) {
	res, err := libs.DBUpdateGetAffected("delete from permissions where permission_id=?", id)
	if err != nil { return res, err }
	if res == 0 { return res, nil }
	_, err = libs.DBUpdate("delete from user_permissions where permission_id=?", id)
	if err != nil { return 0, err }
	_, err = libs.DBUpdate("delete from problem_permissions where permission_id=?", id)
	if err != nil { return 0, err }
	return res, err
}

func PMQuery(bound, pagesize int, isleft bool) ([]Permission, bool, error) {
	pagesize += 1
	var p []Permission
	var err error
	if isleft {
		err = libs.DBSelectAll(&p, "select * from permissions where permission_id>=? order by permission_id limit ?", bound, pagesize)
	} else {
		err = libs.DBSelectAll(&p, "select * from permissions where permission_id<=? order by permission_id desc limit ?", bound, pagesize)
	}
	if err != nil {
		return nil, false, err
	} else {
		isfull := len(p) == pagesize
		if isfull { p = p[: pagesize - 1] }
		if !isleft { libs.Reverse(p) }
		return p, isfull, nil
	}
}

func PMQueryUser(id int) ([]User, error) {
	var users []User
	err := libs.DBSelectAll(&users, "select user_info.user_id, user_name, motto, rating from (user_info join user_permissions on user_info.user_id=user_permissions.user_id) where permission_id=?", id)
	return users, err
}

func PMAddUser(ids []int, id int) (int64, error) {
	query := strings.Builder{}
	for i, j := range ids {
		query.WriteString(fmt.Sprintf("(%d,%d)", j, id))
		if i + 1 < len(ids) {
			query.WriteString(",")
		}
	}
	res, err := libs.DBUpdateGetAffected("insert ignore into user_permissions values " + query.String())
	if err != nil {
		return res, err
	} else {
		_, err = libs.DBUpdate("update permissions set count = count + ? where permission_id=?", res, id)
		return res, err
	}
}

func PMDeleteUser(pid, uid int) (int64, error) {
	res, err := libs.DBUpdateGetAffected("delete from user_permissions where user_id=? and permission_id=?", uid, pid)
	if err != nil { return 0, err }
	if res == 1 {
		libs.DBUpdate("update permissions set count = count - 1 where permission_id=?", pid)
	}
	return res, err
}

func PMExists(permission_id int) bool {
	count, _ := libs.DBSelectSingleInt("select count(*) from permissions where permission_id=?", permission_id)
	return count > 0
}