package topology

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func ParseCPUList(s string) ([]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	seen := map[int]struct{}{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid cpu range %q", part)
			}
			start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid cpu %q: %w", bounds[0], err)
			}
			end, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid cpu %q: %w", bounds[1], err)
			}
			if start > end {
				return nil, fmt.Errorf("invalid cpu range %q", part)
			}
			for cpu := start; cpu <= end; cpu++ {
				seen[cpu] = struct{}{}
			}
			continue
		}
		cpu, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid cpu %q: %w", part, err)
		}
		seen[cpu] = struct{}{}
	}

	out := make([]int, 0, len(seen))
	for cpu := range seen {
		out = append(out, cpu)
	}
	sort.Ints(out)
	return out, nil
}

func FormatCPUList(cpus []int) string {
	if len(cpus) == 0 {
		return ""
	}
	uniq := make([]int, 0, len(cpus))
	seen := map[int]struct{}{}
	for _, cpu := range cpus {
		if _, ok := seen[cpu]; ok {
			continue
		}
		seen[cpu] = struct{}{}
		uniq = append(uniq, cpu)
	}
	sort.Ints(uniq)

	var b strings.Builder
	start := uniq[0]
	prev := uniq[0]
	flush := func(s, e int) {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		if s == e {
			b.WriteString(strconv.Itoa(s))
			return
		}
		b.WriteString(strconv.Itoa(s))
		b.WriteByte('-')
		b.WriteString(strconv.Itoa(e))
	}
	for i := 1; i < len(uniq); i++ {
		cpu := uniq[i]
		if cpu == prev+1 {
			prev = cpu
			continue
		}
		flush(start, prev)
		start = cpu
		prev = cpu
	}
	flush(start, prev)
	return b.String()
}

func ContainsCPU(cpus []int, cpu int) bool {
	for _, c := range cpus {
		if c == cpu {
			return true
		}
	}
	return false
}

func CanonicalizeCPUList(s string) (string, []int, error) {
	cpus, err := ParseCPUList(s)
	if err != nil {
		return "", nil, err
	}
	return FormatCPUList(cpus), cpus, nil
}
