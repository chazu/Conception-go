// Package u6 implements funcs for comparing working directories and branches in vcs repositories.
package u6

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/shurcooL-legacy/Conception-go/pkg/exp13"
	"github.com/shurcooL-legacy/Conception-go/pkg/gist7480523"
	"github.com/shurcooL-legacy/Conception-go/pkg/legacyvcs"
	"github.com/shurcooL/go/pipeutil"
	"gopkg.in/pipe.v2"
)

// GoPackageWorkingDiff shows the difference between the working directory and the most recent commit.
// Precondition is that goPackage.Dir.Repo is not nil, and VcsLocal is updated.
func GoPackageWorkingDiff(goPackage *gist7480523.GoPackage) string {
	if goPackage.Dir.Repo.VcsLocal.Status == "" {
		return ""
	}

	switch goPackage.Dir.Repo.Vcs.Type() {
	case legacyvcs.Git:
		newFileDiff := func(line []byte) []byte {
			cmd := exec.Command("git", "diff", "--no-ext-diff", "--", "/dev/null", strings.TrimSuffix(string(line), "\n"))
			cmd.Dir = goPackage.Dir.Repo.Vcs.RootPath()
			out, err := cmd.Output()
			if len(out) > 0 {
				// diff exits with a non-zero status when the files don't match.
				// Ignore that failure as long as we get output.
				err = nil
			}
			if err != nil {
				return []byte(err.Error())
			}
			return out
		}

		p := pipe.Script(
			pipe.Exec("git", "diff", "--no-ext-diff", "--find-renames", "HEAD"),
			pipe.Line(
				pipe.Exec("git", "ls-files", "--others", "--exclude-standard"),
				pipe.Replace(newFileDiff),
			),
		)

		out, err := pipeutil.OutputDir(p, goPackage.Dir.Repo.Vcs.RootPath())
		if err != nil {
			return err.Error()
		}
		return string(out)
	default:
		return ""
	}
}

// GoPackageWorkingDiffMaster shows the difference between the working directory and master branch.
// Precondition is that goPackage.Dir.Repo is not nil, and VcsLocal is updated.
func GoPackageWorkingDiffMaster(goPackage *gist7480523.GoPackage) string {
	if goPackage.Dir.Repo.VcsLocal.Status == "" && goPackage.Dir.Repo.VcsLocal.LocalBranch == goPackage.Dir.Repo.Vcs.GetDefaultBranch() {
		return ""
	}

	switch goPackage.Dir.Repo.Vcs.Type() {
	case legacyvcs.Git:
		newFileDiff := func(line []byte) []byte {
			cmd := exec.Command("git", "diff", "--no-ext-diff", "--", "/dev/null", strings.TrimSuffix(string(line), "\n"))
			cmd.Dir = goPackage.Dir.Repo.Vcs.RootPath()
			out, err := cmd.Output()
			if len(out) > 0 {
				// diff exits with a non-zero status when the files don't match.
				// Ignore that failure as long as we get output.
				err = nil
			}
			if err != nil {
				return []byte(err.Error())
			}
			return out
		}

		p := pipe.Script(
			pipe.Exec("git", "diff", "--no-ext-diff", "--find-renames", "master"),
			pipe.Line(
				pipe.Exec("git", "ls-files", "--others", "--exclude-standard"),
				pipe.Replace(newFileDiff),
			),
		)

		out, err := pipeutil.OutputDir(p, goPackage.Dir.Repo.Vcs.RootPath())
		if err != nil {
			return err.Error()
		}
		return string(out)
	default:
		return ""
	}
}

// GoPackageWorkingDiffOriginMaster shows the difference between the working directory and origin/master branch.
// Precondition is that goPackage.Dir.Repo is not nil, and VcsLocal is updated.
func GoPackageWorkingDiffOriginMaster(goPackage *gist7480523.GoPackage) string {
	if goPackage.Dir.Repo.VcsLocal.Status == "" && (goPackage.Dir.Repo.VcsLocal.Remote == "" || goPackage.Dir.Repo.VcsLocal.LocalRev == goPackage.Dir.Repo.VcsLocal.LocalRemoteRev) {
		return ""
	}

	switch goPackage.Dir.Repo.Vcs.Type() {
	case legacyvcs.Git:
		newFileDiff := func(line []byte) []byte {
			cmd := exec.Command("git", "diff", "--no-ext-diff", "--", "/dev/null", strings.TrimSuffix(string(line), "\n"))
			cmd.Dir = goPackage.Dir.Repo.Vcs.RootPath()
			out, err := cmd.Output()
			if len(out) > 0 {
				// diff exits with a non-zero status when the files don't match.
				// Ignore that failure as long as we get output.
				err = nil
			}
			if err != nil {
				return []byte(err.Error())
			}
			return out
		}

		p := pipe.Script(
			pipe.Exec("git", "diff", "--no-ext-diff", "--find-renames", "origin/master"),
			pipe.Line(
				pipe.Exec("git", "ls-files", "--others", "--exclude-standard"),
				pipe.Replace(newFileDiff),
			),
		)

		out, err := pipeutil.OutputDir(p, goPackage.Dir.Repo.Vcs.RootPath())
		if err != nil {
			return err.Error()
		}
		return string(out)
	default:
		return ""
	}
}

// BranchesOptions are options for Branches.
type BranchesOptions struct {
	Base string // Base branch to compare against (if blank, defaults to "master").
}

func (bo *BranchesOptions) defaults() {
	if bo.Base == "" {
		bo.Base = "master"
	}
}

// Branches returns a Markdown table of branches with ahead/behind information relative to master branch.
func Branches(repo *exp13.VcsState, opt BranchesOptions) string {
	opt.defaults()
	switch repo.Vcs.Type() {
	case legacyvcs.Git:
		branchInfo := func(line []byte) []byte {
			branch := strings.TrimSuffix(string(line), "\n")
			branchDisplay := branch
			if branch == repo.VcsLocal.LocalBranch {
				branchDisplay = "**" + branch + "**"
			}

			cmd := exec.Command("git", "rev-list", "--count", "--left-right", opt.Base+"..."+branch)
			cmd.Dir = repo.Vcs.RootPath()
			out, err := cmd.Output()
			if err != nil {
				log.Printf("error running %v: %v\n", cmd.Args, err)
				return []byte(fmt.Sprintf("%s | ? | ?\n", branchDisplay))
			}

			behindAhead := strings.Split(strings.TrimSuffix(string(out), "\n"), "\t")
			return []byte(fmt.Sprintf("%s | %s | %s\n", branchDisplay, behindAhead[0], behindAhead[1]))
		}

		p := pipe.Script(
			pipe.Println("Branch | Behind | Ahead"),
			pipe.Println("-------|-------:|:-----"),
			pipe.Line(
				pipe.Exec("git", "for-each-ref", "--format=%(refname:short)", "refs/heads"),
				pipe.Replace(branchInfo),
			),
		)

		out, err := pipeutil.OutputDir(p, repo.Vcs.RootPath())
		if err != nil {
			return err.Error()
		}
		return string(out)
	default:
		return ""
	}
}

// Input is a line containing tab-separated local branch and remote branch.
// For example, "master\torigin/master".
func branchRemoteInfo(repo *exp13.VcsState) func(line []byte) []byte {
	return func(line []byte) []byte {
		branchRemote := strings.Split(strings.TrimSuffix(string(line), "\n"), "\t")
		if len(branchRemote) != 2 {
			return []byte("error: len(branchRemote) != 2")
		}

		branch := branchRemote[0]
		branchDisplay := branch
		if branch == repo.VcsLocal.LocalBranch {
			branchDisplay = "**" + branch + "**"
		}

		remote := branchRemote[1]
		if remote == "" {
			return []byte(fmt.Sprintf("%s | | | \n", branchDisplay))
		}

		cmd := exec.Command("git", "rev-list", "--count", "--left-right", remote+"..."+branch)
		cmd.Dir = repo.Vcs.RootPath()
		out, err := cmd.Output()
		if err != nil {
			// This usually happens when the remote branch is gone.
			remoteDisplay := "~~" + remote + "~~"
			return []byte(fmt.Sprintf("%s | %s | | \n", branchDisplay, remoteDisplay))
		}

		behindAhead := strings.Split(strings.TrimSuffix(string(out), "\n"), "\t")
		return []byte(fmt.Sprintf("%s | %s | %s | %s\n", branchDisplay, remote, behindAhead[0], behindAhead[1]))
	}
}

// BranchesRemote returns a Markdown table of branches with ahead/behind information relative to remote.
func BranchesRemote(repo *exp13.VcsState) string {
	switch repo.Vcs.Type() {
	case legacyvcs.Git:
		p := pipe.Script(
			pipe.Println("Branch | Remote | Behind | Ahead"),
			pipe.Println("-------|--------|-------:|:-----"),
			pipe.Line(
				pipe.Exec("git", "for-each-ref", "--format=%(refname:short)\t%(upstream:short)", "refs/heads"),
				pipe.Replace(branchRemoteInfo(repo)),
			),
		)

		out, err := pipeutil.OutputDir(p, repo.Vcs.RootPath())
		if err != nil {
			return err.Error()
		}
		return string(out)
	default:
		return ""
	}
}

// BranchesRemoteCustom returns a Markdown table of branches with ahead/behind information relative to the specified remote.
func BranchesRemoteCustom(repo *exp13.VcsState, remote string) string {
	switch repo.Vcs.Type() {
	case legacyvcs.Git:
		p := pipe.Script(
			pipe.Println("Branch | Remote | Behind | Ahead"),
			pipe.Println("-------|--------|-------:|:-----"),
			pipe.Line(
				pipe.Exec("git", "for-each-ref", "--format=%(refname:short)\t"+remote+"/%(refname:short)", "refs/heads"),
				pipe.Replace(branchRemoteInfo(repo)),
			),
		)

		out, err := pipeutil.OutputDir(p, repo.Vcs.RootPath())
		if err != nil {
			return err.Error()
		}
		return string(out)
	default:
		return ""
	}
}
