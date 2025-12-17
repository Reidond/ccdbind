package topology

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Result struct {
	OSCPUs   string
	GameCPUs string
	Lists    []string
}

// SelectOSAndGame picks OS CPUs as the list containing CPU0 and GAME CPUs as the
// union of all other lists.
func SelectOSAndGame(lists []string) (osCPUs string, gameCPUs string, canonicalLists []string, err error) {
	uniq := map[string]struct{}{}
	for _, s := range lists {
		canonical, _, err := CanonicalizeCPUList(s)
		if err != nil || canonical == "" {
			continue
		}
		uniq[canonical] = struct{}{}
	}
	if len(uniq) == 0 {
		return "", "", nil, errors.New("no valid cpu lists")
	}

	canonicalLists = make([]string, 0, len(uniq))
	for s := range uniq {
		canonicalLists = append(canonicalLists, s)
	}
	sort.Strings(canonicalLists)

	osIdx := -1
	for i, s := range canonicalLists {
		_, cpus, err := CanonicalizeCPUList(s)
		if err != nil {
			continue
		}
		if ContainsCPU(cpus, 0) {
			osIdx = i
			break
		}
	}
	if osIdx == -1 {
		return "", "", canonicalLists, fmt.Errorf("no cpu list contains CPU0: %v", canonicalLists)
	}
	osCPUs = strings.TrimSpace(canonicalLists[osIdx])

	other := make([]int, 0, 64)
	for i, s := range canonicalLists {
		if i == osIdx {
			continue
		}
		_, cpus, err := CanonicalizeCPUList(s)
		if err != nil {
			continue
		}
		if ContainsCPU(cpus, 0) {
			continue
		}
		other = append(other, cpus...)
	}
	gameCPUs = strings.TrimSpace(FormatCPUList(other))
	return osCPUs, gameCPUs, canonicalLists, nil
}

func Detect() (Result, error) {
	files, err := filepath.Glob("/sys/devices/system/cpu/cpu*/cache/index3/shared_cpu_list")
	if err != nil {
		return Result{}, err
	}
	if len(files) == 0 {
		return Result{}, errors.New("no index3 shared_cpu_list files found")
	}

	raw := make([]string, 0, len(files))
	for _, path := range files {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		raw = append(raw, string(b))
	}
	if len(raw) == 0 {
		return Result{}, errors.New("failed to read any cpu lists")
	}

	osCPUs, gameCPUs, lists, err := SelectOSAndGame(raw)
	if err != nil {
		return Result{}, err
	}
	return Result{OSCPUs: osCPUs, GameCPUs: gameCPUs, Lists: lists}, nil
}
