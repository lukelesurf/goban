package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerPrompts(s *server.MCPServer) {
	s.AddPrompt(mcp.NewPrompt("plan_project",
		mcp.WithPromptDescription("Create a kanban board for a project and break it into tasks"),
		mcp.WithArgument("project_name", mcp.ArgumentDescription("Name of the project"), mcp.RequiredArgument()),
		mcp.WithArgument("description", mcp.ArgumentDescription("Brief description of the project")),
	), func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		name := req.Params.Arguments["project_name"]
		desc := req.Params.Arguments["description"]
		return mcp.NewGetPromptResult(
			"Plan a project on the kanban board",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(
						"Create a kanban board for the project \""+name+"\".\n\n"+
							"Description: "+desc+"\n\n"+
							"Use the create_board tool to create the board, then use create_task to break the project into well-defined tasks. "+
							"Assign appropriate priorities and labels. Place all tasks in the 'To Do' column to start.",
					),
				),
			},
		), nil
	})

	s.AddPrompt(mcp.NewPrompt("resume_work",
		mcp.WithPromptDescription("Find the board, identify the next task, and continue work"),
		mcp.WithArgument("board_id", mcp.ArgumentDescription("Board ID to resume work on"), mcp.RequiredArgument()),
	), func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		boardID := req.Params.Arguments["board_id"]
		return mcp.NewGetPromptResult(
			"Resume work on a board",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(
						"Use get_board_summary to check the current state of board "+boardID+". "+
							"Then use list_tasks to find tasks in 'In Progress'. "+
							"If there are tasks in progress, continue working on the highest priority one. "+
							"If not, pick the next highest priority task from 'To Do' and move it to 'In Progress'. "+
							"Add a comment to the task explaining what you're doing.",
					),
				),
			},
		), nil
	})
}
