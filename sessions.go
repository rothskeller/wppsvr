package main

import (
	"log"
	"time"

	"github.com/rothskeller/wppsvr/report"
	"github.com/rothskeller/wppsvr/retrieve"
	"github.com/rothskeller/wppsvr/store"
)

// closeSessions closes any sessions that are past their end time and sends
// reports for them.
func closeSessions(st *store.Store) {
	var now = time.Now()

	for _, session := range st.GetRunningSessions() {
		if session.End.Before(now) {
			session.Flags &^= store.Running
			st.UpdateSession(session)
			log.Printf("Closed session for %s ending %s.", session.Name, session.End.Format("2006-01-02 15:04"))
			if len(session.ReportToText) != 0 || len(session.ReportToHTML) != 0 || st.SessionHasMessages(session.ID) {
				var conn = retrieve.ConnectToBBS(session.ToBBSes[0], session.CallSign)
				report.Send(st, conn, session)
				conn.Close()
			}
		}
	}
}

// openSessions marks as "running" any sessions that encompass the current time
// and are not already running.
func openSessions(st *store.Store) {
	// Sessions generally run for a week, so it suffices to look for an end
	// date between now and a week from now.  However, for safety's sake,
	// we'll make it a month.
	start := time.Now()
	end := start.AddDate(0, 1, 0)
	for _, session := range st.GetSessions(start, end) {
		if session.Flags&store.Running != 0 || session.Start.After(start) {
			continue
		}
		// We found a session that should be started.
		session.Flags |= store.Running
		st.UpdateSession(session)
		log.Printf("Opened session for %s ending %s.", session.Name, session.End.Format("2006-01-02 15:04"))
	}
}
