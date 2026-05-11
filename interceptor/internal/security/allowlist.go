package security

import (
	"regexp"
	"strings"
)

// Classification describes whether a command is safe, requires approval, or is forbidden.
type Classification int

const (
	// ReadOnly commands never modify system state. Always permitted.
	ReadOnly Classification = iota
	// Mutating commands change system state. Require allow_mutations=true.
	Mutating
	// Forbidden commands are never permitted regardless of flags.
	Forbidden
)

func (c Classification) String() string {
	switch c {
	case ReadOnly:
		return "read-only"
	case Mutating:
		return "mutating"
	case Forbidden:
		return "forbidden"
	default:
		return "unknown"
	}
}

// CheckResult is the output of Classify.
type CheckResult struct {
	Class  Classification
	Reason string // human-readable explanation returned to the LLM
}

// forbiddenPatterns matches commands that are unconditionally blocked.
// These are catastrophic or irreversible operations.
var forbiddenPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\brm\s+-[a-z]*r[a-z]*f[a-z]*\s+/`),  // rm -rf /
	regexp.MustCompile(`(?i)\brm\s+-[a-z]*f[a-z]*r[a-z]*\s+/`),  // rm -fr /
	regexp.MustCompile(`(?i)\bmkfs\b`),                            // mkfs.*
	regexp.MustCompile(`(?i)\bdd\s+if=`),                          // dd if=...
	regexp.MustCompile(`(?i)>\s*/dev/sd[a-z]`),                    // > /dev/sda
	regexp.MustCompile(`(?i):\(\)\s*\{`),                          // fork bomb
	regexp.MustCompile(`(?i)\bshred\b`),                           // shred
	regexp.MustCompile(`(?i)\bwipefs\b`),                          // wipefs
	regexp.MustCompile(`(?i)\bfdisk\b`),                           // fdisk
	regexp.MustCompile(`(?i)\bparted\b`),                          // parted
	regexp.MustCompile(`(?i)\bcrypt\w*\s+luksFormat`),             // cryptsetup luksFormat
	regexp.MustCompile(`(?i)\biptables\s+-F\b`),                   // iptables -F (flush all rules)
	regexp.MustCompile(`(?i)\bnft\s+flush\b`),                     // nft flush
	regexp.MustCompile(`(?i)\bpasswd\b`),                          // passwd (change passwords)
	regexp.MustCompile(`(?i)\buseradd\b|\buserdel\b|\busermod\b`), // user management
	regexp.MustCompile(`(?i)\bvisudo\b`),                          // sudoers edit
}

// readOnlyPrefixes are command prefixes that are always safe.
// Order matters — checked first.
var readOnlyPrefixes = []string{
	"systemctl status",
	"systemctl list-units",
	"systemctl is-active",
	"systemctl is-enabled",
	"journalctl",
	"cat ",
	"ls",
	"ps ",
	"ps\n",
	"top -b",
	"htop",
	"df ",
	"df\n",
	"free",
	"ss ",
	"ss\n",
	"ip addr",
	"ip link",
	"ip route",
	"ip neigh",
	"ping ",
	"traceroute ",
	"tracepath ",
	"mtr ",
	"curl ",
	"wget ",
	"nslookup ",
	"dig ",
	"host ",
	"nmap ",
	"uname",
	"uptime",
	"who",
	"w\n",
	"last",
	"id\n",
	"whoami",
	"hostname",
	"date",
	"find ",
	"locate ",
	"which ",
	"whereis ",
	"stat ",
	"file ",
	"head ",
	"tail ",
	"less ",
	"more ",
	"grep ",
	"egrep ",
	"fgrep ",
	"awk ",
	"sed ",
	"sort ",
	"uniq ",
	"wc ",
	"cut ",
	"tr ",
	"diff ",
	"md5sum ",
	"sha256sum ",
	"lsof ",
	"netstat ",
	"ifconfig",
	"route ",
	"arp ",
	"dmesg",
	"sysctl -a",
	"sysctl -p",
	"env",
	"printenv",
	"echo ",
	"docker ps",
	"docker images",
	"docker logs",
	"docker inspect",
	"docker stats",
	"docker network ls",
	"docker volume ls",
	"kubectl get",
	"kubectl describe",
	"kubectl logs",
	"kubectl top",
	"kubectl explain",
	"kubectl version",
	"kubectl cluster-info",
	"timedatectl",
	"hostnamectl",
	"localectl",
	"lsblk",
	"lspci",
	"lsusb",
	"lscpu",
	"lsmem",
	"dmidecode",
	"smartctl",
	"iostat",
	"vmstat",
	"sar ",
	"mpstat",
	"pidstat",
	"strace -p",
	"ltrace -p",
	"rpm -q",
	"dpkg -l",
	"dpkg -s",
	"apt-cache",
	"yum info",
	"yum list",
	"dnf info",
	"dnf list",
	"snap list",
	"flatpak list",
	"git log",
	"git status",
	"git diff",
	"git show",
	"git branch",
	"git remote",
	"crontab -l",
	"at -l",
	"systemctl cat",
}

// mutatingPrefixes are command prefixes that change system state.
// Permitted only when allow_mutations=true.
var mutatingPrefixes = []string{
	"systemctl start",
	"systemctl stop",
	"systemctl restart",
	"systemctl reload",
	"systemctl enable",
	"systemctl disable",
	"systemctl mask",
	"systemctl unmask",
	"service ",
	"apt-get install",
	"apt-get remove",
	"apt-get purge",
	"apt-get upgrade",
	"apt install",
	"apt remove",
	"apt purge",
	"apt upgrade",
	"apt update",
	"yum install",
	"yum remove",
	"yum update",
	"yum upgrade",
	"dnf install",
	"dnf remove",
	"dnf update",
	"dnf upgrade",
	"pip install",
	"pip uninstall",
	"npm install",
	"npm uninstall",
	"cp ",
	"mv ",
	"mkdir ",
	"rmdir ",
	"touch ",
	"chmod ",
	"chown ",
	"chgrp ",
	"ln ",
	"tee ",
	"truncate ",
	"sed -i",
	"awk -i",
	"rm ",
	"docker start",
	"docker stop",
	"docker restart",
	"docker rm",
	"docker rmi",
	"docker pull",
	"docker run",
	"docker exec",
	"docker build",
	"docker-compose",
	"kubectl apply",
	"kubectl delete",
	"kubectl scale",
	"kubectl rollout",
	"kubectl patch",
	"kubectl edit",
	"kubectl cordon",
	"kubectl drain",
	"kubectl uncordon",
	"kubectl taint",
	"kubectl label",
	"kubectl annotate",
	"iptables -A",
	"iptables -D",
	"iptables -I",
	"iptables -R",
	"ufw allow",
	"ufw deny",
	"ufw enable",
	"ufw disable",
	"ufw delete",
	"firewall-cmd",
	"sysctl -w",
	"ulimit ",
	"crontab ",
	"at ",
	"reboot",
	"shutdown",
	"halt",
	"poweroff",
	"kill ",
	"killall ",
	"pkill ",
	"nice ",
	"renice ",
	"mount ",
	"umount ",
}

// Classify determines whether a command is read-only, mutating, or forbidden.
// The command string should be the raw shell command as it will be executed.
func Classify(command string) CheckResult {
	cmd := strings.TrimSpace(command)
	lower := strings.ToLower(cmd)

	// 1. Check forbidden patterns first — unconditional.
	for _, re := range forbiddenPatterns {
		if re.MatchString(cmd) {
			return CheckResult{
				Class:  Forbidden,
				Reason: "command matches a permanently blocked pattern (catastrophic/irreversible operation)",
			}
		}
	}

	// 2. Check read-only allowlist.
	for _, prefix := range readOnlyPrefixes {
		if strings.HasPrefix(lower, strings.ToLower(prefix)) {
			return CheckResult{Class: ReadOnly, Reason: ""}
		}
	}

	// 3. Check known mutating prefixes.
	for _, prefix := range mutatingPrefixes {
		if strings.HasPrefix(lower, strings.ToLower(prefix)) {
			return CheckResult{
				Class:  Mutating,
				Reason: "command modifies system state — re-run with allow_mutations=true after user approval",
			}
		}
	}

	// 4. Default: treat unrecognised commands as mutating (fail-safe).
	return CheckResult{
		Class:  Mutating,
		Reason: "unrecognised command — treated as mutating by default; re-run with allow_mutations=true after user approval",
	}
}
