package main

import "os/exec"

type IntermediateErr struct {
	error
}

func runJob(id string) error {
	const jobBinPath = "bad/job/binary"
	isExecutable, err := isGloballyExec(jobBinPath)
	if err != nil {
		return IntermediateErr{wrapError(
			err,
			"cannot run job %q: requisite binaries not available",
			id,
		)}
	} else if isExecutable == false {
		return wrapError(nil, "job binary is not executable")
	}

	return exec.Command(jobBinPath, "--id="+id).Run()
}
