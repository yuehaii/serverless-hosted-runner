package common

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Fa1k3n/resolvconf"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

// DONE: https://github.com/ingka-group-digital/serverless-hosted-runner/issues/14
type ISysErr interface {
	IsSysBusy(string) bool
	IsFileBusy(string) bool
}

type ISysFunc interface {
	SetResolvers() error
	StartProcess(string, ...string) error
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
	oom_kill                   string
	oom                        string
	file_busy                  string
	plugin_not_install         string
	plugin_timeout_start       string
	could_not_connect_registry string
	fail_to_read_schema        string
	fail_to_read_provider      string
	could_not_query_registry   string
	resolvers                  []string
	bin_path                   string
	busy_count                 int
	busy_count_total           int
	busy_count_reload          int
	busy_process               string
	busy_process_duration      string
}

type WindowsSysCtl struct {
}

func CreateUnixSysCtl() IUnixSysCtl {
	return &UnixSysCtl{"signal: killed", "out of memory", "text file busy",
		"plugins are not installed", "timeout while waiting for plugin to start",
		"could not connect to registry", "failed to read schema",
		"failed to read provider", "could not query provider registry",
		[]string{"10.82.31.69", "10.82.31.116"}, "/usr/bin/", 0, 0, 10,
		"terraform-provi", "1800"}
}

func (ctl *UnixSysCtl) ReloadPlugin() error {
	logrus.Warnf("busy_count %v, busy_count_total %v",
		ctl.busy_count, ctl.busy_count_total)
	if ctl.busy_count > ctl.busy_count_reload {
		ctl.busy_count = 0
		logrus.Warnf("reloading terraform plugin longer than specified")
		return ctl.cleanProcessCmd(ctl.busy_process, ctl.busy_process_duration)
	} else {
		return nil
	}
}

func (ctl *UnixSysCtl) ExceedReload() bool {
	return ctl.busy_count_total > ctl.busy_count_reload
}

func (ctl *UnixSysCtl) IsSysBusy(sysmsg string) bool {
	busy := strings.Contains(sysmsg, ctl.oom) || strings.Contains(sysmsg, ctl.oom_kill) ||
		strings.Contains(sysmsg, ctl.plugin_timeout_start) ||
		strings.Contains(sysmsg, ctl.could_not_connect_registry) ||
		strings.Contains(sysmsg, ctl.fail_to_read_schema) ||
		strings.Contains(sysmsg, ctl.fail_to_read_provider) ||
		strings.Contains(sysmsg, ctl.could_not_query_registry)
	if busy {
		ctl.busy_count += 1
		ctl.busy_count_total += 1
	}
	return busy
}

func (ctl *UnixSysCtl) IsFileBusy(sysmsg string) bool {
	busy := strings.Contains(sysmsg, ctl.file_busy) || strings.Contains(sysmsg, ctl.plugin_not_install)
	if busy {
		ctl.busy_count += 1
		ctl.busy_count_total += 1
	}
	return busy
}

func (ctl UnixSysCtl) SetResolvers() error {
	return ctl.setWithCmd()
}

func (ctl UnixSysCtl) DockerStorageDriver(driver string) error {
	return ctl.enableStorageDriver(driver)
}

func (ctl UnixSysCtl) StartProcess(name string, args ...string) error {
	if err := ctl.sysConf(name); err != nil {
		logrus.Errorf("fail to implement sys config for %s process, %v", name, err)
	}
	var proc_attr os.ProcAttr
	proc_attr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	p, err := os.StartProcess(ctl.bin_path+name, args, &proc_attr)
	if err != nil {
		logrus.Errorf("fail to start process, %s", err)
	} else {
		logrus.Warnf("success start process %s, pid %v", name, p.Pid)
	}
	return err
}

func (ctl UnixSysCtl) Addr() string {
	addr := ""
	if dis_addr, err := net.InterfaceAddrs(); err == nil {
		for _, item := range dis_addr {
			logrus.Infof("Loop Addr go is: %s", item.String())
		}
		for _, item := range dis_addr {
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
	if res_out, err := exec.Command("/usr/bin/hostname", "-I").Output(); err == nil {
		return string(res_out)
	}
	return ""
}

func (ctl UnixSysCtl) filterAddr(addr string) string {
	split_addr := strings.Split(addr, "/")
	return strings.TrimSpace(split_addr[0])
}

func (ctl UnixSysCtl) NetworkConnectivity() {
	ctl.httpsConnectivity("https://registry.terraform.io/.well-known/terraform.json", 3)
	ctl.httpsConnectivity("https://git.build.ingka.ikea.com/api/v3/repos/labrador/sentry-exporter", 2)
	ctl.httpsConnectivity("https://www.google.com", 1)

}

func (ctl UnixSysCtl) httpsConnectivity(conn_url string, conn_times int) {
	logrus.Warnf("NetworkConnectivity conn_url: %s", conn_url)
	sum := 1

	for sum < conn_times {
		cmd := exec.Command("/bin/bash", "-c", "wget "+conn_url)
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
	res_out, err := exec.Command(ctl.bin_path+"cat", "/etc/resolv.conf").Output()
	if err != nil {
		logrus.Errorf("fail to get resolver with cmd %v, output %s", err, string(res_out))
	}
	compose_para := ""
	for _, ser := range ctl.resolvers {
		compose_para += "nameserver " + ser + "\n"
	}
	compose_para = "'" + compose_para + "\n" + string(res_out) + "'"
	logrus.Infof("compose_para %s", compose_para)

	cmd := exec.Command("/bin/bash", "-c",
		"echo -e "+compose_para+" > /etc/resolv.conf")
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
	conf.Write(logrus.StandardLogger().Writer())
	return err
}

func (ctl UnixSysCtl) setWithResolverFile() error {
	append_file, err := os.Create("/etc/resolv.conf.add")
	if err != nil {
		logrus.Errorf("fail to create a new resolver file, %s", err)
		return err
	}
	defer append_file.Close()
	read_file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		logrus.Errorf("fail to open resolver file, %s", err)
		return err
	}
	defer read_file.Close()
	for _, resolver := range ctl.resolvers {
		_, err = append_file.WriteString("nameserver " + resolver)
		if err != nil {
			logrus.Errorf("fail to add line to new resolver file, %s", err)
			return err
		}
		_, err = append_file.WriteString("\n")
		if err != nil {
			logrus.Errorf("fail to add newline to new resolver file, %s", err)
			return err
		}
	}
	scanner := bufio.NewScanner(read_file)
	for scanner.Scan() {
		_, err = append_file.WriteString(scanner.Text())
		if err != nil {
			logrus.Errorf("fail to add ori resolver to new resolver file, %s", err)
			return err
		}
		_, err = append_file.WriteString("\n")
		if err != nil {
			logrus.Errorf("fail to add newline to new resolver file, %s", err)
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		logrus.Errorf("fail to scan the ori resolver file, %s", err)
		return err
	}
	append_file.Sync()
	err = os.Rename("/etc/resolv.conf.add", "/etc/resolv.conf")
	if err != nil {
		logrus.Errorf("fail to update the resolver file, %s", err)
		return err
	}
	return nil
}

func (ctl UnixSysCtl) enableIpForward() (err error) {
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
			ctl.enableIpForward()
			ctl.enableStorageDriver("overlay2")
		} else {
			logrus.Warnf("/etc/docker/daemon.json exists")
		}
	} // other sys conf
	return nil
}

func (ctl UnixSysCtl) cleanProcess(p string, duration int) (err error) {
	cmd := exec.Command("/bin/bash", "-c",
		"kill -9 $(ps -axfo pid,comm,etimes | grep "+p+" | awk '\\$NF > "+string(duration)+"' | awk '{print \\$1}')")
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

func (ctl UnixSysCtl) cleanProcessPipe(p string, duration int) (err error) {
	var b bytes.Buffer
	if err := ctl.execWithPipe(&b,
		exec.Command("ps", "-axfo", "pid,comm,etimes"),
		exec.Command("grep", p),
		exec.Command("awk", "'\\$NF > "+string(duration)+"'"),
		exec.Command("awk", "'{print \\$1}'"),
		exec.Command("xargs", "kill", "-9"),
	); err != nil {
		logrus.Errorf("fail to execWithPipe %s, err:%v", p, err)
	}
	// io.Copy(os.Stdout, &b)
	io.Copy(logrus.StandardLogger().Writer(), &b)
	logrus.Warnf("execWithPipe buffer:%v", b)
	return err
}

func (ctl UnixSysCtl) execWithPipe(output_buffer *bytes.Buffer, stack ...*exec.Cmd) (err error) {
	var error_buffer bytes.Buffer
	pipe_stack := make([]*io.PipeWriter, len(stack)-1)
	i := 0
	for ; i < len(stack)-1; i++ {
		stdin_pipe, stdout_pipe := io.Pipe()
		stack[i].Stdout = stdout_pipe
		stack[i].Stderr = &error_buffer
		stack[i+1].Stdin = stdin_pipe
		pipe_stack[i] = stdout_pipe
	}
	stack[i].Stdout = output_buffer
	stack[i].Stderr = &error_buffer

	if err := ctl.callStack(stack, pipe_stack); err != nil {
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
				pipes[0].Close()
				err = ctl.callStack(stack[1:], pipes[1:])
			} else {
				stack[1].Wait()
			}
		}()
	}
	return stack[0].Wait()
}
