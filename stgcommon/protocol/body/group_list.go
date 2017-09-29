package body

import (
	set "github.com/deckarep/golang-set"
)

// GroupList分组集合
// Author rongzhihong
// Since 2017/9/19
type GroupList struct {
	GroupList set.Set `json:"groupList"`
}

func NewGroupList() *GroupList {
	groupList := new(GroupList)
	groupList.GroupList = set.NewSet()
	return groupList
}
