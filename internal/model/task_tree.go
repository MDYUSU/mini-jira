package model

type TaskTreeNode struct {
	Task
	Depth    int             `json:"depth"`
	Children []*TaskTreeNode `json:"children"`
}