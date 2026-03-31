package mcpserver

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/yhy0/ChYing/mitmproxy"
	"github.com/yhy0/ChYing/pkg/db"
)

// --- get_traffic_detail ---

func getTrafficDetailTool() mcp.Tool {
	return mcp.NewTool("get_traffic_detail",
		mcp.WithDescription("Get the full HTTP request and response raw data for a specific traffic entry. Provide either 'hid' or 'id'. Use 'hid' from get_http_history/get_traffic_by_host results, or 'id' from query_by_dsl results."),
		mcp.WithNumber("hid",
			mcp.Description("The Hid (history ID) of the traffic entry. Available from get_http_history and get_traffic_by_host results."),
		),
		mcp.WithNumber("id",
			mcp.Description("The database ID of the traffic entry. Available from query_by_dsl results."),
		),
	)
}

func handleGetTrafficDetail(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	hid := req.GetInt("hid", 0)
	id := req.GetInt("id", 0)

	if hid == 0 && id == 0 {
		return errorResult("either 'hid' or 'id' is required"), nil
	}

	// 如果提供了 id 而非 hid，先通过 id 查找 hid
	if hid == 0 && id > 0 {
		history, err := db.GetHistoryByID(int64(id))
		if err != nil || history == nil {
			return errorResult("traffic not found for id: %d", id), nil
		}
		hid = int(history.Hid)
	}

	request, response := db.GetTraffic(hid)
	if request == nil {
		return errorResult("traffic not found for hid: %d", hid), nil
	}

	result := fmt.Sprintf("=== REQUEST ===\n%s", request.RequestRaw)
	if response != nil {
		result += fmt.Sprintf("\n\n=== RESPONSE ===\n%s", response.ResponseRaw)
	}

	return textResult(result), nil
}

// --- query_by_dsl ---

func queryByDSLTool() mcp.Tool {
	return mcp.NewTool("query_by_dsl",
		mcp.WithDescription(`Query HTTP traffic history using DSL expressions.

Available fields: id, url, path, method, host, status, length, content_type, timestamp, request, request_body, response, response_body, request_headers, response_headers, status_reason

Available functions and operators: contains(), regex(), ==, !=, &&, ||

Examples:
- status == 200 && contains(response_body, "admin")
- contains(host, "api.example.com") && method == "POST"
- regex(path, "/api/v[0-9]+/users")
- status != 200 && contains(request_headers, "Authorization")`),
		mcp.WithString("dsl_query",
			mcp.Required(),
			mcp.Description("The DSL expression to query traffic"),
		),
	)
}

func handleQueryByDSL(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dslQuery, err := req.RequireString("dsl_query")
	if err != nil {
		return errorResult("dsl_query is required"), nil
	}

	results, queryErr := mitmproxy.QueryHistoryByDSL(dslQuery)
	if queryErr != nil {
		return errorResult("DSL query failed: %v", queryErr), nil
	}

	items := make([]historyItem, 0, len(results))
	for _, h := range results {
		items = append(items, historyItem{
			ID:        h.Id,
			Host:      h.Host,
			Method:    h.Method,
			FullUrl:   h.FullUrl,
			Path:      h.Path,
			Status:    h.Status,
			Length:    h.Length,
			MIMEType:  h.MIMEType,
			Extension: h.Extension,
			Title:     h.Title,
			IP:        h.IP,
		})
	}

	return jsonResult(items), nil
}
