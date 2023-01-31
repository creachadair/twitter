// Copyright (C) 2021 Michael J. Fromberger. All Rights Reserved.

package lists

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/jape"
)

// An Edit is a query to edit or delete a list or list membership.
type Edit struct {
	*jape.Request
	tag       string
	encodeErr error
}

// Invoke executes the query on the given context and client. A successful
// response reports whether the edit took effect.
func (e Edit) Invoke(ctx context.Context, cli *twitter.Client) (bool, error) {
	if e.encodeErr != nil {
		return false, e.encodeErr // deferred encoding error
	}
	rsp, err := cli.Call(ctx, e.Request)
	if err != nil {
		return false, err
	}
	m := make(map[string]*bool)
	if err := json.Unmarshal(rsp.Data, &m); err != nil {
		return false, &jape.Error{Data: rsp.Data, Message: "decoding response", Err: err}
	}
	if v := m[e.tag]; v != nil {
		return *v, nil
	}
	return false, fmt.Errorf("tag %q not found", e.tag)
}

// Delete constructs a query to delete an existing list.
//
// API: DELETE 2/lists/:id
func Delete(id string) Edit {
	req := &jape.Request{
		Method:     "2/lists/" + id,
		HTTPMethod: "DELETE",
	}
	return Edit{Request: req, tag: "deleted"}
}

// Update constructs a query to update an existing list.
//
// API: PUT 2/lists/:id
func Update(id string, opts UpdateOpts) Edit {
	req := &jape.Request{
		Method:     "2/lists/" + id,
		HTTPMethod: "PUT",
	}
	body, err := json.Marshal(opts)
	req.Data = body
	req.ContentType = "application/json"
	return Edit{Request: req, tag: "updated", encodeErr: err}
}

// AddMember constructs a query to add a member to an existing list.
//
// API: POST 2/lists/:id/members
func AddMember(listID, userID string) Edit {
	req := &jape.Request{
		Method:     "2/lists/" + listID + "/members",
		HTTPMethod: "POST",
	}
	body, err := json.Marshal(struct {
		U string `json:"user_id"`
	}{U: userID})
	req.Data = body
	req.ContentType = "application/json"
	return Edit{Request: req, tag: "is_member", encodeErr: err}
}

// RemoveMember constructs a query to remove a member from a list.
//
// API: DELETE 2/lists/:id/members/:userid
func RemoveMember(listID, userID string) Edit {
	req := &jape.Request{
		Method:     "2/lists/" + listID + "/members/" + userID,
		HTTPMethod: "DELETE",
	}
	return Edit{Request: req, tag: "is_member"}
}

// UpdateOpts provide parameters for list update queries.  The fields that are
// non-nil are modified to the given values. Fields that are nil are not
// changed from their existing settings.
type UpdateOpts struct {
	Name    *string `json:"name,omitempty"`
	Desc    *string `json:"description,omitempty"`
	Private *bool   `json:"private,omitempty"`
}

// SetName sets the option to update the name of the list.
func (u *UpdateOpts) SetName(name string) *UpdateOpts { u.Name = &name; return u }

// SetDescription sets the option to update the description of the list.
func (u *UpdateOpts) SetDescription(desc string) *UpdateOpts { u.Desc = &desc; return u }

// SetPrivate sets the option to mark a list private or non-private.
func (u *UpdateOpts) SetPrivate(private bool) *UpdateOpts { u.Private = &private; return u }
