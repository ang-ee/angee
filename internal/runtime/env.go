package runtime

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// ReadEnvFile parses an env file into KEY=value entries suitable for a child
// process's environment. An empty path or a missing file yields nil, nil.
// Blank lines and `#` comments are skipped, each remaining line is split on the
// first `=`, key and value are trimmed, and a quoted value is unquoted.
func ReadEnvFile(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var env []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}
		env = append(env, strings.TrimSpace(key)+"="+value)
	}
	return env, scanner.Err()
}

// ChildEnviron builds the environment for a runtime child process by appending
// the env-file entries after the operator's own inherited environment.
//
// The operator may run for days with the stack's secrets exported into its own
// environment at start (process-compose starts it that way). A runtime child
// must see the env file's CURRENT values, not those stale copies, so the file
// entries are appended last: os/exec keeps the last value for a duplicate key,
// and Docker Compose / process-compose then read the fresh value from the file
// instead of the inherited copy. The returned slice is a fresh copy that does
// not alias fileEnv.
func ChildEnviron(fileEnv []string) []string {
	return append(os.Environ(), fileEnv...)
}
