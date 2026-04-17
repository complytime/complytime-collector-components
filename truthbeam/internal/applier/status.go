package applier

import gemara "github.com/gemaraproj/go-gemara"

// Status defines the compliance status value
type Status int

const (
	// Unknown is the default status when a status is not explicit defined.
	Unknown Status = iota
	// Compliant defines then status when a  resource in compliant
	Compliant
	// NotCompliant define the status when a resource is not compliant.
	NotCompliant
	// NotApplicable define the status when a resource does not fall into any applicability category.
	NotApplicable
	// Exempt defines the status when a resource has an active compliance exception.
	Exempt
)

var toString = map[Status]string{
	Compliant:     "Compliant",
	NotCompliant:  "Non-Compliant",
	NotApplicable: "Not Applicable",
	Exempt:        "Exempt",
	Unknown:       "Unknown",
}

func (s Status) String() string {
	return toString[s]
}

func parseResult(resultStr string) gemara.Result {
	switch resultStr {
	case "Not Run":
		return gemara.NotRun
	case "Not Applicable":
		return gemara.NotApplicable
	case "Passed":
		return gemara.Passed
	case "Failed":
		return gemara.Failed
	default:
		return gemara.Unknown
	}
}

func mapResult(resultStr string) Status {
	result := parseResult(resultStr)
	switch result {
	case gemara.Passed:
		return Compliant
	case gemara.Failed:
		return NotCompliant
	case gemara.NotApplicable, gemara.NotRun:
		return NotApplicable
	default:
		return Unknown
	}
}
