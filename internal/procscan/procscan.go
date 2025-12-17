package procscan

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type GameProcess struct {
	PID       int
	StartTime uint64
	Exe       string
	GameID    string
	IDSource  string
}

type Scanner struct {
	UID int

	envKeyOrder []string
	envKeyIndex map[string]int

	exeAllowlist map[string]struct{}
	ignoreExe    map[string]struct{}
}

func NewScanner(uid int, envKeys, exeAllowlist, ignoreExe []string) *Scanner {
	keys := make([]string, 0, len(envKeys))
	idx := make(map[string]int, len(envKeys))
	for _, k := range envKeys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if _, ok := idx[k]; ok {
			continue
		}
		idx[k] = len(keys)
		keys = append(keys, k)
	}

	return &Scanner{
		UID:          uid,
		envKeyOrder:  keys,
		envKeyIndex:  idx,
		exeAllowlist: toSetLower(exeAllowlist),
		ignoreExe:    toSetLower(ignoreExe),
	}
}

func (s *Scanner) Scan() (map[string][]GameProcess, error) {
	ents, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	results := map[string][]GameProcess{}
	for _, ent := range ents {
		if !ent.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(ent.Name())
		if err != nil || pid <= 0 {
			continue
		}
		owned, err := isOwnedByUID(pid, s.UID)
		if err != nil || !owned {
			continue
		}

		exeBase := exeBasenameLower(pid)
		if exeBase == "" {
			continue
		}
		if _, ignored := s.ignoreExe[exeBase]; ignored {
			continue
		}

		id, src := s.gameIDFromEnviron(pid)
		if id == "" {
			if _, ok := s.exeAllowlist[exeBase]; ok {
				id = exeBase
				src = "exe_allowlist"
			}
		}
		if id == "" {
			continue
		}

		startTime, err := procStartTime(pid)
		if err != nil {
			startTime = 0
		}
		gp := GameProcess{PID: pid, StartTime: startTime, Exe: exeBase, GameID: id, IDSource: src}
		results[id] = append(results[id], gp)
	}
	return results, nil
}

func procStartTime(pid int) (uint64, error) {
	path := filepath.Join("/proc", strconv.Itoa(pid), "stat")
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	line := strings.TrimSpace(string(data))
	if line == "" {
		return 0, fmt.Errorf("empty stat")
	}

	idx := strings.LastIndexByte(line, ')')
	if idx == -1 {
		return 0, fmt.Errorf("invalid stat format")
	}
	if idx+2 >= len(line) {
		return 0, fmt.Errorf("invalid stat format")
	}

	fields := strings.Fields(line[idx+2:])
	// fields[0] is state (field 3), starttime is field 22 => index 19 here.
	if len(fields) <= 19 {
		return 0, fmt.Errorf("stat too short")
	}
	return strconv.ParseUint(fields[19], 10, 64)
}

func toSetLower(in []string) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for _, s := range in {
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" {
			continue
		}
		out[s] = struct{}{}
	}
	return out
}

func exeBasenameLower(pid int) string {
	path := filepath.Join("/proc", strconv.Itoa(pid), "exe")
	target, err := os.Readlink(path)
	if err != nil {
		return ""
	}
	base := filepath.Base(target)
	base = strings.TrimSpace(base)
	if base == "" || base == "." || base == "/" {
		return ""
	}
	return strings.ToLower(base)
}

func (s *Scanner) gameIDFromEnviron(pid int) (string, string) {
	if len(s.envKeyOrder) == 0 {
		return "", ""
	}
	path := filepath.Join("/proc", strconv.Itoa(pid), "environ")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ""
	}

	bestIdx := len(s.envKeyOrder) + 1
	bestKey := ""
	bestVal := ""

	start := 0
	for start < len(data) {
		end := bytes.IndexByte(data[start:], 0)
		if end == -1 {
			end = len(data) - start
		}
		entry := data[start : start+end]
		start += end + 1
		if len(entry) == 0 {
			continue
		}
		eq := bytes.IndexByte(entry, '=')
		if eq <= 0 {
			continue
		}
		k := string(entry[:eq])
		idx, ok := s.envKeyIndex[k]
		if !ok || idx >= bestIdx {
			continue
		}
		v := strings.TrimSpace(string(entry[eq+1:]))
		if v == "" {
			continue
		}
		bestIdx = idx
		bestKey = k
		bestVal = v
		if bestIdx == 0 {
			break
		}
	}
	return bestVal, bestKey
}

func isOwnedByUID(pid int, uid int) (bool, error) {
	path := filepath.Join("/proc", strconv.Itoa(pid), "status")
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "Uid:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return false, fmt.Errorf("unexpected Uid line: %q", line)
		}
		parsed, err := strconv.Atoi(fields[1])
		if err != nil {
			return false, err
		}
		return parsed == uid, nil
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}
	return false, errors.New("uid line not found")
}
