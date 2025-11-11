package app

import (
	"github.com/google/uuid"
	"strings"
)

type CommentNode struct {
	Comment
	Children []CommentNode
}

func BuildTree(comments []Comment, parentID *uuid.UUID) []CommentNode {
	var result []CommentNode
	for _, c := range comments {
		if (c.ParentID == nil && parentID == nil) || (c.ParentID != nil && parentID != nil && *c.ParentID == *parentID) {
			node := CommentNode{
				Comment:  c,
				Children: BuildTree(comments, &c.ID),
			}
			result = append(result, node)
		}
	}
	return result
}

func FilterTreeByText(nodes []CommentNode, text string) []CommentNode {
	var result []CommentNode
	for _, n := range nodes {
		filteredChildren := FilterTreeByText(n.Children, text)
		if strings.Contains(strings.ToLower(n.Text), strings.ToLower(text)) || len(filteredChildren) > 0 {
			result = append(result, CommentNode{
				Comment:  n.Comment,
				Children: filteredChildren,
			})
		}
	}
	return result
}
