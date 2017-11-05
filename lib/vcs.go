package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var commitPat = regexp.MustCompile(`^\[(.+) ([0-9a-f]+)\] (.+)\n`)

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

// Add stages a new file
func (r *GitRepo) Add(path string) error {
	defer r.resetCmd()
	if strings.HasPrefix(path, "/") {
		newPath, errr := filepath.Rel(r.cmd.Dir, path)
		if errr != nil {
			return errr
		} else if strings.HasPrefix(newPath, "../") {
			return fmt.Errorf(
				"Path must be relative to repository root (%s)", r.cmd.Dir)
		}
		path = newPath
	}
	r.cmd.Args = append(r.cmd.Args, "add", path)
	stdout, stderr, err := r.run()
	if err != nil {
		return fmt.Errorf("%+v\n%q\n%q", err, stdout, stderr)
	}
	return nil
}

// Commit the staged changes
func (r *GitRepo) Commit(message string, author string) (string, error) {
	defer r.resetCmd()
	r.cmd.Args = append(
		r.cmd.Args, "commit", "-m", message)
	if author != "" {
		r.cmd.Args = append(r.cmd.Args, "--author", author)
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
	defer r.resetCmd()
	r.cmd.Args = append(r.cmd.Args, "reset")
	if stdout, stderr, err := r.run(); err != nil {
		return fmt.Errorf("%q\n%q", stdout, stderr)
	}
	r.resetCmd()
	r.cmd.Args = append(r.cmd.Args, "clean", "-fd")
	if stdout, stderr, err := r.run(); err != nil {
		return fmt.Errorf("%q\n%q", stdout, stderr)
	}
	return nil
}

func writeLineData(volumeIdent string, basePath string, line OCRLine, repo *GitRepo) (string, string, error) {
	lineID := MakeLineIdentifier(volumeIdent, line)
	cachedPath := LineCache.GetLinePath(lineID)
	if cachedPath == "" {
		path, err := LineCache.CacheLine(line.ImageURL, lineID)
		if err != nil {
			return "", "", err
		}
		cachedPath = path
	}

	// Move line image from cache into repository
	imgPath := filepath.Join(basePath, lineID+".png")
	in, err := os.Open(cachedPath)
	if err != nil {
		return "", "", err
	}
	out, err := os.Create(imgPath)
	if err != nil {
		return "", "", err
	}
	io.Copy(out, in)
	in.Close()
	out.Close()
	if err := os.Remove(cachedPath); err != nil {
		return "", "", err
	}
	if err := repo.Add(imgPath); err != nil {
		return "", "", err
	}

	// Write transcription
	transPath := filepath.Join(basePath, lineID+".txt")
	transOut, err := os.Create(transPath)
	if err != nil {
		return "", "", err
	}
	if _, err = transOut.WriteString(line.Transcription + "\n"); err != nil {
		return "", "", err
	}
	transOut.Close()
	if err := repo.Add(transPath); err != nil {
		return "", "", err
	}
	return imgPath, Sha1Digest([]byte(line.ImageURL)), nil
}

// GitWatcher TODO
// TODO: Progress channel?
func GitWatcher(repoPath string, taskChan chan TaskDefinition) {
	log.Printf("Launched GitWatcher")
	repo, _ := GitOpen(repoPath)
outer:
	for {
		task, more := <-taskChan
		if !more {
			return
		}
		log.Printf("Committing transcriptions...")
		if err := repo.CleanUp(); err != nil {
			task.ResultChan <- SubmitResult{Error: err}
			continue
		}
		if err := repo.Pull("origin", "master", true); err != nil {
			task.ResultChan <- SubmitResult{Error: err}
			continue
		}
		ident := task.Identifier
		yearPath := filepath.Join(
			repoPath, "transcriptions", task.Metadata.Get("year").MustString())
		os.MkdirAll(yearPath, 0755)
		lineMapping := make(map[string]string)
		for _, line := range task.Lines {
			_, lineHash, err := writeLineData(ident, yearPath, line, repo)
			if err != nil {
				task.ResultChan <- SubmitResult{Error: err}
				continue outer
			}
			lineMapping[lineHash] = line.ImageURL
		}
		if err := LineCache.PurgeLines(ident); err != nil {
			task.ResultChan <- SubmitResult{Error: err}
			continue outer
		}
		task.Metadata.Set("lines", lineMapping)
		metaPath := filepath.Join(yearPath, ident+".json")
		metaOut, _ := os.Create(metaPath)
		json.NewEncoder(metaOut).Encode(task.Metadata)
		metaOut.Close()
		if err := repo.Add(metaPath); err != nil {
			task.ResultChan <- SubmitResult{Error: err}
			continue
		}
		readme := createReadme(repoPath)
		readmePath := filepath.Join(repoPath, "README.md")
		readmeOut, _ := os.Create(readmePath)
		readmeOut.WriteString(readme)
		readmeOut.Close()
		if err := repo.Add(readmePath); err != nil {
			task.ResultChan <- SubmitResult{Error: err}
			continue
		}
		commitMessage := fmt.Sprintf(
			"Transcribed %d lines from %s (%s)", len(task.Lines), task.Identifier,
			task.Metadata.Get("year").MustString())
		if task.Comment != "" {
			commitMessage += ("\n" + task.Comment)
		}
		if commitHash, err := repo.Commit(commitMessage, task.Author); err != nil {
			task.ResultChan <- SubmitResult{Error: err}
			continue
		} else {
			res := SubmitResult{CommitSha: commitHash}
			task.ResultChan <- res
		}
		repo.Push("origin", "master")
	}
}
