package mwemu

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// handleAPI is the main router for MediaWiki API requests
func (e *Emulator) handleAPI(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		e.sendError(w, "invalid-request", "Failed to parse request: "+err.Error())
		return
	}

	action := r.FormValue("action")

	// Convert form values to map for recording
	params := make(map[string]string)
	for key, values := range r.Form {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// Record API call
	e.state.RecordAPICall(action, params)

	// Route to appropriate handler
	switch action {
	case "query":
		e.handleQuery(w, r)
	case "revisiondelete":
		e.handleRevisionDelete(w, r)
	default:
		e.sendError(w, "unknown-action", "Unrecognized value for parameter 'action': "+action)
	}
}

// handleQuery routes query requests based on meta/list/prop parameters
func (e *Emulator) handleQuery(w http.ResponseWriter, r *http.Request) {
	meta := r.FormValue("meta")
	list := r.FormValue("list")
	prop := r.FormValue("prop")

	if meta == "userinfo" {
		e.handleUserInfo(w, r)
	} else if meta == "tokens" {
		e.handleTokens(w, r)
	} else if list == "recentchanges" {
		e.handleRecentChanges(w, r)
	} else if prop == "revisions" {
		e.handleRevisions(w, r)
	} else {
		e.sendError(w, "unknown-query", "Unrecognized query type")
	}
}

// handleUserInfo handles action=query&meta=userinfo
func (e *Emulator) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	uiprop := r.FormValue("uiprop")

	userinfo := map[string]interface{}{
		"id":   e.config.User.UserID,
		"name": e.config.User.Name,
	}

	// Add rights if requested
	if strings.Contains(uiprop, "rights") {
		userinfo["rights"] = e.config.User.Rights
	}

	e.sendJSON(w, map[string]interface{}{
		"query": map[string]interface{}{
			"userinfo": userinfo,
		},
	})
}

// handleTokens handles action=query&meta=tokens
func (e *Emulator) handleTokens(w http.ResponseWriter, r *http.Request) {
	tokenType := r.FormValue("type")

	// Validate token type
	if tokenType != "csrf" && tokenType != "" {
		e.sendError(w, "unknown-token-type", "Unrecognized token type: "+tokenType)
		return
	}

	e.sendJSON(w, map[string]interface{}{
		"query": map[string]interface{}{
			"tokens": map[string]interface{}{
				"csrftoken": e.state.GetCSRFToken(),
			},
		},
	})
}

// handleRecentChanges handles action=query&list=recentchanges
func (e *Emulator) handleRecentChanges(w http.ResponseWriter, r *http.Request) {
	rcstart := r.FormValue("rcstart")
	rclimit := r.FormValue("rclimit")
	rcdir := r.FormValue("rcdir")
	rctype := r.FormValue("rctype")
	_ = r.FormValue("rcshow") // rcshow filtering not implemented yet

	// Validate rcdir
	if rcdir != "" && rcdir != "newer" && rcdir != "older" {
		e.sendError(w, "invalid-parameter", "Invalid value for rcdir: "+rcdir)
		return
	}

	// Validate rctype
	if rctype != "" && rctype != "edit" && rctype != "new" && rctype != "log" {
		e.sendError(w, "invalid-parameter", "Invalid value for rctype: "+rctype)
		return
	}

	// Get recent changes from state
	changes := e.state.GetRecentChanges(rcstart, rclimit)

	// Filter by rctype if specified
	if rctype != "" {
		filtered := []interface{}{}
		for _, change := range changes {
			if changeMap, ok := change.(map[string]interface{}); ok {
				if changeType, ok := changeMap["type"].(string); ok && changeType == rctype {
					filtered = append(filtered, change)
				}
			}
		}
		changes = filtered
	}

	// Filter by rcshow if specified (e.g., "!bot" means not bots)
	// For simplicity, we'll just ignore this for now since our test data doesn't track bot status

	e.sendJSON(w, map[string]interface{}{
		"query": map[string]interface{}{
			"recentchanges": changes,
		},
	})
}

// handleRevisions handles action=query&prop=revisions
func (e *Emulator) handleRevisions(w http.ResponseWriter, r *http.Request) {
	titles := r.FormValue("titles")
	rvprop := r.FormValue("rvprop")
	rvlimitStr := r.FormValue("rvlimit")

	// Validate titles parameter
	if titles == "" {
		e.sendError(w, "missingparam", "The 'titles' parameter must be set")
		return
	}

	// Parse rvlimit
	rvlimit := 5000
	if rvlimitStr != "" {
		if limit, err := strconv.Atoi(rvlimitStr); err == nil {
			rvlimit = limit
		}
	}

	// Get page from state
	page := e.state.GetPageByTitle(titles)
	if page == nil {
		// Page doesn't exist
		e.sendJSON(w, map[string]interface{}{
			"query": map[string]interface{}{
				"pages": map[string]interface{}{
					"-1": map[string]interface{}{
						"title":   titles,
						"missing": "",
					},
				},
			},
		})
		return
	}

	// Build revisions array
	revisions := []map[string]interface{}{}
	for i, rev := range page.Revisions {
		if i >= rvlimit {
			break
		}

		revData := map[string]interface{}{
			"revid":    rev.RevID,
			"parentid": rev.ParentID,
		}

		// Add properties based on rvprop
		if strings.Contains(rvprop, "ids") || rvprop == "" {
			// ids already included
		}
		if strings.Contains(rvprop, "timestamp") {
			revData["timestamp"] = rev.Timestamp.UTC().Format("2006-01-02T15:04:05Z")
		}
		if strings.Contains(rvprop, "user") {
			if rev.UserHidden {
				revData["userhidden"] = ""
			} else {
				revData["user"] = rev.User
			}
		}
		if strings.Contains(rvprop, "comment") {
			if rev.CommentHidden {
				revData["commenthidden"] = ""
			} else {
				revData["comment"] = rev.Comment
			}
		}
		if strings.Contains(rvprop, "content") {
			revData["content"] = rev.Content
			revData["*"] = rev.Content // MediaWiki uses * for content
		}

		// Add suppressed marker if revision is suppressed
		if rev.Suppressed {
			revData["suppressed"] = ""
		}

		revisions = append(revisions, revData)
	}

	// Special handling for suppression list page
	if titles == e.config.ListPageName {
		// Generate content based on current suppression list
		listContent := strings.Join(e.state.GetSuppressionList(), "\n")

		// If requesting content, include it
		if strings.Contains(rvprop, "content") {
			if len(revisions) > 0 {
				revisions[0]["content"] = listContent
				revisions[0]["*"] = listContent
			}
		}
	}

	e.sendJSON(w, map[string]interface{}{
		"query": map[string]interface{}{
			"pages": map[string]interface{}{
				fmt.Sprint(page.PageID): map[string]interface{}{
					"pageid":    page.PageID,
					"ns":        page.Namespace,
					"title":     page.Title,
					"revisions": revisions,
				},
			},
		},
	})
}

// handleRevisionDelete handles action=revisiondelete
func (e *Emulator) handleRevisionDelete(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	revType := r.FormValue("type")
	ids := r.FormValue("ids")
	hide := r.FormValue("hide")
	suppress := r.FormValue("suppress")

	// Validate token
	if token != e.state.GetCSRFToken() {
		e.sendError(w, "badtoken", "Invalid token")
		return
	}

	// Validate type
	if revType != "revision" {
		e.sendError(w, "invalid-parameter", "Invalid value for type parameter")
		return
	}

	// Validate ids
	if ids == "" {
		e.sendError(w, "missingparam", "The 'ids' parameter must be set")
		return
	}

	// Validate hide parameter
	if hide != "user|comment" && hide != "comment|user" {
		e.sendError(w, "invalid-parameter", "Invalid value for hide parameter")
		return
	}

	// Validate suppress parameter
	if suppress != "yes" && suppress != "" {
		e.sendError(w, "invalid-parameter", "Invalid value for suppress parameter")
		return
	}

	// Parse revision IDs (pipe-separated or comma-separated)
	idsStr := strings.ReplaceAll(ids, "|", ",")
	idList := strings.Split(idsStr, ",")

	suppressedCount := 0
	for _, idStr := range idList {
		idStr = strings.TrimSpace(idStr)
		revID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			e.sendError(w, "invalid-parameter", "Invalid revision ID: "+idStr)
			return
		}

		// Suppress the revision
		if e.state.SuppressRevision(revID) {
			suppressedCount++
		}
	}

	e.t.Logf("Suppressed %d revisions: %s", suppressedCount, ids)

	e.sendJSON(w, map[string]interface{}{
		"revisiondelete": map[string]interface{}{
			"status": "Success",
			"items": []map[string]interface{}{
				{
					"id":            ids,
					"userhidden":    true,
					"commenthidden": true,
				},
			},
		},
	})
}
