package config

import (
	"github.com/rothskeller/packet/message"
)

// ComputeRecommendedHandlingOrder computes the recommended handling order for a
// message.  Only message types with computed (non-static) recommended handling
// orders are handled by this function.
func ComputeRecommendedHandlingOrder(msg message.Message) string {
	switch msg.Base().Type.Tag {
	case "ICS213":
		for _, f := range msg.Base().Fields {
			if f.Label == "Severity" {
				switch *f.Value {
				case "EMERGENCY":
					return "IMMEDIATE"
				case "URGENT":
					return "PRIORITY"
				case "OTHER":
					return "ROUTINE"
				}
				break
			}
		}
	case "EOC213RR":
		for _, f := range msg.Base().Fields {
			if f.Label == "Priority" {
				switch *f.Value {
				case "Now", "High":
					return "IMMEDIATE"
				case "Medium":
					return "PRIORITY"
				case "Low":
					return "ROUTINE"
				}
				break
			}
		}
	}
	return ""
}
