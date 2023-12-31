// Package report generates reports on the messages in a practice session.
package report

import (
	"time"

	"github.com/rothskeller/wppsvr/store"
)

// Store is an interface covering those methods of store.Store that are used in
// generating reports.
type Store interface {
	GetSessionMessages(int) []*store.Message
	GetSessions(start, end time.Time) []*store.Session
	UpdateSession(*store.Session)
	NextMessageID(string) string
}

// A Report contains all of the information that goes into a report about a
// practice session.  (This can include information from multiple sessions when
// a weekly summary is part of the report.)
type Report struct {
	SessionName         string
	SessionDate         string
	Preliminary         bool
	MessageTypes        []string
	HasModel            bool
	SentTo              string
	SentBefore          string
	SentAfter           string
	NotSentFrom         string
	Modified            bool
	ValidCount          int
	InvalidCount        int
	ReplacedCount       int
	DroppedCount        int
	AverageValidScore   int
	uniqueCallSigns     map[string]struct{}
	UniqueCallSigns     int
	UniqueCallSignsWeek int
	Sources             []*Source
	Jurisdictions       []*Count
	MTypeCounts         []*Count
	Messages            []*Message
	Participants        []string
	GenerationInfo      string
}

// A Source contains the information about a single source of messages in a
// Report.
type Source struct {
	Name          string
	Count         int
	SimulatedDown bool
}

// A Count contains a name/count pair.
type Count struct {
	Name  string
	Count int
}

// A Message contains the information about a single message in a Report.
type Message struct {
	ID           string
	Hash         string
	FromCallSign string
	Prefix       string
	Suffix       string
	Source       string
	Multiple     bool
	Jurisdiction string
	Score        int
	Summary      string
}
