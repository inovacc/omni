package scan

import (
	"math"
	"strconv"
	"strings"
)

// rawSeverity is one OSV severity record: a CVSS type tag and either a numeric
// base score or a full CVSS vector string in the Score field.
type rawSeverity struct {
	Type  string `json:"type"`
	Score string `json:"score"`
}

// severityLabel normalizes an OSV severity[] slice to an ordered Severity.
//
// It prefers a CVSS_V4 record, then CVSS_V3. For the chosen record it first
// tries a producer-supplied numeric base score; failing that it computes a
// CVSS v3.1 base score from the vector metrics. CVSS v4.0 has no closed-form
// equation, so a v4 record contributes only its numeric score (band) or, when
// none is present, SeverityUnknown. When no usable severity exists the result
// is SeverityUnknown (always reported, below SeverityLow for gating).
func severityLabel(sevs []rawSeverity) Severity {
	s, ok := pickSeverity(sevs)
	if !ok {
		return SeverityUnknown
	}
	if f, err := strconv.ParseFloat(strings.TrimSpace(s.Score), 64); err == nil {
		return band(f)
	}
	if base, ok := cvssBaseScore(s.Type, s.Score); ok {
		return band(base)
	}
	return SeverityUnknown
}

// pickSeverity returns the first CVSS_V4 record, else the first CVSS_V3 record.
func pickSeverity(sevs []rawSeverity) (rawSeverity, bool) {
	for _, s := range sevs {
		if s.Type == "CVSS_V4" {
			return s, true
		}
	}
	for _, s := range sevs {
		if s.Type == "CVSS_V3" {
			return s, true
		}
	}
	return rawSeverity{}, false
}

// band maps a CVSS base score to the canonical CVSS v3.1/v4.0 qualitative band.
func band(score float64) Severity {
	switch {
	case score <= 0.0:
		return SeverityNone
	case score < 4.0:
		return SeverityLow
	case score < 7.0:
		return SeverityMedium
	case score < 9.0:
		return SeverityHigh
	default:
		return SeverityCritical
	}
}

// cvssBaseScore computes a CVSS base score from a vector string. Only CVSS v3.x
// (CVSS_V3) is computed via the published closed-form v3.1 formula; v4.0 has no
// closed-form equation and is not hand-rolled here. Any missing required metric
// yields (0, false).
func cvssBaseScore(typ, vector string) (float64, bool) {
	if typ != "CVSS_V3" {
		return 0, false
	}
	return cvss31BaseScore(vector)
}

// cvss31BaseScore implements the CVSS v3.1 base-score equation from the
// AV/AC/PR/UI/S/C/I/A metrics in vector. PR weighting depends on Scope.
func cvss31BaseScore(vector string) (float64, bool) {
	m := parseVector(vector)

	av, ok := metricWeight(cvss31AV, m["AV"])
	if !ok {
		return 0, false
	}
	ac, ok := metricWeight(cvss31AC, m["AC"])
	if !ok {
		return 0, false
	}
	ui, ok := metricWeight(cvss31UI, m["UI"])
	if !ok {
		return 0, false
	}
	scope, ok := m["S"]
	if !ok {
		return 0, false
	}
	changed := scope == "C"
	prTable := cvss31PRUnchanged
	if changed {
		prTable = cvss31PRChanged
	}
	pr, ok := metricWeight(prTable, m["PR"])
	if !ok {
		return 0, false
	}
	c, ok := metricWeight(cvss31CIA, m["C"])
	if !ok {
		return 0, false
	}
	i, ok := metricWeight(cvss31CIA, m["I"])
	if !ok {
		return 0, false
	}
	a, ok := metricWeight(cvss31CIA, m["A"])
	if !ok {
		return 0, false
	}

	iss := 1 - (1-c)*(1-i)*(1-a)
	var impact float64
	if changed {
		impact = 7.52*(iss-0.029) - 3.25*math.Pow(iss-0.02, 15)
	} else {
		impact = 6.42 * iss
	}
	if impact <= 0 {
		return 0, true
	}
	exploitability := 8.22 * av * ac * pr * ui
	var base float64
	if changed {
		base = roundup(math.Min(1.08*(impact+exploitability), 10))
	} else {
		base = roundup(math.Min(impact+exploitability, 10))
	}
	return base, true
}

// parseVector splits a CVSS vector ("CVSS:3.1/AV:N/AC:L/...") into metric=>value.
// A leading "CVSS:x.y" prefix is ignored.
func parseVector(vector string) map[string]string {
	m := map[string]string{}
	for _, part := range strings.Split(vector, "/") {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		if kv[0] == "CVSS" {
			continue
		}
		m[kv[0]] = kv[1]
	}
	return m
}

// metricWeight looks up a metric value's numeric weight; ok is false if absent.
func metricWeight(table map[string]float64, value string) (float64, bool) {
	w, ok := table[value]
	return w, ok
}

var (
	cvss31AV  = map[string]float64{"N": 0.85, "A": 0.62, "L": 0.55, "P": 0.2}
	cvss31AC  = map[string]float64{"L": 0.77, "H": 0.44}
	cvss31UI  = map[string]float64{"N": 0.85, "R": 0.62}
	cvss31CIA = map[string]float64{"H": 0.56, "L": 0.22, "N": 0.0}
	// Privileges Required weights differ by Scope per the v3.1 spec.
	cvss31PRUnchanged = map[string]float64{"N": 0.85, "L": 0.62, "H": 0.27}
	cvss31PRChanged   = map[string]float64{"N": 0.85, "L": 0.68, "H": 0.5}
)

// roundup applies the CVSS v3.1 Roundup function: round to one decimal place,
// always rounding up unless the value is already an exact tenth.
func roundup(input float64) float64 {
	intInput := int(math.Round(input * 100000))
	if intInput%10000 == 0 {
		return float64(intInput) / 100000.0
	}
	return (math.Floor(float64(intInput)/10000.0) + 1) / 10.0
}
