package git

import (
	"sort"
	"strings"
)

type ChangeSet struct {
	Commit string
	Files  []string
}

type PairWeight struct {
	Left   string
	Right  string
	Count  int
	Weight float64
}

func ParseNameOnlyLog(raw string) []ChangeSet {
	lines := strings.Split(raw, "\n")
	out := make([]ChangeSet, 0)
	var current *ChangeSet
	seen := map[string]struct{}{}

	flush := func() {
		if current == nil {
			return
		}
		if len(current.Files) > 0 {
			out = append(out, *current)
		}
		current = nil
		seen = map[string]struct{}{}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "/") && !strings.Contains(line, ".") {
			flush()
			current = &ChangeSet{Commit: line}
			continue
		}
		if current == nil {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		current.Files = append(current.Files, line)
	}
	flush()
	return out
}

func CoChangeWeights(changeSets []ChangeSet) []PairWeight {
	type pair struct {
		left  string
		right string
	}

	counts := map[pair]int{}
	maxCount := 0
	for _, changeSet := range changeSets {
		if len(changeSet.Files) < 2 {
			continue
		}
		files := append([]string(nil), changeSet.Files...)
		sort.Strings(files)
		for i := 0; i < len(files); i++ {
			for j := i + 1; j < len(files); j++ {
				key := pair{left: files[i], right: files[j]}
				counts[key]++
				if counts[key] > maxCount {
					maxCount = counts[key]
				}
			}
		}
	}
	if maxCount == 0 {
		return nil
	}

	out := make([]PairWeight, 0, len(counts))
	for key, count := range counts {
		out = append(out, PairWeight{
			Left:   key.left,
			Right:  key.right,
			Count:  count,
			Weight: float64(count) / float64(maxCount),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		if out[i].Left != out[j].Left {
			return out[i].Left < out[j].Left
		}
		return out[i].Right < out[j].Right
	})
	return out
}
