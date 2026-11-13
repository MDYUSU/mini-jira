package service

import (
	"context"

	"github.com/MDYUSU/mini-jira/internal/model"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

func (s *TaskService) GetTree(ctx context.Context, rootID int64) (*model.TaskTreeNode, error) {
	tasks, err := s.tasks.ListSubtree(ctx, rootID)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, repository.ErrNotFound
	}

	nodes := make(map[int64]*model.TaskTreeNode, len(tasks))
	var root *model.TaskTreeNode

	for _, t := range tasks {
		node := &model.TaskTreeNode{Task: t, Children: []*model.TaskTreeNode{}}
		nodes[t.ID] = node

		if t.ID == rootID {
			root = node
			continue
		}

		parent := nodes[*t.ParentTaskID]
		node.Depth = parent.Depth + 1
		parent.Children = append(parent.Children, node)
	}

	return root, nil
}
