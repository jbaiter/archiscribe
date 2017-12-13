package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

var commitPat = regexp.MustCompile(`^\[(.+) ([0-9a-f]+)\] (.+)\n`)
var logEscapePat = regexp.MustCompile(`\s*\"(.*?)\"\s*[,}:]`)

// LogEntry encodes a git log entry
type LogEntry struct {
	Author struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
	Date    time.Time `json:"date"`
	Commit  string    `json:"commit"`
	Subject string    `json:"subject"`
	Body    string    `json:"body,omitempty"`
}

// FileStatus encodes the status of a file
type FileStatus rune

// Status constants from git diff output
const (
	StatusModified FileStatus = 'M'
	StatusAdded    FileStatus = 'A'
	StatusDeleted  FileStatus = 'D'
)

// GitRepo represents a Git repository
type GitRepo struct {
	cmd *exec.Cmd
}

// GitOpen a repository
func GitOpen(path string) (*GitRepo, error) {
	// TODO: Check if path exists
	// TODO: Check if path contains a git repo
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(gitPath)
	cmd.Dir = path
	return &GitRepo{
		cmd: cmd,
	}, nil
}

func (r *GitRepo) run() (stdout string, stderr string, err error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	r.cmd.Stdout = &stdoutBuf
	r.cmd.Stderr = &stderrBuf
	err = r.cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func (r *GitRepo) resetCmd() {
	dir := r.cmd.Dir
	r.cmd = exec.Command(r.cmd.Path)
	r.cmd.Dir = dir
}

// Pull from remote and optionally rebase
func (r *GitRepo) Pull(remote string, branch string, rebase bool) error {
	defer r.resetCmd()
	r.cmd.Args = append(r.cmd.Args, "pull", remote, branch)
	if rebase {
		r.cmd.Args = append(r.cmd.Args, "--rebase")
	}
	stdout, stderr, err := r.run()
	if err != nil {
		return fmt.Errorf("%q\n%q", stdout, stderr)
	}
	return nil
}

func (r *GitRepo) adjustPath(path string) (string, error) {
	if strings.HasPrefix(path, "/") {
		newPath, errr := filepath.Rel(r.cmd.Dir, path)
		if errr != nil {
			return "", errr
		} else if strings.HasPrefix(newPath, "../") {
			return "", fmt.Errorf(
				"Path must be relative to repository root (%s)", r.cmd.Dir)
		}
		return newPath, nil
	}
	return path, nil
}

// Add stages a new file
func (r *GitRepo) Add(path string) error {
	defer r.resetCmd()
	p, err := r.adjustPath(path)
	if err != nil {
		return err
	}
	r.cmd.Args = append(r.cmd.Args, "add", p)
	stdout, stderr, err := r.run()
	if err != nil {
		return fmt.Errorf("%+v\n%q\n%q", err, stdout, stderr)
	}
	return nil
}

// Remove removes a file
func (r *GitRepo) Remove(path string) error {
	defer r.resetCmd()
	p, err := r.adjustPath(path)
	if err != nil {
		return err
	}
	r.cmd.Args = append(r.cmd.Args, "rm", "-rf", p)
	stdout, stderr, err := r.run()
	if err != nil {
		return fmt.Errorf("%+v\n%q\n%q", err, stdout, stderr)
	}
	return nil
}

// Commit the staged changes
func (r *GitRepo) Commit(message string, author string, email string) (string, error) {
	defer r.resetCmd()
	r.cmd.Args = append(
		r.cmd.Args, "commit", "-m", message)
	if author != "" {
		r.cmd.Args = append(
			r.cmd.Args, "--author", fmt.Sprintf("%s <%s>", author, email))
	}
	stdout, stderr, err := r.run()
	if err != nil {
		return "", fmt.Errorf("%+v, %q\n%q", err, stdout, stderr)
	}
	commitSha := commitPat.FindAllStringSubmatch(stdout, 1)[0][2]
	return commitSha, nil
}

// Push changes to remote
func (r *GitRepo) Push(remote string, branch string) error {
	defer r.resetCmd()
	r.cmd.Args = append(r.cmd.Args, "push", remote, branch)
	stdout, stderr, err := r.run()
	if err != nil {
		return fmt.Errorf("%q\n%q", stdout, stderr)
	}
	return nil
}

// CleanUp residual modifications
func (r *GitRepo) CleanUp() error {
	r.cmd.Args = append(r.cmd.Args, "reset")
	if stdout, stderr, err := r.run(); err != nil {
		return fmt.Errorf("%q\n%q", stdout, stderr)
	}
	r.resetCmd()
	r.cmd.Args = append(r.cmd.Args, "checkout", "--", ".")
	if stdout, stderr, err := r.run(); err != nil {
		return fmt.Errorf("%q\n%q", stdout, stderr)
	}
	r.resetCmd()
	r.cmd.Args = append(r.cmd.Args, "clean", "-fd")
	if stdout, stderr, err := r.run(); err != nil {
		return fmt.Errorf("%q\n%q", stdout, stderr)
	}
	r.resetCmd()
	return nil
}

// Diff lists modified files
func (r *GitRepo) Diff(cached bool) (map[string]FileStatus, error) {
	defer r.resetCmd()
	r.cmd.Args = append(
		r.cmd.Args, "diff", "--name-status")
	if cached {
		r.cmd.Args = append(r.cmd.Args, "--cached")
	}
	stdout, stderr, err := r.run()
	if err != nil {
		return nil, fmt.Errorf("%q\n%q", stdout, stderr)
	}
	out := make(map[string]FileStatus)
	for _, line := range strings.Split(stdout, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if strings.Index("AMD", parts[0]) == -1 {
			// Unknown status, skipping
			continue
		}
		out[parts[1]] = FileStatus([]rune(parts[0])[0])
	}
	return out, nil
}

// Log returns the git log of a given file
func (r *GitRepo) Log(fpaths ...string) ([]LogEntry, error) {
	defer r.resetCmd()
	r.cmd.Args = append(
		r.cmd.Args, "log", `--pretty=format:{"commit":"%H","subject":"%s","body":"%b","author": {"name":"%aN","email":"%aE"},"date":"%aI"}`)
	if len(fpaths) > 0 {
		r.cmd.Args = append(r.cmd.Args, fpaths...)
	}
	stdout, stderr, err := r.run()
	if err != nil {
		return nil, fmt.Errorf("%q\n%q", stdout, stderr)
	}
	// Escape double quotes inside of values
	cleanedLog := logEscapePat.ReplaceAllStringFunc(stdout, func(match string) string {
		start := strings.Index(match, `"`)
		end := strings.LastIndex(match, `"`)
		escaped := strings.Replace(string(match[start+1:end]), `"`, `\"`, -1)
		escaped = strings.Replace(escaped, `\n`, `\\n`, -1)
		return match[:start+1] + escaped + match[end:]
	})
	logJsons := strings.Split(cleanedLog, "\n")
	logEntries := make([]LogEntry, 0, len(logJsons))
	for _, entryJSON := range logJsons {
		if entryJSON == "" {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal([]byte(entryJSON), &entry); err != nil {
			log.Error().
				Err(err).
				Str("logEntry", entryJSON).
				Msg("Failed to parse Git log entry")
			return nil, err
		}
		logEntries = append(logEntries, entry)
	}
	return logEntries, nil
}
