// Copyright (c) 2017-2019, AT&T Intellectual Property.
// All rights reserved.
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: LGPL-2.1-only

package session

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/danos/config/commit"
	"github.com/danos/config/data"
	"github.com/danos/config/schema"
	"github.com/danos/configd"
	"github.com/danos/mgmterror"
	"github.com/danos/utils/exec"
	spawn "os/exec"
)

type commitctx struct {
	debug     bool
	effective *Session
	candidate *data.Node
	running   *data.Node
	schema    schema.Node
	sid       string
	sctx      *configd.Context
	ctx       *configd.Context
	message   string
}

func newctx(sid string, sctx *configd.Context, effective *Session, candidate, running *data.Node, sch schema.Node, message string, debug bool) *commitctx {
	return &commitctx{
		debug:     debug,
		effective: effective,
		candidate: candidate,
		running:   running,
		schema:    sch,
		sid:       sid,
		sctx:      sctx,
		ctx: &configd.Context{
			Configd: true,
			Config:  sctx.Config,
			Pid:     int32(configd.SYSTEM),
			Auth:    sctx.Auth,
			Dlog:    sctx.Dlog,
			Elog:    sctx.Elog,
			Noexec:  true,
		},
		message: message,
	}
}

const (
	commitLogMsgPrefix = "COMMIT"
	padToLength        = 50
	// 50 + 3 extra just in case
	padding = "                                                     "
)

func pad(msg string) string {
	msgLen := len(msg)
	padLen := 0
	if msgLen < padToLength {
		padLen = padToLength - msgLen
	}
	return msg + ": " + padding[:padLen]
}

func (c *commitctx) LogCommitMsg(msg string) {
	c.sctx.Dlog.Println(fmt.Sprintf("%s: %s", commitLogMsgPrefix, msg))
}

func (c *commitctx) LogCommitTime(msg string, startTime time.Time) {
	c.sctx.Dlog.Println(fmt.Sprintf("%s: %s%s", commitLogMsgPrefix, pad(msg),
		time.Since(startTime).Round(time.Millisecond)))
}

func (c *commitctx) Log(msgs ...interface{}) {
	c.sctx.Dlog.Println(msgs...)
}

func (c *commitctx) LogError(msgs ...interface{}) {
	c.sctx.Elog.Println(msgs...)
}

func (c *commitctx) LogAudit(msg string) {
	c.ctx.Auth.AuditLog(msg)
}

func (c *commitctx) Debug() bool {
	return c.debug
}

func (c *commitctx) Sid() string {
	return c.sid
}

func (c *commitctx) Uid() uint32 {
	return c.sctx.Uid
}

func (c *commitctx) Running() *data.Node {
	return c.running
}

func (c *commitctx) Candidate() *data.Node {
	return c.candidate
}

func (c *commitctx) Schema() schema.Node {
	return c.schema
}

func (c *commitctx) RunDeferred() bool {
	return false
}

func (c *commitctx) Effective() commit.EffectiveDatabase {
	return c
}

func (c *commitctx) Set(path []string) error {
	return c.effective.Set(c.ctx, path)
}

func (c *commitctx) Delete(path []string) error {
	return c.effective.Delete(c.ctx, path)
}

func (c *commitctx) validate() ([]*exec.Output, []error, bool) {
	return commit.Validate(c)
}

// Original implementation ignores the result of the hooks
func (c *commitctx) execute_hooks(hookdir string, env []string) (*exec.Output, error) {
	out := new(bytes.Buffer)
	err := new(bytes.Buffer)
	cmd := spawn.Command("/bin/run-parts", "--regex=^[a-zA-Z0-9._-]+$", "--", hookdir)
	cmd.Stdout = out
	cmd.Stderr = err
	if env != nil {
		cmd.Env = append(cmd.Env, env...)
	}

	c.sctx.Dlog.Printf("Executing %s hooks\n", hookdir)
	if cmd.Run() != nil {
		cerr := mgmterror.NewOperationFailedApplicationError()
		cerr.Message = err.String()
		return &exec.Output{Output: out.String()}, cerr
	}
	return &exec.Output{Output: out.String()}, nil
}

func (c *commitctx) send_notify() {
	pid := strconv.FormatInt(int64(c.sctx.Pid), 10)
	uid := strconv.FormatUint(uint64(c.sctx.Uid), 10)
	spawn.Command("/opt/vyatta/sbin/vyatta-cfg-notify", uid, pid).Run()
}

func (c *commitctx) commit(env *[]string) ([]*exec.Output, []error, bool) {
	outs, errs, successes, failures := commit.Commit(c)

	if successes > 0 {
		c.send_notify()
	}

	status := "COMMIT_STATUS="
	if failures > 0 {
		if successes > 0 {
			status += "PARTIAL"
		} else {
			status += "FAILURE"
		}
	} else {
		status += "SUCCESS"
	}
	*env = append(*env, status)
	return outs, errs, failures == 0
}