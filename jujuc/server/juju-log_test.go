package server_test

import (
	"bytes"
	"fmt"
	"launchpad.net/gnuflag"
	. "launchpad.net/gocheck"
	"launchpad.net/juju-core/cmd"
	"launchpad.net/juju-core/cmd/jujuc/server"
	"launchpad.net/juju-core/log"
	stdlog "log"
)

type JujuLogSuite struct{}

var _ = Suite(&JujuLogSuite{})

func pushLog(debug bool) (buf *bytes.Buffer, pop func()) {
	oldTarget, oldDebug := log.Target, log.Debug
	buf = new(bytes.Buffer)
	log.Target, log.Debug = stdlog.New(buf, "", 0), debug
	return buf, func() {
		log.Target, log.Debug = oldTarget, oldDebug
	}
}

func dummyFlagSet() *gnuflag.FlagSet {
	return gnuflag.NewFlagSet("", gnuflag.ContinueOnError)
}

var commonLogTests = []struct {
	debugEnabled bool
	debugFlag    bool
	target       string
}{
	{false, false, "JUJU"},
	{false, true, ""},
	{true, false, "JUJU"},
	{true, true, "JUJU:DEBUG"},
}

func assertLogs(c *C, ctx *server.ClientContext, badge string) {
	msg1 := "the chickens"
	msg2 := "are 110% AWESOME"
	com, err := ctx.NewCommand("juju-log")
	c.Assert(err, IsNil)
	for _, t := range commonLogTests {
		buf, pop := pushLog(t.debugEnabled)
		defer pop()

		var args []string
		if t.debugFlag {
			args = []string{"--debug", msg1, msg2}
		} else {
			args = []string{msg1, msg2}
		}
		code := cmd.Main(com, &cmd.Context{}, args)
		c.Assert(code, Equals, 0)

		if t.target == "" {
			c.Assert(buf.String(), Equals, "")
		} else {
			expect := fmt.Sprintf("%s %s: %s %s\n", t.target, badge, msg1, msg2)
			c.Assert(buf.String(), Equals, expect)
		}
	}
}

func (s *JujuLogSuite) TestBadges(c *C) {
	local := &server.ClientContext{LocalUnitName: "minecraft/0"}
	assertLogs(c, local, "minecraft/0")
	relation := &server.ClientContext{LocalUnitName: "minecraft/0", RelationName: "bot"}
	assertLogs(c, relation, "minecraft/0 bot")
}

func (s *JujuLogSuite) TestRequiresMessage(c *C) {
	ctx := &server.ClientContext{}
	com, err := ctx.NewCommand("juju-log")
	c.Assert(err, IsNil)
	err = com.Init(dummyFlagSet(), nil)
	c.Assert(err, ErrorMatches, "no message specified")
}