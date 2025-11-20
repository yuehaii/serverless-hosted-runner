package common

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Fa1k3n/resolvconf"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

var SysTimeZoneP = map[string]string{
	"00:00": "Europe/Dublin",
	"01:00": "Europe/Busingen",
	"02:00": "Europe/Helsinki",
	"03:00": "Europe/Minsk",
	"04:00": "Europe/Samara",
	"05:00": "Asia/Qostanay",
	"06:00": "Asia/Omsk",
	"07:00": "Asia/Saigon",
	"08:00": "Asia/Shanghai",
	"09:00": "Asia/Seoul",
	"10:00": "Asia/Sakhalin",
	"11:00": "Asia/Srednekolymsk",
	"12:00": "NZ",
	"13:00": "Pacific/Apia",
	"14:00": "Pacific/Kiritimati",
}
var SysTimeZoneS = map[string]string{
	"00:00": "Europe/Dublin",
	"01:00": "Atlantic/Cape_Verde",
	"02:00": "Atlantic/South_Georgia",
	"03:00": "America/Mendoza",
	"04:00": "America/Martinique",
	"05:00": "America/Louisville",
	"06:00": "America/Guatemala",
	"07:00": "America/Inuvik",
	"08:00": "Pacific/Pitcairn",
	"09:00": "Pacific/Gambier",
	"10:00": "Pacific/Rarotonga",
	"11:00": "Pacific/Midway",
	"12:00": "Etc/GMT+12",
}

// ISysErr DONE: https://github.com/ingka-group-digital/serverless-hosted-runner/issues/14
type ISysErr interface {
	IsSysBusy(string) bool
	IsFileBusy(string) bool
}

type ISysFunc interface {
	SetResolvers()
	StartProcess(string, ...string)
	DockerStorageDriver(string) error
	Addr() string
}

type ISysCtl interface {
	ISysErr
	ISysFunc
}

type IUnixSysCtl interface {
	ISysCtl
	ReloadPlugin() error
	ExceedReload() bool
	NetworkConnectivity()
}

type UnixSysCtl struct {
	oomKill                 string
	oom                     string
	fileBusy                string
	pluginNotInstall        string
	pluginTimeoutStart      string
	couldNotConnectRegistry string
	failToReadSchema        string
	failToReadProvider      string
	couldNotQueryRegistry   string
	resolvers               []string
	binPath                 string
	busyCount               int
	busyCountTotal          int
	busyCountReload         int
	busyProcess             string
	busyProcessDuration     string
	processCmd              bool
}

type WindowsSysCtl struct {
}

func CreateUnixSysCtl() IUnixSysCtl {
	return &UnixSysCtl{"signal: killed", "out of memory", "text file busy",
		"plugins are not installed", "timeout while waiting for plugin to start",
		"could not connect to registry", "failed to read schema",
		"failed to read provider", "could not query provider registry",
		[]string{"10.82.31.69", "10.82.31.116"}, "/usr/bin/", 0, 0, 10,
		"terraform-provi", "1800", true}
}

func ParseTimeLocation(exp string) (time.Time, error) {
	fmtTime := "2006-01-02T15:04:05"
	timeAndNano := strings.Split(exp, ".")
	expTime := timeAndNano[0]
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logrus.Errorf("fail to parse default location: %v", err)
		return time.Time{}, fmt.Errorf("fail to parse default location")
	}
	if len(timeAndNano[1]) > 0 && strings.Contains(timeAndNano[1], "+") {
		nanoAndTz := strings.Split(exp, "+")
		expTime += "+" + nanoAndTz[1]
		fmtTime += "+" + nanoAndTz[1]
		if l, ok := SysTimeZoneP[nanoAndTz[1]]; ok {
			if newLoc, err := time.LoadLocation(l); err == nil {
				location = newLoc
			}
		}
	} else if len(timeAndNano[1]) > 0 && strings.Contains(timeAndNano[1], "-") {
		nanoAndTz := strings.Split(exp, "-")
		expTime += "-" + nanoAndTz[1]
		fmtTime += "-" + nanoAndTz[1]
		if l, ok := SysTimeZoneS[nanoAndTz[1]]; ok {
			if newLoc, err := time.LoadLocation(l); err == nil {
				location = newLoc
			}
		}
	}
	return time.ParseInLocation(fmtTime, expTime, location)

}

func (ctl *UnixSysCtl) ReloadPlugin() error {
	logrus.Warnf("busyCount %v, busyCountTotal %v",
		ctl.busyCount, ctl.busyCountTotal)
	if ctl.busyCount > ctl.busyCountReload {
		ctl.busyCount = 0
		logrus.Warnf("reloading terraform plugin longer than specified")
		return TernaryComparable(ctl.processCmd,
			ctl.cleanProcessCmd(ctl.busyProcess, ctl.busyProcessDuration),
			ctl.cleanProcessPipe(ctl.busyProcess, ctl.busyProcessDuration))
	} else {
		return nil
	}
}

func (ctl *UnixSysCtl) ExceedReload() bool {
	return ctl.busyCountTotal > ctl.busyCountReload
}

func (ctl *UnixSysCtl) IsSysBusy(sysmsg string) bool {
	busy := strings.Contains(sysmsg, ctl.oom) || strings.Contains(sysmsg, ctl.oomKill) ||
		strings.Contains(sysmsg, ctl.pluginTimeoutStart) ||
		strings.Contains(sysmsg, ctl.couldNotConnectRegistry) ||
		strings.Contains(sysmsg, ctl.failToReadSchema) ||
		strings.Contains(sysmsg, ctl.failToReadProvider) ||
		strings.Contains(sysmsg, ctl.couldNotQueryRegistry)
	if busy {
		ctl.busyCount += 1
		ctl.busyCountTotal += 1
	}
	return busy
}

func (ctl *UnixSysCtl) IsFileBusy(sysmsg string) bool {
	busy := strings.Contains(sysmsg, ctl.fileBusy) || strings.Contains(sysmsg, ctl.pluginNotInstall)
	if busy {
		ctl.busyCount += 1
		ctl.busyCountTotal += 1
	}
	return busy
}

func (ctl UnixSysCtl) SetResolvers() {
	if err := TernaryComparable(ctl.processCmd, ctl.setWithCmd(), ctl.setWithResolvConf()); err != nil {
		logrus.Errorf("fail to set resolver, %v", err)
	}
}

func (ctl UnixSysCtl) DockerStorageDriver(driver string) error {
	return ctl.enableStorageDriver(driver)
}

func (ctl UnixSysCtl) StartProcess(name string, args ...string) {
	if err := ctl.sysConf(name); err != nil {
		logrus.Errorf("fail to implement sys config for %s process, %v", name, err)
	}
	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	p, err := os.StartProcess(ctl.binPath+name, args, &procAttr)
	if err != nil {
		logrus.Errorf("fail to start process, %s", err)
	} else {
		logrus.Warnf("success start process %s, pid %v", name, p.Pid)
	}
}

func (ctl UnixSysCtl) Addr() string {
	addr := ""
	if disAddr, err := net.InterfaceAddrs(); err == nil {
		for _, item := range disAddr {
			logrus.Infof("Loop Addr go is: %s", item.String())
		}
		for _, item := range disAddr {
			if len(item.String()) > 0 && !strings.Contains(item.String(), "127.0.0.1") {
				addr = item.String()
				logrus.Infof("Addr go is: %s", addr)
				break
			}
		}
	}
	if len(addr) == 0 {
		addr = ctl.cmdAddr()
		logrus.Infof("Addr cmd is: %s", addr)
	}
	return ctl.filterAddr(addr)
}

func (ctl UnixSysCtl) cmdAddr() string {
	if resOut, err := exec.Command("/usr/bin/hostname", "-I").Output(); err == nil {
		return string(resOut)
	}
	return ""
}

func (ctl UnixSysCtl) filterAddr(addr string) string {
	splitAddr := strings.Split(addr, "/")
	return strings.TrimSpace(splitAddr[0])
}

func (ctl UnixSysCtl) NetworkConnectivity() {
	ctl.httpsConnectivity("https://registry.terraform.io/.well-known/terraform.json", 3)
	ctl.httpsConnectivity("https://git.build.ingka.ikea.com/api/v3/repos/labrador/sentry-exporter", 2)
	ctl.httpsConnectivity("https://www.google.com", 1)

}

func (ctl UnixSysCtl) httpsConnectivity(connURL string, connTimes int) {
	logrus.Warnf("NetworkConnectivity connURL: %s", connURL)
	sum := 1

	for sum < connTimes {
		cmd := exec.Command("/bin/bash", "-c", "wget "+connURL)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			logrus.Errorf("fail to start network connectivity %v", err)
			return
		}
		err = cmd.Wait()
		if err != nil {
			logrus.Errorf("fail to wait network connectivity %v", err)
		}
		time.Sleep(10 * time.Second)
		sum += 1
	}
}

func (ctl UnixSysCtl) setWithCmd() error {
	resOut, err := exec.Command(ctl.binPath+"cat", "/etc/resolv.conf").Output()
	if err != nil {
		logrus.Errorf("fail to get resolver with cmd %v, output %s", err, string(resOut))
	}
	composePara := ""
	for _, ser := range ctl.resolvers {
		composePara += "nameserver " + ser + "\n"
	}
	composePara = "'" + composePara + "\n" + string(resOut) + "'"
	logrus.Infof("composePara %s", composePara)

	cmd := exec.Command("/bin/bash", "-c",
		"echo -e "+composePara+" > /etc/resolv.conf")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		logrus.Errorf("fail to start cmd %v", err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		logrus.Errorf("fail to wait cmd %v", err)
		return err
	}
	return err
}

func (ctl UnixSysCtl) setWithResolvConf() (err error) {
	conf := resolvconf.New()
	nameservers := conf.GetNameservers()
	for _, ns := range nameservers {
		if err = conf.Remove(ns); err != nil {
			logrus.Errorf("fail to remove sys resolver %s, %v", ns.IP, err)
		}
	}
	for _, resolver := range ctl.resolvers {
		if err = conf.Add(resolvconf.NewNameserver(net.ParseIP(resolver))); err != nil {
			logrus.Errorf("fail to add new resolver %s, %v", resolver, err)
		}
	}
	for _, ns := range nameservers {
		if err = conf.Add(ns); err != nil {
			logrus.Errorf("fail to add back sys resolver %s, %v", ns.IP, err)
		}
	}
	if err = conf.Write(logrus.StandardLogger().Writer()); err != nil {
		logrus.Errorf("fail to write sys resolver, %v", err)
		if err = ctl.setWithResolverFile(); err != nil {
			logrus.Errorf("fail to set with resolver file, %v", err)
		}
	}
	return err
}

func (ctl UnixSysCtl) setWithResolverFile() error {
	appendFile, err := os.Create("/etc/resolv.conf.add")
	if err != nil {
		logrus.Errorf("fail to create a new resolver file, %s", err)
	}
	readFile, err := os.Open("/etc/resolv.conf")
	if err != nil {
		logrus.Errorf("fail to open resolver file, %s", err)
	}
	fileClose := func() {
		if err := readFile.Close(); err != nil {
			logrus.Errorf("fail to close resolver read file, %v", err)
		}
		if err := appendFile.Close(); err != nil {
			logrus.Errorf("fail to close resolver append file, %v", err)
		}
	}
	defer fileClose()
	for _, resolver := range ctl.resolvers {
		_, err = appendFile.WriteString("nameserver " + resolver)
		if err != nil {
			logrus.Errorf("fail to add line to new resolver file, %s", err)
			return err
		}
		_, err = appendFile.WriteString("\n")
		if err != nil {
			logrus.Errorf("fail to add newline to new resolver file, %s", err)
			return err
		}
	}
	scanner := bufio.NewScanner(readFile)
	for scanner.Scan() {
		_, err = appendFile.WriteString(scanner.Text())
		if err != nil {
			logrus.Errorf("fail to add ori resolver to new resolver file, %s", err)
			return err
		}
		_, err = appendFile.WriteString("\n")
		if err != nil {
			logrus.Errorf("fail to add newline to new resolver file, %s", err)
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		logrus.Errorf("fail to scan the ori resolver file, %s", err)
		return err
	}
	if err := appendFile.Sync(); err != nil {
		logrus.Errorf("fail to sync resolver file, %v", err)
	}
	err = os.Rename("/etc/resolv.conf.add", "/etc/resolv.conf")
	if err != nil {
		logrus.Errorf("fail to update the resolver file, %s", err)
		return err
	}
	return nil
}

func (ctl UnixSysCtl) enableIPForward() (err error) {
	cmd := exec.Command("/bin/bash", "-c",
		"echo -e net.ipv4.ip_forward=1 >> /etc/sysctl.conf")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		logrus.Errorf("fail to start ipfwd cmd %v", err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		logrus.Errorf("fail to wait ipfwd cmd %v", err)
		return err
	}
	return nil
}

func (ctl UnixSysCtl) enableStorageDriver(driver string) (err error) {
	if err := os.MkdirAll("/etc/docker", os.ModePerm); err != nil {
		logrus.Errorf("fail to create docker dir: %s", err)
	}
	cmd := exec.Command("/bin/bash", "-c",
		"echo -e  {\\\"storage-driver\\\": \\\""+driver+"\\\"} > /etc/docker/daemon.json")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		logrus.Errorf("fail to start storage driver cmd %v", err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		logrus.Errorf("fail to wait storage driver cmd %v", err)
		return err
	}
	return nil
}

func (ctl UnixSysCtl) sysConf(name string) error {
	logrus.Infof("sysConf %s", name)
	if name == "dockerd" {
		_, err := os.Stat("/etc/docker/daemon.json")
		if err != nil {
			if err = ctl.enableIPForward(); err != nil {
				logrus.Errorf("sysConf fail to enableIPForward %v", err)
			}
			if err = ctl.enableStorageDriver("overlay2"); err != nil {
				logrus.Errorf("sysConf fail to enableStorageDriver %v", err)
			}
		} else {
			logrus.Warnf("/etc/docker/daemon.json exists")
		}
	} // other sys conf
	return nil
}

func (ctl UnixSysCtl) cleanProcess(p string, duration string) (err error) {
	cmd := exec.Command("/bin/bash", "-c",
		"kill -9 $(ps -axfo pid,comm,etimes | grep "+p+" | awk '\\$NF > "+duration+"' | awk '{print \\$1}')")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		logrus.Errorf("fail to start clean process %s, %v", p, err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		logrus.Errorf("fail to wait clean process %s, %v", p, err)
		return err
	}
	logrus.Warnf("success clean process %s", p)
	return nil
}

func (ctl UnixSysCtl) cleanProcessCmd(p string, duration string) (err error) {
	err = ctl.buildProcessCmd(p, duration)
	if err != nil {
		logrus.Errorf("fail to build process cmd %v", err)
		errp := ctl.cleanProcess(p, duration)
		if errp != nil {
			logrus.Errorf("fail to clean with default process cmd %v", errp)
		}
		return err
	}
	err = ctl.startProcessCmd()
	if err != nil {
		logrus.Errorf("fail to start process cmd %v", err)
		return err
	}
	return err
}

func (ctl UnixSysCtl) startProcessCmd() (err error) {
	out, err := exec.Command("/bin/bash", "./cleanp.sh").Output()
	if err != nil {
		logrus.Errorf("fail to start proces cmd %v, out: %s", err, string(out))
		return err
	}
	logrus.Warnf("success start proces cmd out: %s", string(out))
	return err
}

func (ctl UnixSysCtl) buildProcessCmd(p string, duration string) (err error) {
	cmd := exec.Command("/bin/bash", "-c",
		"echo -e \"ps -axfo pid,comm,etimes | grep "+p+
			" | awk '\\$NF > "+duration+
			"' | awk '{print \\$1}' | xargs kill -9\" > /go/bin/cleanp.sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		logrus.Errorf("fail to start building cmd %v", err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		logrus.Errorf("fail to wait building cmd %v", err)
		return err
	}
	return nil
}

func (ctl UnixSysCtl) cleanProcessPipe(p string, duration string) (err error) {
	var b bytes.Buffer
	if err := ctl.execWithPipe(&b,
		exec.Command("ps", "-axfo", "pid,comm,etimes"),
		exec.Command("grep", p),
		exec.Command("awk", "'\\$NF > "+duration+"'"),
		exec.Command("awk", "'{print \\$1}'"),
		exec.Command("xargs", "kill", "-9"),
	); err != nil {
		logrus.Errorf("fail to execWithPipe %s, err:%v", p, err)
	}
	// io.Copy(os.Stdout, &b)
	if _, err := io.Copy(logrus.StandardLogger().Writer(), &b); err != nil {
		logrus.Errorf("fail to copy to std, err:%v", err)
	}
	logrus.Warnf("execWithPipe buffer:%v", b)
	return err
}

func (ctl UnixSysCtl) execWithPipe(outputBuffer *bytes.Buffer, stack ...*exec.Cmd) (err error) {
	var errorBuffer bytes.Buffer
	pipeStack := make([]*io.PipeWriter, len(stack)-1)
	i := 0
	for ; i < len(stack)-1; i++ {
		stdinPipe, stdoutPipe := io.Pipe()
		stack[i].Stdout = stdoutPipe
		stack[i].Stderr = &errorBuffer
		stack[i+1].Stdin = stdinPipe
		pipeStack[i] = stdoutPipe
	}
	stack[i].Stdout = outputBuffer
	stack[i].Stderr = &errorBuffer

	if err := ctl.callStack(stack, pipeStack); err != nil {
		logrus.Errorf("fail to call pipe stack %v", err)
	}
	return err
}

func (ctl UnixSysCtl) callStack(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}
	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}
		defer func() {
			if err == nil {
				if errp := pipes[0].Close(); errp != nil {
					logrus.Errorf("callStack fail close pipe %v", errp)
				}
				err = ctl.callStack(stack[1:], pipes[1:])
			} else {
				if errw := stack[1].Wait(); errw != nil {
					logrus.Errorf("callStack fail wait stack %v", errw)
				}
			}
		}()
	}
	return stack[0].Wait()
}
