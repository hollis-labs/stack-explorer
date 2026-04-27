package audits

import (
	"fmt"
	"sort"
	"strings"
)

type DiffBucket struct {
	Label    string    `json:"label"`
	Findings []Finding `json:"findings"`
}

type DiffResult struct {
	Resolved     []Finding `json:"resolved"`
	StillOpen    []Finding `json:"still_open"`
	New          []Finding `json:"new"`
	Regressed    []Finding `json:"regressed"`
	PossiblySame []string  `json:"possibly_same"`
}

func DiffBundles(a, b *AuditBundle) (*DiffResult, error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("both audits are required")
	}
	if a.Audit.RepoID != b.Audit.RepoID {
		return nil, fmt.Errorf("audits must belong to the same repo")
	}
	left := map[string]Finding{}
	for _, finding := range a.Findings {
		left[findingKey(a.Audit.Scope, finding)] = finding
	}
	right := map[string]Finding{}
	for _, finding := range b.Findings {
		right[findingKey(b.Audit.Scope, finding)] = finding
	}
	result := &DiffResult{}
	for key, leftFinding := range left {
		rightFinding, ok := right[key]
		if !ok {
			result.Resolved = append(result.Resolved, leftFinding)
			continue
		}
		if severityRank(rightFinding.Severity) < severityRank(leftFinding.Severity) {
			result.Regressed = append(result.Regressed, rightFinding)
		} else {
			result.StillOpen = append(result.StillOpen, rightFinding)
		}
		delete(right, key)
	}
	for _, finding := range right {
		result.New = append(result.New, finding)
	}
	for _, leftFinding := range a.Findings {
		for _, rightFinding := range b.Findings {
			if normalizeTitle(leftFinding.Title) == normalizeTitle(rightFinding.Title) && primaryFile(leftFinding) != "" && primaryFile(leftFinding) == primaryFile(rightFinding) {
				leftKey := findingKey(a.Audit.Scope, leftFinding)
				rightKey := findingKey(b.Audit.Scope, rightFinding)
				if leftKey != rightKey {
					result.PossiblySame = append(result.PossiblySame, fmt.Sprintf("%s <-> %s", leftFinding.Title, rightFinding.Title))
				}
			}
		}
	}
	sortFindings := func(items []Finding) {
		sort.Slice(items, func(i, j int) bool {
			if severityRank(items[i].Severity) != severityRank(items[j].Severity) {
				return severityRank(items[i].Severity) < severityRank(items[j].Severity)
			}
			return items[i].Title < items[j].Title
		})
	}
	sortFindings(result.Resolved)
	sortFindings(result.StillOpen)
	sortFindings(result.New)
	sortFindings(result.Regressed)
	sort.Strings(result.PossiblySame)
	return result, nil
}

func findingKey(scope string, finding Finding) string {
	return strings.Join([]string{scope, normalizeTitle(finding.Title), primaryFile(finding)}, "|")
}

func normalizeTitle(title string) string {
	title = strings.ToLower(title)
	replacer := strings.NewReplacer("`", "", "'", "", "\"", "", "—", "-", ":", "", ",", "", ".", "", "(", "", ")", "")
	title = replacer.Replace(title)
	title = strings.Join(strings.Fields(title), " ")
	return title
}

func primaryFile(f Finding) string {
	if len(f.CodeRefs) == 0 {
		return ""
	}
	return f.CodeRefs[0].FilePath
}

func severityRank(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}
