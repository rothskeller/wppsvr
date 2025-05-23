// Package retrieve handles connecting to BBSes, retrieving messages from them,
// and sending the responses to them.
package retrieve

import (
	"log"
	"time"

	"github.com/rothskeller/packet/jnos"
	"github.com/rothskeller/packet/jnos/kpc3plus"
	"github.com/rothskeller/packet/jnos/telnet"
	"github.com/rothskeller/wppsvr/analyze"
	"github.com/rothskeller/wppsvr/config"
	"github.com/rothskeller/wppsvr/store"
)

// ForRunningSessions retrieves and responds to new messages in all running
// practice sessions.
func ForRunningSessions(st *store.Store) {
	now := time.Now()

	for _, session := range st.GetRunningSessions() {
		for _, ret := range session.Retrieve {
			point := now
			if point.Equal(point.Truncate(time.Minute)) {
				point = point.Add(-time.Minute)
			} else {
				point = point.Truncate(time.Minute)
			}
			for point.After(ret.LastRun) && !session.RetrieveInterval.Match(point) {
				point = point.Add(-time.Minute)
			}
			if point.After(ret.LastRun) {
				checkBBS(st, session, ret)
			}
		}
	}
}

// ForSession retrieves and responds to new messages in the specified practice
// session.
func ForSession(st *store.Store, session *store.Session) {
	now := time.Now()

	for _, ret := range session.Retrieve {
		point := now
		if point.Equal(point.Truncate(time.Minute)) {
			point = point.Add(-time.Minute)
		} else {
			point = point.Truncate(time.Minute)
		}
		for point.After(ret.LastRun) && !session.RetrieveInterval.Match(point) {
			point = point.Add(-time.Minute)
		}
		if point.After(ret.LastRun) {
			checkBBS(st, session, ret)
		}
	}
}

// checkBBS retrieves and responds to new check-in messages on a specific BBS.
func checkBBS(st *store.Store, session *store.Session, retrieval *store.Retrieval) {
	var (
		conn   *jnos.Conn
		err    error
		msgnum = 1
		start  = time.Now()
	)
	if conn = ConnectToBBS(retrieval.BBS, session.CallSign); conn == nil {
		return
	}
	defer func() {
		if err = conn.Close(); err != nil {
			log.Printf("ERROR: closing connection to %s@%s: %s", session.CallSign, retrieval.BBS, err)
		}
	}()
	for {
		var message string

		if message, err = conn.Read(msgnum); err != nil {
			log.Printf("ERROR: reading messages to %s@%s: %s", session.CallSign, retrieval.BBS, err)
			return
		} else if message == "" { // no more messages
			break
		}
		handleMessage(st, conn, session, retrieval, message, msgnum)
		msgnum++
	}
	retrieval.LastRun = start
	st.UpdateSession(session)
}

// ConnectToBBS connects to the specified mailbox on the specified BBS, in the
// manner dictated by the BBS configuration.
func ConnectToBBS(bbsname, mailbox string) (conn *jnos.Conn) {
	// This function is exported because it is also used by
	// wppsvr/sessions.go to connect to the BBS to send end-of-session
	// reports.
	var (
		bbs *config.BBSConfig
		err error
	)
	bbs = config.Get().BBSes[bbsname]
	switch bbs.Transport {
	case "disable":
		log.Printf("ERROR: can't connect to %s@%s: connections to %s are disabled", mailbox, bbsname, bbsname)
		return nil
	case "kpc3plus":
		conn, err = kpc3plus.Connect("/dev/tty.usbserial-1410", bbs.AX25, mailbox, "KC6RSC", nil)
	case "telnet":
		conn, err = telnet.Connect(bbs.TCP, mailbox, bbs.Passwords[mailbox], nil)
	}
	if err != nil {
		log.Printf("ERROR: can't connect to %s@%s via %s: %s", mailbox, bbsname, bbs.Transport, err)
		return nil
	}
	return conn
}

// handleMessage handles a single incoming message.
func handleMessage(st *store.Store, conn *jnos.Conn, session *store.Session, retrieval *store.Retrieval, message string, msgnum int) {
	var (
		analysis  *analyze.Analysis
		responses []*store.Response
		err       error
	)
	analysis = analyze.Analyze(st, session, retrieval.BBS, message)
	responses = analysis.Responses(st)
	for _, response := range responses {
		if session.Flags&store.DontSendResponses == 0 {
			if err = conn.Send(response.Subject, response.Body, response.To); err != nil {
				log.Printf("ERROR: sending message from %s@%s: %s", session.CallSign, retrieval.BBS, err)
				return
			}
		}
		response.SendTime = time.Now()
	}
	if session.Flags&store.DontKillMessages == 0 {
		if err = conn.Kill(msgnum); err != nil {
			log.Printf("ERROR: killing message %d at %s@%s: %s", msgnum, session.CallSign, retrieval.BBS, err)
			return
		}
	}
	analysis.Commit(st)
	for _, response := range responses {
		st.SaveResponse(response)
	}
}
