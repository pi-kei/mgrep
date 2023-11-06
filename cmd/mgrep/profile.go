package main

import (
	"errors"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
)

func getProfile(profile string) (func(), error) {
	fileExt := ".prof"
	if profile == "cpu" {
		cpuProf, err := os.Create(profile + fileExt)
		if err != nil {
			return nil, err
		}
		err = pprof.StartCPUProfile(cpuProf)
		if err != nil {
			cpuProf.Close()
			return nil, err
		}
		return func() {
			pprof.StopCPUProfile()
			cpuProf.Close()
		}, nil
	} else if profile == "heap" {
		return func() {
			heapProf, err := os.Create(profile + fileExt)
			if err != nil {
				return
			}
			defer heapProf.Close()
			p := pprof.Lookup("heap")
			if p == nil {
				return
			}
			p.WriteTo(heapProf, 0)
		}, nil
	} else if profile == "block" {
		runtime.SetBlockProfileRate(1)
		return func() {
			blockProf, err := os.Create(profile + fileExt)
			if err != nil {
				return
			}
			defer blockProf.Close()
			p := pprof.Lookup("block")
			if p == nil {
				return
			}
			p.WriteTo(blockProf, 0)
		}, nil
	} else if profile == "mutex" {
		runtime.SetMutexProfileFraction(1)
		return func() {
			mutexProf, err := os.Create(profile + fileExt)
			if err != nil {
				return
			}
			defer mutexProf.Close()
			p := pprof.Lookup("mutex")
			if p == nil {
				return
			}
			p.WriteTo(mutexProf, 0)
		}, nil
	} else if profile == "trace" {
		traceProf, err := os.Create(profile + fileExt)
		if err != nil {
			return nil, err
		}
		err = trace.Start(traceProf)
		if err != nil {
			traceProf.Close()
			return nil, err
		}
		return func() {
			trace.Stop()
			traceProf.Close()
		}, nil
	} else if len(profile) > 0 {
		return nil, errors.New("unknown profile")
	}

	return nil, nil
}